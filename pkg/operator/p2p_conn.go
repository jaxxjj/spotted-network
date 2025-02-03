package operator

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerAction represents the action to take for a peer
type PeerAction int

const (
	PeerDisconnect PeerAction = iota // Need to disconnect
	PeerKeep                         // Keep connected
	PeerConnect                      // Need to connect
)

// UpdateActiveConnections updates connections based on new active operators
func (n *Node) UpdateActiveConnections(ctx context.Context, newOperators map[peer.ID]*OperatorState) {
	log.Printf("[Node] Updating active connections with %d new operators", len(newOperators))
	
	// Create action map for all peers
	peerActions := make(map[peer.ID]struct{
		action PeerAction
		state  *OperatorState
	})
	
	// Mark all current connections as disconnect (except registry)
	for _, peerID := range n.host.Network().Peers() {
		if peerID == n.registryID {
			continue
		}
		peerActions[peerID] = struct{
			action PeerAction
			state  *OperatorState
		}{action: PeerDisconnect}
	}
	
	// Process new operators
	for peerID, state := range newOperators {
		// Skip self and registry
		if peerID == n.host.ID() || peerID == n.registryID {
			continue
		}
		
		if n.host.Network().Connectedness(peerID) == network.Connected {
			// Already connected, keep it
			peerActions[peerID] = struct{
				action PeerAction
				state  *OperatorState
			}{action: PeerKeep}
			log.Printf("[Node] Keeping connection to active peer %s", peerID)
		} else {
			// Need to connect
			peerActions[peerID] = struct{
				action PeerAction
				state  *OperatorState
			}{
				action: PeerConnect,
				state:  state,
			}
			log.Printf("[Node] Marked new peer %s for connection", peerID)
		}
	}
	
	// Execute actions
	for peerID, action := range peerActions {
		switch action.action {
		case PeerDisconnect:
			log.Printf("[Node] Disconnecting inactive peer %s", peerID)
			if err := n.disconnectPeer(peerID); err != nil {
				log.Printf("[Node] Error disconnecting peer %s: %v", peerID, err)
			}
			
		case PeerConnect:
			log.Printf("[Node] Connecting to new peer %s", peerID)
			if err := n.host.Connect(ctx, peer.AddrInfo{
				ID:    peerID,
				Addrs: action.state.Multiaddrs,
			}); err != nil {
				log.Printf("[Node] Failed to connect to peer %s: %v", peerID, err)
			}
		}
	}

	log.Printf("[Node] Finished updating active connections")
}

func (n *Node) disconnectPeer(targetPeer peer.ID) error {
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
	n.host.Peerstore().ClearAddrs(peer)
	log.Printf("[Node] Cleaned up resources for peer %s", peer.String())
}
