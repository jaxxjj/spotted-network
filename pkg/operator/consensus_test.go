package operator

import (
	"context"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
)

func (s *OperatorTestSuite) TestCheckConsensus() {
	tests := []struct {
		name         string
		initialState string
		setupState   func(tp *TaskProcessor)
		taskID       string
		wantErr      bool
		goldenFile   string
	}{
		{
			name:         "successful_consensus",
			initialState: "TestCheckConsensus.tasks.input.json",
			setupState: func(tp *TaskProcessor) {
				// 设置响应和权重
				tp.responses = map[string]map[string]*task_responses.TaskResponses{
					"task1": {
						"op1": &task_responses.TaskResponses{
							TaskID:    "task1",
							Epoch:     1,
							Value:     pgtype.Numeric{Int: big.NewInt(100), Valid: true},
							Signature: []byte("sig1"),
						},
						"op2": &task_responses.TaskResponses{
							TaskID:    "task1",
							Epoch:     1,
							Value:     pgtype.Numeric{Int: big.NewInt(100), Valid: true},
							Signature: []byte("sig2"),
						},
					},
				}
				tp.taskWeights = map[string]map[string]*big.Int{
					"task1": {
						"op1": big.NewInt(60),
						"op2": big.NewInt(40),
					},
				}

				// 设置epoch state
				threshold := big.NewInt(80)
				_, err := s.epochStateQuerier.UpsertEpochState(s.ctx, epoch_states.UpsertEpochStateParams{
					EpochNumber:     1,
					ThresholdWeight: pgtype.Numeric{Int: threshold, Valid: true},
					MinimumWeight:   pgtype.Numeric{Int: big.NewInt(10), Valid: true},
				})
				s.Require().NoError(err)
			},
			taskID:     "task1",
			wantErr:    false,
			goldenFile: "consensus_successful",
		},
		{
			name:         "threshold_not_reached",
			initialState: "TestCheckConsensus.tasks.input.json",
			setupState: func(tp *TaskProcessor) {
				// 设置响应和权重
				tp.responses = map[string]map[string]*task_responses.TaskResponses{
					"task1": {
						"op1": &task_responses.TaskResponses{
							TaskID:    "task1",
							Epoch:     1,
							Value:     pgtype.Numeric{Int: big.NewInt(100), Valid: true},
							Signature: []byte("sig1"),
						},
					},
				}
				tp.taskWeights = map[string]map[string]*big.Int{
					"task1": {
						"op1": big.NewInt(30),
					},
				}

				// 设置更高的阈值
				threshold := big.NewInt(80)
				_, err := s.epochStateQuerier.UpsertEpochState(s.ctx, epoch_states.UpsertEpochStateParams{
					EpochNumber:     1,
					ThresholdWeight: pgtype.Numeric{Int: threshold, Valid: true},
					MinimumWeight:   pgtype.Numeric{Int: big.NewInt(10), Valid: true},
				})
				s.Require().NoError(err)
			},
			taskID:     "task1",
			wantErr:    false,
			goldenFile: "consensus_threshold_not_reached",
		},
		{
			name:         "no_responses",
			initialState: "TestCheckConsensus.tasks.input.json",
			setupState: func(tp *TaskProcessor) {
				// Empty responses
				tp.responses = map[string]map[string]*task_responses.TaskResponses{}
				tp.taskWeights = map[string]map[string]*big.Int{}
			},
			taskID:     "task1",
			wantErr:    false,
			goldenFile: "consensus_no_responses",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "tasks.truncate", "TRUNCATE TABLE tasks")
			s.Require().NoError(err)
			_, err = s.GetPool().WConn().WExec(s.ctx, "consensus_responses.truncate", "TRUNCATE TABLE consensus_responses")
			s.Require().NoError(err)
			_, err = s.GetPool().WConn().WExec(s.ctx, "epoch_states.truncate", "TRUNCATE TABLE epoch_states")
			s.Require().NoError(err)

			// 加载初始数据
			serde := &TasksTableSerde{tasksQuerier: s.tasksQuerier}
			s.LoadState(tt.initialState, serde)

			// 创建TaskProcessor
			tp := &TaskProcessor{
				tasks:             s.tasksQuerier,
				consensusResponse: s.consensusResponseQuerier,
				epochState:        s.epochStateQuerier,
				responses:         make(map[string]map[string]*task_responses.TaskResponses),
				taskWeights:       make(map[string]map[string]*big.Int),
			}

			// 设置状态
			if tt.setupState != nil {
				tt.setupState(tp)
			}

			// 执行测试
			err = tp.checkConsensus(tt.taskID)

			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}

			// 验证状态
			s.Golden(tt.goldenFile, serde)
		})
	}
}

func (s *OperatorTestSuite) TestGetConsensusThreshold() {
	tests := []struct {
		name       string
		setupState func(context.Context)
		wantErr    bool
		wantValue  *big.Int
	}{
		{
			name: "successful_get_threshold",
			setupState: func(ctx context.Context) {
				threshold := big.NewInt(80)
				_, err := s.epochStateQuerier.UpsertEpochState(ctx, epoch_states.UpsertEpochStateParams{
					EpochNumber:     1,
					ThresholdWeight: pgtype.Numeric{Int: threshold, Valid: true},
					MinimumWeight:   pgtype.Numeric{Int: big.NewInt(10), Valid: true},
				})
				s.Require().NoError(err)
			},
			wantErr:   false,
			wantValue: big.NewInt(80),
		},
		{
			name: "nil_epoch_state",
			setupState: func(ctx context.Context) {
				// 不插入任何数据,让表为空
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "epoch_states.truncate", "TRUNCATE TABLE epoch_states")
			s.Require().NoError(err)

			// 创建完整的TaskProcessor
			tp := &TaskProcessor{
				node:              s.node,
				tasks:             s.tasksQuerier,
				taskResponse:      s.taskResponseQuerier,
				consensusResponse: s.consensusResponseQuerier,
				epochState:        s.epochStateQuerier,
				chainManager:      s.mockChainManager,
				responses:         make(map[string]map[string]*task_responses.TaskResponses),
				taskWeights:       make(map[string]map[string]*big.Int),
			}

			// 设置测试状态
			if tt.setupState != nil {
				tt.setupState(s.ctx)
			}

			// 执行测试
			threshold, err := tp.getConsensusThreshold(s.ctx)

			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(0, threshold.Cmp(tt.wantValue))
			}
		})
	}
}
