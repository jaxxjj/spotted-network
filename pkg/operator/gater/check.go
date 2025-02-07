package gater

import (
	"context"
	"fmt"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/libp2p/go-libp2p/core/peer"
)

// isAllowed checks if the peer is in the active operators map
func (g *ConnectionGater) isBlocked(peerID peer.ID) (bool, error) {
	blocked, err := g.blacklistRepo.IsBlocked(context.Background(), peerID.String())
	if err != nil {
		return false, err
	}
	if blocked == nil {
		return false, nil
	}
	return *blocked, nil
}

// isActiveOperator checks if the peer is an active operator
func (g *ConnectionGater) isActiveOperator(peerID peer.ID) (bool, error) {
	// convert peerID to p2p key
	p2pKey, err := utils.PeerIDToP2PKey(peerID)
	if err != nil {
		return false, fmt.Errorf("failed to convert peerID to p2p key: %w", err)
	}

	// check if the peer is an active operator
	isActive, err := g.operatorRepo.IsActiveOperator(context.Background(), p2pKey)
	if err != nil {
		return false, err
	}
	if isActive == nil {
		return false, nil
	}
	return *isActive, nil
}