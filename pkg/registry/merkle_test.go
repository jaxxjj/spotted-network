package registry

import (
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func (s *RegistryTestSuite) TestMerkleStateManagement() {
	tests := []struct {
		name         string
		initialState string
		operations   func(node *Node)
		wantErr     bool
		goldenFile  string
	}{
		{
			name:         "compute_root_with_initial_operators",
			initialState: "TestMerkleStateManagement.operators.input.json",
			operations: func(node *Node) {
				// Compute and verify merkle root
				root := node.computeActiveOperatorsRoot()				
				node.setActiveOperatorsRoot(root)
				storedRoot := node.getActiveOperatorsRoot()
				s.Equal(root, storedRoot)
			},
			wantErr:    false,
			goldenFile: "merkle_initial_state",
		},
		{
			name:         "update_operators_and_recompute",
			initialState: "TestMerkleStateManagement.operators.input.json",
			operations: func(node *Node) {
				// Add new operator
				id, err := peer.Decode("12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk")
				s.Require().NoError(err)
				multiaddrs, err := utils.StringsToMultiaddrs([]string{"/ip4/127.0.0.1/tcp/9012"})		
				s.Require().NoError(err)
				params := operators.UpsertOperatorParams{
					Address:     "0x0000000000000000000000000000000000000005",
					SigningKey:  "0x0000000000000000000000000000000000000006",
					RegisteredAtBlockNumber: 1,
					RegisteredAtTimestamp: 1,
					ActiveEpoch: 1,
					Weight: pgtype.Numeric{
						Int:    big.NewInt(200),
						Exp:    0,
						Valid:  true,
					},
					ExitEpoch: 4294967295,
				}
				log.Printf("Created operator params: %+v", params)
			
				// Update operator in database
				_, err = node.opQuerier.UpsertOperator(s.ctx, params, &params.Address)
				s.Require().NoError(err)
				// Update active operators
				node.activeOperators.mu.Lock()
				node.activeOperators.active[id] = &OperatorPeerInfo{
					Address:    "0x0000000000000000000000000000000000000005",
					PeerID:     id,
					LastSeen:   time.Now(),
					Multiaddrs: multiaddrs,
				}
				node.activeOperators.mu.Unlock()

				// Recompute root
				root := node.computeActiveOperatorsRoot()
				s.NotNil(root)
			},
			wantErr:    false,
			goldenFile: "merkle_updated_state",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "operators.truncate", "TRUNCATE TABLE operators")
			s.Require().NoError(err)

			// 加载初始数据到数据库
			serde := &OperatorsTableSerde{opQuerier: s.opQuerier}
			s.LoadState(tt.initialState, serde)

			// 从数据库加载 operators
			operators, err := s.opQuerier.ListAllOperators(s.ctx)
			s.Require().NoError(err)
			validPeerIDs := []string{
				"12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr",
				"12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk",
			}
			// 创建 activeOperators - 使用指针！
			active := make(map[peer.ID]*OperatorPeerInfo)
			// 填充 activeOperators
			for index, op := range operators {
				// 为每个 operator 生成测试用的 PeerID
				peerIDStr := validPeerIDs[index%len(validPeerIDs)]
				peerID, err := peer.Decode(peerIDStr)
				s.Require().NoError(err)

				// 生成测试用的 multiaddrs
				addrStr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 9000+index)
				addr, err := multiaddr.NewMultiaddr(addrStr)
				s.Require().NoError(err)

				// 创建 OperatorPeerInfo
				active[peerID] = &OperatorPeerInfo{
					Address:    op.Address,
					PeerID:    peerID,
					Multiaddrs: []multiaddr.Multiaddr{addr},
					LastSeen:  time.Now(),
				}
			}

			// 创建 node - 直接使用指针
			node := &Node{
				opQuerier:       s.opQuerier,
				activeOperators: ActiveOperatorPeers{
					active:    active,
					stateRoot: nil,
					mu:       sync.RWMutex{},
				},  
			}
			log.Println("activeOperators", node.activeOperators.active)

			// 执行测试操作
			if tt.operations != nil {
				tt.operations(node)
			}

			s.Golden(tt.goldenFile, serde)
		})
	}
}