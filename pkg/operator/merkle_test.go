package operator

import (
	"math/big"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func (s *OperatorTestSuite) TestMerkleStateManagement() {
	tests := []struct {
		name       string
		setupNode  func() *Node
		operations func(node *Node)
		wantErr    bool
	}{
		{
			name: "compute_root_with_initial_operators",
			setupNode: func() *Node {
				// 创建初始operator
				id, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)
				addr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000")
				s.Require().NoError(err)

				active := make(map[peer.ID]*OperatorState)
				active[id] = &OperatorState{
					PeerID:     id,
					Multiaddrs: []multiaddr.Multiaddr{addr},
					Address:    "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
					SigningKey: "0x2345678901234567890123456789012345678901",
					Weight:     big.NewInt(100),
				}

				return &Node{
					activeOperators: ActivePeerStates{
						active:    active,
						stateRoot: nil,
						mu:       sync.RWMutex{},
					},
				}
			},
			operations: func(node *Node) {
				// 计算并验证merkle root
				root := node.computeActiveOperatorsRoot()
				node.setCurrentStateRoot(root)
				storedRoot := node.getCurrentStateRoot()
				s.Equal(root, storedRoot)
			},
			wantErr: false,
		},
		{
			name: "update_operators_and_recompute",
			setupNode: func() *Node {
				// 创建初始operator
				id1, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)
				addr1, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000")
				s.Require().NoError(err)

				active := make(map[peer.ID]*OperatorState)
				active[id1] = &OperatorState{
					PeerID:     id1,
					Multiaddrs: []multiaddr.Multiaddr{addr1},
					Address:    "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
					SigningKey: "0x2345678901234567890123456789012345678901",
					Weight:     big.NewInt(100),
				}

				return &Node{
					activeOperators: ActivePeerStates{
						active:    active,
						stateRoot: nil,
						mu:       sync.RWMutex{},
					},
				}
			},
			operations: func(node *Node) {
				// 添加新的operator
				id2, err := peer.Decode("12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk")
				s.Require().NoError(err)
				addr2, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9012")
				s.Require().NoError(err)

				// 更新active operators
				node.activeOperators.mu.Lock()
				node.activeOperators.active[id2] = &OperatorState{
					PeerID:     id2,
					Multiaddrs: []multiaddr.Multiaddr{addr2},
					Address:    "0x0000000000000000000000000000000000000005",
					SigningKey: "0x0000000000000000000000000000000000000006",
					Weight:     big.NewInt(200),
				}
				node.activeOperators.mu.Unlock()

				// 重新计算root
				root := node.computeActiveOperatorsRoot()
				s.NotNil(root)

				// 验证root已更新
				node.setCurrentStateRoot(root)
				newRoot := node.getCurrentStateRoot()
				s.Equal(root, newRoot)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 创建node
			node := tt.setupNode()

			// 执行测试操作
			if tt.operations != nil {
				tt.operations(node)
			}
		})
	}
} 