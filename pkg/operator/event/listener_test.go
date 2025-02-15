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
	regSub              event.Subscription
	deregSub            event.Subscription
	regChan             chan *ethereum.OperatorRegisteredEvent
	deregChan           chan *ethereum.OperatorDeregisteredEvent
	watchRegError       error
	watchDeregError     error
	blockNumberError    error
	watchRegCallCount   int
	watchDeregCallCount int
	closed              bool
	mu                  sync.Mutex
}

func NewMockChainClient() *MockChainClient {
	return &MockChainClient{
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

func (m *MockChainClient) WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchRegCallCount++
	if m.watchRegError != nil {
		return nil, m.watchRegError
	}

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

func (m *MockChainClient) BlockNumber(ctx context.Context) (uint64, error) {
	if m.blockNumberError != nil {
		return 0, m.blockNumberError
	}
	return 100, nil
}

type OperatorTestSuite struct {
	*testsuite.WPgxTestSuite
	eventListener *eventListener
	operatorRepo  OperatorRepo
	mockClient    *MockChainClient

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

	s.mockClient = NewMockChainClient()

	operatorRepo := operators.New(conn, s.DCache)
	s.Require().NotNil(operatorRepo, "Operator repository should not be nil")
	s.operatorRepo = operatorRepo

	listener, err := NewEventListener(context.Background(), &Config{
		MainnetClient: s.mockClient,
		OperatorRepo:  s.operatorRepo,
	})
	s.Require().NoError(err, "Failed to create event listener")
	s.eventListener = listener.(*eventListener)
}

func (s *OperatorTestSuite) TearDownTest() {
	if s.eventListener != nil {
		s.eventListener.Stop()
	}
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

	memCache := freecache.NewCache(100 * 1024 * 1024)

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
func (s *OperatorTestSuite) TestEventListenerInitialization() {
	// Test successful initialization
	listener, err := NewEventListener(context.Background(), &Config{
		MainnetClient: s.mockClient,
		OperatorRepo:  s.operatorRepo,
	})
	s.Require().NoError(err)
	s.NotNil(listener)

	// Test initialization with nil config
	listener, err = NewEventListener(context.Background(), nil)
	s.Error(err)
	s.Nil(listener)

	// Test initialization with missing MainnetClient
	listener, err = NewEventListener(context.Background(), &Config{
		OperatorRepo: s.operatorRepo,
	})
	s.Error(err)
	s.Nil(listener)

	// Test initialization with missing OperatorRepo
	listener, err = NewEventListener(context.Background(), &Config{
		MainnetClient: s.mockClient,
	})
	s.Error(err)
	s.Nil(listener)

}

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
		{
			name: "duplicate_registration",
			setupMock: func(m *MockChainClient) {
			},
			events: []interface{}{
				&ethereum.OperatorRegisteredEvent{
					Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
					SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
					P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
					BlockNumber: big.NewInt(2),
				},
				&ethereum.OperatorRegisteredEvent{
					Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
					SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
					P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
					BlockNumber: big.NewInt(3),
				},
			},
			validateState: func(s *OperatorTestSuite) {
				op, err := s.operatorRepo.GetOperatorByAddress(context.Background(), "0x0000000000000000000000000000000000000001")
				s.Require().NoError(err)
				s.Equal(uint64(3), op.RegisteredAtBlockNumber)
			},
			goldenFile: "duplicate_registration",
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

func (s *OperatorTestSuite) TestConcurrentRegistration() {
	const numOperators = 5
	var wg sync.WaitGroup

	for i := 0; i < numOperators; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			addr := fmt.Sprintf("0x%040x", index+1)
			event := &ethereum.OperatorRegisteredEvent{
				Operator:    common.HexToAddress(addr),
				SigningKey:  common.HexToAddress(fmt.Sprintf("0x%040x", index+1000)),
				P2PKey:      common.HexToAddress(fmt.Sprintf("0x%040x", index+2000)),
				BlockNumber: big.NewInt(int64(index + 1)),
			}

			s.mockClient.regChan <- event
		}(i)
	}

	wg.Wait()
	time.Sleep(3 * time.Second)

	for i := 0; i < numOperators; i++ {
		addr := fmt.Sprintf("0x%040x", i+1)
		op, err := s.operatorRepo.GetOperatorByAddress(context.Background(), addr)
		s.Require().NoError(err)
		s.NotNil(op)
	}
}

func (s *OperatorTestSuite) TestEpochBoundaries() {
	testCases := []struct {
		name          string
		blockNumber   uint64
		expectedEpoch uint32
	}{
		{
			name:          "min_block_number",
			blockNumber:   0,
			expectedEpoch: 1,
		},
		{
			name:          "epoch_transition",
			blockNumber:   27,
			expectedEpoch: 3,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			event := &ethereum.OperatorRegisteredEvent{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
				SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
				P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
				BlockNumber: new(big.Int).SetUint64(tc.blockNumber),
			}

			s.mockClient.regChan <- event
			time.Sleep(1 * time.Second)

			op, err := s.operatorRepo.GetOperatorByAddress(context.Background(),
				"0x0000000000000000000000000000000000000001")
			s.Require().NoError(err)
			s.Equal(tc.expectedEpoch, op.ActiveEpoch)
		})
	}
}

func (s *OperatorTestSuite) TestErrorRecoveryAndRetry() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	event := &ethereum.OperatorRegisteredEvent{
		Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
		SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
		P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
		BlockNumber: big.NewInt(1),
	}

	s.mockClient.blockNumberError = fmt.Errorf("temporary block number error")

	s.mockClient.regChan <- event

	time.Sleep(2 * time.Second)
	s.mockClient.blockNumberError = nil

	time.Sleep(1 * time.Second)
	op, err := s.operatorRepo.GetOperatorByAddress(ctx, "0x0000000000000000000000000000000000000001")
	s.NoError(err)
	s.NotNil(op)
	s.Equal(uint64(1), op.RegisteredAtBlockNumber)
}

