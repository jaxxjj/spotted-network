package event

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx/testsuite"
)

type TimePointSet struct {
	Now  time.Time
	P12H string
	P24H string
}

type operatorStatesTableSerde struct {
	operatorStatesQuerier *operators.Queries
}

func (o operatorStatesTableSerde) Load(data []byte) error {
	if err := o.operatorStatesQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load operator states data: %w", err)
	}
	return nil
}

func (o operatorStatesTableSerde) Dump() ([]byte, error) {
	return o.operatorStatesQuerier.Dump(context.Background(), func(m *operators.Operators) {
		m.CreatedAt = time.Unix(0, 0).UTC()
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
}

// MockChainClient implements ChainClient interface for testing
type MockChainClient struct {
	weight              *big.Int
	regSub              event.Subscription
	deregSub            event.Subscription
	regChan             chan *ethereum.OperatorRegisteredEvent
	deregChan           chan *ethereum.OperatorDeregisteredEvent
	getWeightError      error
	watchRegError       error
	watchDeregError     error
	watchRegCallCount   int
	watchDeregCallCount int
	closed              bool
	mu                  sync.Mutex // 添加互斥锁保护状态
}

func NewMockChainClient() *MockChainClient {
	return &MockChainClient{
		weight:    big.NewInt(100),
		regChan:   make(chan *ethereum.OperatorRegisteredEvent),
		deregChan: make(chan *ethereum.OperatorDeregisteredEvent),
		regSub: event.NewSubscription(func(quit <-chan struct{}) error {
			<-quit
			return nil
		}),
		deregSub: event.NewSubscription(func(quit <-chan struct{}) error {
			<-quit
			return nil
		}),
		closed: false,
	}
}

func (m *MockChainClient) GetOperatorWeight(ctx context.Context, address common.Address) (*big.Int, error) {
	return m.weight, m.getWeightError
}

func (m *MockChainClient) WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchRegCallCount++
	if m.watchRegError != nil {
		return nil, m.watchRegError
	}

	// 创建一个新的订阅，它会在context取消时关闭
	sub := event.NewSubscription(func(quit <-chan struct{}) error {
		for {
			select {
			case evt := <-m.regChan:
				select {
				case sink <- evt:
				case <-quit:
					return nil
				}
			case <-quit:
				return nil
			}
		}
	})

	return sub, nil
}

func (m *MockChainClient) WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorDeregisteredEvent) (event.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchDeregCallCount++
	if m.watchDeregError != nil {
		return nil, m.watchDeregError
	}

	// 创建一个新的订阅，它会在context取消时关闭
	sub := event.NewSubscription(func(quit <-chan struct{}) error {
		for {
			select {
			case evt := <-m.deregChan:
				select {
				case sink <- evt:
				case <-quit:
					return nil
				}
			case <-quit:
				return nil
			}
		}
	})

	return sub, nil
}

func (m *MockChainClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		close(m.regChan)
		close(m.deregChan)
		m.closed = true
	}
	return nil
}

// OperatorTestSuite 是主测试套件
type OperatorTestSuite struct {
	*testsuite.WPgxTestSuite
	eventListener *EventListener
	operatorRepo  OperatorRepo
	mockClient    *MockChainClient

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

	s.mockClient = NewMockChainClient()

	// 确保 operatorRepo 正确初始化
	operatorRepo := operators.New(conn, s.DCache)
	s.Require().NotNil(operatorRepo, "Operator repository should not be nil")
	s.operatorRepo = operatorRepo

	// 创建 EventListener
	eventListener, err := NewEventListener(context.Background(), &Config{
		MainnetClient: s.mockClient,
		OperatorRepo:  s.operatorRepo,
	})
	s.Require().NoError(err, "Failed to create event listener")
	s.eventListener = eventListener
}

