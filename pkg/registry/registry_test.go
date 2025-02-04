package registry

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/coocood/freecache"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"

	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
)

// TimePointSet 用于模板中的时间点
type TimePointSet struct {
	Now  time.Time
	P12H string // 12小时前
	P24H string // 24小时前
}

// BlacklistState 用于序列化/反序列化黑名单状态
type BlacklistState struct {
	Blacklist []struct {
		PeerID    string     `json:"peer_id"`
		IP        string     `json:"ip"`
		Reason    string     `json:"reason"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
	} `json:"blacklist"`
}

// BlacklistStateManager 实现 Loader 和 Dumper 接口
type BlacklistStateManager struct {
	suite *RegistryTestSuite
}

func (m *BlacklistStateManager) Load(data []byte) error {
	return m.suite.loadBlacklistState(data)
}

func (m *BlacklistStateManager) Dump() ([]byte, error) {
	return m.suite.dumpBlacklistState()
}

// BlacklistSerde 实现 Serde 接口
type BlacklistSerde struct {
	blacklist *blacklist.Queries
}

func (b BlacklistSerde) Load(data []byte) error {
	var state BlacklistState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// 将状态应用到数据库
	for _, item := range state.Blacklist {
		_, err := b.blacklist.BlockNode(context.Background(), blacklist.BlockNodeParams{
			PeerID:    item.PeerID,
			Ip:        item.IP,
			Reason:    &item.Reason,
			ExpiresAt: item.ExpiresAt,
		}, &blacklist.IsBlockedParams{
			PeerID: item.PeerID,
			Ip:     item.IP,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (b BlacklistSerde) Dump() ([]byte, error) {
	// 从数据库查询黑名单状态
	blacklists, err := b.blacklist.ListBlacklist(context.Background(), blacklist.ListBlacklistParams{
		Offset: 0,
		Limit:  100,
	})
	if err != nil {
		return nil, err
	}

	state := BlacklistState{
		Blacklist: make([]struct {
			PeerID    string     `json:"peer_id"`
			IP        string     `json:"ip"`
			Reason    string     `json:"reason"`
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
			CreatedAt time.Time  `json:"created_at"`
			UpdatedAt time.Time  `json:"updated_at"`
		}, len(blacklists)),
	}

	for i, b := range blacklists {
		reason := ""
		if b.Reason != nil {
			reason = *b.Reason
		}
		state.Blacklist[i] = struct {
			PeerID    string     `json:"peer_id"`
			IP        string     `json:"ip"`
			Reason    string     `json:"reason"`
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
			CreatedAt time.Time  `json:"created_at"`
			UpdatedAt time.Time  `json:"updated_at"`
		}{
			PeerID:    b.PeerID,
			IP:        b.Ip,
			Reason:    reason,
			ExpiresAt: b.ExpiresAt,
			CreatedAt: time.Unix(0, 0).UTC(), // 使用固定时间戳
			UpdatedAt: time.Unix(0, 0).UTC(), // 使用固定时间戳
		}
	}

	// 按照 peer_id 和 ip 排序，以确保输出顺序一致
	sort.Slice(state.Blacklist, func(i, j int) bool {
		if state.Blacklist[i].PeerID != state.Blacklist[j].PeerID {
			return state.Blacklist[i].PeerID < state.Blacklist[j].PeerID
		}
		return state.Blacklist[i].IP < state.Blacklist[j].IP
	})
	
	return json.MarshalIndent(state, "", "    ")
}

// Serde 定义了序列化和反序列化接口
type Serde interface {
	Load(data []byte) error
	Dump() ([]byte, error)
}

// RegistryTestSuite 是主测试套件
type RegistryTestSuite struct {
	*testsuite.WPgxTestSuite
	// 核心组件
	node          *Node
	
	// Mock 组件
	mockHost      *MockHost
	mockPubSub    *MockPubSub
	redisConn     redis.UniversalClient
	dcache        *dcache.DCache
	blacklistDB   *blacklist.Queries
	
	// 测试辅助
	ctx           context.Context
	timePoints    *TimePointSet
	stateManager  *BlacklistStateManager
}

// MockHost 实现 host.Host 接口
type MockHost struct {
	mock.Mock
}

func (m *MockHost) ID() peer.ID {
	args := m.Called()
	return args.Get(0).(peer.ID)
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

// TestRegistrySuite 运行测试套件
func TestRegistrySuite(t *testing.T) {
	suite.Run(t, NewRegistryTestSuite())
}

// NewRegistryTestSuite 创建新的测试套件实例
func NewRegistryTestSuite() *RegistryTestSuite {
	s := &RegistryTestSuite{
		WPgxTestSuite: testsuite.NewWPgxTestSuiteFromEnv("registry_test", []string{
			// 添加 blacklist schema
			`CREATE TABLE IF NOT EXISTS blacklist (
				id              BIGSERIAL       PRIMARY KEY,
				peer_id         VARCHAR(255)    NOT NULL,    -- libp2p peer ID
				ip              VARCHAR(45)     NOT NULL,    -- 节点的公网IP
				reason          TEXT           NULL,         -- 封禁原因
				created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
				expires_at      TIMESTAMPTZ    NULL CHECK (expires_at > created_at),
				CONSTRAINT valid_peer_id CHECK (
					peer_id ~ '^12D3KooW[1-9A-HJ-NP-Za-km-z]{44,48}$'  -- libp2p peer ID
				),
				CONSTRAINT valid_ip CHECK (
					ip ~ '^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$' OR          -- IPv4
					ip ~ '^[0-9a-fA-F:]+$'                             -- IPv6
				),
				CONSTRAINT unique_peer_ip UNIQUE (peer_id, ip)
			);
			CREATE INDEX IF NOT EXISTS blacklist_peer_id_idx ON blacklist(peer_id);
			CREATE INDEX IF NOT EXISTS blacklist_ip_idx ON blacklist(ip);
			CREATE INDEX IF NOT EXISTS blacklist_created_at_idx ON blacklist(created_at);
			CREATE INDEX IF NOT EXISTS blacklist_expires_at_idx ON blacklist(expires_at);`,
		}),
	}
	s.stateManager = &BlacklistStateManager{suite: s}
	return s
}

// SetupTest 在每个测试前运行
func (s *RegistryTestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.ctx = context.Background()
	
	// 初始化时间点
	now := time.Now().UTC()
	s.timePoints = &TimePointSet{
		Now:  now,
		P12H: now.Add(-12 * time.Hour).Format(time.RFC3339),
		P24H: now.Add(-24 * time.Hour).Format(time.RFC3339),
	}
	
	// 初始化 Redis 和 DCache
	s.redisConn = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{"localhost:6379"},
		DB:    0,
	})
	var err error
	s.dcache, err = dcache.NewDCache(
		"test_registry",
		s.redisConn,
		freecache.NewCache(50*1024*1024), // 50MB
		500*time.Millisecond,
		true,
		true,
	)
	s.Require().NoError(err)
	
	// 初始化 mocks
	s.mockHost = new(MockHost)
	s.mockPubSub = new(MockPubSub)
	
	// 初始化 blacklist 仓库
	s.blacklistDB = blacklist.New(s.GetPool().WConn(), s.dcache)
	
	// 设置基本的 mock 行为
	s.mockHost.On("ID").Return(peer.ID("test-host"))
	s.mockPubSub.On("Join", mock.Anything).Return(new(pubsub.Topic), nil)

	// 清理数据库
	_, err = s.GetPool().WConn().WExec(s.ctx, "blacklist.truncate", "TRUNCATE TABLE blacklist")
	s.Require().NoError(err)
}

// loadBlacklistState 从 JSON 加载黑名单状态
func (s *RegistryTestSuite) loadBlacklistState(data []byte) error {
	var state BlacklistState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// 将状态应用到数据库
	for _, item := range state.Blacklist {
		_, err := s.blacklistDB.BlockNode(s.ctx, blacklist.BlockNodeParams{
			PeerID:    item.PeerID,
			Ip:        item.IP,
			Reason:    &item.Reason,
			ExpiresAt: item.ExpiresAt,
		}, &blacklist.IsBlockedParams{
			PeerID: item.PeerID,
			Ip:     item.IP,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// dumpBlacklistState 导出黑名单状态为 JSON
func (s *RegistryTestSuite) dumpBlacklistState() ([]byte, error) {
	// 从数据库查询黑名单状态
	blacklists, err := s.blacklistDB.ListBlacklist(s.ctx, blacklist.ListBlacklistParams{
		Offset: 0,
		Limit:  100,
	})
	if err != nil {
		return nil, err
	}

	state := BlacklistState{
		Blacklist: make([]struct {
			PeerID    string     `json:"peer_id"`
			IP        string     `json:"ip"`
			Reason    string     `json:"reason"`
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
			CreatedAt time.Time  `json:"created_at"`
			UpdatedAt time.Time  `json:"updated_at"`
		}, len(blacklists)),
	}

	for i, b := range blacklists {
		reason := ""
		if b.Reason != nil {
			reason = *b.Reason
		}
		state.Blacklist[i] = struct {
			PeerID    string     `json:"peer_id"`
			IP        string     `json:"ip"`
			Reason    string     `json:"reason"`
			ExpiresAt *time.Time `json:"expires_at,omitempty"`
			CreatedAt time.Time  `json:"created_at"`
			UpdatedAt time.Time  `json:"updated_at"`
		}{
			PeerID:    b.PeerID,
			IP:        b.Ip,
			Reason:    reason,
			ExpiresAt: b.ExpiresAt,
			CreatedAt: time.Unix(0, 0).UTC(), // 使用固定时间戳
			UpdatedAt: time.Unix(0, 0).UTC(), // 使用固定时间戳
		}
	}

	// 按照 peer_id 和 ip 排序，以确保输出顺序一致
	sort.Slice(state.Blacklist, func(i, j int) bool {
		if state.Blacklist[i].PeerID != state.Blacklist[j].PeerID {
			return state.Blacklist[i].PeerID < state.Blacklist[j].PeerID
		}
		return state.Blacklist[i].IP < state.Blacklist[j].IP
	})
	
	return json.MarshalIndent(state, "", "    ")
}

// TearDownTest 在每个测试后运行
func (s *RegistryTestSuite) TearDownTest() {
	s.WPgxTestSuite.TearDownTest()
	if s.dcache != nil {
		s.dcache.Close()
	}
	if s.redisConn != nil {
		err := s.redisConn.Close()
		s.Require().NoError(err)
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