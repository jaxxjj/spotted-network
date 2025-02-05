package registry

import (
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/mock"
)

func (s *RegistryTestSuite) TestHandleRegister() {
	// 生成一个固定的私钥用于测试
	privateKey, err := crypto.HexToECDSA("1234567890123456789012345678901234567890123456789012345678901234")
	s.Require().NoError(err)
	
	// 从私钥获取对应的地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	s.Require().True(ok)
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	
	tests := []struct {
		name         string
		initialState string
		setupMocks   func()
		request      *pb.RegisterMessage
		wantSuccess  bool
		wantMessage  string
		goldenFile   string
	}{
		{
			name:         "valid_registration",
			initialState: "TestHandleRegister.operators.input.json",
			setupMocks: func() {
				// 设置 operator 为 active 状态
				op := &operators.Operators{
					Address: address.Hex(), // 使用实际生成的地址
					Status:  "active",
				}
				s.mockOpQuerier.On("GetOperatorByAddress", mock.Anything, address.Hex()).Return(op, nil)
			},
			request: &pb.RegisterMessage{
				Address:   address.Hex(), // 使用实际生成的地址
				Version:   CurrentVersion,
				Timestamp: uint64(time.Now().Unix()),
				Multiaddrs: []string{"/ip4/127.0.0.1/tcp/9000"},
			},
			wantSuccess: true,
			wantMessage: "auth successful",
			goldenFile:  "registry_valid_registration",
		},
		{
			name:         "invalid_version",
			initialState: "TestHandleRegister.operators.input.json",
			request: &pb.RegisterMessage{
				Address:   "0x1234567890123456789012345678901234567890",
				Version:   "0.0.1",
				Timestamp: uint64(time.Now().Unix()),
				Multiaddrs: []string{"/ip4/127.0.0.1/tcp/9000"},
			},
			wantSuccess: false,
			wantMessage: fmt.Sprintf("version mismatch: registry version %s, operator version 0.0.1", CurrentVersion),
			goldenFile:  "registry_invalid_version",
		},
		{
			name:         "expired_timestamp",
			initialState: "TestHandleRegister.operators.input.json",
			request: &pb.RegisterMessage{
				Address:   "0x1234567890123456789012345678901234567890",
				Version:   CurrentVersion,
				Timestamp: uint64(time.Now().Add(-AuthTimeout*2).Unix()),
				Multiaddrs: []string{"/ip4/127.0.0.1/tcp/9000"},
			},
			wantSuccess: false,
			wantMessage: "auth request expired",
			goldenFile:  "registry_expired_timestamp",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			// 加载初始数据
			serde := &OperatorsTableSerde{opQuerier: s.opQuerier}
			s.LoadState(tt.initialState, serde)

			// 设置 mocks
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			// 使用固定的私钥签名
			message := crypto.Keccak256(
				[]byte(tt.request.Address),
				utils.Uint64ToBytes(tt.request.Timestamp),
				[]byte(tt.request.Version),
			)
			signature, err := crypto.Sign(crypto.Keccak256Hash(message).Bytes(), privateKey)
			s.Require().NoError(err)
			tt.request.Signature = fmt.Sprintf("0x%x", signature)

			// 创建 handler
			handler := NewRegistryHandler(&RegistryHandlerConfig{
				node:      s.node,
				opQuerier: s.opQuerier,
			})

			// 验证请求
			success, msg := handler.verifyAuthRequest(s.ctx, tt.request)

			// 验证结果
			s.Equal(tt.wantSuccess, success)
			s.Equal(tt.wantMessage, msg)

			// 验证数据库状态
			s.Golden(tt.goldenFile, serde)
		})
	}
}

func (s *RegistryTestSuite) TestHandleDisconnect() {
	tests := []struct {
		name         string
		initialState string
		setupState   func(node *Node)
		peerID      string
		goldenFile  string
	}{
		{
			name:         "disconnect_active_operator",
			initialState: "TestHandleDisconnect.operators.input.json",
			setupState: func(node *Node) {
				// 初始化 activeOperators
				node.activeOperators = ActiveOperatorPeers{
					active: make(map[peer.ID]*OperatorPeerInfo),
					mu:     sync.RWMutex{},
				}
				
				// 添加一个活跃的 operator
				id, _ := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				addr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9000")
				node.UpdateOperatorState(id, &OperatorPeerInfo{
					Address:    "0x1234567890123456789012345678901234567890",
					PeerID:     id,
					Multiaddrs: []multiaddr.Multiaddr{addr},
					LastSeen:   time.Now(),
				})
			},
			peerID:     "12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr",
			goldenFile: "registry_disconnect_active",
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

			// 设置初始状态
			if tt.setupState != nil {
				tt.setupState(s.node)
			}

			// 创建 handler
			handler := NewRegistryHandler(&RegistryHandlerConfig{
				node:      s.node,
				opQuerier: s.opQuerier,
			})

			// 执行断开连接
			id, err := peer.Decode(tt.peerID)
			s.Require().NoError(err)
			handler.handleDisconnect(id)

			// 验证状态
			s.Nil(s.node.GetOperatorState(id))
			s.Golden(tt.goldenFile, serde)
		})
	}
} 