package registry

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/coocood/freecache"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v5"
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
	"github.com/stumble/wpgx"
	"github.com/stumble/wpgx/testsuite"

	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
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

// 1. 为每个表定义 TableSerde
type OperatorsTableSerde struct {
	opQuerier *operators.Queries
}

func (o OperatorsTableSerde) Load(data []byte) error {
	if err := o.opQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load operators data: %w", err)
	}
	return nil
}

func (o OperatorsTableSerde) Dump() ([]byte, error) {
	return o.opQuerier.Dump(context.Background(), func(m *operators.Operators) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

type BlacklistTableSerde struct {
	blacklistQuerier *blacklist.Queries
}

func (b BlacklistTableSerde) Load(data []byte) error {
	err := b.blacklistQuerier.Load(context.Background(), data)
	if err != nil {
		return fmt.Errorf("failed to load blacklist data: %w", err)
	}
	
	return b.blacklistQuerier.RefreshIDSerial(context.Background())
}

func (b BlacklistTableSerde) Dump() ([]byte, error) {
	return b.blacklistQuerier.Dump(context.Background(), func(m *blacklist.Blacklist) {
		m.CreatedAt = time.Unix(0, 0).UTC()
	})
}

// RegistryTestSuite 是主测试套件
type RegistryTestSuite struct {
	*testsuite.WPgxTestSuite
	// 核心组件
	node          *Node
	
	// 数据库组件
	blacklistQuerier *blacklist.Queries
	opQuerier     *operators.Queries
	
	// 缓存组件
	redisConn     redis.UniversalClient
	freeCache     *freecache.Cache
	dcache        *dcache.DCache
	
	// Mock 组件
	mockHost      *MockHost
	mockPubSub    *MockPubSub
	mockMainnet   *MockMainnetClient
	mockTxManager *MockTxManager
	mockOpQuerier *MockOperatorsQuerier
	mockStateSync *MockStateSyncNotifier
	mockP2PHost   *MockP2PHost
	
	// 测试辅助
	ctx           context.Context
	timePoints    *TimePointSet
}

// MockPubSub 实现 PubSubService 接口
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) BlacklistPeer(p peer.ID) {
	m.Called(p)
}

func (m *MockPubSub) Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	args := m.Called(topic, opts)
	return args.Get(0).(*pubsub.Topic), args.Error(1)
}

// MockStateSyncNotifier 实现 StateSyncNotifier 接口
type MockStateSyncNotifier struct {
	mock.Mock
}

func (m *MockStateSyncNotifier) NotifyEpochUpdate(ctx context.Context, epoch uint32) error {
	args := m.Called(ctx, epoch)
	return args.Error(0)
}

// MockMainnetClient 实现 MainnetClient 接口
type MockMainnetClient struct {
	mock.Mock
}

func (m *MockMainnetClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockMainnetClient) GetOperatorWeight(ctx context.Context, addr ethcommon.Address) (*big.Int, error) {
	args := m.Called(ctx, addr)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockMainnetClient) GetOperatorSigningKey(ctx context.Context, addr ethcommon.Address, epoch uint32) (ethcommon.Address, error) {
	args := m.Called(ctx, addr, epoch)
	return args.Get(0).(ethcommon.Address), args.Error(1)
}

func (m *MockMainnetClient) GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error) {
	args := m.Called(ctx, blockNumber)
	return args.Get(0).(uint32), args.Error(1)
}

func (m *MockMainnetClient) IsOperatorRegistered(ctx context.Context, addr ethcommon.Address) (bool, error) {
	args := m.Called(ctx, addr)
	return args.Get(0).(bool), args.Error(1)
}

// MockTxManager 实现 TxManager 接口
type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) Transact(ctx context.Context, txOptions pgx.TxOptions, f func(context.Context, *wpgx.WTx) (any, error)) (any, error) {
	args := m.Called(ctx, txOptions, f)
	return args.Get(0), args.Error(1)
}

// MockOperatorsQuerier 实现 OperatorsQuerier 接口
type MockOperatorsQuerier struct {
	mock.Mock
	*operators.Queries
}

func (m *MockOperatorsQuerier) ListAllOperators(ctx context.Context) ([]operators.Operators, error) {
	args := m.Called(ctx)
	return args.Get(0).([]operators.Operators), args.Error(1)
}

