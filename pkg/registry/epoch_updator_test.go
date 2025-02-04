package registry

import (
	"context"
	"time"

	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/stretchr/testify/mock"
)

// EpochState 用于序列化/反序列化 epoch 状态
type EpochState struct {
	Operators []struct {
		Address     string `json:"address"`
		ActiveEpoch uint32 `json:"active_epoch"`
		ExitEpoch   uint32 `json:"exit_epoch"`
		Status      types.OperatorStatus `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"operators"`
}


// SetupEpochTest 在每个 epoch 测试前运行
func (s *RegistryTestSuite) SetupEpochTest() {
	// 初始化 mock 组件
	s.mockMainnet = &MockMainnetClient{}
	s.mockTxManager = &MockTxManager{}
	s.mockOpQuerier = NewMockOperatorsQuerier()
	s.mockStateSync = &MockStateSyncNotifier{}

	// 清理数据库
	_, err := s.GetPool().WConn().WExec(s.ctx, "operators.truncate", "TRUNCATE TABLE operators")
	s.Require().NoError(err)
}

// TestEpochUpdator_Stop 测试 epoch updator 的停止功能
func (s *RegistryTestSuite) TestEpochUpdator_Stop() {
	s.SetupEpochTest()

	// Setup expectations
	s.mockMainnet.On("BlockNumber", mock.Anything).Return(uint64(100), nil).Maybe()

	// Create updator
	ctx, cancel := context.WithCancel(context.Background())
	updator := &EpochUpdator{
		node:          s.mockStateSync,
		mainnetClient: s.mockMainnet,
		txManager:     s.mockTxManager,
		opQuerier:     s.mockOpQuerier,
		cancel:        cancel,
	}

	// Start the updator
	updator.wg.Add(1)
	go func() {
		defer updator.wg.Done()
		err := updator.start(ctx)
		s.NoError(err)
	}()

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Stop the updator
	updator.Stop()

	// Verify that the wait group is done
	done := make(chan struct{})
	go func() {
		updator.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - wait group completed
	case <-time.After(1 * time.Second):
		s.Fail("Timeout waiting for updator to stop")
	}

	s.mockMainnet.AssertExpectations(s.T())
}

