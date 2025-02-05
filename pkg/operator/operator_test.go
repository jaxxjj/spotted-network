package operator

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/coocood/freecache"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
)

type MockHost struct {
	mock.Mock
}

func (m *MockHost) ID() peer.ID {
	args := m.Called()
	return args.Get(0).(peer.ID)
}

func (m *MockHost) Network() network.Network {
	args := m.Called()
	return args.Get(0).(network.Network)
}

// 添加必要的 mock 实现
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	args := m.Called(topic)
	return args.Get(0).(*pubsub.Topic), args.Error(1)
}

func (m *MockPubSub) BlacklistPeer(p peer.ID) {
	m.Called(p)
}

// 需要添加新的mock实现
type MockChainManager struct {
	mock.Mock
}

func (m *MockChainManager) GetChainID() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockChainManager) GetClientByChainId(chainId uint32) (*ethereum.ChainClient, error) {
	args := m.Called(chainId)
	return args.Get(0).(*ethereum.ChainClient), args.Error(1)
}

func (m *MockChainManager) GetMainnetClient() (*ethereum.ChainClient, error) {
	args := m.Called()
	return args.Get(0).(*ethereum.ChainClient), args.Error(1)
}

// MockChainClient mocks ethereum.ChainClient
type MockChainClient struct {
	mock.Mock
}

func (m *MockChainClient) GetMinimumWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockChainClient) GetTotalWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockChainClient) GetThresholdWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}


// TimePointSet 用于模板中的时间点
type TimePointSet struct {
	Now  time.Time
	P12H string // 12小时前
	P24H string // 24小时前
}

// Serde 定义了序列化和反序列化接口
type Serde interface {
	Load(data []byte) error
	Dump() ([]byte, error)
}

type TasksTableSerde struct {
	tasksQuerier *tasks.Queries
}

func (o TasksTableSerde) Load(data []byte) error {
	if err := o.tasksQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load tasks data: %w", err)
	}
	return nil
}

func (o TasksTableSerde) Dump() ([]byte, error) {
	return o.tasksQuerier.Dump(context.Background(), func(m *tasks.Tasks) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

// MockP2PHost implements P2PHost interface
type MockP2PHost struct {
	mock.Mock
}

func (m *MockP2PHost) ID() peer.ID {
	args := m.Called()
	return args.Get(0).(peer.ID)
}

func (m *MockP2PHost) Addrs() []multiaddr.Multiaddr {
	args := m.Called()
	return args.Get(0).([]multiaddr.Multiaddr)
}

func (m *MockP2PHost) Connect(ctx context.Context, pi peer.AddrInfo) error {
	args := m.Called(ctx, pi)
	return args.Error(0)
}

func (m *MockP2PHost) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	args := m.Called(ctx, p, pids)
	return args.Get(0).(network.Stream), args.Error(1)
}

func (m *MockP2PHost) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	m.Called(pid, handler)
}

func (m *MockP2PHost) RemoveStreamHandler(pid protocol.ID) {
	m.Called(pid)
}

func (m *MockP2PHost) Network() network.Network {
	args := m.Called()
	return args.Get(0).(network.Network)
}

func (m *MockP2PHost) Peerstore() peerstore.Peerstore {
	args := m.Called()
	return args.Get(0).(peerstore.Peerstore)
}

func (m *MockP2PHost) Close() error {
	args := m.Called()
	return args.Error(0)
}

// OperatorTestSuite 是主测试套件
type OperatorTestSuite struct {
	*testsuite.WPgxTestSuite
	// 核心组件
	node          *Node
	
	// 数据库组件
	tasksQuerier  *tasks.Queries
	taskResponseQuerier *task_responses.Queries
	consensusResponseQuerier *consensus_responses.Queries
	epochStateQuerier *epoch_states.Queries
	
	// 缓存组件
	redisConn     redis.UniversalClient
	freeCache     *freecache.Cache
	dcache        *dcache.DCache
	
	// Mock 组件
	mockHost      *MockHost
	mockPubSub    *MockPubSub
	mockChainManager *MockChainManager
	mockP2PHost   *MockP2PHost
	
	// 测试辅助
	ctx           context.Context
}


func TestOperatorSuite(t *testing.T) {
	suite.Run(t, NewOperatorTestSuite())
} 

// SetupTest 在每个测试前运行
func (s *OperatorTestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.ctx = context.Background()
	
	// 清理缓存
	s.Require().NoError(s.redisConn.FlushAll(context.Background()).Err())
	s.freeCache.Clear()
	
	// 初始化 queriers
	s.tasksQuerier = tasks.New(s.GetPool().WConn(), s.dcache)
	s.taskResponseQuerier = task_responses.New(s.GetPool().WConn(), s.dcache)
	s.consensusResponseQuerier = consensus_responses.New(s.GetPool().WConn(), s.dcache)
	s.epochStateQuerier = epoch_states.New(s.GetPool().WConn(), s.dcache)
	
	// 初始化 mocks
	s.mockP2PHost = new(MockP2PHost)
	s.mockChainManager = new(MockChainManager)

	
	// 创建节点
	s.node = &Node{
		host:            s.mockP2PHost,
		chainManager:   s.mockChainManager,
		tasksQuerier:   s.tasksQuerier,
		taskResponseQuerier: s.taskResponseQuerier,
		consensusResponseQuerier: s.consensusResponseQuerier,
		epochStateQuerier: s.epochStateQuerier,
		activeOperators: ActivePeerStates{
			active: make(map[peer.ID]*OperatorState),
		},
	}
}

// LoadState 从文件加载状态
func (s *OperatorTestSuite) LoadState(filename string, serde Serde) {
	s.WPgxTestSuite.LoadState(filename, serde)
}

// Golden 验证状态并生成 golden 文件
func (s *OperatorTestSuite) Golden(name string, serde Serde) {
	s.WPgxTestSuite.Golden(name, serde)
}

// GoldenVarJSON 验证 JSON 变量并生成 golden 文件
func (s *OperatorTestSuite) GoldenVarJSON(name string, v interface{}) {
	s.WPgxTestSuite.GoldenVarJSON(name, v)
}

// NewOperatorTestSuite 创建新的测试套件实例
func NewOperatorTestSuite() *OperatorTestSuite {
	// 初始化 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		ReadTimeout: 3 * time.Second,
		PoolSize:    50,
		Password:    "",
	})
	
	if redisClient.Ping(context.Background()).Err() != nil {
		panic(fmt.Errorf("redis connection failed to ping"))
	}
	
	// 创建 freecache 实例
	memCache := freecache.NewCache(100 * 1024 * 1024)
	
	// 创建 dcache
	dCache, err := dcache.NewDCache(
		"test", redisClient, memCache, 100*time.Millisecond, true, true)
	if err != nil {
		panic(err)
	}
	
	s := &OperatorTestSuite{
		WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("operator_test", []string{
			tasks.Schema,
			epoch_states.Schema,
			consensus_responses.Schema, 
			task_responses.Schema,
		}),
		redisConn:  redisClient,
		freeCache:  memCache,
		dcache:     dCache,
		mockHost:   new(MockHost),
		mockPubSub: new(MockPubSub),
	}
	
	// 设置基本的 mock 行为
	s.mockHost.On("ID").Return(peer.ID("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr"))
	s.mockPubSub.On("Join", mock.Anything).Return(new(pubsub.Topic), nil)
	
	return s
}