func (m *MockOperatorsQuerier) WithTx(tx *wpgx.WTx) *operators.Queries {
	args := m.Called(tx)
	return args.Get(0).(*operators.Queries)
}

func (m *MockOperatorsQuerier) UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, addr *string) (*operators.Operators, error) {
	args := m.Called(ctx, arg, addr)
	return args.Get(0).(*operators.Operators), args.Error(1)
}

func (m *MockOperatorsQuerier) GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(*operators.Operators), args.Error(1)
}

func (m *MockOperatorsQuerier) UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams, addr *string) (*operators.Operators, error) {
	args := m.Called(ctx, arg, addr)
	return args.Get(0).(*operators.Operators), args.Error(1)
}

func (m *MockOperatorsQuerier) UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams, addr *string) (*operators.Operators, error) {
	args := m.Called(ctx, arg, addr)
	return args.Get(0).(*operators.Operators), args.Error(1)
}

func NewMockOperatorsQuerier() *MockOperatorsQuerier {
	return &MockOperatorsQuerier{
		Mock:    mock.Mock{},
		Queries: &operators.Queries{},
	}
}

// MockNode 实现 Node 接口用于测试
type MockNode struct {
	mock.Mock
}

func (m *MockNode) getActivePeerIDs() []peer.ID {
	args := m.Called()
	return args.Get(0).([]peer.ID)
}

func (m *MockNode) GetOperatorState(p peer.ID) *OperatorPeerInfo {
	args := m.Called(p)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*OperatorPeerInfo)
}

func (m *MockNode) UpdateOperatorState(p peer.ID, info *OperatorPeerInfo) {
	m.Called(p, info)
}

func (m *MockNode) disconnectPeer(p peer.ID) error {
	args := m.Called(p)
	return args.Error(0)
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



// SetupTest 在每个测试前运行
func (s *RegistryTestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.ctx = context.Background()
	s.Require().NoError(s.redisConn.FlushAll(context.Background()).Err())
	s.freeCache.Clear()
	s.blacklistQuerier = blacklist.New(s.GetPool().WConn(), s.dcache)
	s.opQuerier = operators.New(s.GetPool().WConn(), s.dcache)
	
	// 初始化 mockP2PHost
	s.mockP2PHost = new(MockP2PHost)
	
	s.node = &Node{
		blacklistQuerier: s.blacklistQuerier,
		opQuerier:       s.opQuerier,
		pubsub:          s.mockPubSub,
		mainnetClient:   s.mockMainnet,
		host:            s.mockP2PHost,  // 添加 host
	}
}

// LoadState 从文件加载状态
func (s *RegistryTestSuite) LoadState(filename string, serde Serde) {
	s.WPgxTestSuite.LoadState(filename, serde)
}

// Golden 验证状态并生成 golden 文件
func (s *RegistryTestSuite) Golden(name string, serde Serde) {
	s.WPgxTestSuite.Golden(name, serde)
}

// GoldenVarJSON 验证 JSON 变量并生成 golden 文件
func (s *RegistryTestSuite) GoldenVarJSON(name string, v interface{}) {
	s.WPgxTestSuite.GoldenVarJSON(name, v)
}


func TestOperatorSuite(t *testing.T) {
	suite.Run(t, NewRegistryTestSuite())
} 

// NewRegistryTestSuite 创建新的测试套件实例
func NewRegistryTestSuite() *RegistryTestSuite {
	// 初始化基础设施
	redisClient := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		ReadTimeout: 3 * time.Second,
		PoolSize:    50,
		Password:    "",
	})
	
	if redisClient.Ping(context.Background()).Err() != nil {
		panic(fmt.Errorf("redis connection failed to ping"))
	}
	
	// Create freecache instance
	memCache := freecache.NewCache(100 * 1024 * 1024)
	
	dCache, err := dcache.NewDCache(
		"test", redisClient, memCache, 100*time.Millisecond, true, true)
	if err != nil {
		panic(err)
	}
	
	s := &RegistryTestSuite{
		WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("registry_test", []string{
			blacklist.Schema,
			operators.Schema,
		}),
		redisConn: redisClient,
		freeCache: memCache,
		dcache:    dCache,
		mockHost:  new(MockHost),
		mockPubSub: new(MockPubSub),
	}
	
	// 设置基本的 mock 行为
	s.mockHost.On("ID").Return(peer.ID("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr"))
	s.mockPubSub.On("Join", mock.Anything).Return(new(pubsub.Topic), nil)
	
	return s
} 