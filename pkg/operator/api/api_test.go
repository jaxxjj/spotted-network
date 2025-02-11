package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/dcache"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/galxe/spotted-network/pkg/testsuite"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

// APITestSuite is the main test suite for API
type APITestSuite struct {
	*testsuite.WPgxTestSuite
	server *httptest.Server
	api    *Server

	operatorRepo          OperatorRepo
	taskRepo              TaskRepo
	blacklistRepo         BlacklistRepo
	consensusResponseRepo ConsensusResponseRepo

	signer        signer.Signer
	taskProcessor *mockTaskProcessor
	RedisConn     redis.UniversalClient
	FreeCache     *freecache.Cache
	DCache        *dcache.DCache
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

	// Initialize mock task processor
	s.taskProcessor = new(mockTaskProcessor)
	s.taskProcessor.On("ProcessTask", mock.Anything, mock.Anything).Return(nil).Maybe()
	s.taskProcessor.On("Start", mock.Anything, mock.Anything).Return(nil).Maybe()
	s.taskProcessor.On("Stop").Return(nil).Maybe()

	// Initialize API server
	api, err := NewServer(&Config{
		TaskProcessor:         s.taskProcessor,
		TaskRepo:              taskRepo,
		ConsensusResponseRepo: consensusResponseRepo,
		Port:                  8080,
	})
	s.Require().NoError(err)
	s.api = api

	// Create test HTTP server
	s.server = httptest.NewServer(s.api.router)
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
	// TODO: Implement request helper
	return nil
}
