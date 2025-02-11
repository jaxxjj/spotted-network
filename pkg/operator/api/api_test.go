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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

// mockChainClient mocks the chain client
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
	// Set default behaviors
	mc.On("BlockNumber", mock.Anything).Return(mc.currentBlock, nil).Maybe()
	mc.On("GetStateAtBlock", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(big.NewInt(1), nil).Maybe()
	mc.On("BlockByNumber", mock.Anything, mock.Anything).Return(&types.Block{}, nil).Maybe()
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

func (m *mockChainClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	if v := args.Get(0); v != nil {
		return v.(*types.Block), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// mockChainManager mocks the chain manager
type mockChainManager struct {
	mock.Mock
	clients       map[uint32]*mockChainClient
	mainnetClient *mockChainClient
}

func NewMockChainManager() *mockChainManager {
	mm := &mockChainManager{
		clients:       make(map[uint32]*mockChainClient),
		mainnetClient: NewMockChainClient(),
	}
	// Add default client for chain ID 1
	mm.clients[1] = NewMockChainClient()

	// Set default behaviors
	mm.On("GetMainnetClient").Return(mm.mainnetClient, nil).Maybe()
	mm.On("GetClientByChainId", mock.Anything).Return(mm.clients[1], nil).Maybe()
	mm.On("Close").Return(nil).Maybe()

	return mm
}

func (m *mockChainManager) GetMainnetClient() (ethereum.ChainClient, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(ethereum.ChainClient), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChainManager) GetClientByChainId(chainId uint32) (ethereum.ChainClient, error) {
	args := m.Called(chainId)
	if v := args.Get(0); v != nil {
		return v.(ethereum.ChainClient), args.Error(1)
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
	chainClient   *mockChainClient
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
	s.chainClient = NewMockChainClient()

	// Initialize mock task processor
	s.taskProcessor = new(mockTaskProcessor)
	s.taskProcessor.On("ProcessTask", mock.Anything, mock.Anything).Return(nil).Maybe()
	s.taskProcessor.On("Start", mock.Anything, mock.Anything).Return(nil).Maybe()
	s.taskProcessor.On("Stop").Return(nil).Maybe()

	// Load config from file
	cfg, err := config.LoadConfig("../../../config/operator.yaml")
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
			goldenFile:   "TestAPISuite/TestSendRequest/new_task.golden",
			request: SendRequestParams{
				ChainID:       1,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode: http.StatusOK,
			setupMocks: func() {
				s.chainClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
			},
		},
		{
			name:         "existing task",
			initialState: "TestAPISuite/TestSendRequest/existing_task.json",
			goldenFile:   "TestAPISuite/TestSendRequest/existing_task.golden",
			request: SendRequestParams{
				ChainID:       1,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   100,
			},
			expectedCode: http.StatusOK,
			setupMocks: func() {
				s.chainClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
			},
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
			errorContains: "chain not found",
			setupMocks: func() {
				s.chainManager.On("GetClientByChainId", uint32(999)).Return(nil, fmt.Errorf("chain not found")).Once()
			},
		},
		{
			name:         "invalid target address",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			request: SendRequestParams{
				ChainID:       1,
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
				ChainID:       1,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "invalid-key",
				BlockNumber:   100,
			},
			expectedCode:  http.StatusBadRequest,
			errorContains: "invalid key format",
		},
		{
			name:         "block number with confirmations",
			initialState: "TestAPISuite/TestSendRequest/new_task.json",
			goldenFile:   "TestAPISuite/TestSendRequest/confirming_task.golden",
			request: SendRequestParams{
				ChainID:       1,
				TargetAddress: "0x1234567890123456789012345678901234567890",
				Key:           "1000000000000000000",
				BlockNumber:   90,
			},
			expectedCode: http.StatusOK,
			setupMocks: func() {
				s.chainClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset mock expectations
			s.chainManager.ExpectedCalls = nil
			s.chainClient.ExpectedCalls = nil

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
			s.chainClient.AssertExpectations(s.T())
		})
	}
}
