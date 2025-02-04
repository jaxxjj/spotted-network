package operator

import (
	"context"
	"fmt"
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
	"github.com/stumble/dcache"
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
func (s *OperatorTestSuite) SetupTest() {
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
	
	s := &OperatorTestSuite{
		WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("operator_test", []string{
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