func (s *OperatorTestSuite) TestEventProcessingOrder() {
	events := []struct {
		eventType string
		event     interface{}
		delay     time.Duration
	}{
		{
			eventType: "register",
			event: &ethereum.OperatorRegisteredEvent{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
				SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
				P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
				BlockNumber: big.NewInt(1),
			},
			delay: 100 * time.Millisecond,
		},
		{
			eventType: "deregister",
			event: &ethereum.OperatorDeregisteredEvent{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
				BlockNumber: big.NewInt(2),
			},
			delay: 50 * time.Millisecond,
		},
		{
			eventType: "register",
			event: &ethereum.OperatorRegisteredEvent{
				Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
				SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
				P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
				BlockNumber: big.NewInt(3),
			},
			delay: 0,
		},
	}

	for _, evt := range events {
		time.Sleep(evt.delay)
		switch e := evt.event.(type) {
		case *ethereum.OperatorRegisteredEvent:
			s.mockClient.regChan <- e
		case *ethereum.OperatorDeregisteredEvent:
			s.mockClient.deregChan <- e
		}
	}

	time.Sleep(2 * time.Second)

	op, err := s.operatorRepo.GetOperatorByAddress(context.Background(),
		"0x0000000000000000000000000000000000000001")
	s.NoError(err)
	s.Equal(uint64(3), op.RegisteredAtBlockNumber)
}

func (s *OperatorTestSuite) TestResourceLeakage() {
	listeners := make([]*eventListener, 0)
	for i := 0; i < 10; i++ {
		listener, err := NewEventListener(context.Background(), &Config{
			MainnetClient: s.mockClient,
			OperatorRepo:  s.operatorRepo,
		})
		s.NoError(err)
		listeners = append(listeners, listener.(*eventListener))
	}

	event := &ethereum.OperatorRegisteredEvent{
		Operator:    common.HexToAddress("0x0000000000000000000000000000000000000001"),
		SigningKey:  common.HexToAddress("0x0000000000000000000000000000000000000002"),
		P2PKey:      common.HexToAddress("0x0000000000000000000000000000000000000003"),
		BlockNumber: big.NewInt(1),
	}
	s.mockClient.regChan <- event

	for _, l := range listeners {
		l.Stop()
	}

	time.Sleep(1 * time.Second)
	s.True(s.mockClient.closed)
}

