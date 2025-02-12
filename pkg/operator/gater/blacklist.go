package gater

import (
	"context"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/libp2p/go-libp2p/core/peer"
)

// BlacklistRepo is a repository for managing the blacklist
type BlacklistRepo interface {
	IncrementViolationCount(ctx context.Context, arg blacklist.IncrementViolationCountParams, isBlocked *string) (*blacklist.Blacklist, error)
	UnblockNode(ctx context.Context, peerID string, isBlocked *string) error
	IsBlocked(ctx context.Context, peerID string) (*bool, error)
}

// ViolationParams contains parameters for incrementing violation count
type ViolationParams struct {
	PeerID         peer.ID
	ViolationCount int32
	ExpiresAt      *time.Time // optional expiration time
}

// IncrementViolationCount increments the violation count for a peer
func (g *connectionGater) IncrementViolationCount(ctx context.Context, params ViolationParams) error {
	dbParams := blacklist.IncrementViolationCountParams{
		PeerID:         params.PeerID.String(),
		ViolationCount: params.ViolationCount,
		ExpiresAt:      params.ExpiresAt, // could be nil
	}

	_, err := g.blacklistRepo.IncrementViolationCount(ctx, dbParams, nil)
	if err != nil {
		return err
	}

	return nil
}

// UnblockNode removes a peer from the blacklist
func (g *connectionGater) UnblockNode(ctx context.Context, peerID peer.ID) error {
	err := g.blacklistRepo.UnblockNode(ctx, peerID.String(), nil)
	if err != nil {
		return err
	}

	return nil
}
