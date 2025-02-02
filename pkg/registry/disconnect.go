package registry

import (
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) DisconnectPeer(targetPeer peer.ID) error {
	log.Printf("[Node] Disconnecting peer %s", targetPeer.String())

	// 1. close streams
	if err := n.closeStreams(targetPeer); err != nil {
		log.Printf("[Node] Error closing streams with peer %s: %v", targetPeer, err)
	}

	// 2. disconnect p2p connection
	if err := n.closeConnection(targetPeer); err != nil {
		log.Printf("[Node] Error disconnecting from peer %s: %v", targetPeer, err)
	}
	
	// 3. clean up resources
	n.cleanupPeerResources(targetPeer)
	
	log.Printf("[Node] Successfully disconnected peer %s", targetPeer.String())
	return nil
}

// closeStreams close all streams
func (n *Node) closeStreams(peer peer.ID) error {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()

	// get all streams related to the peer
	if info, exists := n.activeOperators.active[peer]; exists {
		// close registry stream
		if info.RegistryStream != nil {
			if err := info.RegistryStream.Close(); err != nil {
				log.Printf("[Node] Error closing registry stream: %v", err)
			}
			info.RegistryStream.Reset()
		}

		// close state sync stream
		if info.StateSyncStream != nil {
			if err := info.StateSyncStream.Close(); err != nil {
				log.Printf("[Node] Error closing state sync stream: %v", err)
			}
			info.StateSyncStream.Reset()
		}
	}
	
	return nil
}

// closeConnection disconnect p2p connection
func (n *Node) closeConnection(peer peer.ID) error {
	// check if the connection exists
	if n.host.Network().Connectedness(peer) != network.Connected {
		return nil
	}
	
	// close all connections to the peer
	conns := n.host.Network().ConnsToPeer(peer)
	for _, conn := range conns {
		if err := conn.Close(); err != nil {
			log.Printf("[Node] Error closing connection: %v", err)
		}
	}
	
	return nil
}

// cleanupPeerResources clean up resources
func (n *Node) cleanupPeerResources(peer peer.ID) {
	// remove from activeOperators
	n.activeOperators.mu.Lock()
	delete(n.activeOperators.active, peer)
	n.activeOperators.mu.Unlock()

	// remove from peerstore
	n.host.Peerstore().ClearAddrs(peer)
	
	log.Printf("[Node] Cleaned up resources for peer %s", peer.String())
}
