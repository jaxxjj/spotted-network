package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func (s *RegistryTestSuite) TestBuildOperatorStates() {
	tests := []struct {
		name         string
		initialState string
		setupState   func(node *Node)
		goldenFile   string
	}{
		{
			name:         "build_active_operators",
			initialState: "TestBuildOperatorStates.operators.input.json",
			setupState: func(node *Node) {
				// 初始化 activeOperators
				node.activeOperators = ActiveOperatorPeers{
					active: make(map[peer.ID]*OperatorPeerInfo),
					mu:     sync.RWMutex{},
				}

				// 添加两个活跃的 operators
				peerIDs := []string{
					"12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr",
					"12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk",
				}
				
				for i, peerIDStr := range peerIDs {
					id, err := peer.Decode(peerIDStr)
					s.Require().NoError(err)
					
					// 正确的端口号转换
					addr, err := multiaddr.NewMultiaddr(
						fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 9000+i),
					)
					s.Require().NoError(err)
					
					node.UpdateOperatorState(id, &OperatorPeerInfo{
						Address:    "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
						PeerID:     id,
						Multiaddrs: []multiaddr.Multiaddr{addr},
						LastSeen:   time.Now(),
					})
				}
			},
			goldenFile: "state_active_operators",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "operators.truncate", "TRUNCATE TABLE operators")
			s.Require().NoError(err)

			// 加载初始数据
			serde := &OperatorsTableSerde{opQuerier: s.opQuerier}
			s.LoadState(tt.initialState, serde)

			// 创建节点
			node := &Node{
				opQuerier: s.opQuerier,
			}

			// 设置初始状态
			if tt.setupState != nil {
				tt.setupState(node)
			}

			// 构建状态
			states := node.buildOperatorPeerStates()
			s.NotEmpty(states)

			// 验证状态
			s.Golden(tt.goldenFile, serde)
		})
	}
} 