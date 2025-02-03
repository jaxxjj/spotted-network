package registry

import (
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) disconnectPeer(targetPeer peer.ID) error {
	log.Printf("[Node] Disconnecting peer %s", targetPeer.String())
	
	// 1. disconnect all connection
	if err := n.closeConnection(targetPeer); err != nil {
		log.Printf("[Node] Error disconnecting from peer %s: %v", targetPeer, err)
	}
	
	// 2. clean up resources
	n.cleanupPeerResources(targetPeer)
	n.sp.broadcastStateUpdate(nil)
	log.Printf("[Node] Successfully disconnected peer %s", targetPeer.String())
	return nil
}


// closeConnection disconnect p2p connection
func (n *Node) closeConnection(peer peer.ID) error {
	// check if the connection exists
	if n.host.Network().Connectedness(peer) != network.Connected {
		return nil
	}
	
	// close all connections to the peer
	if err := n.host.Network().ClosePeer(peer); err != nil {
		log.Printf("[Node] Error closing connection: %v", err)
	}
	
	return nil
}

// cleanupPeerResources clean up resources
func (n *Node) cleanupPeerResources(peer peer.ID) {
	n.RemoveOperator(peer)	
	// remove from peerstore
	n.host.Peerstore().ClearAddrs(peer)
	
	log.Printf("[Node] Cleaned up resources for peer %s", peer.String())
}
