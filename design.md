## 背景

这个node script 是服务于spotted oracle (cross chain state proof) AVS的，spotted oracle 提供一个以密码学为基础的state proof。它在每一个支持的链上都会部署一个statemanager, 主要有两个功能：

1. 任何user都可以在这个contract里面set自己想存储的key-value pair (msg.sender → key → value 和 msg.sender → key → History[])
2. 为每一个key  维护一个历史记录的array 存储着这个key的历史value变化（block number and timestamp)
3. statemanager 提供了查询方法可以查询每一个key value pair 的历史记录：有getHistoryAtBlock 和 getHistoryAtTimestamp
4. chainId, user, key, value, blocknumber(or timestamp) 是唯一的

这个AVS 背靠eigenlayer，它利用eigenlayer 的再质押经济安全模型来保证去中心化和安全。每个operator 需要注册到AVS 参与证明。他们会sign每一个正确的proof （query statemanager后确认value正确）然后集成一个超过 2/3 weight的ECDSA 签名集合。

用户可以在链上利用我们另一个contract （EIP1271标准）的 `isValidSignature` function来验证这个签名集合是否是valid 如果是，那么用户就得到了其他链上的状态证明

## 数据库Tables

使用sqlc 实现 (postgres)：

### 1. operator 记录operators的数据

注册的时候链上会 emit OperatorRegistered event

```solidity
    event OperatorRegistered(
        address indexed _operator,
        uint256 indexed blockNumber,
        address indexed _signingKey,
        uint256 timestamp,
        address _avs
    );
```

1. address: operator注册的address 可以通过event中 _operator 获得。 作为operator的UID
2. signingKey（address）: operator用来sign proof的Publick key 通过event中 _signingKey 获得
3. registeredAtBlocknumber （uint256）: 通过event中 block.number获得
4. registeredAtTimestamp （uint256）：通过event中timestamp获得
5. activeEpoch （uint32): operator 生效的epoch number  通过调用 `EpochManager` 中的 function getEffectiveEpochForBlock(uint64 blockNumber) public view returns (uint32) 获得，输入 registeredAtBlocknumber 得到 activeEpoch
6. exitEpoch (uint32): operator unactive的epochnumber 。如果operator在链上注销了 那么会emit event  emit OperatorDeregistered(_operator, block.number address(SERVICE_MANAGER)); avs server会监听这个event 并且把他更新到数据库里。
7. status: waitingJoin/waitingActive/active/unactive/suspended waitingJoin是指成功在链上注册后，avs server 监听到了event后存储的状态。waitingActive是指通过成功调用JoinNetwork之后 加入的operator 如果join时候的epoch还没有到 operator的activeEpoch 证明这个operator需要等待到 ≥ activeEpoch 才能开始工作。active是指正常运行的operator （JoinNetwork 并且当前epoch大于等于activeEpoch 并且没有deregister)。unactive: deregister的operator 通过监听 event emit OperatorDeregistered(_operator, address(SERVICE_MANAGER)); 收到event后 mark operator as unactive。suspended指missing次数过多的operator
8. weight(uint256):  当前epoch operator的weight。通过 `ECDSAStakeRegistry` function getOperatorWeightAtEpoch(
address _operator,
uint32 _epochNumber
) external view returns (uint256) 获得，每个epoch AVS 会和链上记录的状态同步。
9. missing: 在接到后没有在 3s 内返回任务response (either signed or 说明为什么refused to sign) 的次数，达到5次后把状态mark成为suspended 需要联系avs解决
10. successfulResponseCount: 记录operator成功response的task数量 用于之后奖励

### 2. Epoch State

维护epoch中operators的状态。监控mainnet的blocknumber，每当新的epoch的 start block number被mined的时候，trigger更新

1. epochNumber: primary key 唯一的自增ID 
2. operator[]: one to many reference。每当epoch advance的时候，先更新operators的状态：对于所有activeEpoch等于current epoch的operator把他们status mark为active；对于所有exitEpoch等于current epoch的operator 把他们status mark 为unactive。接着更新所有operators 的weight 通过 `ECDSAStakeRegistry` function getOperatorWeightAtEpoch(
address _operator,
uint32 _epochNumber
) external view returns (uint256)。同样的，更新他们的signingkey，通过function getOperatorSigningKeyAtEpoch(
address _operator,
uint32 _epochNumber
) external view returns (address) 获得。然后选取所有active status的operators添加到这个epoch 的reference下面。
3. blockNumber (uint64): 这个epoch 的start block number
4. minimumWeight (uint256): 通过query ECDSAStakeRegistry:: function minimumWeight() external view returns (uint256) 获得。
5. totalWeight(uint256): 通过query function getTotalWeightAtEpoch(
uint32 _epochNumber
) external view returns (uint256) 获得。
6. thresholdWeight(uint256): 通过query function getThresholdStake(
uint32 _referenceEpoch
) external view returns (uint256) 获得。
7. updatedAt(timestamp)：更新时候的timestamp

### 3. Task

user在sendRequest时候会指明 chainid, targetAddress, key，block number 或者 timestamp (二选一）

1. taskId (bytes32): 需要AVS server 自己先query对应 chainId 上的 targetAddress set 的 key value pair。如果用户指定的是blocknumber，那么就调用 statemanager的 function getHistoryAtBlock(
    address user,
    uint256 key,
    uint256 blockNumber
) external view returns (History memory); 如果指定的是timestamp 那么就调用   function getHistoryAtTimestamp(
        address user,
        uint256 key,
        uint256 timestamp
    ) external view returns (History memory);  这里History的struct是     struct History {
        uint256 value; 
        uint64 blockNumber; 
        uint48 timestamp; 
    }。然后根据  keccak256(abi.encodePacked(user, chainId, blockNumber, timestamp, epoch （通过 `EpochManager::`function getCurrentEpoch() external view returns (uint32)获得） , key, value)) 作为task 的UID 

2. targetAddress (address): 是 statemanager上存储key value pair 的msg.sender 等待证明他存储的状态 通过 user 的request 获取
3. chainId （uint32):  查询的chainid 通过 user 的request 获取
4. blockNumber (uint64): user二选一 如果user选择的是blocknumber 那么通过 user 的request 获取  如果不是，通过返回的history.blockNumber
5. timestamp (uint48)： user二选一 如果user选择的是timestamp 那么通过 user 的request 获取  如果不是，通过返回的history.timestamp
6. epoch (uint32): 通过 `EpochManager::getCurrentEpoch`  获得
7. key (uint256): 查询的key 通过user request获得
8. value (uint256): 查询的value  通过History.value 获得
9. expireTime (timestamp)：任务的过期时间，如果在这个时间点之前没有成功返回给用户的话就会timeout。15s
10. retries：记录retires的次数。retry发生在确认一个一个operators set之后（见下文确认需要签名的operators介绍）因为有些operator拒绝签名 或者超时，所以需要剔除这些operator然后重新选择set。如果这样的retry达到3次还是不成功，则mark as failed。
11. status: 链下定义 用于区分任务状态 三个状态：pending/completed/expired/failed pending是指AVS server 成功收到任务，等待operatos响应 和 返回任务。completed是指任务成功响应并且返回给user后。expired是指在expireTime前没有返回给用户的任务。failed是指因其他错误而导致任务失败的任务。
12. createdAt: 任务发布时的Timestamp

### 4. Operator Task Response

单个operator返回给avs server的response 包含 

1. ID: uid 由数据库生成。
2. Task 作为 One-to-One Relationship 
3. operator(address)：返回response的operator 
4. signature: 用operator的signingkey签了名之后的signed state proof
5. epoch (uint32): 通过Task.epoch获得
6. submittedAt: 返回时的timestamp

### 5. AVS Task Response

记录返回给用户成功的task response 包含：

1. taskID (bytes32):  跟Task 里面的taskid一样
2. Task：One-to-One Relationship with Shared Primary Key
3. epoch (uint32): 通过Task.epoch获得
4. operatorsResponses[]: 记录所有这个task的operator response（见上）
5. signature[]: 所有response了这个task 的operators的合并签名
6. respondAt: 返回给user的timestamp

### 6. rewards record （TODO 暂时不实现 配合interface）

记录分配给operators的奖励记录。每个interval可以设置完成一个task的奖励 所有task奖励是一致的

1. ID: 自增的UID 作为primary key

## sqlc 工具使用说明

1. 安装 

```
# 注意：需要启用 cgo，因为依赖 pg_query_go
git clone https://github.com/Stumble/sqlc.git
cd sqlc/
git checkout v2.3.0
make install
sqlc version  # 应该显示: v2.3.0-wicked-fork
```

1. 基本项目结构设置 

```
.
├── go.mod
└── pkg
    └── repos
        ├── your_table_name/
        │   ├── query.sql    # 存放SQL查询
        │   └── schema.sql   # 存放表结构定义
        └── sqlc.yaml        # sqlc配置文件
```

1. 配置文件设置（sqlc.yaml）：

```yaml
version: '2'
sql:
  - schema: your_table_name/schema.sql
    queries: your_table_name/query.sql
    engine: postgresql
    gen:
      go:
        sql_package: wpgx
        package: your_table_name
        out: your_table_name
```

