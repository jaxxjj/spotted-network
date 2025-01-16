CREATE TABLE IF NOT EXISTS tasks (
    task_id TEXT PRIMARY KEY,  -- keccak256(abi.encodePacked(user, chainId, blockNumber, timestamp, epoch, key, value))
    target_address TEXT NOT NULL,  -- statemanager上存储key value pair的msg.sender
    chain_id INT NOT NULL,  -- 查询的chainid
    block_number NUMERIC(78),  -- 如果用户选择block number
    timestamp NUMERIC(78),  -- 如果用户选择timestamp
    epoch INT NOT NULL,  -- 通过EpochManager::getCurrentEpoch获得
    key NUMERIC(78) NOT NULL,  -- 查询的key
    value NUMERIC(78),  -- 查询的value，通过History.value获得
    expire_time TIMESTAMP NOT NULL,  -- 任务过期时间，15s
    retries INT DEFAULT 0,  -- 重试次数，最多3次
    status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'expired', 'failed')),  -- 任务状态
    created_at TIMESTAMP NOT NULL DEFAULT NOW()  -- 任务创建时间
);

-- 添加索引以提高查询性能
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_expire_time ON tasks(expire_time); 