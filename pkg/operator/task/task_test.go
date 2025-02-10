package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	"math/big"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"
)

type mockEpochStateQuerier struct {
	mock.Mock
	thresholdWeight *big.Int
	currentEpoch    uint32
}

func NewMockEpochStateQuerier() *mockEpochStateQuerier {
	m := &mockEpochStateQuerier{
		thresholdWeight: big.NewInt(100),
		currentEpoch:    1,
	}
	// 设置默认行为
	m.On("GetThresholdWeight").Return(big.NewInt(100), nil).Maybe()
	m.On("GetCurrentEpochNumber").Return(uint32(1)).Maybe()
	return m
}

func (m *mockEpochStateQuerier) GetThresholdWeight() *big.Int {
	args := m.Called()
	return args.Get(0).(*big.Int)
}

func (m *mockEpochStateQuerier) GetCurrentEpochNumber() uint32 {
	args := m.Called()
	return args.Get(0).(uint32)
}

type mockChainClient struct {
	mock.Mock
	currentBlock uint64
	stateData    map[string]*big.Int
	chainID      uint32
}

func NewMockChainClient() *mockChainClient {
	mc := &mockChainClient{
		currentBlock: 100,
		stateData:    make(map[string]*big.Int),
		chainID:      1,
	}
	// 添加一些模拟的状态数据
	mc.stateData["default"] = big.NewInt(1)

	mc.On("BlockNumber", mock.Anything).Return(mc.currentBlock, nil).Maybe()
	mc.On("GetStateAtBlock", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mc.stateData["default"], nil).Maybe()
	mc.On("Close").Return(nil).Maybe()
	return mc
}