1. 使用步骤
    1. 首先在 schema.sql 中定义你的表结构，例如： 
    
    ```sql
    CREATE TABLE IF NOT EXISTS users (
       id          INT            GENERATED ALWAYS AS IDENTITY,
       name        VARCHAR(255)   NOT NULL,
       created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
       CONSTRAINT users_id_pkey PRIMARY KEY (id)
    );
    ```
    
    b.  在 query.sql 中编写你的查询，例如： 
    
    ```sql
    -- name: GetUserByID :one
    SELECT * FROM users
    WHERE id = $1;
    
    -- name: ListUsers :many
    SELECT * FROM users
    ORDER BY id;
    
    -- name: CreateUser :one
    INSERT INTO users (name)
    VALUES ($1)
    RETURNING *;
    ```
    
    c.  生成代码: sqlc generate
    
2. 

重要使用注意事项：

- 重要使用注意事项：
- 每个查询都需要添加注释来指定名称和返回类型：
- :one - 返回单行
- :many - 返回多行
- :exec - 不返回结果
- :execrows - 返回影响的行数

使用生成的代码时需要先初始化数据库连接：

```go
import (
    "context"
    "your_project/pkg/repos/users"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    ctx := context.Background()
    db, err := pgxpool.Connect(ctx, "postgresql://user:password@localhost:5432/dbname")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // 创建查询客户端
    queries := users.New(db)

    // 使用生成的方法
    user, err := queries.GetUserByID(ctx, 1)
    if err != nil {
        panic(err)
    }
}
```

1. 高级功能使用：

缓存配置：在查询注释中添加缓存配置 

```sql
-- name: GetUserByID :one
-- cache_ttl: 5m
-- cache_key: user:id:{}
SELECT * FROM users WHERE id = $1;
```

超时配置： 

```sql
-- name: GetUserByID :one
-- timeout: 500ms
SELECT * FROM users WHERE id = $1;
```

最佳实践：

- 将相关的表和查询组织在同一个目录下
- 为每个查询添加适当的注释和文档
- 明确列出所有表的约束和索引
- 避免过度使用动态SQL生成
- 合理使用缓存配置来优化性能

调试和测试：

- 使用生成的 Load 和 Dump 函数进行测试数据管理
- 利用内置的监控功能追踪性能问题
- 使用生成的类型安全的接口来避免运行时错误

这个工具的使用相对直接，主要流程是：定义表结构 → 编写SQL查询 → 生成代码 → 使用生成的代码。关键是要理解配置文件的结构和查询注释的格式，这样才能充分利用工具的所有功能。

## 存储设计

### registry node

1. 存储完整的网络状态
2. 维护operator life cycle信息
3. 记录所有epoch状态
4. Registry Node之间同步完整状态

```sql
-- 1. operators表 (完整的operator信息)
CREATE TABLE operators (
    address TEXT PRIMARY KEY,
    signing_key TEXT NOT NULL,
    registered_at_block_number NUMERIC(78) NOT NULL,
    registered_at_timestamp NUMERIC(78) NOT NULL,
    active_epoch INT NOT NULL,
    exit_epoch INT,
    status TEXT NOT NULL CHECK (status IN ('waitingJoin', 'waitingActive', 'active', 'unactive', 'suspended')),
    weight NUMERIC(78),
    missing INT DEFAULT 0,
    successful_response_count INT DEFAULT 0
);

-- 2. epoch_states表
CREATE TABLE epoch_states (
    epoch_number INT PRIMARY KEY,
    block_number NUMERIC(78) NOT NULL,
    minimum_weight NUMERIC(78) NOT NULL,
    total_weight NUMERIC(78) NOT NULL,
    threshold_weight NUMERIC(78) NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- 3. epoch_operators表 (epoch和operator的多对多关系)
CREATE TABLE epoch_operators (
    epoch_number INT REFERENCES epoch_states(epoch_number),
    operator_address TEXT REFERENCES operators(address),
    weight NUMERIC(78) NOT NULL,
    signing_key TEXT NOT NULL,
    PRIMARY KEY (epoch_number, operator_address)
);

-- 4. registry_events表 (记录重要事件)
CREATE TABLE registry_events (
    id SERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    operator_address TEXT,
    block_number NUMERIC(78),
    timestamp TIMESTAMP,
    data JSONB
);
```

### sqlc 应用到当前设计的 示例

1. pkg/repos/registry/operators/schema.sql: 

```sql
CREATE TABLE IF NOT EXISTS operators (
    address TEXT PRIMARY KEY,
    signing_key TEXT NOT NULL,
    registered_at_block_number NUMERIC(78) NOT NULL,
    registered_at_timestamp NUMERIC(78) NOT NULL,
    active_epoch INT NOT NULL,
    exit_epoch INT,
    status TEXT NOT NULL CHECK (status IN ('waitingJoin', 'waitingActive', 'active', 'unactive', 'suspended')),
    weight NUMERIC(78),
    missing INT DEFAULT 0,
    successful_response_count INT DEFAULT 0
);

CREATE INDEX idx_operators_status ON operators(status);
CREATE INDEX idx_operators_active_epoch ON operators(active_epoch);
```

1. pkg/repos/registry/operators/query.sql: 

```sql
-- name: GetOperatorByAddress :one
-- 获取单个operator信息
SELECT * FROM operators
WHERE address = $1;

-- name: ListActiveOperators :many
-- 获取所有active状态的operators
SELECT * FROM operators
WHERE status = 'active';

-- name: UpdateOperatorStatus :exec
-- 更新operator状态
UPDATE operators
SET status = $2,
    updated_at = NOW()
WHERE address = $1;

-- name: UpdateOperatorWeight :one
-- 更新operator权重
UPDATE operators
SET weight = $2,
    updated_at = NOW()
WHERE address = $1
RETURNING *;

-- name: IncrementMissingCount :one
-- 增加missing计数
UPDATE operators
SET missing = missing + 1,
    status = CASE 
        WHEN missing + 1 >= 5 THEN 'suspended'
        ELSE status 
    END
WHERE address = $1
RETURNING *;
```

1. pkg/repos/sqlc.yaml: 

```sql
version: "2"
sql:
  - engine: "postgresql"
    queries: 
      - "registry/operators/query.sql"
      - "registry/epoch_states/query.sql"
      - "registry/epoch_operators/query.sql"
      - "registry/registry_events/query.sql"
    schema: 
      - "registry/operators/schema.sql"
      - "registry/epoch_states/schema.sql"
      - "registry/epoch_operators/schema.sql"
      - "registry/registry_events/schema.sql"
    gen:
      go:
        package: "registry"
        out: "registry"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: true

  - engine: "postgresql"
    queries:
      - "operator/local_operator/query.sql"
      - "operator/current_epoch/query.sql"
      - "operator/tasks/query.sql"
      - "operator/task_responses/query.sql"
      - "operator/peer_cache/query.sql"
    schema:
      - "operator/local_operator/schema.sql"
      - "operator/current_epoch/schema.sql"
      - "operator/tasks/schema.sql"
      - "operator/task_responses/schema.sql"
      - "operator/peer_cache/schema.sql"
    gen:
      go:
        package: "operator"
        out: "operator"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: true
```

1. 使用示例: 

```sql
// pkg/db/connection.go
package db

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return nil, err
    }
    
    // 设置连接池参数
    config.MaxConns = 10
    config.MinConns = 2
    
    return pgxpool.NewWithConfig(ctx, config)
}

// pkg/registry/node.go
import (
    "your_project/pkg/repos/registry"
)

type RegistryNode struct {
    db      *pgxpool.Pool
    queries *registry.Queries
}

func (n *RegistryNode) UpdateOperatorStatus(ctx context.Context, address, status string) error {
    return n.queries.UpdateOperatorStatus(ctx, registry.UpdateOperatorStatusParams{
        Address: address,
        Status:  status,
    })
}
```

### sqlc 对于当前项目可能得优化

1. 配置文件优化： 

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: 
      - "registry/operators/query.sql"
      - "registry/epoch_states/query.sql"
      # ... 其他查询文件
    schema: 
      - "registry/operators/schema.sql"
      - "registry/epoch_states/schema.sql"
      # ... 其他schema文件
    gen:
      go:
        package: "registry"
        out: "registry"
        sql_package: "wpgx"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: true
        # 建议添加以下配置
        emit_exact_table_names: true  # 保持表名不变
        emit_empty_slices: true       # 返回空切片而不是nil
        emit_exported_queries: true    # 导出查询方法
```

1. 查询优化 

```sql
-- operators/query.sql
-- 建议添加以下查询：

-- name: GetOperatorWithLock :one
-- 获取operator并加锁，用于事务
SELECT * FROM operators
WHERE address = $1
FOR UPDATE;

-- name: ListOperatorsByEpoch :many
-- 获取特定epoch的operators
SELECT o.* 
FROM operators o
JOIN epoch_operators eo ON o.address = eo.operator_address
WHERE eo.epoch_number = $1;

