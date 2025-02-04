package operator

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
)

func (s *OperatorTestSuite) TestUpdateEpochState() {
	tests := []struct {
		name         string
		epochNumber  uint32
		setupMocks   func(mockChainManager *MockChainManager, mockClient *MockChainClient)
		wantErr      bool
		goldenFile   string
	}{
		{
			name:        "successful_update",
			epochNumber: 1,
			setupMocks: func(mockChainManager *MockChainManager, mockClient *MockChainClient) {
				// 设置mock client返回值
				mockClient.On("GetMinimumWeight", s.ctx).Return(big.NewInt(10), nil)
				mockClient.On("GetTotalWeight", s.ctx).Return(big.NewInt(100), nil)
				mockClient.On("GetThresholdWeight", s.ctx).Return(big.NewInt(80), nil)
				
				// 设置chain manager返回mock client
				mockChainManager.On("GetMainnetClient").Return(mockClient, nil)
			},
			wantErr:     false,
			goldenFile:  "TestUpdateEpochState/successful_update.golden",
		},
		{
			name:        "mainnet_client_error",
			epochNumber: 1,
			setupMocks: func(mockChainManager *MockChainManager, mockClient *MockChainClient) {
				mockChainManager.On("GetMainnetClient").Return(nil, fmt.Errorf("mainnet client error"))
			},
			wantErr:     true,
		},
		{
			name:        "get_minimum_weight_error",
			epochNumber: 1,
			setupMocks: func(mockChainManager *MockChainManager, mockClient *MockChainClient) {
				mockChainManager.On("GetMainnetClient").Return(mockClient, nil)
				mockClient.On("GetMinimumWeight", s.ctx).Return(nil, fmt.Errorf("minimum weight error"))
			},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "epoch_states.truncate", "TRUNCATE TABLE epoch_states")
			s.Require().NoError(err)

			// 创建mocks
			mockChainManager := new(MockChainManager)
			mockClient := new(MockChainClient)

			// 设置mock行为
			if tt.setupMocks != nil {
				tt.setupMocks(mockChainManager, mockClient)
			}

			// 创建node
			node := &Node{
				chainManager:      mockChainManager,
				epochStateQuerier: s.epochStateQuerier,
			}

			// 执行测试
			err = node.updateEpochState(s.ctx, tt.epochNumber)

			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)

			// 验证golden file
			if tt.goldenFile != "" {
				serde := &EpochStateTableSerde{epochStateQuerier: s.epochStateQuerier}
				s.Golden(tt.goldenFile, serde)
			}

			// 验证mock调用
			mockChainManager.AssertExpectations(s.T())
			mockClient.AssertExpectations(s.T())
		})
	}
}

// EpochStateTableSerde 用于序列化/反序列化epoch state
type EpochStateTableSerde struct {
	epochStateQuerier *epoch_states.Queries
}

func (e EpochStateTableSerde) Load(data []byte) error {
	if err := e.epochStateQuerier.Load(context.Background(), data); err != nil {
		return fmt.Errorf("failed to load epoch state data: %w", err)
	}
	return nil
}

func (e EpochStateTableSerde) Dump() ([]byte, error) {
	return e.epochStateQuerier.Dump(context.Background(), func(m *epoch_states.EpochState) {
		m.UpdatedAt = time.Unix(0, 0).UTC()
	})
} 