func (m *mockChainClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockChainClient) GetStateAtBlock(ctx context.Context, target common.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	args := m.Called(ctx, target, key, blockNumber)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockChainManager struct {
	mock.Mock
	clients       map[uint32]*mockChainClient
	mainnetClient *mockChainClient
	networkState  string
}

func NewMockChainManager() *mockChainManager {
	m := &mockChainManager{
		clients:      make(map[uint32]*mockChainClient),
		networkState: "healthy",
	}
	// 创建mainnet client
	m.mainnetClient = NewMockChainClient()
	// 为一些常用chain ID预设client
	m.clients[1] = NewMockChainClient()
	m.clients[31337] = NewMockChainClient()

	m.On("GetMainnetClient").Return(m.mainnetClient, nil).Maybe()
	m.On("GetClientByChainId", mock.Anything).Return(m.clients[1], nil).Maybe()
	m.On("Close").Return(nil).Maybe()
	return m
}

func (m *mockChainManager) GetMainnetClient() (*ethereum.ChainClient, error) {
	args := m.Called()
	if client := args.Get(0); client != nil {
		return client.(*ethereum.ChainClient), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainManager) GetClientByChainId(chainId uint32) (*ethereum.ChainClient, error) {
	args := m.Called(chainId)
	if client := args.Get(0); client != nil {
		return client.(*ethereum.ChainClient), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockTopic struct {
	mock.Mock
	peers     []peer.ID
	messages  [][]byte
	topicName string
}

func NewMockTopic() *mockTopic {
	mt := &mockTopic{
		peers:     []peer.ID{peer.ID("peer1"), peer.ID("peer2")},
		messages:  make([][]byte, 0),
		topicName: "test_topic",
	}
	mt.On("Subscribe", mock.Anything).Return(&pubsub.Subscription{}, nil).Maybe()
	mt.On("String").Return(mt.topicName).Maybe()
	mt.On("ListPeers").Return(mt.peers).Maybe()
	mt.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return mt
}

func (m *mockTopic) Subscribe(opts ...pubsub.SubOpt) (*pubsub.Subscription, error) {
	args := m.Called(opts)
	if sub := args.Get(0); sub != nil {
		return sub.(*pubsub.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockTopic) Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error {
	args := m.Called(ctx, data, opts)
	return args.Error(0)
}

func (m *mockTopic) ListPeers() []peer.ID {
	args := m.Called()
	return args.Get(0).([]peer.ID)
}

func (m *mockTopic) String() string {
	args := m.Called()
	return args.String(0)
}

type operatorTableSerde struct {
	operatorStatesQuerier *operators.Queries
}

func (o operatorTableSerde) Load(data []byte) error {
	if err := o.operatorStatesQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load operator states data: %w", err)
	}
	return nil
}

func (o operatorTableSerde) Dump() ([]byte, error) {
	return o.operatorStatesQuerier.Dump(context.Background(), func(m *operators.Operators) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

type blacklistTableSerde struct {
	blacklistQuerier *blacklist.Queries
}

func (o blacklistTableSerde) Load(data []byte) error {
	if err := o.blacklistQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load blacklist states data: %w", err)
	}
	return nil
}

func (o blacklistTableSerde) Dump() ([]byte, error) {
	return o.blacklistQuerier.Dump(context.Background(), func(m *blacklist.Blacklist) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

type consensusResponseTableSerde struct {
	consensusResponseQuerier *consensus_responses.Queries
}

func (o consensusResponseTableSerde) Load(data []byte) error {
	if err := o.consensusResponseQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load consensus response states data: %w", err)
	}
	return nil
}

func (o consensusResponseTableSerde) Dump() ([]byte, error) {
	return o.consensusResponseQuerier.Dump(context.Background(), func(m *consensus_responses.ConsensusResponse) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

type taskTableSerde struct {
	taskQuerier *tasks.Queries
}

func (o taskTableSerde) Load(data []byte) error {
	if err := o.taskQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load task states data: %w", err)
	}
	return nil
}

func (o taskTableSerde) Dump() ([]byte, error) {
	return o.taskQuerier.Dump(context.Background(), func(m *tasks.Tasks) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

// OperatorTestSuite 是主测试套件
type OperatorTestSuite struct {
	*testsuite.WPgxTestSuite
	processor *TaskProcessor

	operatorRepo          OperatorRepo
	taskRepo              TaskRepo
	blacklistRepo         BlacklistRepo
	consensusResponseRepo ConsensusResponseRepo

	signer         OperatorSigner
	mockStateQuery *mockEpochStateQuerier
	mockChainMgr   *mockChainManager
	mockTopic      *mockTopic

	RedisConn redis.UniversalClient
	FreeCache *freecache.Cache
	DCache    *dcache.DCache
}

func TestOperatorSuite(t *testing.T) {
	suite.Run(t, newOperatorTestSuite())
}

// SetupTest 在每个测试前运行
func (s *OperatorTestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.Require().NoError(s.RedisConn.FlushAll(context.Background()).Err())
	s.FreeCache.Clear()

	// 确保数据库连接正确初始化
	pool := s.GetPool()
	s.Require().NotNil(pool, "Database pool should not be nil")
	conn := pool.WConn()
	s.Require().NotNil(conn, "Database connection should not be nil")

	// 确保 operatorRepo 正确初始化
	operatorRepo := operators.New(conn, s.DCache)
	s.Require().NotNil(operatorRepo, "Operator repository should not be nil")
	s.operatorRepo = operatorRepo

	blacklistRepo := blacklist.New(conn, s.DCache)
	s.Require().NotNil(blacklistRepo, "Blacklist repository should not be nil")
	s.blacklistRepo = blacklistRepo

	consensusResponseRepo := consensus_responses.New(conn, s.DCache)
	s.Require().NotNil(consensusResponseRepo, "Blacklist repository should not be nil")
	s.consensusResponseRepo = consensusResponseRepo

	taskRepo := tasks.New(conn, s.DCache)
	s.Require().NotNil(taskRepo, "Blacklist repository should not be nil")
	s.taskRepo = taskRepo

	signer, err := signer.NewLocalSigner(&signer.Config{
		SigningKeyPath: "../../../keys/signing/operator1.key.json",
		Password:       "testpassword",
	})
	s.Require().NoError(err)
	s.signer = signer

	mockStateQuery := NewMockEpochStateQuerier()
	s.mockStateQuery = mockStateQuery
	mockChainMgr := NewMockChainManager()
	s.mockChainMgr = mockChainMgr
	mockTopic := NewMockTopic()
	s.mockTopic = mockTopic

	// Create mock subscription
	mockSubscription, err := mockTopic.Subscribe()
	s.Require().NoError(err)

	processor, err := NewTaskProcessor(&Config{
		Signer:                signer,
		EpochStateQuerier:     mockStateQuery,
		ConsensusResponseRepo: consensusResponseRepo,
		BlacklistRepo:         blacklistRepo,
		TaskRepo:              taskRepo,
		OperatorRepo:          operatorRepo,
		ChainManager:          mockChainMgr,
		ResponseTopic:         mockTopic,
		ResponseSubscription:  mockSubscription,
	})
	s.Require().NoError(err)
	s.processor = processor
}

// NewOperatorTestSuite 创建新的测试套件实例
func newOperatorTestSuite() *OperatorTestSuite {
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
			consensus_responses.Schema,
			blacklist.Schema,
			operators.Schema,
		}),
		RedisConn: redisClient,
		FreeCache: memCache,
		DCache:    dCache,
	}

	return s
}

// TestTaskChecking tests the task checking functionality
func (s *OperatorTestSuite) TestTaskChecking() {
	// 如果需要,可以为这个测试用例设置特定的 mock 行为
	// s.mockPubSub.On("Join", TaskResponseTopic).Return(s.mockTopic, nil).Once()
	// s.mockTopic.On("Subscribe").Return(&pubsub.Subscription{}, nil).Once()

	// Load initial task states
	taskSerde := taskTableSerde{
		taskQuerier: s.taskRepo.(*tasks.Queries),
	}
	s.LoadState("TestOperatorSuite/TestTaskChecking/initial_tasks.json", taskSerde)

	// Test case 1: Check pending tasks
	// Trigger check
	s.processor.checkTimeouts(context.Background())

	// Verify state after timeout check
	s.Golden("after_timeout_check", taskSerde)

	// Test case 2: Check confirming tasks
	mockChainClient := NewMockChainClient()
	s.mockChainMgr.On("GetClientByChainId", uint32(1)).Return(mockChainClient, nil).Once()
	mockChainClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil).Once()

	// Trigger check
	s.processor.checkConfirmations(context.Background())

	// Verify state after confirmation check
	s.Golden("after_confirmation_check", taskSerde)

	// Test case 3: Check cleanup of max retried tasks
	// Load state with max retried tasks
	s.LoadState("TestOperatorSuite/TestTaskChecking/max_retry_tasks.json", taskSerde)

	// Trigger check
	s.processor.checkTimeouts(context.Background())

	// Verify final state
	s.Golden("after_cleanup", taskSerde)

	s.mockChainMgr.AssertExpectations(s.T())
	mockChainClient.AssertExpectations(s.T())
}

// TestTaskCleanup tests the task cleanup functionality
func (s *OperatorTestSuite) TestTaskCleanup() {
	// 如果需要,可以为这个测试用例设置特定的 mock 行为
	// s.mockPubSub.On("Join", TaskResponseTopic).Return(s.mockTopic, nil).Once()
	// s.mockTopic.On("Subscribe").Return(&pubsub.Subscription{}, nil).Once()

	// Load initial task states
	taskSerde := taskTableSerde{
		taskQuerier: s.taskRepo.(*tasks.Queries),
	}
	s.LoadState("TestOperatorSuite/TestTaskCleanup/initial_tasks.json", taskSerde)

	// Add some test data
	response := taskResponse{
		taskID:        "test-task-1",
		signature:     []byte("test-signature"),
		epoch:         1,
		chainID:       1,
		targetAddress: "0x1234567890123456789012345678901234567890",
		key:           "1",
		value:         "100",
		blockNumber:   100,
	}

	s.processor.storeResponseAndWeight(response, "0x1234", big.NewInt(50))
	s.processor.storeResponseAndWeight(response, "0x5678", big.NewInt(60))

	// Verify data exists
	s.processor.taskResponseTrack.mu.RLock()
	s.Equal(2, len(s.processor.taskResponseTrack.responses[response.taskID]))
	s.Equal(2, len(s.processor.taskResponseTrack.weights[response.taskID]))
	s.processor.taskResponseTrack.mu.RUnlock()

	// Trigger cleanup
	s.processor.cleanupAllTasks()

	// Verify data was cleaned up
	s.processor.taskResponseTrack.mu.RLock()
	s.Equal(0, len(s.processor.taskResponseTrack.responses))
	s.Equal(0, len(s.processor.taskResponseTrack.weights))
	s.processor.taskResponseTrack.mu.RUnlock()

	// Verify final state
	s.Golden("after_cleanup", taskSerde)
}
