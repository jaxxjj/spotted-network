package task

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"math/big"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	protob "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"
	"google.golang.org/protobuf/proto"
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

	// 设置默认行为
	mc.On("BlockNumber", mock.Anything).Return(mc.currentBlock, nil).Maybe()
	mc.On("GetStateAtBlock", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mc.stateData["default"], nil).Maybe()
	mc.On("GetLatestState", mock.Anything, mock.Anything, mock.Anything).Return(mc.stateData["default"], nil).Maybe()
	mc.On("GetStateAtTimestamp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mc.stateData["default"], nil).Maybe()
	mc.On("GetCurrentEpoch", mock.Anything).Return(uint32(1), nil).Maybe()
	mc.On("GetEffectiveEpochForBlock", mock.Anything, mock.Anything).Return(uint32(1), nil).Maybe()
	mc.On("GetOperatorWeight", mock.Anything, mock.Anything).Return(big.NewInt(100), nil).Maybe()
	mc.On("GetTotalWeight", mock.Anything).Return(big.NewInt(1000), nil).Maybe()
	mc.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(10), nil).Maybe()
	mc.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(500), nil).Maybe()
	mc.On("IsOperatorRegistered", mock.Anything, mock.Anything).Return(true, nil).Maybe()
	mc.On("GetOperatorSigningKey", mock.Anything, mock.Anything, mock.Anything).Return(ethcommon.HexToAddress("0x1"), nil).Maybe()
	mc.On("GetOperatorP2PKey", mock.Anything, mock.Anything, mock.Anything).Return(ethcommon.HexToAddress("0x2"), nil).Maybe()
	mc.On("WatchOperatorRegistered", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	mc.On("WatchOperatorDeregistered", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	mc.On("Close").Return(nil).Maybe()

	return mc
}

func (m *mockChainClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockChainClient) GetStateAtBlock(ctx context.Context, target ethcommon.Address, key *big.Int, blockNumber uint64) (*big.Int, error) {
	args := m.Called(ctx, target, key, blockNumber)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) GetLatestState(ctx context.Context, target ethcommon.Address, key *big.Int) (*big.Int, error) {
	args := m.Called(ctx, target, key)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) GetStateAtTimestamp(ctx context.Context, target ethcommon.Address, key *big.Int, timestamp uint64) (*big.Int, error) {
	args := m.Called(ctx, target, key, timestamp)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) GetCurrentEpoch(ctx context.Context) (uint32, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint32), args.Error(1)
}

func (m *mockChainClient) GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error) {
	args := m.Called(ctx, blockNumber)
	return args.Get(0).(uint32), args.Error(1)
}

func (m *mockChainClient) GetOperatorWeight(ctx context.Context, operator ethcommon.Address) (*big.Int, error) {
	args := m.Called(ctx, operator)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) GetTotalWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) GetMinimumWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) GetThresholdWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) IsOperatorRegistered(ctx context.Context, operator ethcommon.Address) (bool, error) {
	args := m.Called(ctx, operator)
	return args.Bool(0), args.Error(1)
}

func (m *mockChainClient) GetOperatorSigningKey(ctx context.Context, operator ethcommon.Address, epoch uint32) (ethcommon.Address, error) {
	args := m.Called(ctx, operator, epoch)
	return args.Get(0).(ethcommon.Address), args.Error(1)
}

func (m *mockChainClient) GetOperatorP2PKey(ctx context.Context, operator ethcommon.Address, epoch uint32) (ethcommon.Address, error) {
	args := m.Called(ctx, operator, epoch)
	return args.Get(0).(ethcommon.Address), args.Error(1)
}

