package node

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
)

// Node defines the interface for an operator node
type Node interface {
	// Stop stops the node and cleans up resources
	Stop(ctx context.Context) error

	// PrintConnectedPeers prints information about connected peers
	PrintConnectedPeers()

	// PrintPeerInfo prints the node's peer information
	PrintPeerInfo()

	// GetConnectedPeers returns a list of connected peer IDs
	GetConnectedPeers() []peer.ID

	// DisconnectPeer disconnects and removes a peer
	DisconnectPeer(targetPeer peer.ID) error
}
