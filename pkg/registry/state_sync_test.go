package registry

import (
	"bytes"
	"context"
	"sync"
	"time"

	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// Mock PubSubService
type MockPubSubService struct {
	mock.Mock
}

func (m *MockPubSubService) Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error) {
	args := m.Called(topic)
	return args.Get(0).(*pubsub.Topic), args.Error(1)
}

func (m *MockPubSubService) BlacklistPeer(p peer.ID) {
	m.Called(p)
}

// Mock PubsubTopic
type MockPubsubTopic struct {
	mock.Mock
}

func (m *MockPubsubTopic) Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

// Mock Stream
type MockStream struct {
	mock.Mock
	network.Stream
	buf bytes.Buffer
}

func (m *MockStream) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockStream) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStream) Reset() error {
	args := m.Called()
	return args.Error(0)
}

func (s *RegistryTestSuite) TestStateSyncProcessor() {
	tests := []struct {
		name         string
		initialState string
		setupMocks   func(node *Node, mockPubSub *MockPubSubService, mockTopic *MockPubsubTopic)
		operation    func(sp *StateSyncProcessor)
		wantErr      bool
		goldenFile   string
	}{
		{
			name: "successful_state_sync_creation",
			setupMocks: func(node *Node, mockPubSub *MockPubSubService, mockTopic *MockPubsubTopic) {
				mockPubSub.On("Join", StateSyncTopic).Return(mockTopic, nil)
			},
			operation: func(sp *StateSyncProcessor) {
				// 只验证创建是否成功
			},
			wantErr: false,
		},
		{
			name:         "broadcast_state_update",
			initialState: "TestStateSyncProcessor.operators.input.json",
			setupMocks: func(node *Node, mockPubSub *MockPubSubService, mockTopic *MockPubsubTopic) {
				mockPubSub.On("Join", StateSyncTopic).Return(mockTopic, nil)
				mockTopic.On("Publish", mock.Anything, mock.Anything).Return(nil)
				
				// 设置活跃操作者
				node.activeOperators = ActiveOperatorPeers{
					active: make(map[peer.ID]*OperatorPeerInfo),
					mu:     sync.RWMutex{},
				}

				id, _ := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000")
				
				node.UpdateOperatorState(id, &OperatorPeerInfo{
					Address:    "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
					PeerID:     id,
					Multiaddrs: []multiaddr.Multiaddr{addr},
					LastSeen:   time.Now(),
				})
			},
			operation: func(sp *StateSyncProcessor) {
				epoch := uint32(1)
				err := sp.broadcastStateUpdate(&epoch)
				s.NoError(err)
			},
			wantErr:    false,
			goldenFile: "state_sync_broadcast",
		},
		{
			name:         "handle_state_verify_match",
			initialState: "TestStateSyncProcessor.operators.input.json",
			setupMocks: func(node *Node, mockPubSub *MockPubSubService, mockTopic *MockPubsubTopic) {
				mockPubSub.On("Join", StateSyncTopic).Return(mockTopic, nil)
				
				// 设置活跃操作者
				node.activeOperators = ActiveOperatorPeers{
					active: make(map[peer.ID]*OperatorPeerInfo),
					mu:     sync.RWMutex{},
				}

				// 添加一个活跃操作者
				id, _ := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000")
				
				node.UpdateOperatorState(id, &OperatorPeerInfo{
					Address:    "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
					PeerID:     id,
					Multiaddrs: []multiaddr.Multiaddr{addr},
					LastSeen:   time.Now(),
				})
			},
			operation: func(sp *StateSyncProcessor) {
				// 获取实际的状态根
				actualRoot := sp.node.getActiveOperatorsRoot()
				
				mockStream := &MockStream{}
				
				// 使用实际的状态根构造验证消息
				msg := &pb.StateVerifyMessage{
					StateRoot: actualRoot,
				}
				data, _ := proto.Marshal(msg)
				mockStream.On("Read", mock.Anything).Return(len(data), nil)
				mockStream.buf.Write(data)
				
				// 模拟发送响应
				mockStream.On("Write", mock.Anything).Return(0, nil)
				mockStream.On("Close").Return(nil)
				
				sp.handleStateVerifyStream(mockStream)
				
				mockStream.AssertExpectations(s.T())
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "operators.truncate", "TRUNCATE TABLE operators")
			s.Require().NoError(err)

			// 创建mocks
			mockPubSub := new(MockPubSubService)
			mockTopic := new(MockPubsubTopic)
			
			// 创建节点
			node := &Node{
				opQuerier: s.opQuerier,
			}

			// 如果有初始状态，加载它
			if tt.initialState != "" {
				serde := &OperatorsTableSerde{opQuerier: s.opQuerier}
				s.LoadState(tt.initialState, serde)
			}

			// 设置mocks
			if tt.setupMocks != nil {
				tt.setupMocks(node, mockPubSub, mockTopic)
			}

			// 创建处理器
			sp, err := NewStateSyncProcessor(&PubsubConfig{
				node:   node,
				pubsub: mockPubSub,
			})

			if tt.wantErr {
				s.Error(err)
				return
			}
			s.NoError(err)
			s.NotNil(sp)

			// 执行测试操作
			if tt.operation != nil {
				tt.operation(sp)
			}

			// 验证golden file
			if tt.goldenFile != "" {
				serde := &OperatorsTableSerde{opQuerier: s.opQuerier}
				s.Golden(tt.goldenFile, serde)
			}

			// 验证mock调用
			mockPubSub.AssertExpectations(s.T())
			mockTopic.AssertExpectations(s.T())
		})
	}
}