-- name: UpdateOperatorStatusBatch :execrows
-- 批量更新operator状态
UPDATE operators
SET status = $1
WHERE address = ANY($2::text[]);

-- name: GetOperatorStats :one
-- 获取operator统计信息
SELECT 
    COUNT(*) as total_count,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active_count,
    COUNT(CASE WHEN status = 'suspended' THEN 1 END) as suspended_count
FROM operators;
```

1. 错误处理和约束 

```sql
-- operators/schema.sql
CREATE TABLE IF NOT EXISTS operators (
    -- ... 现有字段
    status TEXT NOT NULL CHECK (status IN ('waitingJoin', 'waitingActive', 'active', 'unactive', 'suspended')),
    weight NUMERIC(78) CHECK (weight >= 0), -- 添加非负约束
    missing INT DEFAULT 0 CHECK (missing >= 0),
    successful_response_count INT DEFAULT 0 CHECK (successful_response_count >= 0),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 添加更多索引
CREATE INDEX idx_operators_weight ON operators(weight) WHERE status = 'active';
CREATE INDEX idx_operators_updated_at ON operators(updated_at);
```

1. 缓存和超时配置 

```sql
-- 为频繁访问的查询添加缓存
-- name: GetOperatorByAddress :one
-- -- timeout: 500ms
-- -- cache: 5m
-- -- cache_key: operator:{}
SELECT * FROM operators
WHERE address = $1;

-- name: ListActiveOperators :many
-- -- timeout: 1s
-- -- cache: 1m
SELECT * FROM operators
WHERE status = 'active';
```

1. 事务支持 

```sql
// pkg/registry/node.go
type RegistryNode struct {
    db      *pgxpool.Pool
    queries *registry.Queries
}

// 添加事务支持的方法
func (n *RegistryNode) UpdateOperatorStatusWithTx(ctx context.Context, address, status string) error {
    tx, err := n.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    qtx := n.queries.WithTx(tx)
    
    // 先获取并锁定
    op, err := qtx.GetOperatorWithLock(ctx, address)
    if err != nil {
        return err
    }
    
    // 检查状态转换是否合法
    if !isValidStatusTransition(op.Status, status) {
        return fmt.Errorf("invalid status transition: %s -> %s", op.Status, status)
    }
    
    // 更新状态
    if err := qtx.UpdateOperatorStatus(ctx, address, status); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

1. 监控和指标 

```sql
-- registry_events/query.sql
-- name: RecordMetrics :exec
INSERT INTO registry_events (
    event_type,
    operator_address,
    block_number,
    timestamp,
    data
) VALUES (
    'metrics',
    $1,
    $2,
    NOW(),
    $3
);

-- name: GetOperatorMetrics :many
SELECT * FROM registry_events
WHERE event_type = 'metrics'
  AND operator_address = $1
  AND timestamp >= $2
ORDER BY timestamp DESC;
```

### operator node

1. 只存储本地必要信息
2. 关注当前epoch状态
3. 记录任务处理历史
4. 缓存网络peer信息
5. Operator Node通过P2P网络获取必要更新

```sql
-- 1. local_operator表 (只存储自己的信息)
CREATE TABLE local_operator (
    address TEXT PRIMARY KEY,
    signing_key TEXT NOT NULL,
    active_epoch INT NOT NULL,
    current_status TEXT NOT NULL,
    current_weight NUMERIC(78),
    last_updated_at TIMESTAMP
);

-- 2. current_epoch_state表 (只存储当前epoch信息)
CREATE TABLE current_epoch_state (
    epoch_number INT PRIMARY KEY,
    block_number NUMERIC(78) NOT NULL,
    minimum_weight NUMERIC(78) NOT NULL,
    total_weight NUMERIC(78) NOT NULL,
    threshold_weight NUMERIC(78) NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- 3. tasks表 (存储处理过的任务)
CREATE TABLE tasks (
    task_id BYTES PRIMARY KEY,
    target_address TEXT NOT NULL,
    chain_id INT NOT NULL,
    block_number NUMERIC(78),
    timestamp NUMERIC(78),
    epoch INT NOT NULL,
    key NUMERIC(78) NOT NULL,
    value NUMERIC(78) NOT NULL,
    expire_time TIMESTAMP NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- 4. task_responses表 (存储本地响应)
CREATE TABLE task_responses (
    id SERIAL PRIMARY KEY,
    task_id BYTES REFERENCES tasks(task_id),
    signature BYTES NOT NULL,
    submitted_at TIMESTAMP NOT NULL
);

-- 5. peer_cache表 (缓存活跃peer信息)
CREATE TABLE peer_cache (
    peer_id TEXT PRIMARY KEY,
    operator_address TEXT,
    last_seen TIMESTAMP,
    status TEXT
);
```

## AVS：Registry Node

1. **network引导与维护** 

```go
type RegistryNode struct {
    host host.Host
    db *Database
    
    // 节点管理
    nodeManager *NodeManager
    
    // 事件监听
    eventListener *EventListener
    
    // 状态广播
    stateBroadcaster *StateBroadcaster
}
```

1. 作为boostrap节点 

```go
func (r *RegistryNode) ServeBootstrap() {
    // 1. 提供初始节点列表
    r.host.SetStreamHandler("/bootstrap/1.0.0", func(s network.Stream) {
        // 返回活跃operator列表
        activeOperators := r.nodeManager.GetActiveOperators()
        sendOperatorList(s, activeOperators)
    })
    
    // 2. 协助节点发现
    r.host.SetStreamHandler("/discovery/1.0.0", func(s network.Stream) {
        // 提供节点发现服务
    })
}
```

1. **Operator 注册注销**
    1. listen event 
    
    ```go
    func (r *RegistryNode) MonitorRegistration() {
        events := r.eventListener.ListenOperatorEvents()
        for event := range events {
            switch event.Type {
            case OperatorRegistered:
                // 记录新注册的operator
                r.handleNewOperator(event)
            case OperatorDeregistered:
                // 处理operator注销
                r.handleOperatorDeregistration(event)
            }
        }
    }
    ```
    
    b. 处理Join请求 
    
    ```go
    func (r *RegistryNode) HandleJoinRequest(req *JoinRequest) error {
        // 1. 验证operator状态
        if !r.nodeManager.IsWaitingJoin(req.OperatorAddress) {
            return ErrInvalidStatus
        }
        
        // 2. 验证ECDSA签名
        if err := r.verifySignature(req); err != nil {
            return err
        }
        
        // 3. 更新状态
        r.nodeManager.UpdateOperatorStatus(req.OperatorAddress, StatusWaitingActive)
        
        // 4. 广播状态更新
        r.broadcastOperatorUpdate(req.OperatorAddress)
        
        return nil
    }
    ```
    
2. **状态维护与broadcast**
    1. **维护可信节点列表** 
    
    ```go
    type NodeManager struct {
        // operator状态映射
        operatorStates map[string]*OperatorState
        
        // 状态更新通道
        updates chan StateUpdate
    }
    
    func (nm *NodeManager) MaintainNodeList() {
        // 1. 定期清理过期状态
        go nm.cleanupExpiredStates()
        
        // 2. 处理状态更新
        for update := range nm.updates {
            nm.handleStateUpdate(update)
        }
    }
    ```
    
    b. 状态broadcast 
    
    ```go
    func (r *RegistryNode) BroadcastState() {
        // 1. 广播operator状态更新
        r.stateBroadcaster.BroadcastOperatorUpdates()
        
        // 2. 同步Registry Node之间的状态
        r.syncRegistryStates()
    }
    ```
    
3. health check 

```go
type HealthCheck struct {
    // 检查间隔
    checkInterval time.Duration
    
    // operator健康状态
    healthStatus map[string]HealthStatus
}

func (hc *HealthCheck) MonitorOperators() {
    ticker := time.NewTicker(hc.checkInterval)
    for range ticker.C {
        // 检查所有active operators
        hc.checkOperatorsHealth()
        
        // 更新状态
        hc.updateHealthStatus()
    }
}
```

1. 查询接口 

```go
type QueryService struct {
    // 处理operator查询
    HandleOperatorQuery(query *OperatorQuery) (*OperatorInfo, error)
    
    // 处理网络状态查询
    HandleNetworkQuery(query *NetworkQuery) (*NetworkStatus, error)
}
```

1. registry nodes 之间的同步 

```go
type RegistrySync struct {
    // 其他Registry Node列表
    peers []peer.ID
    
    // 状态同步
    syncState() error
    
    // 处理冲突
    resolveConflict(conflict *StateConflict) error
}
```

## Operator node

1. 核心架构  

```go
type OperatorNode struct {
    // libp2p host
    host host.Host
    
    // 节点配置
    config *Config
    
    // 密钥管理
    signer *Signer
    
    // 任务处理器
    taskProcessor *TaskProcessor
    
    // P2P网络管理
    networkManager *NetworkManager
    
    // epoch管理
    epochManager *EpochManager
    
    // 状态管理
    stateManager *StateManager
    
    // 健康检查
    healthCheck *HealthCheck
}

type Config struct {
    // operator地址
    OperatorAddress string
    
    // 密钥路径
    KeystorePath string
    
    // bootstrap节点列表
    BootstrapPeers []multiaddr.Multiaddr
    
    // 支持的链ID列表
    SupportedChains []uint32
}
```

1. p2p network
    1. 节点初始化 
    
    ```go
    func NewOperatorNode(cfg *Config) (*OperatorNode, error) {
        // 1. 创建libp2p host
        h, err := libp2p.New(
            libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
            libp2p.Identity(cfg.PrivateKey),
            libp2p.EnableRelay(),
            libp2p.EnableAutoRelay(),
        )
        
        // 2. 初始化各个组件
        node := &OperatorNode{
            host: h,
            config: cfg,
            signer: NewSigner(cfg.KeystorePath),
            taskProcessor: NewTaskProcessor(),
            networkManager: NewNetworkManager(h),
            epochManager: NewEpochManager(),
            stateManager: NewStateManager(),
            healthCheck: NewHealthCheck(),
        }
        
        return node, nil
    }
    ```
    
    b. network管理 
    
    ```go
    type NetworkManager struct {
        // libp2p host
        host host.Host
        
        // DHT for peer discovery
        dht *dht.IpfsDHT
        
        // peer store
        peerstore peerstore.Peerstore
        
        // 连接管理
        connManager connmgr.ConnManager
    }
    
    func (nm *NetworkManager) Start() error {
        // 1. 连接bootstrap节点
        nm.connectBootstrapPeers()
        
        // 2. 启动DHT
        nm.startDHT()
        
        // 3. 设置协议处理器
        nm.setupProtocolHandlers()
        
        return nil
    }
    ```
    
2. 任务处理
    1. 任务protocol 
    
    ```go
    const (
        TaskProtocol = "/spotted-oracle/task/1.0.0"
        TaskResponseProtocol = "/spotted-oracle/response/1.0.0"
    )
    
    type TaskProcessor struct {
        // 任务处理队列
        taskQueue chan *Task
        
        // 状态验证器
        stateValidator *StateValidator
        
        // 签名处理器
        signatureProcessor *SignatureProcessor
    }
    ```
    
    b. 任务处理流程 
    
    ```go
    func (tp *TaskProcessor) handleTask(task *Task) error {
        // 1. 验证状态
        if err := tp.stateValidator.ValidateState(task); err != nil {
            return err
        }
        
        // 2. 生成签名
        signature, err := tp.signatureProcessor.Sign(task)
        if err != nil {
            return err
        }
        
        // 3. 广播响应
        return tp.broadcastResponse(&TaskResponse{
            TaskID: task.ID,
            Signature: signature,
            Value: task.Value,
        })
    }
    ```
    
3. 状态管理
    1. epoch 管理 
    
    ```go
    // Epoch管理
    type EpochManager struct {
        // 创世区块号
        genesisBlock uint64
        
        // 每个epoch的区块间隔
        blockInterval uint64
        
        // 当前epoch
        currentEpoch uint32
        
        // epoch状态
        epochState *EpochState
    }
    
    // 计算epoch的start block
    func (em *EpochManager) calculateEpochStartBlock(epoch uint32) uint64 {
        return em.genesisBlock + (uint64(epoch) * em.blockInterval)
    }
    
    // 在epoch start block被mine时更新状态
    func (em *EpochManager) handleNewBlock(blockNumber uint64) error {
        // 检查是否是epoch的start block
        epoch := (blockNumber - em.genesisBlock) / em.blockInterval
        if blockNumber == em.calculateEpochStartBlock(uint32(epoch)) {
            // 是新epoch的start block,更新状态
            return em.updateEpochState(uint32(epoch))
        }
        return nil
    }
    
    // 更新epoch状态
    func (em *EpochManager) updateEpochState(epoch uint32) error {
        // 1. 获取新的operator集合
        operators, err := em.getOperatorsForEpoch(epoch)
        if err != nil {
            return err
        }
        
        // 2. 更新operator权重
        for _, op := range operators {
            weight, err := em.getOperatorWeight(op.Address, epoch)
            if err != nil {
                return err
            }
            op.Weight = weight
        }
        
        // 3. 更新epoch状态
        em.epochState = &EpochState{
            EpochNumber: epoch,
            Operators: operators,
            StartBlock: em.calculateEpochStartBlock(epoch),
            UpdatedAt: time.Now(),
        }
        
        return nil
    }
    
    // 获取operator权重
    func (em *EpochManager) getOperatorWeight(operator string, epoch uint32) (uint256, error) {
        return em.stakeRegistry.GetOperatorWeightAtEpoch(operator, epoch)
    }
    
    // 启动epoch管理
    func (em *EpochManager) Start(ctx context.Context) error {
        // 1. 订阅新区块
        blockChan, err := em.subscribeNewBlocks(ctx)
        if err != nil {
            return err
        }
        
        // 2. 处理新区块
        go func() {
            for {
                select {
                case block := <-blockChan:
                    if err := em.handleNewBlock(block.Number); err != nil {
                        log.Error("Failed to handle new block", "error", err)
                    }
                case <-ctx.Done():
                    return
                }
            }
        }()
        
        return nil
    }
    ```
    
    b. 状态同步 
    
    ```go
    type StateManager struct {
        // 本地状态
        localState *State
        
        // 状态同步协议
        syncProtocol *SyncProtocol
        
        // 状态更新通道
        stateUpdates chan *StateUpdate
    }
    
    func (sm *StateManager) SyncState() error {
        // 1. 获取网络状态
        networkState := sm.syncProtocol.GetNetworkState()
        
        // 2. 更新本地状态
        return sm.updateLocalState(networkState)
    }
    ```
    
4. health check 

```go
type HealthCheck struct {
    // 健康状态
    status HealthStatus
    
    // 检查间隔
    checkInterval time.Duration
    
    // 最后检查时间
    lastCheck time.Time
}

func (hc *HealthCheck) Start() {
    // 1. 定期检查
    go hc.runPeriodicCheck()
    
    // 2. 响应健康检查请求
    hc.handleHealthCheck()
}
```

1. 协议实现 

```go
func (node *OperatorNode) setupProtocols() {
    // 1. 任务处理协议
    node.host.SetStreamHandler(TaskProtocol, node.handleTaskStream)
    
    // 2. 响应处理协议
    node.host.SetStreamHandler(TaskResponseProtocol, node.handleResponseStream)
    
    // 3. 状态同步协议
    node.host.SetStreamHandler(StateSyncProtocol, node.handleStateSync)
    
    // 4. 健康检查协议
    node.host.SetStreamHandler(HealthCheckProtocol, node.handleHealthCheck)
}
```

1. 启动流程  

```go
func (node *OperatorNode) Start() error {
    // 1. 初始化signer
    if err := node.signer.Init(); err != nil {
        return err
    }
    
    // 2. 启动网络
    if err := node.networkManager.Start(); err != nil {
        return err
    }
    
    // 3. 设置协议处理器
    node.setupProtocols()
    
    // 4. 启动状态管理
    node.stateManager.Start()
    
    // 5. 启动epoch管理
    node.epochManager.Start()
    
    // 6. 启动健康检查
    node.healthCheck.Start()
    
    // 7. 连接Registry Node并加入网络
    return node.joinNetwork()
}
```

## 通信

1.  Operator Nodes之间的通信 (使用libp2p)

使用libp2p的pub/sub进行任务和响应广播

使用libp2p streams进行直接通信

支持去中心化的节点发现

```go
// 定义通信协议
const (
    // 任务广播协议
    TaskBroadcastProtocol = "/spotted-oracle/task/1.0.0"
    // 响应共享协议
    ResponseShareProtocol = "/spotted-oracle/response/1.0.0"
    // 状态同步协议
    StateSyncProtocol    = "/spotted-oracle/state/1.0.0"
)

type P2PCommunication struct {
    host host.Host
    pubsub *pubsub.PubSub
    
    // 订阅的topics
    taskSub *pubsub.Subscription
    responseSub *pubsub.Subscription
    stateSub *pubsub.Subscription
}

// 设置P2P通信
func (p *P2PCommunication) Setup() error {
    // 1. 初始化pubsub
    ps, err := pubsub.NewGossipSub(context.Background(), p.host)
    if err != nil {
        return err
    }
    p.pubsub = ps
    
    // 2. 订阅topics
    if err := p.subscribeTopics(); err != nil {
        return err
    }
    
    // 3. 设置流处理器
    p.setupStreamHandlers()
    
    return nil
}

// 处理任务广播
func (p *P2PCommunication) handleTaskBroadcast() {
    for {
        msg, err := p.taskSub.Next(context.Background())
        if err != nil {
            continue
        }
        // 处理收到的任务
        p.processTask(msg.Data)
    }
}
```

1.  **Operator与Registry Node之间的通信 (使用gRPC)**

使用gRPC进行可靠的服务调用

支持双向流通信

便于实现认证和负载均衡

```protobuf
// registry.proto
service RegistryService {
    // Join Network请求
    rpc JoinNetwork(JoinNetworkRequest) returns (JoinNetworkResponse);
    
    // 健康检查
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
    
    // 状态更新流
    rpc SubscribeStateUpdates(StateSubscribeRequest) returns (stream StateUpdate);
}

message JoinNetworkRequest {
    string operator_address = 1;
    bytes signature = 2;
    string node_id = 3;
}

message StateUpdate {
    string operator_address = 1;
    string status = 2;
    uint64 timestamp = 3;
}
```

```go
// Registry Node端实现
type RegistryServer struct {
    pb.UnimplementedRegistryServiceServer
    
    operatorManager *OperatorManager
    stateManager    *StateManager
}

func (s *RegistryServer) JoinNetwork(ctx context.Context, req *pb.JoinNetworkRequest) (*pb.JoinNetworkResponse, error) {
    // 处理Join请求
    return s.operatorManager.HandleJoin(req)
}

// Operator Node端实现
type RegistryClient struct {
    client pb.RegistryServiceClient
    conn   *grpc.ClientConn
}

func (c *RegistryClient) Join(ctx context.Context) error {
    req := &pb.JoinNetworkRequest{
        OperatorAddress: c.address,
        Signature: c.signature,
        NodeID: c.nodeID,
    }
    
    _, err := c.client.JoinNetwork(ctx, req)
    return err
}
```

**3. Registry Nodes之间的通信 (使用gRPC)**

使用gRPC进行状态同步

支持流式更新

便于实现一致性保证

```protobuf
// registry_sync.proto
service RegistrySyncService {
    // 状态同步
    rpc SyncState(SyncStateRequest) returns (SyncStateResponse);
    
    // 操作者更新
    rpc UpdateOperator(OperatorUpdate) returns (OperatorUpdateResponse);
    
    // 状态流同步
    rpc StreamStateUpdates(StreamRequest) returns (stream StateUpdate);
}

message OperatorUpdate {
    string operator_address = 1;
    string status = 2;
    uint64 timestamp = 3;
    string registry_id = 4;
}
```

```protobuf
// Registry Node同步实现
type RegistrySyncer struct {
    clients map[string]pb.RegistrySyncServiceClient
    
    // 本地状态
    localState *State
    
    // 更新通道
    updates chan *OperatorUpdate
}

func (s *RegistrySyncer) Start() error {
    // 1. 连接其他Registry Nodes
    if err := s.connectPeers(); err != nil {
        return err
    }
    
    // 2. 开始状态同步
    go s.syncStates()
    
    // 3. 处理更新
    go s.handleUpdates()
    
    return nil
}

func (s *RegistrySyncer) handleUpdates() {
    for update := range s.updates {
        // 广播更新给其他Registry Nodes
        s.broadcastUpdate(update)
    }
}
```

## WorkFlow

## operator 注册

1.  **Registry Node职责** 

```go
type RegistryNode struct {
    // libp2p host
    host host.Host
    
    // 数据库连接 (存储完整operator信息)
    db *Database
    
    // 事件监听器
    eventListener *EventListener
    
    // operator管理
    operatorManager *OperatorManager
}

// 监听链上事件
func (r *RegistryNode) ListenChainEvents() {
    events := r.eventListener.Listen()
    for event := range events {
        switch event.Type {
        case OperatorRegistered:
            // 1. 记录到数据库
            r.db.SaveOperator(&Operator{
                Address: event.Operator,
                SigningKey: event.SigningKey,
                Status: "waitingJoin",
                RegisteredAt: event.BlockNumber,
            })
            
            // 2. 广播到P2P网络
            r.broadcastOperatorEvent(event)
        }
    }
}
```

1. **Operator Node实现** 

```go
type OperatorNode struct {
    // libp2p host
    host host.Host
    
    // 本地数据库 (只存储必要状态)
    localDB *LocalDB
    
    // 密钥管理
    signer *Signer
    
    // P2P网络管理
    networkManager *NetworkManager
}

// 本地数据库结构
type LocalDB struct {
    // operator信息
    OperatorInfo *OperatorInfo
    
    // 当前epoch状态
    EpochState *EpochState
    
    // 活跃节点列表
    ActivePeers map[peer.ID]*PeerInfo
}
```

1. 链上注册 监听 

```go
// 1. operator自行完成链上注册
// 2. Registry Node监听到事件后记录
func (r *RegistryNode) handleOperatorRegistered(event *OperatorRegisteredEvent) {
    // 1. 验证事件
    if err := r.validateEvent(event); err != nil {
        return
    }
    
    // 2. 记录到数据库
    operator := &Operator{
        Address: event.Operator,
        SigningKey: event.SigningKey,
        Status: "waitingJoin",
        RegisteredAt: event.BlockNumber,
    }
    r.db.SaveOperator(operator)
    
    // 3. 广播到P2P网络
    r.broadcastToPeers(&OperatorUpdate{
        Type: OperatorRegistered,
        Operator: operator,
    })
}
```

1. join network 

```go
// Operator Node加入网络
func (node *OperatorNode) JoinNetwork() error {
    // 1. 验证本地keystore
    if err := node.signer.ValidateKeystore(); err != nil {
        return err
    }
    
    // 2. 连接bootstrap节点(Registry Node)
    if err := node.connectBootstrap(); err != nil {
        return err
    }
    
    // 3. 准备Join请求
    joinReq := &JoinRequest{
        OperatorAddress: node.config.Address,
        NodeID: node.host.ID(),
        // 使用operator私钥签名
        Signature: node.signer.Sign(joinMessage),
    }
    
    // 4. 发送Join请求到Registry Node
    resp, err := node.sendJoinRequest(joinReq)
    if err != nil {
        return err
    }
    
    // 5. 初始化本地状态
    return node.initLocalState(resp)
}

// Registry Node处理Join请求
func (r *RegistryNode) handleJoinRequest(req *JoinRequest) error {
    // 1. 验证operator状态
    operator, err := r.db.GetOperator(req.OperatorAddress)
    if err != nil {
        return err
    }
    if operator.Status != "waitingJoin" {
        return ErrInvalidStatus
    }
    
    // 2. 验证签名
    if err := r.verifySignature(req); err != nil {
        return err
    }
    
    // 3. 更新状态
    operator.Status = "waitingActive"
    operator.NodeID = req.NodeID
    r.db.UpdateOperator(operator)
    
    // 4. 广播状态更新
    r.broadcastOperatorUpdate(operator)
    
    return nil
}
```

1. 状态同步 

```go
// Operator Node加入网络
func (node *OperatorNode) JoinNetwork() error {
    // 1. 验证本地keystore
    if err := node.signer.ValidateKeystore(); err != nil {
        return err
    }
    
    // 2. 连接bootstrap节点(Registry Node)
    if err := node.connectBootstrap(); err != nil {
        return err
    }
    
    // 3. 准备Join请求
    joinReq := &JoinRequest{
        OperatorAddress: node.config.Address,
        NodeID: node.host.ID(),
        // 使用operator私钥签名
        Signature: node.signer.Sign(joinMessage),
    }
    
    // 4. 发送Join请求到Registry Node
    resp, err := node.sendJoinRequest(joinReq)
    if err != nil {
        return err
    }
    
    // 5. 初始化本地状态
    return node.initLocalState(resp)
}

// Registry Node处理Join请求
func (r *RegistryNode) handleJoinRequest(req *JoinRequest) error {
    // 1. 验证operator状态
    operator, err := r.db.GetOperator(req.OperatorAddress)
    if err != nil {
        return err
    }
    if operator.Status != "waitingJoin" {
        return ErrInvalidStatus
    }
    
    // 2. 验证签名
    if err := r.verifySignature(req); err != nil {
        return err
    }
    
    // 3. 更新状态
    operator.Status = "waitingActive"
    operator.NodeID = req.NodeID
    r.db.UpdateOperator(operator)
    
    // 4. 广播状态更新
    r.broadcastOperatorUpdate(operator)
    
    return nil
}
```

## operator 注销

1. 监听event `OperatorDeregistered(_operator, block.number, address(SERVICE_MANAGER));`
2. 数据库中operator 的exitEpoch 会被更新为对应的epoch number
3. epoch 更新的时候会把所有到达exitEpoch 的operator mark as unactive

## task generation and response

**1. 任务生成和广播**

```go
// 任务处理协议
const (
    TaskTopicName = "/spotted-oracle/task"
    ResponseTopicName = "/spotted-oracle/response"
    ResultTopicName = "/spotted-oracle/result"
)

// 任务处理器
type TaskProcessor struct {
    host host.Host
    pubsub *pubsub.PubSub
    
    // 订阅的topics
    taskSub *pubsub.Subscription
    responseSub *pubsub.Subscription
    resultSub *pubsub.Subscription
    
    // 本地任务状态
    tasks map[string]*TaskState
    
    // 用户连接管理
    userStreams map[string]network.Stream
}

// 任务状态
type TaskState struct {
    Task *Task
    Responses map[string]*TaskResponse // operator address -> response
    CurrentWeight *big.Int
    Status TaskStatus
    ExpireTime time.Time
    RetryCount int
}
```

1. 处理用户请求 

```go
func (tp *TaskProcessor) HandleUserRequest(stream network.Stream) {
    // 1. 读取用户请求
    req := readUserRequest(stream)
    
    // 2. 查询链上状态
    history, err := tp.queryStateManager(req)
    if err != nil {
        sendErrorResponse(stream, err)
        return
    }
    
    // 3. 生成任务ID
    taskID := generateTaskID(req, history)
    
    // 4. 创建任务
    task := &Task{
        ID: taskID,
        TargetAddress: req.TargetAddress,
        ChainID: req.ChainID,
        BlockNumber: history.BlockNumber,
        Timestamp: history.Timestamp,
        Key: req.Key,
        Value: history.Value,
        ExpireTime: time.Now().Add(15 * time.Second),
    }
    
    // 5. 保存用户stream用于返回结果
    tp.userStreams[taskID] = stream
    
    // 6. 广播任务
    tp.broadcastTask(task)
}
```

b.  任务broadcast 

```go
func (tp *TaskProcessor) broadcastTask(task *Task) error {
    // 1. 序列化任务
    taskBytes, err := proto.Marshal(task)
    if err != nil {
        return err
    }
    
    // 2. 发布到任务topic
    return tp.pubsub.Publish(TaskTopicName, taskBytes)
}
```

1. **任务处理和响应**
    1. **处理收到的task** 
    
    ```go
    func (tp *TaskProcessor) handleIncomingTask() {
        for {
            msg, err := tp.taskSub.Next(context.Background())
            if err != nil {
                continue
            }
            
            // 1. 解析任务
            task := &Task{}
            if err := proto.Unmarshal(msg.Data, task); err != nil {
                continue
            }
            
            // 2. 验证状态
            if err := tp.validateState(task); err != nil {
                continue
            }
            
            // 3. 生成签名
            signature, err := tp.signer.Sign(task)
            if err != nil {
                continue
            }
            
            // 4. 广播响应
            tp.broadcastResponse(&TaskResponse{
                TaskID: task.ID,
                OperatorAddress: tp.config.Address,
                Signature: signature,
                Epoch: task.Epoch,
            })
        }
    }
    ```
    
    b. 处理响应
    
    ```go
    func (tp *TaskProcessor) handleTaskResponse() {
        for {
            msg, err := tp.responseSub.Next(context.Background())
            if err != nil {
                continue
            }
            
            // 1. 解析响应
            response := &TaskResponse{}
            if err := proto.Unmarshal(msg.Data, response); err != nil {
                continue
            }
            
            // 2. 验证签名
            if err := tp.verifySignature(response); err != nil {
                tp.updateMissingCount(response.OperatorAddress)
                continue
            }
            
            // 3. 更新任务状态
            tp.updateTaskState(response)
        }
    }
    
    func (tp *TaskProcessor) updateTaskState(response *TaskResponse) {
        taskState := tp.tasks[response.TaskID]
        if taskState == nil {
            return
        }
        
        // 1. 添加响应
        taskState.Responses[response.OperatorAddress] = response
        
        // 2. 更新权重
        weight := tp.getOperatorWeight(response.OperatorAddress, response.Epoch)
        taskState.CurrentWeight.Add(taskState.CurrentWeight, weight)
        
        // 3. 检查是否达到阈值
        if tp.checkThreshold(taskState) {
            tp.finalizeTask(taskState)
        }
    }
    ```
    
2. 结果aggregate 和 返回
    1. 完成任务 
    
    ```go
    func (tp *TaskProcessor) finalizeTask(taskState *TaskState) {
        // 1. 聚合签名
        signatures := make([][]byte, 0)
        for _, resp := range taskState.Responses {
            signatures = append(signatures, resp.Signature)
        }
        
        // 2. 创建最终结果
        result := &TaskResult{
            TaskID: taskState.Task.ID,
            Signatures: signatures,
            EpochNumber: taskState.Task.Epoch,
            Value: taskState.Task.Value,
        }
        
        // 3. 广播结果
        tp.broadcastResult(result)
        
        // 4. 返回给用户
        if stream := tp.userStreams[taskState.Task.ID] {
            sendResultToUser(stream, result)
            delete(tp.userStreams, taskState.Task.ID)
        }
    }
    ```
    
    b. 超时处理 
    
    ```go
    func (tp *TaskProcessor) handleTaskTimeout() {
        ticker := time.NewTicker(time.Second)
        for range ticker.C {
            now := time.Now()
            for taskID, state := range tp.tasks {
                if now.After(state.ExpireTime) {
                    if state.RetryCount < 3 {
                        // 重试
                        tp.retryTask(state)
                    } else {
                        // 标记失败
                        tp.markTaskFailed(state)
                    }
                }
            }
        }
    }
    ```
    
3. 本地状态管理  

```go
type LocalTaskManager struct {
    db *LocalDB
    
    // 活跃任务缓存
    activeTasks map[string]*TaskState
    
    // 任务清理
    cleanupInterval time.Duration
}

func (ltm *LocalTaskManager) Start() {
    // 1. 定期清理过期任务
    go ltm.cleanupTasks()
    
    // 2. 维护本地状态
    go ltm.maintainTaskStates()
}
```

总结：

任何operator节点都可以接收用户请求

使用pubsub广播任务和响应

所有节点维护任务状态

共识：

所有active operator参与验证

保持weight阈值机制

分布式结果聚合

结果返回：

接收请求的节点负责返回结果

其他节点也维护完整状态

支持故障转移

容错机制：

任务超时重试

签名验证

状态同步

### epoch 更新

1. **Registry Node的Epoch更新职责** 

```sql
type RegistryNode struct {
    // ... 其他字段
    
    // epoch管理
    epochManager *EpochManager
    
    // 状态广播
    stateBroadcaster *StateBroadcaster
}

type EpochManager struct {
    // 创世区块
    genesisBlock uint64
    
    // 区块间隔
    blockInterval uint64
    
    // 当前epoch
    currentEpoch uint32
    
    // 数据库连接
    db *Database
}

// 监听区块并更新epoch
func (em *EpochManager) MonitorEpochUpdate() {
    // 计算下一个epoch的开始区块
    nextEpochBlock := em.genesisBlock + (uint64(em.currentEpoch+1) * em.blockInterval)
    
    // 监听新区块
    for block := range em.blockFeed {
        if block.Number == nextEpochBlock {
            // 1. 更新epoch状态
            newState, err := em.updateEpochState(em.currentEpoch + 1)
            if err != nil {
                continue
            }
            
            // 2. 保存到数据库
            if err := em.db.SaveEpochState(newState); err != nil {
                continue
            }
            
            // 3. 广播更新
            em.broadcastEpochUpdate(newState)
            
            // 4. 更新下一个目标区块
            nextEpochBlock += em.blockInterval
        }
    }
}

// 更新epoch状态
func (em *EpochManager) updateEpochState(epoch uint32) (*EpochState, error) {
    // 1. 获取所有operator
    operators, err := em.db.GetAllOperators()
    if err != nil {
        return nil, err
    }
    
    // 2. 更新operator状态和权重
    for _, op := range operators {
        // 更新状态
        if op.ActiveEpoch == epoch {
            op.Status = "active"
        }
        if op.ExitEpoch == epoch {
            op.Status = "unactive"
        }
        
        // 获取新权重
        weight, err := em.stakeRegistry.GetOperatorWeightAtEpoch(op.Address, epoch)
        if err != nil {
            continue
        }
        op.Weight = weight
        
        // 更新签名密钥
        signingKey, err := em.stakeRegistry.GetOperatorSigningKeyAtEpoch(op.Address, epoch)
        if err != nil {
            continue
        }
        op.SigningKey = signingKey
    }
    
    // 3. 创建新的epoch状态
    state := &EpochState{
        EpochNumber: epoch,
        BlockNumber: em.genesisBlock + (uint64(epoch) * em.blockInterval),
        MinimumWeight: em.getMinimumWeight(),
        TotalWeight: em.calculateTotalWeight(operators),
        ThresholdWeight: em.getThresholdWeight(epoch),
        UpdatedAt: time.Now(),
        Operators: operators,
    }
    
    return state, nil
}
```

1. p2p 状态同步机制

```sql
// 定义epoch更新topic
const EpochUpdateTopic = "/spotted-oracle/epoch-update/1.0.0"

type StateBroadcaster struct {
    pubsub *pubsub.PubSub
    host host.Host
}

// 广播epoch更新
func (sb *StateBroadcaster) BroadcastEpochUpdate(state *EpochState) error {
    // 1. 序列化状态
    stateBytes, err := proto.Marshal(state)
    if err != nil {
        return err
    }
    
    // 2. 发布到topic
    return sb.pubsub.Publish(EpochUpdateTopic, stateBytes)
}
```

1. operator node 状态同步 

```sql
type OperatorNode struct {
    // ... 其他字段
    
    // epoch状态管理
    epochState *EpochState
    
    // 本地数据库
    localDB *LocalDB
}

// 订阅epoch更新
func (node *OperatorNode) SubscribeEpochUpdates() error {
    // 1. 订阅topic
    sub, err := node.pubsub.Subscribe(EpochUpdateTopic)
    if err != nil {
        return err
    }
    
    // 2. 处理更新
    go func() {
        for {
            msg, err := sub.Next(context.Background())
            if err != nil {
                continue
            }
            
            // 解析状态
            state := &EpochState{}
            if err := proto.Unmarshal(msg.Data, state); err != nil {
                continue
            }
            
            // 验证更新来源
            if !node.isRegistryNode(msg.From) {
                continue
            }
            
            // 更新本地状态
            node.updateLocalEpochState(state)
        }
    }()
    
    return nil
}

// 更新本地状态
func (node *OperatorNode) updateLocalEpochState(state *EpochState) error {
    // 1. 更新内存状态
    node.epochState = state
    
    // 2. 保存到本地数据库
    return node.localDB.SaveCurrentEpochState(state)
}
```

1. 状态验证和恢复机制 

```sql
type StateValidator struct {
    // Registry Node列表
    registryNodes []peer.ID
    
    // 最近的状态hash
    latestStateHash []byte
}

// 验证状态一致性
func (sv *StateValidator) ValidateState() error {
    // 1. 获取所有Registry Node的状态hash
    hashes := make(map[string]int)
    for _, node := range sv.registryNodes {
        hash, err := sv.getStateHashFromNode(node)
        if err != nil {
            continue
        }
        hashes[string(hash)]++
    }
    
    // 2. 检查一致性
    if !sv.checkConsistency(hashes) {
        // 3. 如果不一致，从多数Registry Node同步
        return sv.syncFromMajority()
    }
    
    return nil
}
```

### calculate and submit rewards

TODO 暂时留接口不实现

## go-p2plib 可能的应用示例

1. 基础网络结构  

```go
// pkg/p2p/host.go

type P2PHost struct {
    host host.Host
    pubsub *pubsub.PubSub
    dht *dht.IpfsDHT
    
    // 配置
    config *Config
}

type Config struct {
    // 监听地址
    ListenAddrs []string
    // Bootstrap节点列表
    BootstrapPeers []string
    // 私钥路径
    PrivateKeyPath string
}

// 创建新的P2P Host
func NewP2PHost(cfg *Config) (*P2PHost, error) {
    // 1. 创建libp2p host选项
    opts := []libp2p.Option{
        libp2p.ListenAddrStrings(cfg.ListenAddrs...),
        libp2p.Identity(loadPrivateKey(cfg.PrivateKeyPath)),
        libp2p.EnableRelay(),
        libp2p.EnableAutoRelay(),
        libp2p.NATPortMap(),
    }
    
    // 2. 创建host
    h, err := libp2p.New(opts...)
    if err != nil {
        return nil, err
    }
    
    // 3. 创建DHT
    dht, err := dht.New(context.Background(), h)
    if err != nil {
        return nil, err
    }
    
    // 4. 创建pubsub
    ps, err := pubsub.NewGossipSub(context.Background(), h)
    if err != nil {
        return nil, err
    }
    
    return &P2PHost{
        host:    h,
        pubsub:  ps,
        dht:     dht,
        config:  cfg,
    }, nil
}
```

1. **Registry Node 网络实现** 

```go
// pkg/registry/node.go

type RegistryNode struct {
    p2p *p2p.P2PHost
    db  *Database
    
    // 协议处理器
    protocols *ProtocolHandlers
}

// 协议处理器
type ProtocolHandlers struct {
    // 节点发现协议
    discovery *DiscoveryProtocol
    // 注册协议
    registration *RegistrationProtocol
    // 状态同步协议
    stateSync *StateSyncProtocol
}

// 设置协议处理器
func (n *RegistryNode) setupProtocols() error {
    // 1. 设置节点发现协议
    n.p2p.host.SetStreamHandler("/discovery/1.0.0", func(s network.Stream) {
        n.protocols.discovery.HandleDiscoveryRequest(s)
    })
    
    // 2. 设置注册协议
    n.p2p.host.SetStreamHandler("/registration/1.0.0", func(s network.Stream) {
        n.protocols.registration.HandleRegistration(s)
    })
    
    // 3. 设置状态同步协议
    n.p2p.host.SetStreamHandler("/state-sync/1.0.0", func(s network.Stream) {
        n.protocols.stateSync.HandleStateSync(s)
    })
    
    return nil
}

// 启动Registry Node
func (n *RegistryNode) Start(ctx context.Context) error {
    // 1. 启动P2P服务
    if err := n.p2p.Start(ctx); err != nil {
        return err
    }
    
    // 2. 设置协议处理器
    if err := n.setupProtocols(); err != nil {
        return err
    }
    
    // 3. 启动状态广播
    go n.startStateAdvertise(ctx)
    
    return nil
}
```

1. **Operator Node 网络实现** 

```go
// pkg/operator/node.go

type OperatorNode struct {
    p2p *p2p.P2PHost
    db  *LocalDatabase
    
    // 协议处理器
    protocols *ProtocolHandlers
    // Registry Node连接管理
    registryConn *RegistryConnection
}

// Registry Node连接管理
type RegistryConnection struct {
    registryPeers []peer.ID
    activeConns   map[peer.ID]network.Stream
}

// 连接Registry Node
func (n *OperatorNode) connectToRegistry() error {
    // 1. 发现Registry Node
    peers, err := n.findRegistryNodes()
    if err != nil {
        return err
    }
    
    // 2. 建立连接
    for _, p := range peers {
        if err := n.connectPeer(p); err != nil {
            continue
        }
        // 3. 发送Join请求
        if err := n.sendJoinRequest(p); err != nil {
            continue
        }
    }
    
    return nil
}

// 设置协议处理器
func (n *OperatorNode) setupProtocols() error {
    // 1. 设置任务处理协议
    n.p2p.host.SetStreamHandler("/task/1.0.0", func(s network.Stream) {
        n.protocols.task.HandleTask(s)
    })
    
    // 2. 设置响应处理协议
    n.p2p.host.SetStreamHandler("/response/1.0.0", func(s network.Stream) {
        n.protocols.response.HandleResponse(s)
    })
    
    return nil
}
```

4. p2p 协议实现 

```go
// pkg/p2p/protocols/discovery.go

type DiscoveryProtocol struct {
    host host.Host
    dht  *dht.IpfsDHT
}

// 处理发现请求
func (d *DiscoveryProtocol) HandleDiscoveryRequest(s network.Stream) {
    // 1. 读取请求
    req := &pb.DiscoveryRequest{}
    if err := readProto(s, req); err != nil {
        s.Reset()
        return
    }
    
    // 2. 查找节点
    peers := d.findPeers(req.ServiceName)
    
    // 3. 返回响应
    resp := &pb.DiscoveryResponse{
        Peers: peersToProto(peers),
    }
    if err := writeProto(s, resp); err != nil {
        s.Reset()
        return
    }
}

// pkg/p2p/protocols/registration.go

type RegistrationProtocol struct {
    host host.Host
    db   *Database
}

// 处理注册请求
func (r *RegistrationProtocol) HandleRegistration(s network.Stream) {
    // 1. 读取请求
    req := &pb.RegistrationRequest{}
    if err := readProto(s, req); err != nil {
        s.Reset()
        return
    }
    
    // 2. 验证请求
    if err := r.validateRequest(req); err != nil {
        sendError(s, err)
        return
    }
    
    // 3. 处理注册
    if err := r.processRegistration(req); err != nil {
        sendError(s, err)
        return
    }
    
    // 4. 返回响应
    resp := &pb.RegistrationResponse{
        Status: "success",
    }
    if err := writeProto(s, resp); err != nil {
        s.Reset()
        return
    }
}
```

1. pubsub 实现 

```go
// pkg/p2p/pubsub.go

type PubSubManager struct {
    ps *pubsub.PubSub
    
    // 订阅的topics
    topics map[string]*pubsub.Topic
    subs   map[string]*pubsub.Subscription
}

// 创建新的topic
func (m *PubSubManager) JoinTopic(name string) (*pubsub.Topic, error) {
    topic, err := m.ps.Join(name)
    if err != nil {
        return nil, err
    }
    
    m.topics[name] = topic
    return topic, nil
}

// 订阅topic
func (m *PubSubManager) Subscribe(topicName string) (*pubsub.Subscription, error) {
    topic, ok := m.topics[topicName]
    if !ok {
        var err error
        topic, err = m.JoinTopic(topicName)
        if err != nil {
            return nil, err
        }
    }
    
    sub, err := topic.Subscribe()
    if err != nil {
        return nil, err
    }
    
    m.subs[topicName] = sub
    return sub, nil
}

// 发布消息
func (m *PubSubManager) Publish(topicName string, data []byte) error {
    topic, ok := m.topics[topicName]
    if !ok {
        return fmt.Errorf("topic not found: %s", topicName)
    }
    
    return topic.Publish(context.Background(), data)
}
```

1. 健康检查实现  

```go
// pkg/p2p/healthcheck.go

type HealthCheck struct {
    host host.Host
    
    // 节点状态
    peerStatus map[peer.ID]*PeerStatus
    // 检查间隔
    checkInterval time.Duration
}

type PeerStatus struct {
    LastSeen time.Time
    Status   string
    Latency  time.Duration
}

// 启动健康检查
func (hc *HealthCheck) Start(ctx context.Context) {
    ticker := time.NewTicker(hc.checkInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            hc.checkPeers()
        }
    }
}

// 检查节点健康状态
func (hc *HealthCheck) checkPeers() {
    for _, p := range hc.host.Network().Peers() {
        // 1. 发送ping
        latency, err := hc.ping(p)
        if err != nil {
            hc.updatePeerStatus(p, "offline", 0)
            continue
        }
        
        // 2. 更新状态
        hc.updatePeerStatus(p, "online", latency)
    }
}
```

# directory tree

```markdown
.
├── README.md                     # 项目说明文档
├── go.mod                       # Go模块定义
├── go.sum                       # Go依赖版本锁定
│
├── cmd/                         # 可执行程序入口
│   ├── registry/               # Registry节点程序
│   │   └── main.go            # Registry节点入口
│   └── operator/              # Operator节点程序
│       └── main.go            # Operator节点入口
│
├── config/                     # 配置文件目录
│   ├── registry.yaml          # Registry节点配置
│   └── operator.yaml          # Operator节点配置
│
├── pkg/                        # 项目包目录
│   ├── common/                # 公共代码
│   │   ├── contracts/        # 合约相关代码
│   │   │   ├── bindings/    # 合约绑定代码
│   │   │   │   ├── delegation_manager.go
│   │   │   │   ├── ecdsa_stake_registry.go
│   │   │   │   ├── epoch_manager.go
│   │   │   │   └── state_manager.go
│   │   │   │
│   │   │   └── ethereum/   # 以太坊客户端相关
│   │   │       ├── chain_clients.go
│   │   │       ├── client.go
│   │   │       ├── delegation.go
│   │   │       ├── epoch.go
│   │   │       ├── registry.go
│   │   │       ├── state_only_client.go
│   │   │       └── state.go
│   │   │
│   │   ├── types/            # 共享数据类型
│   │   │   ├── task.go       # 任务相关结构
│   │   │   ├── epoch.go      # Epoch相关结构
│   │   │   └── operator.go   # Operator相关结构
│   │   │
│   │   ├── crypto/           # 加密相关
│   │   │   ├── signer.go     # 签名实现
│   │   │   └── validator.go  # 签名验证
│   │   │
│   │   └── utils/            # 工具函数
│   │       ├── config.go     # 配置加载
│   │       └── logger.go     # 日志工具
│   │
│   ├── p2p/                   # P2P网络相关
│   │   ├── host.go           # libp2p host配置
│   │   ├── discovery.go      # 节点发现
│   │   ├── pubsub.go         # 发布订阅
│   │   └── protocols/        # 协议实现
│   │       ├── task.go       # 任务协议
│   │       ├── response.go   # 响应协议
│   │       └── sync.go       # 状态同步协议
│   │
│   ├── registry/             # Registry节点实现
│   │   ├── node.go          # Registry节点核心逻辑
│   │   ├── manager/         # 管理器
│   │   │   ├── operator.go  # Operator管理
│   │   │   └── epoch.go     # Epoch管理
│   │   └── sync/           # Registry同步
│   │       └── state.go    # 状态同步
│   │
│   ├── operator/            # Operator节点实现
│   │   ├── node.go         # Operator节点核心逻辑
│   │   ├── task/          # 任务处理
│   │   │   ├── processor.go # 任务处理器
│   │   │   └── validator.go # 状态验证器
│   │   └── state/         # 状态管理
│   │       └── manager.go  # 本地状态管理
│   │
│   ├── repos/              # 数据库相关代码
│   │   ├── sqlc.yaml      # sqlc主配置文件
│   │   │
│   │   ├── registry/      # Registry Node数据库
│   │   │   ├── operators/     # operators表
│   │   │   │   ├── schema.sql # 表结构
│   │   │   │   └── query.sql  # SQL查询
│   │   │   ├── epoch_states/  # epoch_states表
│   │   │   │   ├── schema.sql
│   │   │   │   └── query.sql
│   │   │   ├── epoch_operators/ # epoch_operators表
│   │   │   │   ├── schema.sql
│   │   │   │   └── query.sql
│   │   │   └── registry_events/ # registry_events表
│   │   │       ├── schema.sql
│   │   │       └── query.sql
│   │   │
│   │   └── operator/      # Operator Node数据库
│   │       ├── local_operator/ # local_operator表
│   │       │   ├── schema.sql
│   │       │   └── query.sql
│   │       ├── current_epoch/  # current_epoch表
│   │       │   ├── schema.sql
│   │       │   └── query.sql
│   │       ├── tasks/         # tasks表
│   │       │   ├── schema.sql
│   │       │   └── query.sql
│   │       ├── task_responses/ # task_responses表
│   │       │   ├── schema.sql
│   │       │   └── query.sql
│   │       └── peer_cache/    # peer_cache表
│   │           ├── schema.sql
│   │           └── query.sql
│   │
│   └── db/                # 数据库连接管理
│       ├── connection.go  # 数据库连接
│       └── migration.go   # 数据库迁移
│
├── proto/                 # Protocol Buffers定义
│   ├── task.proto        # 任务相关消息
│   ├── registry.proto    # Registry服务定义
│   └── sync.proto        # 同步协议消息
│
├── scripts/              # 脚本文件
│   ├── deploy.sh        # 部署脚本
│   └── monitor.sh       # 监控脚本
│
├── test/                # 测试代码
│   ├── integration/    # 集成测试
│   │   ├── registry/  # Registry节点测试
│   │   └── operator/  # Operator节点测试
│   └── unit/          # 单元测试
│       ├── common/    # 公共模块测试
│       ├── p2p/       # P2P模块测试
│       └── db/        # 数据库测试
│
└── docs/               # 文档
    ├── design.md      # 设计文档
    ├── api.md         # API文档
    └── deployment.md  # 部署文档
```

## 初步规划

我来帮您规划项目的开发阶段：

**第一阶段：基础网络搭建 (Network Bootstrap)**

**目标**: 实现基本的P2P网络功能，让Registry Node和Operator Node能够互相发现和通信

1. **基础设施搭建**:
- **基础设施搭建**:
- 搭建基本项目结构
- 配置开发环境
- 设置Docker开发环境
1. **Registry Node基础功能**:
- **Registry Node基础功能**:
- libp2p host初始化
- 基本数据库结构(operators表)
- 简单的节点管理逻辑
- Bootstrap服务实现

3. **Operator Node基础功能**:

- libp2p host初始化
- 本地数据库结构
- 基本的节点发现逻辑
- 连接Registry Node功能
1. **健康检查系统**:
- **健康检查系统**:
- 实现简单的心跳机制
- 基本的节点状态监控
- 节点存活检测
1. **本地测试环境**:

```sql
   # docker-compose.yml
   services:
     postgres:
       image: postgres:latest
       environment:
         POSTGRES_DB: avs
         POSTGRES_USER: user
         POSTGRES_PASSWORD: password
       
     registry-node:
       build: 
         context: .
         dockerfile: cmd/registry/Dockerfile
       depends_on:
         - postgres
       
     operator-node-1:
       build:
         context: .
         dockerfile: cmd/operator/Dockerfile
       depends_on:
         - registry-node
       
     operator-node-2:
       build:
         context: .
         dockerfile: cmd/operator/Dockerfile
       depends_on:
         - registry-node
```

**第二阶段：合约集成与注册流程 (Contract Integration)**

**目标**: 实现与智能合约的交互，完成完整的operator注册流程

1. **合约集成**:
- **合约集成**
- 实现合约绑定代码
- 设置合约事件监听
- 实现合约调用接口
1. **注册流程实现**:
- Operator注册事件监听
- 完整的JoinNetwork流程
- 注销流程实现
1. **状态管理**:
- Registry Node完整状态管理
- Operator Node本地状态管理
- 状态同步机制

**第三阶段：Epoch管理与P2P通信 (Epoch & Communication)**

**目标**: 实现完整的epoch管理和P2P通信协议

1. **Epoch管理**:
- 区块监听机制
- Epoch状态更新
- Operator权重更新
- 状态广播机制

2. **P2P协议实现**:

- 完整的pubsub系统
- 协议消息定义
- 消息处理流程
1. **状态同步**:
- Registry Node间状态同步
- Operator Node状态订阅
- 冲突解决机制

**第四阶段：任务处理系统 (Task Processing)**

**目标**: 实现完整的任务处理流程

1. **任务系统**
- 任务生成逻辑
- 任务分发机制
- 响应收集系统
1. **签名系统**:
- 签名生成
- 签名验证
- 签名聚合
1. **结果处理**:
- 结果聚合
- 用户响应
- 超时处理

**第五阶段：测试与优化 (Testing & Optimization)**

**目标**: 全面测试和性能优化

1. **测试系统**:

- 单元测试
- 集成测试
- 性能测试
1. **监控系统**:
- **监控系统**:
- 性能指标收集
- 日志系统
- 警报机制
1. **文档完善**:
- **文档完善**:
- API文档
- 部署文档
- 运维手册