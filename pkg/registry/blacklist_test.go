package registry

import (
	"time"

	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
	"github.com/libp2p/go-libp2p/core/peer"
)

// BlacklistState 用于序列化/反序列化黑名单状态
type BlacklistState struct {
	Blacklist []struct {
		PeerID    string     `json:"peer_id"`
		IP        string     `json:"ip"`
		Reason    string     `json:"reason"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
	} `json:"blacklist"`
}


func (s *RegistryTestSuite) TestBlacklistPeer() {
	tests := []struct {
		name      string
		peerId    string
		params    *BlacklistParams
		mockSetup func()
		wantErr   bool
	}{
		{
			name:   "successful blacklist",
			peerId: "12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr",
			params: &BlacklistParams{
				IP:        "192.168.1.1",
				Reason:    "test reason",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			mockSetup: func() {
				id, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)
				s.mockPubSub.On("BlacklistPeer", id).Once()
			},
			wantErr: false,
		},
		{
			name:   "blacklist with no expiry",
			peerId: "12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk",
			params: &BlacklistParams{
				IP:     "192.168.1.2",
				Reason: "permanent ban",
			},
			mockSetup: func() {
				id, err := peer.Decode("12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk")
				s.Require().NoError(err)
				s.mockPubSub.On("BlacklistPeer", id).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 重置 mocks
			s.mockPubSub.ExpectedCalls = nil
			
			// 设置 mock 行为
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// 创建测试节点
			node := &Node{
				blacklistQuerier: s.blacklistQuerier,
				pubsub:          s.mockPubSub,
			}

			// 执行测试
			id, err := peer.Decode(tt.peerId)
			s.Require().NoError(err)
			err = node.BlacklistPeer(s.ctx, id, tt.params)
			
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
			
			// 验证所有 mock 调用
			s.mockPubSub.AssertExpectations(s.T())
		})
	}
}

func (s *RegistryTestSuite) TestUnblacklistPeer() {
	tests := []struct {
		name      string
		peerId    string
		ip        string
		mockSetup func()
		wantErr   bool
	}{
		{
			name:   "successful unblacklist",
			peerId: "12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr",
			ip:     "192.168.1.1",
			mockSetup: func() {},
			wantErr: false,
		},
		{
			name:   "unblacklist error",
			peerId: "12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk",
			ip:     "192.168.1.2",
			mockSetup: func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 设置 mock 行为
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// 创建测试节点
			node := &Node{
				blacklistQuerier: s.blacklistQuerier,
			}

			// 执行测试
			id, err := peer.Decode(tt.peerId)
			s.Require().NoError(err)
			err = node.UnblacklistPeer(s.ctx, id, tt.ip)
			
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *RegistryTestSuite) TestIsBlacklisted() {
	tests := []struct {
		name      string
		peerId    string
		ip        string
		mockSetup func()
		want      bool
		wantErr   bool
	}{
		{
			name:   "peer is blacklisted",
			peerId: "12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr",
			ip:     "192.168.1.1",
			mockSetup: func() {
				id, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)
				s.mockPubSub.On("BlacklistPeer", id).Once()
				expiresAt := time.Now().Add(24 * time.Hour)
				reason := "test reason"
				_, err = s.blacklistQuerier.BlockNode(s.ctx, blacklist.BlockNodeParams{
					PeerID:    id.String(),
					Ip:        "192.168.1.1",
					Reason:    &reason,
					ExpiresAt: &expiresAt,
				}, &blacklist.IsBlockedParams{
					PeerID: id.String(),
					Ip:     "192.168.1.1",
				})
				s.Require().NoError(err)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "peer is not blacklisted",
			peerId: "12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk",
			ip:     "192.168.1.2",
			mockSetup: func() {},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 设置 mock 行为
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			// 创建测试节点
			node := &Node{
				blacklistQuerier: s.blacklistQuerier,
			}

			// 执行测试
			id, err := peer.Decode(tt.peerId)
			s.Require().NoError(err)
			isBlacklisted, err := node.IsBlacklisted(s.ctx, id, tt.ip)
			
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.want, isBlacklisted)
			}
		})
	}
}

func (s *RegistryTestSuite) TestBlacklistStateManagement() {
	tests := []struct {
		name          string
		initialState  string
		operations    func(node *Node)
		wantErr      bool
		goldenFile   string
	}{
		{
			name:         "expire_and_remove_blacklist_entries",
			initialState: "TestBlacklistStateManagement.blacklist.input.json",
			operations: func(node *Node) {
				id, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)
				err = node.UnblacklistPeer(s.ctx, id, "192.168.1.1")
				s.NoError(err)
			},
			wantErr:    false,
			goldenFile: "blacklist_after_removal",
		},
		{
			name:         "load_and_verify_blacklist_state",
			initialState: "TestBlacklistStateManagement.blacklist.input.json",
			operations: func(node *Node) {
				// 验证已加载的状态
				id, err := peer.Decode("12D3KooWJHw6WnvPiUPSpJNaRBWn4XpDgRSR1UvZNhJDkRCfehMr")
				s.Require().NoError(err)
				isBlocked, err := node.IsBlacklisted(s.ctx, id, "192.168.1.1")
				s.NoError(err)
				s.True(isBlocked)

				// 添加新的黑名单条目
				id3, err := peer.Decode("12D3KooWRHVDHQBHN9qFP8UfvUd8azdHxvRHqhYMgwMEjxwj7YGk")
				s.Require().NoError(err)
				s.mockPubSub.On("BlacklistPeer", id3).Once()
				err = node.BlacklistPeer(s.ctx, id3, &BlacklistParams{
					IP:     "192.168.1.3",
					Reason: "test reason 3",
				})
				s.NoError(err)
			},
			wantErr:    false,
			goldenFile: "blacklist_final",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// 清理数据库状态
			_, err := s.GetPool().WConn().WExec(s.ctx, "blacklist.truncate", "TRUNCATE TABLE blacklist")
			s.Require().NoError(err)

			// 重置 mocks
			s.mockPubSub.ExpectedCalls = nil

			// 创建测试节点
			node := &Node{
				blacklistQuerier: s.blacklistQuerier,
				pubsub:          s.mockPubSub,
			}

			// 加载初始状态
			serde := BlacklistTableSerde{blacklistQuerier: s.blacklistQuerier}
			s.LoadState(tt.initialState, serde)

			// 执行操作
			if tt.operations != nil {
				tt.operations(node)
			}

			s.Golden(tt.goldenFile, serde)
		})
	}
} 