// TestMultipleRegistrationsWithKeyChanges test the scenario of multiple registrations with key changes
func (s *OperatorTestSuite) TestMultipleRegistrationsWithKeyChanges() {
	operatorAddr := common.HexToAddress("0x0000000000000000000000000000000000000001")

	signingKey1 := common.HexToAddress("0x0000000000000000000000000000000000000002")
	p2pKey1 := common.HexToAddress("0x0000000000000000000000000000000000000003")

	signingKey2 := common.HexToAddress("0x0000000000000000000000000000000000000004")
	p2pKey2 := common.HexToAddress("0x0000000000000000000000000000000000000005")

	signingKey3 := common.HexToAddress("0x0000000000000000000000000000000000000006")
	p2pKey3 := common.HexToAddress("0x0000000000000000000000000000000000000007")

	registerEvents := []*ethereum.OperatorRegisteredEvent{
		{
			Operator:    operatorAddr,
			SigningKey:  signingKey1,
			P2PKey:      p2pKey1,
			BlockNumber: big.NewInt(10),
		},
		{
			Operator:    operatorAddr,
			SigningKey:  signingKey2,
			P2PKey:      p2pKey2,
			BlockNumber: big.NewInt(20),
		},
		{
			Operator:    operatorAddr,
			SigningKey:  signingKey3,
			P2PKey:      p2pKey3,
			BlockNumber: big.NewInt(30),
		},
	}

	for i, evt := range registerEvents {
		s.mockClient.regChan <- evt
		time.Sleep(1 * time.Second)

		// verify the state after each register
		op, err := s.operatorRepo.GetOperatorByAddress(context.Background(), operatorAddr.Hex())
		s.Require().NoError(err)
		s.Equal(evt.BlockNumber.Uint64(), op.RegisteredAtBlockNumber,
			"Should have updated register block number at step %d", i)
		s.Equal(evt.SigningKey.Hex(), op.SigningKey,
			"Should have updated signing key at step %d", i)
		s.Equal(evt.P2PKey.Hex(), op.P2pKey,
			"Should have updated p2p key at step %d", i)
		s.Equal(uint32(4294967295), op.ExitEpoch,
			"Should have max exit epoch at step %d", i)
	}

	// send deregister event
	deregisterEvent := &ethereum.OperatorDeregisteredEvent{
		Operator:    operatorAddr,
		BlockNumber: big.NewInt(40),
	}
	s.mockClient.deregChan <- deregisterEvent
	time.Sleep(1 * time.Second)

	// verify the state after deregister
	op, err := s.operatorRepo.GetOperatorByAddress(context.Background(), operatorAddr.Hex())
	s.Require().NoError(err)
	s.Equal(uint64(30), op.RegisteredAtBlockNumber, "Register block number should not change")
	s.Equal(signingKey3.Hex(), op.SigningKey, "Signing key should remain unchanged")
	s.Equal(p2pKey3.Hex(), op.P2pKey, "P2P key should remain unchanged")
	s.Equal(uint32(4), op.ExitEpoch, "Should have updated exit epoch to 4")

	// register again, using new key
	finalSigningKey := common.HexToAddress("0x0000000000000000000000000000000000000008")
	finalP2PKey := common.HexToAddress("0x0000000000000000000000000000000000000009")

	finalRegisterEvent := &ethereum.OperatorRegisteredEvent{
		Operator:    operatorAddr,
		SigningKey:  finalSigningKey,
		P2PKey:      finalP2PKey,
		BlockNumber: big.NewInt(50),
	}
	s.mockClient.regChan <- finalRegisterEvent
	time.Sleep(1 * time.Second)

	//verify the final state
	op, err = s.operatorRepo.GetOperatorByAddress(context.Background(), operatorAddr.Hex())
	s.Require().NoError(err)
	s.Equal(uint64(50), op.RegisteredAtBlockNumber, "Should have new register block number")
	s.Equal(finalSigningKey.Hex(), op.SigningKey, "Should have final signing key")
	s.Equal(finalP2PKey.Hex(), op.P2pKey, "Should have final p2p key")
	s.Equal(uint32(4294967295), op.ExitEpoch, "Should have max exit epoch again")

	operatorSerde := operatorStatesTableSerde{
		operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
	}
	s.Golden("multiple_registrations_with_key_changes", operatorSerde)
}
