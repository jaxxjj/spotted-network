package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/operator/task"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"

	"github.com/galxe/spotted-network/pkg/config"
)

// mockTaskProcessor mocks the task processor
type mockTaskProcessor struct {
	mock.Mock
}

func (m *mockTaskProcessor) Start(ctx context.Context, responseSubscription interface{}) error {
	args := m.Called(ctx, responseSubscription)
	return args.Error(0)
}

func (m *mockTaskProcessor) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockTaskProcessor) ProcessTask(ctx context.Context, task *tasks.Tasks) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

type mockChainClient struct {
	mock.Mock
	currentBlock uint64
	stateData    map[string]*big.Int
	chainID      uint32
}

func NewMockChainClient(chainID uint32) *mockChainClient {
	mc := &mockChainClient{
		currentBlock: 100,
		stateData:    make(map[string]*big.Int),
		chainID:      chainID,
	}

	// set different default behaviors for different chains
	switch chainID {
	case 31337: // Mainnet
		mc.currentBlock = 1000
		mc.stateData["default"] = big.NewInt(100)
	case 11155111: // Sepolia
		mc.currentBlock = 2000
		mc.stateData["default"] = big.NewInt(200)
	case 5: // Goerli
		mc.currentBlock = 3000
		mc.stateData["default"] = big.NewInt(300)
	case 80001: // Mumbai
		mc.currentBlock = 4000
		mc.stateData["default"] = big.NewInt(400)
	default:
		mc.currentBlock = 100
		mc.stateData["default"] = big.NewInt(1)
	}

	// set common mock behavior
	mc.On("BlockNumber", mock.Anything).Return(mc.currentBlock, nil).Maybe()
	mc.On("GetStateAtBlock", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mc.stateData["default"], nil).Maybe()
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
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

	// 为每个支持的链创建独立的client
	supportedChains := []uint32{31337, 11155111, 5, 80001}
	for _, chainID := range supportedChains {
		m.clients[chainID] = NewMockChainClient(chainID)
	}

	// 设置mainnet client (31337)
	m.mainnetClient = m.clients[31337]

	// 更新mock行为
	m.On("GetMainnetClient").Return(m.mainnetClient, nil).Maybe()

	// 为每个支持的链设置mock行为
	for chainID, client := range m.clients {
		m.On("GetClientByChainId", chainID).Return(client, nil).Maybe()
	}

	// 添加未知chainID的处理
	m.On("GetClientByChainId", uint32(999)).Return(nil, fmt.Errorf("chain not found")).Maybe()

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

// APITestSuite is the main test suite for API
type APITestSuite struct {
	*testsuite.WPgxTestSuite
	server  *httptest.Server
	api     *Server
	handler *Handler
	router  chi.Router

	operatorRepo          task.OperatorRepo
	taskRepo              task.TaskRepo
	blacklistRepo         task.BlacklistRepo
	consensusResponseRepo task.ConsensusResponseRepo

	signer        signer.Signer
	taskProcessor *mockTaskProcessor
	RedisConn     redis.UniversalClient
	FreeCache     *freecache.Cache
	DCache        *dcache.DCache
	chainManager  *mockChainManager
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

func TestAPISuite(t *testing.T) {
	suite.Run(t, newAPITestSuite())
}

// SetupTest runs before each test
func (s *APITestSuite) SetupTest() {
	s.WPgxTestSuite.SetupTest()
	s.Require().NoError(s.RedisConn.FlushAll(context.Background()).Err())
	s.FreeCache.Clear()

	// Initialize database connection
	pool := s.GetPool()
	s.Require().NotNil(pool, "Database pool should not be nil")
	conn := pool.WConn()
	s.Require().NotNil(conn, "Database connection should not be nil")

	// Initialize repositories
	operatorRepo := operators.New(conn, s.DCache)
	s.Require().NotNil(operatorRepo, "Operator repository should not be nil")
	s.operatorRepo = operatorRepo

	blacklistRepo := blacklist.New(conn, s.DCache)
	s.Require().NotNil(blacklistRepo, "Blacklist repository should not be nil")
	s.blacklistRepo = blacklistRepo

	consensusResponseRepo := consensus_responses.New(conn, s.DCache)
	s.Require().NotNil(consensusResponseRepo, "Consensus response repository should not be nil")
	s.consensusResponseRepo = consensusResponseRepo

	taskRepo := tasks.New(conn, s.DCache)
	s.Require().NotNil(taskRepo, "Task repository should not be nil")
	s.taskRepo = taskRepo

	// Initialize signer
	signer, err := signer.NewLocalSigner(&signer.Config{
		SigningKeyPath: "../../../keys/signing/operator1.key.json",
		Password:       "testpassword",
	})
	s.Require().NoError(err)
	s.signer = signer

	// Initialize chain manager and client
	s.chainManager = NewMockChainManager()

	// Initialize mock task processor
	s.taskProcessor = new(mockTaskProcessor)
	s.taskProcessor.On("ProcessTask", mock.Anything, mock.Anything).Return(nil).Maybe()
	s.taskProcessor.On("Start", mock.Anything, mock.Anything).Return(nil).Maybe()
	s.taskProcessor.On("Stop").Return(nil).Maybe()

	// Load config from file
	cfg, err := config.LoadConfig("../../../config/operator.test.yaml")
	s.Require().NoError(err)

	// Initialize API handler
	handler, err := NewHandler(Config{
		ChainManager:          s.chainManager,
		TaskProcessor:         s.taskProcessor,
		TaskRepo:              taskRepo,
		ConsensusResponseRepo: consensusResponseRepo,
		Config:                cfg,
	})
	s.Require().NoError(err)
	s.handler = handler

	// Initialize router and register routes
	router := chi.NewRouter()
	handler.RegisterRoutes(router)
	s.router = router

	// Initialize API server with handler
	s.api = NewServer(handler, 8080)

	// Create test HTTP server
	s.server = httptest.NewServer(router)

}

// TearDownTest runs after each test
func (s *APITestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
	s.WPgxTestSuite.TearDownTest()
}

// newAPITestSuite creates a new test suite instance
func newAPITestSuite() *APITestSuite {
	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:6379",
		ReadTimeout: 3 * time.Second,
		PoolSize:    50,
		Password:    "",
	})

	if redisClient.Ping(context.Background()).Err() != nil {
		panic("redis connection failed to ping")
	}

	// Create freecache instance
	memCache := freecache.NewCache(100 * 1024 * 1024)

	// Create dcache
	dCache, err := dcache.NewDCache(
		"test", redisClient, memCache, 100*time.Millisecond, true, true)
	if err != nil {
		panic(err)
	}

	s := &APITestSuite{
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

// Helper methods for tests
func (s *APITestSuite) makeRequest(method, path string, body interface{}) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		s.Require().NoError(err)
	}

	req, err := http.NewRequest(method, s.server.URL+"/api/v1"+path, bytes.NewBuffer(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	return resp
}

func (s *APITestSuite) TestSendRequest() {
	tests := []struct {
		name          string
		initialState  string
		goldenFile    string
		request       SendRequestParams
		expectedCode  int
		setupMocks    func()
		errorContains string
	}{
		{
			name:         "successful new task",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			goldenFile:   "new_task",
			request: SendRequestParams{
				ChainID:       31337,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "existing task",
			initialState: "TestAPISuite/TestSendRequest/existing_task.json",
			goldenFile:   "existing_task",
			request: SendRequestParams{
				ChainID:       31337,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid chain id",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			request: SendRequestParams{
				ChainID:       999,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "unsupported chain ID 999",
		},
		{
			name:         "invalid target address",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			request: SendRequestParams{
				ChainID:       31337,
				TargetAddress: "invalid-address",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid target address",
		},
		{
			name:         "invalid key format",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			request: SendRequestParams{
				ChainID:       31337,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "invalid-key",
				BlockNumber:   100,
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid key: must be valid uint256",
		},
		{
			name:         "block number with confirmations",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			goldenFile:   "confirming_task",
			request: SendRequestParams{
				ChainID:       31337,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   90,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "successful new task on sepolia",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			goldenFile:   "new_task_sepolia",
			request: SendRequestParams{
				ChainID:       11155111,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "block confirmations on mumbai",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			goldenFile:   "confirming_task_mumbai",
			request: SendRequestParams{
				ChainID:       80001,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   3998,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "timestamp based request on mumbai",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			goldenFile:   "timestamp_task_mumbai",
			request: SendRequestParams{
				ChainID:       80001,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				Timestamp:     uint64(1704067200), // 2024-01-01 00:00:00 UTC
			},
			expectedCode: http.StatusOK,
			setupMocks: func() {
				// Get the Mumbai client
				mumbaiClient := s.chainManager.clients[80001]

				mumbaiClient.On("BlockNumber", mock.Anything).Return(uint64(4000), nil).Times(3)

				// Mock BlockByNumber to handle any block number query
				mumbaiClient.On("BlockByNumber", mock.Anything, mock.AnythingOfType("*big.Int")).Return(
					types.NewBlock(
						&types.Header{
							Number: big.NewInt(4000),
							Time:   uint64(1704067200), // 2024-01-01 00:00:00 UTC
						},
						nil, nil, nil,
					), nil,
				)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()

			// Setup mock behaviors
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			// Load initial state if provided
			if tt.initialState != "" {
				s.LoadState(tt.initialState, taskTableSerde{taskQuerier: tasks.New(s.GetPool().WConn(), s.DCache)})
			}

			// Make request
			resp := s.makeRequest(http.MethodPost, "/tasks", tt.request)
			s.Require().Equal(tt.expectedCode, resp.StatusCode)

			// Check response
			var result SendRequestResponse
			err := json.NewDecoder(resp.Body).Decode(&result)
			s.Require().NoError(err)
			resp.Body.Close()

			if tt.errorContains != "" {
				s.Contains(result.Error, tt.errorContains)
				return
			}

			// Verify final state if golden file provided
			if tt.goldenFile != "" {
				s.Golden(tt.goldenFile, taskTableSerde{taskQuerier: tasks.New(s.GetPool().WConn(), s.DCache)})
			}

			// Verify mock expectations
			s.chainManager.AssertExpectations(s.T())
		})
	}
}

func (s *APITestSuite) TestGenerateTaskID() {
	// Test case 1: Verify same inputs generate same task ID
	params1 := &SendRequestParams{
		ChainID:       80001,
		TargetAddress: "0x1234567890123456789012345678901234567890",
		Key:           "1000000000000000000",
		BlockNumber:   3800,
	}
	value1 := "400"

	taskID1 := s.handler.generateTaskID(params1, value1)
	taskID2 := s.handler.generateTaskID(params1, value1)
	s.Equal(taskID1, taskID2, "Same inputs should generate same task ID")

	// Test case 2: Verify different inputs generate different task IDs
	params2 := &SendRequestParams{
		ChainID:       80001,
		TargetAddress: "0x1234567890123456789012345678901234567890",
		Key:           "2000000000000000000", // Different key
		BlockNumber:   3800,
	}
	value2 := "500" // Different value

	taskID3 := s.handler.generateTaskID(params2, value2)
	s.NotEqual(taskID1, taskID3, "Different inputs should generate different task IDs")

	// Test case 3: Verify task ID format
	s.Regexp("^[0-9a-f]{64}$", taskID1, "Task ID should be 64 hex characters")
}
