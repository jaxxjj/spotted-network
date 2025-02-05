package operator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/stretchr/testify/mock"
)

// MockPingService mocks the PingService interface
type MockPingService struct {
	mock.Mock
}

func (m *MockPingService) Ping(ctx context.Context, p peer.ID) <-chan ping.Result {
	args := m.Called(ctx, p)
	return args.Get(0).(<-chan ping.Result)
}

func (s *OperatorTestSuite) TestHealthChecker() {
	tests := []struct {
		name       string
		setupNode  func() *Node
		setupMocks func(*MockPingService, *MockP2PHost)
		operations func(*HealthChecker)
		wantErr    bool
	}{
		{
			name: "successful_ping",
			setupNode: func() *Node {
				// 创建初始operator
				id, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)

				active := make(map[peer.ID]*OperatorState)
				active[id] = &OperatorState{
					PeerID: id,
					Address: "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
				}

				return &Node{
					activeOperators: ActivePeerStates{
						active: active,
						mu:     sync.RWMutex{},
					},
				}
			},
			setupMocks: func(mockPing *MockPingService, mockHost *MockP2PHost) {
				// 设置mock host
				mockHost.On("ID").Return(peer.ID("local-peer-id"))
				mockHost.On("Network").Return(new(MockNetwork))

				// 设置成功的ping结果
				resultChan := make(chan ping.Result, 1)
				resultChan <- ping.Result{
					RTT:   time.Millisecond * 100,
					Error: nil,
				}
				mockPing.On("Ping", mock.Anything, mock.Anything).Return((<-chan ping.Result)(resultChan))
			},
			operations: func(hc *HealthChecker) {
				// 执行健康检查
				hc.checkOperators(context.Background())
			},
			wantErr: false,
		},
		{
			name: "ping_failure",
			setupNode: func() *Node {
				// 创建初始operator
				id, err := peer.Decode("12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk")
				s.Require().NoError(err)

				active := make(map[peer.ID]*OperatorState)
				active[id] = &OperatorState{
					PeerID: id,
					Address: "0x0000000000000000000000000000000000000005",
				}

				return &Node{
					activeOperators: ActivePeerStates{
						active: active,
						mu:     sync.RWMutex{},
					},
				}
			},
			setupMocks: func(mockPing *MockPingService, mockHost *MockP2PHost) {
				// 设置mock host
				mockHost.On("ID").Return(peer.ID("local-peer-id"))
				mockHost.On("Network").Return(new(MockNetwork))

				// 设置失败的ping结果
				resultChan := make(chan ping.Result, 1)
				resultChan <- ping.Result{
					RTT:   0,
					Error: fmt.Errorf("ping failed"),
				}
				mockPing.On("Ping", mock.Anything, mock.Anything).Return((<-chan ping.Result)(resultChan))
			},
			operations: func(hc *HealthChecker) {
				// 执行健康检查
				hc.checkOperators(context.Background())
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 创建mocks
			mockPing := new(MockPingService)
			mockHost := new(MockP2PHost)

			// 设置mock行为
			if tt.setupMocks != nil {
				tt.setupMocks(mockPing, mockHost)
			}

			// 创建node
			node := tt.setupNode()
			node.host = mockHost

			// 创建health checker
			hc, err := newHealthChecker(context.Background(), node, mockPing)
			s.Require().NoError(err)

			// 执行测试操作
			if tt.operations != nil {
				tt.operations(hc)
			}

			// 验证mock调用
			mockPing.AssertExpectations(s.T())
			mockHost.AssertExpectations(s.T())
		})
	}
}

// MockNetwork mocks the network.Network interface
type MockNetwork struct {
	mock.Mock
	network.Network
}

func (m *MockNetwork) Peers() []peer.ID {
	args := m.Called()
	return args.Get(0).([]peer.ID)
}

func (m *MockNetwork) Peerstore() peerstore.Peerstore {
	args := m.Called()
	return args.Get(0).(peerstore.Peerstore)
} 