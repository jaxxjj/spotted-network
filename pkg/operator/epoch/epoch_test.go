package epoch

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
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

// mockMainnetClient implements MainnetClient interface for testing
type mockMainnetClient struct {
	mock.Mock
}

func newMockMainnetClient() *mockMainnetClient {
	return &mockMainnetClient{}
}

func (m *mockMainnetClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockMainnetClient) GetOperatorWeight(ctx context.Context, address common.Address) (*big.Int, error) {
	args := m.Called(ctx, address)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMainnetClient) GetOperatorSigningKey(ctx context.Context, operator common.Address, epoch uint32) (common.Address, error) {
	args := m.Called(ctx, operator, epoch)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *mockMainnetClient) GetOperatorP2PKey(ctx context.Context, operator common.Address, epoch uint32) (common.Address, error) {
	args := m.Called(ctx, operator, epoch)
	return args.Get(0).(common.Address), args.Error(1)
}

func (m *mockMainnetClient) GetMinimumWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMainnetClient) GetTotalWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMainnetClient) GetThresholdWeight(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if v := args.Get(0); v != nil {
		return v.(*big.Int), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMainnetClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type OperatorTestSuite struct {
	*testsuite.WPgxTestSuite
	updator *epochUpdator

	operatorRepo  OperatorRepo
	mainnetClient *mockMainnetClient

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
	s.Require().NotNil(pool)

	operatorRepo := operators.New(pool.WConn(), s.DCache)
	s.Require().NotNil(operatorRepo)
	s.operatorRepo = operatorRepo

	s.mainnetClient = newMockMainnetClient()
	updator, err := NewEpochUpdator(context.Background(), &Config{
		OperatorRepo:  operatorRepo,
		MainnetClient: s.mainnetClient,
		TxManager:     pool,
	})
	s.Require().NoError(err)
	s.updator = updator.(*epochUpdator)
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

// Test cases for epoch updator
func (s *OperatorTestSuite) TestEpochUpdatorInitialization() {
	updator, err := NewEpochUpdator(context.Background(), &Config{
		OperatorRepo:  s.operatorRepo,
		MainnetClient: s.mainnetClient,
		TxManager:     s.GetPool(),
	})
	s.Require().NoError(err)
	s.NotNil(updator)

	// Test initialization with nil config
	updator, err = NewEpochUpdator(context.Background(), nil)
	s.Error(err)
	s.Nil(updator)

	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestConcurrentEpochUpdates() {
	const numUpdates = 5
	var wg sync.WaitGroup

	// Setup mock expectations
	for i := 0; i < numUpdates; i++ {
		s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
		s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
		s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()
	}

	// Perform concurrent updates
	for i := 0; i < numUpdates; i++ {
		wg.Add(1)
		go func(epoch uint32) {
			defer wg.Done()
			err := s.updator.updateEpochState(context.Background(), epoch)
			s.NoError(err)
		}(uint32(i + 1))
	}

	wg.Wait()

	// Verify final state is valid
	state := s.updator.GetEpochState()
	s.NotNil(state.MinimumWeight)
	s.NotNil(state.TotalWeight)
	s.NotNil(state.ThresholdWeight)

	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestEpochStateUpdate() {
	// Setup mock expectations
	s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
	s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
	s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()

	// Test normal update
	err := s.updator.updateEpochState(context.Background(), 1)
	s.NoError(err)

	// Verify state
	state := s.updator.GetEpochState()
	s.Equal(uint32(1), state.EpochNumber)
	s.Equal(big.NewInt(1), state.MinimumWeight)
	s.Equal(big.NewInt(100), state.TotalWeight)
	s.Equal(big.NewInt(67), state.ThresholdWeight)

	// Test error cases
	s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(nil, fmt.Errorf("mock error")).Once()
	err = s.updator.updateEpochState(context.Background(), 2)
	s.Error(err)

	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestOperatorActiveStatus() {
	testCases := []struct {
		name         string
		currentBlock uint64
		activeEpoch  uint32
		exitEpoch    uint32
		expectActive bool
	}{
		{
			name:         "only_active_epoch_not_reached",
			currentBlock: 5,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			expectActive: false,
		},
		{
			name:         "only_active_epoch_reached",
			currentBlock: 15,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			expectActive: true,
		},
		{
			name:         "exit_after_active_before_active",
			currentBlock: 5,
			activeEpoch:  1,
			exitEpoch:    2,
			expectActive: false,
		},
		{
			name:         "exit_after_active_during_active",
			currentBlock: 15,
			activeEpoch:  1,
			exitEpoch:    2,
			expectActive: true,
		},
		{
			name:         "exit_after_active_after_exit",
			currentBlock: 25,
			activeEpoch:  1,
			exitEpoch:    2,
			expectActive: false,
		},
		{
			name:         "exit_before_active_before_exit",
			currentBlock: 5,
			activeEpoch:  2,
			exitEpoch:    1,
			expectActive: true,
		},
		{
			name:         "exit_before_active_during_inactive",
			currentBlock: 15,
			activeEpoch:  2,
			exitEpoch:    1,
			expectActive: false,
		},
		{
			name:         "exit_before_active_after_reactivation",
			currentBlock: 25,
			activeEpoch:  2,
			exitEpoch:    1,
			expectActive: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			active := IsOperatorActive(tc.currentBlock, tc.activeEpoch, tc.exitEpoch)
			s.Equal(tc.expectActive, active)
		})
	}
}

func (s *OperatorTestSuite) TestOperatorActiveStatusEdgeCases() {
	testCases := []struct {
		name         string
		currentBlock uint64
		activeEpoch  uint32
		exitEpoch    uint32
		expectActive bool
	}{
		{
			name:         "max_uint32_exit_epoch",
			currentBlock: 100,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			expectActive: true,
		},
		{
			name:         "same_active_and_exit_epoch",
			currentBlock: 100,
			activeEpoch:  5,
			exitEpoch:    5,
			expectActive: false,
		},
		{
			name:         "zero_active_epoch",
			currentBlock: 100,
			activeEpoch:  0,
			exitEpoch:    4294967295,
			expectActive: true,
		},
		{
			name:         "zero_current_block",
			currentBlock: 0,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			expectActive: false,
		},
		{
			name:         "max_uint64_current_block",
			currentBlock: 18446744073709551615,
			activeEpoch:  1,
			exitEpoch:    2,
			expectActive: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			active := IsOperatorActive(tc.currentBlock, tc.activeEpoch, tc.exitEpoch)
			s.Equal(tc.expectActive, active)
		})
	}
}

func (s *OperatorTestSuite) TestEpochStateGetters() {
	// Setup initial state
	s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
	s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
	s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()

	err := s.updator.updateEpochState(context.Background(), 42)
	s.NoError(err)

	// Test getters
	s.Equal(uint32(42), s.updator.GetCurrentEpochNumber())
	s.Equal(big.NewInt(67), s.updator.GetThresholdWeight())

	state := s.updator.GetEpochState()
	s.Equal(uint32(42), state.EpochNumber)
	s.Equal(big.NewInt(1), state.MinimumWeight)
	s.Equal(big.NewInt(100), state.TotalWeight)
	s.Equal(big.NewInt(67), state.ThresholdWeight)

	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestEpochStateConsistency() {
	// Setup initial state
	s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
	s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
	s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()

	err := s.updator.updateEpochState(context.Background(), 42)
	s.NoError(err)

	// Test state consistency across all getters
	state := s.updator.GetEpochState()
	s.Equal(state.EpochNumber, s.updator.GetCurrentEpochNumber())
	s.Equal(state.ThresholdWeight, s.updator.GetThresholdWeight())

	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestEpochUpdatorStop() {
	// Create a new updator with a mock client
	updator, err := NewEpochUpdator(context.Background(), &Config{
		OperatorRepo:  s.operatorRepo,
		MainnetClient: s.mainnetClient,
		TxManager:     s.GetPool(),
	})
	s.Require().NoError(err)
	s.NotNil(updator)

	// Setup mock expectations for Close
	s.mainnetClient.On("Close").Return(nil).Once()

	// Stop the updator
	s.updator.Stop()

	// Verify expectations
	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestCompleteEpochUpdateFlow() {
	// Setup test data
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	operatorSerde := operatorStatesTableSerde{
		operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
	}
	s.LoadState("TestOperatorSuite/TestCompleteEpochUpdateFlow/active_operator.json", operatorSerde)
	// Setup mock expectations for first epoch
	s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(17), nil).Once()
	s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
	s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
	s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()
	s.mainnetClient.On("GetOperatorWeight", mock.Anything, addr).Return(big.NewInt(10), nil).Once()
	s.mainnetClient.On("GetOperatorSigningKey", mock.Anything, addr, uint32(1)).Return(
		common.HexToAddress("0x0000000000000000000000000000000000000005"), nil).Once()
	s.mainnetClient.On("GetOperatorP2PKey", mock.Anything, addr, uint32(1)).Return(
		common.HexToAddress("0x0000000000000000000000000000000000000006"), nil).Once()

	// Test first epoch update
	err := s.updator.handleEpochUpdate(context.Background(), 1)
	s.NoError(err)

	// Verify state after first update using golden file
	s.Golden("after_first_update", operatorSerde)

	// Setup mock expectations for second epoch with changes
	s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(26), nil).Once()
	s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(2), nil).Once()
	s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(200), nil).Once()
	s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(134), nil).Once()
	s.mainnetClient.On("GetOperatorWeight", mock.Anything, addr).Return(big.NewInt(20), nil).Once()
	s.mainnetClient.On("GetOperatorSigningKey", mock.Anything, addr, uint32(2)).Return(
		common.HexToAddress("0x0000000000000000000000000000000000000007"), nil).Once()
	s.mainnetClient.On("GetOperatorP2PKey", mock.Anything, addr, uint32(2)).Return(
		common.HexToAddress("0x0000000000000000000000000000000000000008"), nil).Once()

	// Test second epoch update
	err = s.updator.handleEpochUpdate(context.Background(), 2)
	s.NoError(err)

	// Verify final state using golden file
	s.Golden("final_state", operatorSerde)

	// Verify epoch state
	state := s.updator.GetEpochState()
	s.Equal(uint32(2), state.EpochNumber)
	s.Equal(big.NewInt(2), state.MinimumWeight)
	s.Equal(big.NewInt(200), state.TotalWeight)
	s.Equal(big.NewInt(134), state.ThresholdWeight)

	s.mainnetClient.AssertExpectations(s.T())
}

func (s *OperatorTestSuite) TestEpochUpdateErrorHandling() {
	// Setup test data
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	operatorSerde := operatorStatesTableSerde{
		operatorStatesQuerier: s.operatorRepo.(*operators.Queries),
	}
	s.LoadState("TestOperatorSuite/TestEpochUpdateErrorHandling/active_operator.json", operatorSerde)

	testCases := []struct {
		name        string
		setupMocks  func()
		expectError bool
	}{
		{
			name: "block_number_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(0), fmt.Errorf("block number error")).Once()
			},
			expectError: true,
		},
		{
			name: "minimum_weight_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
				s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(nil, fmt.Errorf("minimum weight error")).Once()
			},
			expectError: true,
		},
		{
			name: "total_weight_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
				s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
				s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(nil, fmt.Errorf("total weight error")).Once()
			},
			expectError: true,
		},
		{
			name: "threshold_weight_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
				s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
				s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
				s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(nil, fmt.Errorf("threshold weight error")).Once()
			},
			expectError: true,
		},
		{
			name: "operator_weight_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
				s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
				s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
				s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()
				s.mainnetClient.On("GetOperatorWeight", mock.Anything, addr).Return(nil, fmt.Errorf("operator weight error")).Once()
			},
			expectError: true,
		},
		{
			name: "signing_key_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
				s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
				s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
				s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()
				s.mainnetClient.On("GetOperatorWeight", mock.Anything, addr).Return(big.NewInt(10), nil).Once()
				s.mainnetClient.On("GetOperatorSigningKey", mock.Anything, addr, uint32(1)).Return(
					common.Address{}, fmt.Errorf("signing key error")).Once()
			},
			expectError: true,
		},
		{
			name: "p2p_key_error",
			setupMocks: func() {
				s.mainnetClient.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Once()
				s.mainnetClient.On("GetMinimumWeight", mock.Anything).Return(big.NewInt(1), nil).Once()
				s.mainnetClient.On("GetTotalWeight", mock.Anything).Return(big.NewInt(100), nil).Once()
				s.mainnetClient.On("GetThresholdWeight", mock.Anything).Return(big.NewInt(67), nil).Once()
				s.mainnetClient.On("GetOperatorWeight", mock.Anything, addr).Return(big.NewInt(10), nil).Once()
				s.mainnetClient.On("GetOperatorSigningKey", mock.Anything, addr, uint32(1)).Return(
					common.HexToAddress("0x0000000000000000000000000000000000000002"), nil).Once()
				s.mainnetClient.On("GetOperatorP2PKey", mock.Anything, addr, uint32(1)).Return(
					common.Address{}, fmt.Errorf("p2p key error")).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock expectations
			s.mainnetClient.ExpectedCalls = nil

			// Setup mocks
			tc.setupMocks()

			// Run test
			err := s.updator.handleEpochUpdate(context.Background(), 1)

			if tc.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}

			// Verify all mocks were called
			s.mainnetClient.AssertExpectations(s.T())
		})
	}
}
