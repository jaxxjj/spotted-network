package registry

import (
	"context"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/stretchr/testify/mock"
)

// MockPingService implements PingService for testing
type MockPingService struct {
	mock.Mock
}

func (m *MockPingService) Ping(ctx context.Context, p peer.ID) <-chan ping.Result {
	args := m.Called(ctx, p)
	return args.Get(0).(<-chan ping.Result)
}

func (s *RegistryTestSuite) TestHealthChecker() {
	peer1 := peer.ID("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
	peer2 := peer.ID("12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk")

	tests := []struct {
		name         string
		description  string
		peers        []peer.ID
		mockSetup    func(*MockPingService, *MockNode)
		operations   func(*HealthChecker)
		wantErr      bool
	}{
		{
			name:        "successful_ping",
			description: "All peers respond successfully to ping",
			peers:       []peer.ID{peer1, peer2},
			mockSetup: func(mockPing *MockPingService, mockNode *MockNode) {
				// Mock active peers - called multiple times
				mockNode.On("getActivePeerIDs").Return([]peer.ID{peer1, peer2}).Maybe()

				// Mock operator states
				state1 := &OperatorPeerInfo{LastSeen: time.Now()}
				state2 := &OperatorPeerInfo{LastSeen: time.Now()}
				mockNode.On("GetOperatorState", peer1).Return(state1).Maybe()
				mockNode.On("GetOperatorState", peer2).Return(state2).Maybe()
				mockNode.On("UpdateOperatorState", peer1, state1).Return().Maybe()
				mockNode.On("UpdateOperatorState", peer2, state2).Return().Maybe()

				// Mock successful pings
				for _, p := range []peer.ID{peer1, peer2} {
					resultCh := make(chan ping.Result, 1)
					resultCh <- ping.Result{RTT: time.Millisecond * 100}
					mockPing.On("Ping", mock.Anything, p).Return((<-chan ping.Result)(resultCh)).Maybe()
				}
			},
			operations: func(hc *HealthChecker) {
				// 等待足够长的时间以确保至少一次完整的健康检查
				time.Sleep(2 * time.Second)
			},
		},
		{
			name:        "failed_ping",
			description: "One peer fails to respond to ping",
			peers:       []peer.ID{peer1, peer2},
			mockSetup: func(mockPing *MockPingService, mockNode *MockNode) {
				// Mock active peers - called multiple times
				mockNode.On("getActivePeerIDs").Return([]peer.ID{peer1, peer2}).Maybe()

				// Mock operator states
				state1 := &OperatorPeerInfo{LastSeen: time.Now()}
				mockNode.On("GetOperatorState", peer1).Return(state1).Maybe()
				mockNode.On("GetOperatorState", peer2).Return(nil).Maybe()
				mockNode.On("UpdateOperatorState", peer1, state1).Return().Maybe()

				// Mock ping responses
				successCh := make(chan ping.Result, 1)
				successCh <- ping.Result{RTT: time.Millisecond * 100}
				mockPing.On("Ping", mock.Anything, peer1).Return((<-chan ping.Result)(successCh)).Maybe()

				failCh := make(chan ping.Result, 1)
				failCh <- ping.Result{Error: errors.New("ping failed")}
				mockPing.On("Ping", mock.Anything, peer2).Return((<-chan ping.Result)(failCh)).Maybe()

				// Mock disconnect for failed peer
				mockNode.On("disconnectPeer", peer2).Return(nil).Maybe()
			},
			operations: func(hc *HealthChecker) {
				// 等待足够长的时间以确保至少一次完整的健康检查
				time.Sleep(2 * time.Second)
			},
		},
		{
			name:        "ping_timeout",
			description: "Ping times out for one peer",
			peers:       []peer.ID{peer1},
			mockSetup: func(mockPing *MockPingService, mockNode *MockNode) {
				// Mock active peers - called multiple times
				mockNode.On("getActivePeerIDs").Return([]peer.ID{peer1}).Maybe()

				// Mock operator state
				mockNode.On("GetOperatorState", peer1).Return(nil).Maybe()

				// Mock ping timeout
				timeoutCh := make(chan ping.Result, 1)
				timeoutCh <- ping.Result{Error: context.DeadlineExceeded}
				mockPing.On("Ping", mock.Anything, peer1).Return((<-chan ping.Result)(timeoutCh)).Maybe()

				// Mock disconnect for timed out peer
				mockNode.On("disconnectPeer", peer1).Return(nil).Maybe()
			},
			operations: func(hc *HealthChecker) {
				// 等待足够长的时间以确保至少一次完整的健康检查
				time.Sleep(2 * time.Second)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Create mocks
			mockPing := new(MockPingService)
			mockNode := new(MockNode)

			// Setup mocks
			if tt.mockSetup != nil {
				tt.mockSetup(mockPing, mockNode)
			}

			// Create health checker
			ctx := context.Background()
			hc, err := newHealthChecker(ctx, mockNode, mockPing)
			s.Require().NoError(err)

			// Execute operations
			if tt.operations != nil {
				tt.operations(hc)
			}
			// Verify mock expectations
			mockPing.AssertExpectations(s.T())
			mockNode.AssertExpectations(s.T())
		})
	}
} 