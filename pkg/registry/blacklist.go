package registry

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"

	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
)

type BlacklistQuerier interface {
	BlockNode(ctx context.Context, arg blacklist.BlockNodeParams) (blacklist.Blacklist, error)
	IsBlocked(ctx context.Context, arg blacklist.IsBlockedParams) (bool, error)
	UnblockNode(ctx context.Context, arg blacklist.UnblockNodeParams) error
}
// BlacklistParams contains parameters for blacklisting a peer
type BlacklistParams struct {
	IP        string
	Reason    string
	ExpiresAt time.Time
}

// BlacklistPeer adds a peer to the blacklist
func (n *Node) BlacklistPeer(ctx context.Context, peerId peer.ID, params *BlacklistParams) error {
	var expiresAt *time.Time
	if !params.ExpiresAt.IsZero() {
		expiresAt = &params.ExpiresAt
	}

	invalidateParams := &blacklist.IsBlockedParams{
		PeerID: peerId.String(),
		Ip:     params.IP,
	}

	_, err := n.blacklistQuerier.BlockNode(ctx, blacklist.BlockNodeParams{
		PeerID:    peerId.String(),
		Ip:        params.IP,
		Reason:    &params.Reason,
		ExpiresAt: expiresAt,
	}, invalidateParams)
	if err != nil {
		log.Error().Err(err).
			Str("peer_id", peerId.String()).
			Str("ip", params.IP).
			Msg("failed to add peer to blacklist")
		return err
	}

	// 调用pubsub的blacklist
	n.pubsub.BlacklistPeer(peerId)

	log.Info().
		Str("peer_id", peerId.String()).
		Str("ip", params.IP).
		Str("reason", params.Reason).
		Time("expires_at", params.ExpiresAt).
		Msg("peer added to blacklist")

	return nil
}

// UnblacklistPeer removes a peer from the blacklist
func (n *Node) UnblacklistPeer(ctx context.Context, peerId peer.ID, ip string) error {
	invalidateParams := &blacklist.IsBlockedParams{
		PeerID: peerId.String(),
		Ip:     ip,
	}

	err := n.blacklistQuerier.UnblockNode(ctx, blacklist.UnblockNodeParams{
		PeerID: peerId.String(),
		Ip:     ip,
	}, invalidateParams)
	if err != nil {
		log.Error().Err(err).
			Str("peer_id", peerId.String()).
			Str("ip", ip).
			Msg("failed to remove peer from blacklist")
		return err
	}

	log.Info().
		Str("peer_id", peerId.String()).
		Str("ip", ip).
		Msg("peer removed from blacklist")

	return nil
}

// IsBlacklisted checks if a peer is blacklisted
func (n *Node) IsBlacklisted(ctx context.Context, peerId peer.ID, ip string) (bool, error) {
	isBlocked, err := n.blacklistQuerier.IsBlocked(ctx, blacklist.IsBlockedParams{
		PeerID: peerId.String(),
		Ip:     ip,
	})
	if err != nil {
		return false, err
	}

	return *isBlocked, nil
}