func (m *mockChainClient) WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error) {
	args := m.Called(filterOpts, sink)
	if sub := args.Get(0); sub != nil {
		return sub.(event.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorDeregisteredEvent) (event.Subscription, error) {
	args := m.Called(filterOpts, sink)
	if sub := args.Get(0); sub != nil {
		return sub.(event.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockChainClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	if block := args.Get(0); block != nil {
		return block.(*types.Block), args.Error(1)
	}
	return nil, args.Error(1)
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

func (m *mockChainManager) GetMainnetClient() (ethereum.ChainClient, error) {
	args := m.Called()
	if client := args.Get(0); client != nil {
		return client.(ethereum.ChainClient), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainManager) GetClientByChainId(chainId uint32) (ethereum.ChainClient, error) {
	args := m.Called(chainId)
	if client := args.Get(0); client != nil {
		return client.(ethereum.ChainClient), args.Error(1)
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
	mt.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
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
	operatorStatesQuerier OperatorRepo
}

func (o operatorTableSerde) Load(data []byte) error {
	querier, ok := o.operatorStatesQuerier.(*operators.Queries)
	if !ok {
		return fmt.Errorf("operatorStatesQuerier is not of type *operators.Queries")
	}
	if err := querier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load operator states data: %w", err)
	}
	return nil
}

func (o operatorTableSerde) Dump() ([]byte, error) {
	querier, ok := o.operatorStatesQuerier.(*operators.Queries)
	if !ok {
		return nil, fmt.Errorf("operatorStatesQuerier is not of type *operators.Queries")
	}
	return querier.Dump(context.Background(), func(m *operators.Operators) {
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
	processor *taskProcessor

	operatorRepo          OperatorRepo
	taskRepo              TaskRepo
	blacklistRepo         BlacklistRepo
	consensusResponseRepo ConsensusResponseRepo

	signer           OperatorSigner
	mockStateQuery   *mockEpochStateQuerier
	mockChainMgr     *mockChainManager
	mockTopic        *mockTopic
	mockSubscription *pubsub.Subscription

	RedisConn redis.UniversalClient
	FreeCache *freecache.Cache
	DCache    *dcache.DCache
}

func TestOperatorSuite(t *testing.T) {
	suite.Run(t, newOperatorTestSuite())
}

func (s *OperatorTestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.Require().NoError(s.RedisConn.FlushAll(context.Background()).Err())
	s.FreeCache.Clear()

	pool := s.GetPool()
	s.Require().NotNil(pool, "Database pool should not be nil")
	conn := pool.WConn()
	s.Require().NotNil(conn, "Database connection should not be nil")

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
	s.mockSubscription = mockSubscription

	pendingTaskCheckInterval = 1 * time.Second
	confirmationCheckInterval = 1 * time.Second
	processor, err := NewTaskProcessor(&Config{
		Signer:                signer,
		EpochStateQuerier:     mockStateQuery,
		ConsensusResponseRepo: consensusResponseRepo,
		BlacklistRepo:         blacklistRepo,
		TaskRepo:              taskRepo,
		OperatorRepo:          operatorRepo,
		ChainManager:          mockChainMgr,
		ResponseTopic:         mockTopic,
	})
	s.Require().NoError(err)
	s.processor = processor.(*taskProcessor)
}

func newOperatorTestSuite() *OperatorTestSuite {
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

func (s *OperatorTestSuite) TestProcessorInitialization() {
	tests := []struct {
		name        string
		setupConfig func() *Config
		wantErr     bool
		errMsg      string
	}{
		{
			name: "nil_config",
			setupConfig: func() *Config {
				return nil
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] config is nil",
		},
		{
			name: "nil_signer",
			setupConfig: func() *Config {
				cfg := &Config{
					EpochStateQuerier:     s.mockStateQuery,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					TaskRepo:              s.taskRepo,
					OperatorRepo:          s.operatorRepo,
					ChainManager:          s.mockChainMgr,
					ResponseTopic:         s.mockTopic,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] signer is nil",
		},
		{
			name: "nil_epoch_state_querier",
			setupConfig: func() *Config {
				cfg := &Config{
					Signer:                s.signer,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					TaskRepo:              s.taskRepo,
					OperatorRepo:          s.operatorRepo,
					ChainManager:          s.mockChainMgr,
					ResponseTopic:         s.mockTopic,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] epoch state querier is nil",
		},
		{
			name: "nil_task_repo",
			setupConfig: func() *Config {
				cfg := &Config{
					Signer:                s.signer,
					EpochStateQuerier:     s.mockStateQuery,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					OperatorRepo:          s.operatorRepo,
					ChainManager:          s.mockChainMgr,
					ResponseTopic:         s.mockTopic,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] task repo is nil",
		},
		{
			name: "nil_operator_repo",
			setupConfig: func() *Config {
				cfg := &Config{
					Signer:                s.signer,
					EpochStateQuerier:     s.mockStateQuery,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					TaskRepo:              s.taskRepo,
					ChainManager:          s.mockChainMgr,
					ResponseTopic:         s.mockTopic,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] operator repo is nil",
		},
		{
			name: "nil_consensus_response_repo",
			setupConfig: func() *Config {
				cfg := &Config{
					Signer:            s.signer,
					EpochStateQuerier: s.mockStateQuery,
					BlacklistRepo:     s.blacklistRepo,
					TaskRepo:          s.taskRepo,
					OperatorRepo:      s.operatorRepo,
					ChainManager:      s.mockChainMgr,
					ResponseTopic:     s.mockTopic,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] consensus response repo is nil",
		},
		{
			name: "nil_chain_manager",
			setupConfig: func() *Config {
				cfg := &Config{
					Signer:                s.signer,
					EpochStateQuerier:     s.mockStateQuery,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					TaskRepo:              s.taskRepo,
					OperatorRepo:          s.operatorRepo,
					ResponseTopic:         s.mockTopic,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] chain manager is nil",
		},
		{
			name: "nil_response_topic",
			setupConfig: func() *Config {
				cfg := &Config{
					Signer:                s.signer,
					EpochStateQuerier:     s.mockStateQuery,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					TaskRepo:              s.taskRepo,
					OperatorRepo:          s.operatorRepo,
					ChainManager:          s.mockChainMgr,
				}
				return cfg
			},
			wantErr: true,
			errMsg:  "[TaskProcessor] response topic is nil",
		},
		{
			name: "successful_initialization",
			setupConfig: func() *Config {
				return &Config{
					Signer:                s.signer,
					EpochStateQuerier:     s.mockStateQuery,
					ConsensusResponseRepo: s.consensusResponseRepo,
					BlacklistRepo:         s.blacklistRepo,
					TaskRepo:              s.taskRepo,
					OperatorRepo:          s.operatorRepo,
					ChainManager:          s.mockChainMgr,
					ResponseTopic:         s.mockTopic,
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := tt.setupConfig()
			processor, err := NewTaskProcessor(cfg)

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errMsg)
				s.Nil(processor)
			} else {
				s.NoError(err)
				s.NotNil(processor)
			}
		})
	}
}

func (s *OperatorTestSuite) TestProcessorStop() {
	// Create a new processor with all dependencies
	processor, err := NewTaskProcessor(&Config{
		Signer:                s.signer,
		EpochStateQuerier:     s.mockStateQuery,
		ConsensusResponseRepo: s.consensusResponseRepo,
		BlacklistRepo:         s.blacklistRepo,
		TaskRepo:              s.taskRepo,
		OperatorRepo:          s.operatorRepo,
		ChainManager:          s.mockChainMgr,
		ResponseTopic:         s.mockTopic,
	})
	s.NoError(err)
	s.NotNil(processor)

	// Start the processor
	ctx := context.Background()
	err = processor.Start(ctx, s.mockSubscription)
	s.NoError(err)

	// Add some test data to memory maps
	tp := processor.(*taskProcessor)
	tp.taskResponseTrack.mu.Lock()
	tp.taskResponseTrack.responses["test-task"] = make(map[string]taskResponse)
	tp.taskResponseTrack.weights["test-task"] = make(map[string]*big.Int)
	tp.taskResponseTrack.mu.Unlock()

	// Stop the processor
	err = processor.Stop()
	s.NoError(err)

	// Verify cleanup
	tp.taskResponseTrack.mu.RLock()
	s.Empty(tp.taskResponseTrack.responses)
	s.Empty(tp.taskResponseTrack.weights)
	tp.taskResponseTrack.mu.RUnlock()

	// Verify topic is cleaned up
	s.Nil(tp.responseTopic)
}

// TestTaskChecking tests the task checking functionality
func (s *OperatorTestSuite) TestTaskChecking() {
	pendingTaskCheckInterval = 1 * time.Second
	confirmationCheckInterval = 1 * time.Second
	cleanupInterval = 1 * time.Second

	mockChainClient := NewMockChainClient()
	s.mockChainMgr.On("GetClientByChainId", uint32(1)).Return(mockChainClient, nil).Maybe()
	mockChainClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil).Maybe()

	s.processor.Start(context.Background(), s.mockSubscription)

	taskSerde := taskTableSerde{
		taskQuerier: s.taskRepo.(*tasks.Queries),
	}
	s.LoadState("TestOperatorSuite/TestTaskChecking/initial_tasks.json", taskSerde)

	time.Sleep(4 * time.Second)

	s.Golden("after_timeout_check", taskSerde)

	time.Sleep(2 * time.Second)

	s.Golden("after_confirmation_check", taskSerde)
}

// TestTaskCleanup tests the task cleanup functionality
func (s *OperatorTestSuite) TestTaskCleanup() {
	// Load initial task states
	taskSerde := taskTableSerde{
		taskQuerier: s.taskRepo.(*tasks.Queries),
	}
	s.LoadState("TestOperatorSuite/TestTaskCleanup/initial_tasks.json", taskSerde)

	// Add some test data
	response1 := taskResponse{
		taskID:        "test-task-1",
		signature:     []byte("test-signature-1"),
		epoch:         1,
		chainID:       1,
		targetAddress: "0x1234567890123456789012345678901234567890",
		key:           "1000000000000000000",
		value:         "1000000000000000000",
		blockNumber:   100,
	}

	response2 := taskResponse{
		taskID:        "test-task-1",
		signature:     []byte("test-signature-2"),
		epoch:         1,
		chainID:       1,
		targetAddress: "0x1234567890123456789012345678901234567890",
		key:           "2000000000000000000",
		value:         "2000000000000000000",
		blockNumber:   100,
	}

	s.processor.storeResponseAndWeight(response1, "0x1234", big.NewInt(50))
	s.processor.storeResponseAndWeight(response2, "0x5678", big.NewInt(60))

	// Verify data exists
	s.processor.taskResponseTrack.mu.RLock()
	s.Equal(2, len(s.processor.taskResponseTrack.responses[response1.taskID]))
	s.Equal(2, len(s.processor.taskResponseTrack.weights[response1.taskID]))
	s.processor.taskResponseTrack.mu.RUnlock()

	// Trigger cleanup
	s.processor.cleanupAllTasks()

	// Verify data was cleaned up
	s.processor.taskResponseTrack.mu.RLock()
	s.Equal(0, len(s.processor.taskResponseTrack.responses))
	s.Equal(0, len(s.processor.taskResponseTrack.weights))
	s.processor.taskResponseTrack.mu.RUnlock()

}

// TestTaskMemoryCleanup tests the cleanup of task tracking data in memory
func (s *OperatorTestSuite) TestTaskMemoryCleanup() {
	// Setup test data
	taskID1 := "task1"
	taskID2 := "task2"
	operatorID := "operator1"

	// Initialize test data in memory
	s.processor.taskResponseTrack.mu.Lock()
	s.processor.taskResponseTrack.responses = map[string]map[string]taskResponse{
		taskID1: {
			operatorID: taskResponse{
				taskID:  taskID1,
				epoch:   1,
				chainID: 1,
				value:   "1000000000000000000",
			},
		},
		taskID2: {
			operatorID: taskResponse{
				taskID:  taskID2,
				epoch:   1,
				chainID: 1,
				value:   "2000000000000000000",
			},
		},
	}
	s.processor.taskResponseTrack.weights = map[string]map[string]*big.Int{
		taskID1: {
			operatorID: big.NewInt(100),
		},
		taskID2: {
			operatorID: big.NewInt(200),
		},
	}
	s.processor.taskResponseTrack.mu.Unlock()

	// Verify initial state
	s.processor.taskResponseTrack.mu.RLock()
	s.Require().Len(s.processor.taskResponseTrack.responses[taskID1], 1)
	s.Require().Len(s.processor.taskResponseTrack.responses[taskID2], 1)
	s.Require().Len(s.processor.taskResponseTrack.weights[taskID1], 1)
	s.Require().Len(s.processor.taskResponseTrack.weights[taskID2], 1)
	s.processor.taskResponseTrack.mu.RUnlock()

	// Clean up task1
	s.processor.cleanupTask(taskID1)

	// Verify task1 is cleaned up but task2 remains
	s.processor.taskResponseTrack.mu.RLock()
	s.Require().Nil(s.processor.taskResponseTrack.responses[taskID1])
	s.Require().Len(s.processor.taskResponseTrack.responses[taskID2], 1)
	s.Require().Nil(s.processor.taskResponseTrack.weights[taskID1])
	s.Require().Len(s.processor.taskResponseTrack.weights[taskID2], 1)
	s.processor.taskResponseTrack.mu.RUnlock()

	// Clean up task2
	s.processor.cleanupTask(taskID2)

	// Verify all data is cleaned up
	s.processor.taskResponseTrack.mu.RLock()
	s.Require().Nil(s.processor.taskResponseTrack.responses[taskID1])
	s.Require().Nil(s.processor.taskResponseTrack.responses[taskID2])
	s.Require().Nil(s.processor.taskResponseTrack.weights[taskID1])
	s.Require().Nil(s.processor.taskResponseTrack.weights[taskID2])
	s.processor.taskResponseTrack.mu.RUnlock()
}

// Helper function to set up operator responses in memory
func (s *OperatorTestSuite) setupOperatorResponses(taskID string, responses map[string]taskResponse, weights map[string]*big.Int) {
	s.processor.taskResponseTrack.mu.Lock()
	defer s.processor.taskResponseTrack.mu.Unlock()

	// Initialize maps if they don't exist
	if s.processor.taskResponseTrack.responses[taskID] == nil {
		s.processor.taskResponseTrack.responses[taskID] = make(map[string]taskResponse)
	}
	if s.processor.taskResponseTrack.weights[taskID] == nil {
		s.processor.taskResponseTrack.weights[taskID] = make(map[string]*big.Int)
	}

	// Set responses
	for operatorID, response := range responses {
		s.processor.taskResponseTrack.responses[taskID][operatorID] = response
	}

	// Set weights
	for operatorID, weight := range weights {
		s.processor.taskResponseTrack.weights[taskID][operatorID] = weight
	}
}

// TestConsensus tests the consensus functionality
func (s *OperatorTestSuite) TestConsensus() {
	// Set mock expectations for threshold weight
	s.mockStateQuery.thresholdWeight = big.NewInt(100)

	// Test cases
	tests := []struct {
		name         string
		taskID       string
		responses    map[string]taskResponse
		weights      map[string]*big.Int
		goldenFile   string
		initialState string
	}{
		{
			name:   "successful_consensus",
			taskID: "ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6",
			responses: map[string]taskResponse{
				"0x1111111111111111111111111111111111111111": {
					taskID:        "ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6",
					signature:     []byte("test-signature-1"),
					epoch:         1,
					chainID:       1,
					targetAddress: "0x1234567890123456789012345678901234567890",
					key:           "1000000000000000000",
					value:         "1000000000000000000",
					blockNumber:   100,
				},
				"0x2222222222222222222222222222222222222222": {
					taskID:        "ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6",
					signature:     []byte("test-signature-2"),
					epoch:         1,
					chainID:       1,
					targetAddress: "0x1234567890123456789012345678901234567890",
					key:           "1000000000000000000",
					value:         "1000000000000000000",
					blockNumber:   100,
				},
			},
			weights: map[string]*big.Int{
				"0x1111111111111111111111111111111111111111": big.NewInt(60),
				"0x2222222222222222222222222222222222222222": big.NewInt(50),
			},
			goldenFile:   "after_consensus",
			initialState: "TestOperatorSuite/TestConsensus/initial_tasks.json",
		},
		{
			name:   "insufficient_weight",
			taskID: "dcd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6",
			responses: map[string]taskResponse{
				"0x1111111111111111111111111111111111111111": {
					taskID:        "dcd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6",
					signature:     []byte("test-signature-1"),
					epoch:         1,
					chainID:       1,
					targetAddress: "0x1234567890123456789012345678901234567890",
					key:           "1000000000000000000",
					value:         "1000000000000000000",
					blockNumber:   100,
				},
				"0x2222222222222222222222222222222222222222": {
					taskID:        "dcd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6",
					signature:     []byte("test-signature-2"),
					epoch:         1,
					chainID:       1,
					targetAddress: "0x1234567890123456789012345678901234567890",
					key:           "1000000000000000000",
					value:         "1000000000000000000",
					blockNumber:   100,
				},
			},
			weights: map[string]*big.Int{
				"0x1111111111111111111111111111111111111111": big.NewInt(40),
				"0x2222222222222222222222222222222222222222": big.NewInt(30),
			},
			goldenFile:   "after_consensus",
			initialState: "TestOperatorSuite/TestConsensus/insufficient_weight/initial_task.json",
		},
		{
			name:   "single_response_no_consensus",
			taskID: "fcd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a7",
			responses: map[string]taskResponse{
				"0x1111111111111111111111111111111111111111": {
					taskID:        "fcd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a7",
					signature:     []byte("test-signature-1"),
					epoch:         1,
					chainID:       1,
					targetAddress: "0x1234567890123456789012345678901234567890",
					key:           "2000000000000000000",
					value:         "2000000000000000000",
					blockNumber:   100,
				},
			},
			weights: map[string]*big.Int{
				"0x1111111111111111111111111111111111111111": big.NewInt(60),
			},
			goldenFile:   "after_single_response",
			initialState: "TestOperatorSuite/TestConsensus/single_response/initial_task.json",
		},
		{
			name:         "empty_responses",
			taskID:       "hcd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a9",
			responses:    map[string]taskResponse{}, // Empty responses
			weights:      map[string]*big.Int{},     // Empty weights
			goldenFile:   "after_empty_responses",
			initialState: "TestOperatorSuite/TestConsensus/empty_responses/initial_task.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Clean up database before each test case
			s.SetupTest()

			// Load initial operator states
			operatorSerde := operatorTableSerde{
				operatorStatesQuerier: s.operatorRepo,
			}
			s.LoadState("TestOperatorSuite/TestConsensus/initial_operators.json", operatorSerde)

			// Load initial task state
			taskSerde := taskTableSerde{
				taskQuerier: s.taskRepo.(*tasks.Queries),
			}
			s.LoadState(tt.initialState, taskSerde)

			// Set up operator responses in memory
			s.setupOperatorResponses(tt.taskID, tt.responses, tt.weights)

			// Get the task
			task, err := s.taskRepo.GetTaskByID(context.Background(), tt.taskID)
			s.NoError(err)
			s.NotNil(task)

			// Process one of the responses to trigger consensus check
			var firstResponse taskResponse
			for _, resp := range tt.responses {
				firstResponse = resp
				break
			}
			err = s.processor.checkConsensus(context.Background(), firstResponse)
			s.NoError(err)

			// Verify final state using golden files
			s.Golden(tt.goldenFile, taskSerde)

			// Verify consensus response if needed
			if tt.name == "successful_consensus" {
				consensusSerde := consensusResponseTableSerde{
					consensusResponseQuerier: s.consensusResponseRepo.(*consensus_responses.Queries),
				}
				s.Golden("consensus_response", consensusSerde)
			}
		})
	}
}

// TestResponseHandling tests the response handling functionality
func (s *OperatorTestSuite) TestResponseHandling() {
	// Set up test data
	taskID := "ecd4bb90ee55a19b8bf10e5a44b07d1dcceafb9f82f180be7aaa881e5953f5a6"
	operatorAddr := "0x9F0D8BAC11C5693a290527f09434b86651c66Bf2"
	signingKey := "0x2222222222222222222222222222222222222222"

	// Create task response
	response := taskResponse{
		taskID:        taskID,
		epoch:         1,
		chainID:       1,
		targetAddress: "0x1234567890123456789012345678901234567890",
		key:           "1000000000000000000",
		value:         "1000000000000000000",
		blockNumber:   100,
	}

	// Create sign params
	keyBigInt, ok := new(big.Int).SetString(response.key, 10)
	s.True(ok, "Failed to convert key to big.Int")
	valueBigInt, ok := new(big.Int).SetString(response.value, 10)
	s.True(ok, "Failed to convert value to big.Int")

	params := signer.TaskSignParams{
		User:        ethcommon.HexToAddress(response.targetAddress),
		ChainID:     response.chainID,
		BlockNumber: response.blockNumber,
		Key:         keyBigInt,
		Value:       valueBigInt,
	}

	// Sign response
	signature, err := s.signer.SignTaskResponse(params)
	s.NoError(err)
	response.signature = signature

	// Test storing response
	s.processor.storeResponseAndWeight(response, signingKey, big.NewInt(100))

	// Verify response is stored
	s.processor.taskResponseTrack.mu.RLock()
	storedResponse, exists := s.processor.taskResponseTrack.responses[taskID][signingKey]
	s.processor.taskResponseTrack.mu.RUnlock()
	s.True(exists)
	s.Equal(response, storedResponse)

	// Test duplicate response handling
	s.processor.storeResponseAndWeight(response, signingKey, big.NewInt(100))
	s.True(s.processor.alreadyProcessed(taskID, signingKey))

	// Test response verification with operator address
	err = s.signer.VerifyTaskResponse(params, signature, operatorAddr)
	s.NoError(err)

	// Clean up
	s.processor.cleanupTask(taskID)
}

// TestResponseValidation tests the response validation and blacklist functionality
func (s *OperatorTestSuite) TestResponseValidation() {
	// Set up test data
	taskID := "test-validation-task"
	signingKey := "0x2222222222222222222222222222222222222222"
	peerID := peer.ID("12D3KooWB21ALruLNKbH5vLjDjiG1mM9XVnCJMGdsdo2LKvoqneD")

	// Load initial states from JSON files
	s.LoadState("TestOperatorSuite/TestResponseValidation/initial_operators.json", &operatorTableSerde{s.operatorRepo})
	s.LoadState("TestOperatorSuite/TestResponseValidation/initial_tasks.json", &taskTableSerde{s.taskRepo.(*tasks.Queries)})
	s.LoadState("TestOperatorSuite/TestResponseValidation/initial_blacklist.json", &blacklistTableSerde{s.blacklistRepo.(*blacklist.Queries)})

	// Create mock response
	response := taskResponse{
		taskID:        taskID,
		signature:     []byte("invalid-signature"),
		epoch:         1,
		chainID:       1,
		targetAddress: "0x1234567890123456789012345678901234567890",
		key:           "1000000000000000000",
		value:         "1000000000000000000",
		blockNumber:   100,
	}

	// Test response from inactive operator
	err := s.processor.incrementPeerViolation(context.Background(), peerID, fmt.Errorf("operator not active"))
	s.Error(err)

	// Verify blacklist record was created
	isBlocked, err := s.blacklistRepo.(*blacklist.Queries).IsBlocked(context.Background(), peerID.String())
	s.NoError(err)
	s.NotNil(isBlocked)
	s.False(*isBlocked) // Should be false since violation count is only 1

	// Add two more violations to exceed threshold
	err = s.processor.incrementPeerViolation(context.Background(), peerID, fmt.Errorf("operator not active"))
	s.Error(err)
	err = s.processor.incrementPeerViolation(context.Background(), peerID, fmt.Errorf("operator not active"))
	s.Error(err)

	// Verify peer is now blocked
	isBlocked, err = s.blacklistRepo.(*blacklist.Queries).IsBlocked(context.Background(), peerID.String())
	s.NoError(err)
	s.NotNil(isBlocked)
	s.True(*isBlocked) // Should be true since violation count is now 3

	// Test invalid signature
	err = s.processor.verifyResponse(response, signingKey)
	s.Error(err)

	// Test wrong epoch
	response.epoch = 2 // Different from current epoch (1)
	pbMsg := &protob.TaskResponseMessage{
		TaskId:        response.taskID,
		Signature:     response.signature,
		Epoch:         response.epoch,
		ChainId:       response.chainID,
		TargetAddress: response.targetAddress,
		Key:           response.key,
		Value:         response.value,
		BlockNumber:   response.blockNumber,
	}
	data, err := proto.Marshal(pbMsg)
	s.NoError(err)

	err = s.processor.processResponse(context.Background(), &pubsub.Message{
		Message: &pb.Message{
			Data: data,
		},
		ID:            "test-id",
		ReceivedFrom:  peerID,
		ValidatorData: data,
		Local:         false,
	})
	s.NoError(err) // Should be no error since different epoch just gets skipped

	// Clean up
	s.processor.cleanupTask(taskID)
}

// TestResponseConcurrency tests concurrent response handling
func (s *OperatorTestSuite) TestResponseConcurrency() {
	// Set up test data
	taskID := "test-concurrent-task"
	numOperators := 10

	type testOperator struct {
		addr       string
		signingKey string
		p2pKey     string
	}
	operators := make([]testOperator, numOperators)

	// Load initial operator states
	operatorSerde := operatorTableSerde{
		operatorStatesQuerier: s.operatorRepo,
	}
	s.LoadState("TestOperatorSuite/TestResponseConcurrency/initial_operators.json", operatorSerde)

	// Load initial task state
	taskSerde := taskTableSerde{
		taskQuerier: s.taskRepo.(*tasks.Queries),
	}
	s.LoadState("TestOperatorSuite/TestResponseConcurrency/initial_tasks.json", taskSerde)

	// Initialize operators array with values from loaded state
	for i := 0; i < numOperators; i++ {
		operators[i].addr = fmt.Sprintf("0x%040x", i+1)
		operators[i].signingKey = fmt.Sprintf("0x%040x", i+100)
		operators[i].p2pKey = fmt.Sprintf("0x%040x", i+200)
	}

	// Create a wait group for concurrent operations
	var wg sync.WaitGroup
	wg.Add(numOperators)

	// Concurrently store responses from multiple operators
	for i := 0; i < numOperators; i++ {
		go func(idx int) {
			defer wg.Done()

			response := taskResponse{
				taskID:        taskID,
				signature:     []byte(fmt.Sprintf("signature-%d", idx)),
				epoch:         1,
				chainID:       1,
				targetAddress: "0x1234567890123456789012345678901234567890",
				key:           "1000000000000000000",
				value:         "1000000000000000000",
				blockNumber:   100,
			}

			s.processor.storeResponseAndWeight(response, operators[idx].signingKey, big.NewInt(100))
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify all responses were stored correctly
	s.processor.taskResponseTrack.mu.RLock()
	responses := s.processor.taskResponseTrack.responses[taskID]
	weights := s.processor.taskResponseTrack.weights[taskID]
	s.processor.taskResponseTrack.mu.RUnlock()

	// Verify response count
	s.Equal(numOperators, len(responses))
	s.Equal(numOperators, len(weights))

	// Verify each operator's response and weight
	for i := 0; i < numOperators; i++ {
		signingKey := operators[i].signingKey
		response, exists := responses[signingKey]
		s.True(exists)
		s.Equal(taskID, response.taskID)
		s.Equal([]byte(fmt.Sprintf("signature-%d", i)), response.signature)

		weight, exists := weights[signingKey]
		s.True(exists)
		s.Equal(big.NewInt(100), weight)
	}

	// Clean up
	s.processor.cleanupTask(taskID)
}

// TestProcessTask tests the ProcessTask functionality
func (s *OperatorTestSuite) TestProcessTask() {
	tests := []struct {
		name        string
		task        *tasks.Tasks
		setupMocks  func()
		wantErr     bool
		errContains string
	}{
		{
			name:        "nil_task",
			task:        nil,
			setupMocks:  func() {},
			wantErr:     true,
			errContains: "task is nil",
		},
		{
			name: "zero_block_number",
			task: &tasks.Tasks{
				TaskID:      "test-task",
				BlockNumber: 0,
			},
			setupMocks:  func() {},
			wantErr:     true,
			errContains: "task block number is nil",
		},
		{
			name: "successful_processing",
			task: &tasks.Tasks{
				TaskID:        "test-task",
				BlockNumber:   100,
				ChainID:       1,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           pgtype.Numeric{Int: big.NewInt(1), Exp: 0, Valid: true},
				Value:         pgtype.Numeric{Int: big.NewInt(1), Exp: 0, Valid: true},
				Epoch:         1,
			},
			setupMocks: func() {
				// Load initial operator states
				operatorSerde := operatorTableSerde{
					operatorStatesQuerier: s.operatorRepo,
				}
				s.LoadState("TestOperatorSuite/TestProcessTask/initial_operators.json", operatorSerde)

				mockChainClient := NewMockChainClient()
				s.mockChainMgr.On("GetClientByChainId", uint32(1)).Return(mockChainClient, nil)
				mockChainClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil)
				mockChainClient.On("GetStateAtBlock", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1), nil)
			},
			wantErr: false,
		},
		{
			name: "state_mismatch",
			task: &tasks.Tasks{
				TaskID:        "test-task",
				BlockNumber:   100,
				ChainID:       1,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           pgtype.Numeric{Int: big.NewInt(1000000000000000000), Exp: 0, Valid: true},
				Value:         pgtype.Numeric{Int: big.NewInt(1000000000000000000), Exp: 0, Valid: true},
				Epoch:         1,
			},
			setupMocks: func() {
				// Load initial operator states
				operatorSerde := operatorTableSerde{
					operatorStatesQuerier: s.operatorRepo,
				}
				s.LoadState("TestOperatorSuite/TestProcessTask/initial_operators.json", operatorSerde)

				mockChainClient := NewMockChainClient()
				s.mockChainMgr.On("GetClientByChainId", uint32(1)).Return(mockChainClient, nil)
				s.mockChainMgr.On("GetClientByChainId", uint32(1)).Return(mockChainClient, nil)
				mockChainClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil)
				mockChainClient.On("GetStateAtBlock", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(2000000000000000000), nil)
			},
			wantErr:     true,
			errContains: "state value mismatch",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset state before each test
			s.SetupTest()

			// Setup mocks
			tt.setupMocks()

			// Execute test
			err := s.processor.ProcessTask(context.Background(), tt.task)

			// Verify results
			if tt.wantErr {
				s.Error(err)
				if tt.errContains != "" {
					s.Contains(err.Error(), tt.errContains)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

// TestGetStateWithRetries tests normal getStateWithRetries functionality
func (s *OperatorTestSuite) TestGetStateWithRetries() {
	mockChainClient := NewMockChainClient()

	mockChainClient.On("BlockNumber", mock.Anything).Return(uint64(120), nil)
	mockChainClient.On("GetStateAtBlock",
		mock.MatchedBy(func(addr ethcommon.Address) bool {
			return addr.Hex() == "0x1234567890123456789012345678901234567890"
		}),
		mock.Anything,
		uint64(100),
	).Return(big.NewInt(1), nil)

	state, err := getStateWithRetries(
		context.Background(),
		mockChainClient,
		ethcommon.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1),
		uint64(100),
	)

	s.NoError(err)
	s.NotNil(state)
	s.Equal(big.NewInt(1), state)
}

// TestGetStateWithRetriesError tests error scenarios for getStateWithRetries
func (s *OperatorTestSuite) TestGetStateWithRetriesError() {
	tests := []struct {
		name        string
		blockNumber uint64
		setupMocks  func(*mockChainClient)
		wantErr     bool
		errContains string
	}{
		{
			name:        "future_block",
			blockNumber: 200,
			setupMocks: func(m *mockChainClient) {
				m.On("BlockNumber", mock.Anything).Return(uint64(100), nil)
			},
			wantErr:     true,
			errContains: "block number 200 is in the future",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockChainClient := NewMockChainClient()
			tt.setupMocks(mockChainClient)

			state, err := getStateWithRetries(
				context.Background(),
				mockChainClient,
				ethcommon.HexToAddress("0x1234567890123456789012345678901234567890"),
				big.NewInt(1),
				tt.blockNumber,
			)

			if tt.wantErr {
				s.Error(err)
				s.Contains(err.Error(), tt.errContains)
				s.Nil(state)
			} else {
				s.NoError(err)
				s.NotNil(state)
			}
		})
	}
}

// TestBroadcastResponse tests normal message sending scenarios
func (s *OperatorTestSuite) TestBroadcastResponse() {
	tests := []struct {
		name     string
		response taskResponse
	}{
		{
			name: "normal_message",
			response: taskResponse{
				taskID:        "test-task-1",
				signature:     []byte("test-signature"),
				epoch:         1,
				chainID:       1,
				targetAddress: "0x1234567890123456789012345678901234567890",
				key:           "1000000000000000000",
				value:         "1000000000000000000",
				blockNumber:   100,
			},
		},
		{
			name: "message_with_large_values",
			response: taskResponse{
				taskID:        "test-task-2",
				signature:     []byte("test-signature-2"),
				epoch:         1,
				chainID:       1,
				targetAddress: "0x1234567890123456789012345678901234567890",
				key:           "100000000000000000000000000000",
				value:         "100000000000000000000000000000",
				blockNumber:   200,
			},
		},
		{
			name: "message_with_zero_values",
			response: taskResponse{
				taskID:        "test-task-3",
				signature:     []byte("test-signature-3"),
				epoch:         1,
				chainID:       1,
				targetAddress: "0x1234567890123456789012345678901234567890",
				key:           "0",
				value:         "0",
				blockNumber:   300,
			},
		},
	}

	// Setup mock topic
	s.mockTopic.On("Publish", mock.Anything, mock.Anything).Return(nil)

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Execute test
			err := s.processor.broadcastResponse(tt.response)
			s.NoError(err)
		})
	}
}

// TestCleanupMethods tests cleanup related methods
func (s *OperatorTestSuite) TestCleanupMethods() {
	s.Run("cleanup_single_task", func() {
		// Setup test data
		s.processor.taskResponseTrack.mu.Lock()
		s.processor.taskResponseTrack.responses = map[string]map[string]taskResponse{
			"task1": {
				"operator1": taskResponse{taskID: "task1"},
			},
			"task2": {
				"operator1": taskResponse{taskID: "task2"},
			},
		}
		s.processor.taskResponseTrack.weights = map[string]map[string]*big.Int{
			"task1": {
				"operator1": big.NewInt(100),
			},
			"task2": {
				"operator1": big.NewInt(200),
			},
		}
		s.processor.taskResponseTrack.mu.Unlock()

		// Test cleanup single task
		s.processor.cleanupTask("task1")

		// Verify task1 is removed but task2 remains
		s.processor.taskResponseTrack.mu.RLock()
		_, hasTask1Responses := s.processor.taskResponseTrack.responses["task1"]
		_, hasTask1Weights := s.processor.taskResponseTrack.weights["task1"]
		hasTask2 := s.processor.taskResponseTrack.responses["task2"] != nil
		s.processor.taskResponseTrack.mu.RUnlock()

		s.False(hasTask1Responses)
		s.False(hasTask1Weights)
		s.True(hasTask2)
	})

	s.Run("cleanup_all_tasks", func() {
		// Setup test data
		s.processor.taskResponseTrack.mu.Lock()
		s.processor.taskResponseTrack.responses = map[string]map[string]taskResponse{
			"task1": {
				"operator1": taskResponse{taskID: "task1"},
			},
			"task2": {
				"operator1": taskResponse{taskID: "task2"},
			},
		}
		s.processor.taskResponseTrack.weights = map[string]map[string]*big.Int{
			"task1": {
				"operator1": big.NewInt(100),
			},
			"task2": {
				"operator1": big.NewInt(200),
			},
		}
		s.processor.taskResponseTrack.mu.Unlock()

		// Test cleanup all tasks
		s.processor.cleanupAllTasks()

		// Verify all tasks are removed
		s.processor.taskResponseTrack.mu.RLock()
		responsesLen := len(s.processor.taskResponseTrack.responses)
		weightsLen := len(s.processor.taskResponseTrack.weights)
		s.processor.taskResponseTrack.mu.RUnlock()

		s.Equal(0, responsesLen)
		s.Equal(0, weightsLen)
	})
}

// TestAlreadyProcessed tests the alreadyProcessed method
func (s *OperatorTestSuite) TestAlreadyProcessed() {
	tests := []struct {
		name       string
		setupMocks func()
		taskID     string
		signingKey string
		want       bool
	}{
		{
			name: "task_already_processed",
			setupMocks: func() {
				s.processor.taskResponseTrack.mu.Lock()
				s.processor.taskResponseTrack.responses = map[string]map[string]taskResponse{
					"task1": {
						"operator1": taskResponse{taskID: "task1"},
					},
				}
				s.processor.taskResponseTrack.mu.Unlock()
			},
			taskID:     "task1",
			signingKey: "operator1",
			want:       true,
		},
		{
			name: "task_not_processed",
			setupMocks: func() {
				s.processor.taskResponseTrack.mu.Lock()
				s.processor.taskResponseTrack.responses = map[string]map[string]taskResponse{
					"task1": {
						"operator1": taskResponse{taskID: "task1"},
					},
				}
				s.processor.taskResponseTrack.mu.Unlock()
			},
			taskID:     "task2",
			signingKey: "operator1",
			want:       false,
		},
		{
			name: "different_operator",
			setupMocks: func() {
				s.processor.taskResponseTrack.mu.Lock()
				s.processor.taskResponseTrack.responses = map[string]map[string]taskResponse{
					"task1": {
						"operator1": taskResponse{taskID: "task1"},
					},
				}
				s.processor.taskResponseTrack.mu.Unlock()
			},
			taskID:     "task1",
			signingKey: "operator2",
			want:       false,
		},
		{
			name: "nil_responses_map",
			setupMocks: func() {
				s.processor.taskResponseTrack.mu.Lock()
				s.processor.taskResponseTrack.responses = nil
				s.processor.taskResponseTrack.mu.Unlock()
			},
			taskID:     "task1",
			signingKey: "operator1",
			want:       false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setupMocks()

			got := s.processor.alreadyProcessed(tt.taskID, tt.signingKey)
			s.Equal(tt.want, got)
		})
	}
}
