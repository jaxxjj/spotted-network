package node

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *node) DisconnectPeer(targetPeer peer.ID) error {
	log.Printf("[Node] Disconnecting peer %s", targetPeer.String())

	// 1. disconnect all connection
	if err := n.closeConnection(targetPeer); err != nil {
		log.Printf("[Node] Error disconnecting from peer %s: %v", targetPeer, err)
	}

	// 2. clean up resources
	n.cleanupPeerResources(targetPeer)
	log.Printf("[Node] Successfully disconnected peer %s", targetPeer.String())
	return nil
}

func (n *node) closeConnection(peer peer.ID) error {
	if n.host.Network().Connectedness(peer) != network.Connected {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- n.host.Network().ClosePeer(peer)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("[Node] Error closing connection to peer %s: %v", peer, err)
			return err
		}
	case <-ctx.Done():
		return fmt.Errorf("timeout closing connection to peer %s", peer)
	}

	n.cleanupPeerResources(peer)

	if n.host.Network().Connectedness(peer) == network.Connected {
		return fmt.Errorf("failed to close all connections to peer %s", peer)
	}

	return nil
}

// cleanupPeerResources clean up resources
func (n *node) cleanupPeerResources(peer peer.ID) {
	peerstore := n.host.Network().Peerstore()
	peerstore.ClearAddrs(peer)
	peerstore.RemovePeer(peer)
	log.Printf("[Node] Cleaned up resources for peer %s", peer.String())
}