// TearDownTest 在每个测试后运行
func (s *OperatorTestSuite) TearDownTest() {
	if s.eventListener != nil {
		s.eventListener.Stop()
	}
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

// Test cases for EventListener
func (s *OperatorTestSuite) TestEventListener() {
	for _, tc := range []struct {
		name          string
		setupMock     func(*MockChainClient)
		events        []interface{}
		expectedError error
		validateState func(*OperatorTestSuite)
		goldenFile    string
	}{
		{
			name: "successful_operator_registration",
			setupMock: func(m *MockChainClient) {
				m.weight = big.NewInt(100)
			},
			events: []interface{}{
				&ethereum.OperatorRegisteredEvent{
					Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
					SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
					P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
					BlockNumber: big.NewInt(2),
				},
			},
			expectedError: nil,
			validateState: func(s *OperatorTestSuite) {
				// Verify operator was registered correctly
				op, err := s.operatorRepo.GetOperatorByAddress(context.Background(), "0x0000000000000000000000000000000000000001")
				s.Require().NoError(err)
				s.Equal("0x0000000000000000000000000000000000000002", op.SigningKey)
				s.Equal("0x0000000000000000000000000000000000000003", op.P2pKey)
			},
			goldenFile: "successful_registration",
		},
		{
			name: "successful_operator_deregistration",
			setupMock: func(m *MockChainClient) {
				m.weight = big.NewInt(100)
			},
			events: []interface{}{
				&ethereum.OperatorRegisteredEvent{
					Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
					SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
					P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
					BlockNumber: big.NewInt(2),
				},
				&ethereum.OperatorDeregisteredEvent{
					Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
					BlockNumber: big.NewInt(14),
				},
			},
			expectedError: nil,
			validateState: func(s *OperatorTestSuite) {
				// Verify operator was deregistered correctly
				op, err := s.operatorRepo.GetOperatorByAddress(context.Background(), "0x0000000000000000000000000000000000000001")
				s.Require().NoError(err)
				s.NotEqual(uint64(4294967295), op.ExitEpoch)
			},
			goldenFile: "successful_deregistration",
		},
		{
			name: "registration_subscription_error",
			setupMock: func(m *MockChainClient) {
				m.watchRegError = fmt.Errorf("registration subscription failed")
			},
			events:        []interface{}{},
			expectedError: fmt.Errorf("failed to subscribe after retries: registration subscription failed"),
		},
		{
			name: "deregistration_subscription_error",
			setupMock: func(m *MockChainClient) {
				m.watchDeregError = fmt.Errorf("deregistration subscription failed")
			},
			events:        []interface{}{},
			expectedError: fmt.Errorf("failed to subscribe after retries: deregistration subscription failed"),
		},
	} {
		s.Run(tc.name, func() {
			s.SetupTest()
			if tc.setupMock != nil {
				tc.setupMock(s.mockClient)
			}

			// Send events
			for _, evt := range tc.events {
				switch e := evt.(type) {
				case *ethereum.OperatorRegisteredEvent:
					s.mockClient.regChan <- e
				case *ethereum.OperatorDeregisteredEvent:
					s.mockClient.deregChan <- e
				}
				// Give some time for event processing
				time.Sleep(3 * time.Second)
			}

			// Validate state if needed
			if tc.validateState != nil {
				tc.validateState(s)
			}

			// Compare with golden file if specified
			if tc.goldenFile != "" {
				operatorSerde := operatorStatesTableSerde{
					operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
				}
				s.Golden(tc.goldenFile, operatorSerde)
			}
		})
	}
}

func (s *OperatorTestSuite) TestEventListenerRetryLogic() {
	// 设置一个明确的错误
	expectedErr := fmt.Errorf("temporary error")
	s.mockClient.watchRegError = expectedErr
	s.mockClient.watchDeregError = expectedErr

	// Create event listener with retry logic
	listener, err := NewEventListener(context.Background(), &Config{
		MainnetClient: s.mockClient,
		OperatorRepo:  s.operatorRepo,
	})

	// Should fail after 3 retries
	s.Require().Error(err)
	s.Contains(err.Error(), expectedErr.Error(), "Error should contain the original error message")
	s.Equal(3, s.mockClient.watchRegCallCount, "Should retry 3 times")

	if listener != nil {
		listener.Stop()
	}
}

func (s *OperatorTestSuite) TestEventListenerContextCancellation() {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	listener, err := NewEventListener(ctx, &Config{
		MainnetClient: s.mockClient,
		OperatorRepo:  s.operatorRepo,
	})
	s.Require().NoError(err)

	// Cancel context
	cancel()

	// Give some time for shutdown
	time.Sleep(100 * time.Millisecond)

	// Verify cleanup
	listener.Stop()
}
