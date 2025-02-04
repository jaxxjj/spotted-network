package operator

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerAction represents the action to take for a peer
type PeerAction int

const (
	PeerDisconnect PeerAction = iota // Need to disconnect
	PeerKeep                         // Keep connected
	PeerConnect                      // Need to connect

	maxConnectRetries = 3
	connectTimeout    = 10 * time.Second
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
		
		// Check if already connected
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
			// Create context with timeout
			connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
			defer cancel()

			// Try to connect with retries
			var lastErr error
			for retry := 0; retry < maxConnectRetries; retry++ {
				if retry > 0 {
					log.Printf("[Node] Retry %d connecting to peer %s", retry, peerID)
					time.Sleep(time.Second * time.Duration(retry)) // Exponential backoff
				}

				err := n.host.Connect(connectCtx, peer.AddrInfo{
					ID:    peerID,
					Addrs: action.state.Multiaddrs,
				})
				if err == nil {
					log.Printf("[Node] Successfully connected to peer %s", peerID)
					lastErr = nil
					break
				}
				lastErr = err
				log.Printf("[Node] Failed to connect to peer %s (attempt %d/%d): %v", 
					peerID, retry+1, maxConnectRetries, err)
			}
			if lastErr != nil {
				log.Printf("[Node] All connection attempts failed for peer %s: %v", peerID, lastErr)
			}
		}
	}

	// Log final connection status
	connectedPeers := n.host.Network().Peers()
	log.Printf("[Node] Connected to %d peers:", len(connectedPeers))
	for _, peer := range connectedPeers {
		conns := n.host.Network().ConnsToPeer(peer)
		if len(conns) > 0 {
			log.Printf("[Node] - Peer %s at %v", peer, conns[0].RemoteMultiaddr())
		}
	}
}

// disconnectPeer handles the disconnection of a peer and cleanup
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

// closeConnection disconnects p2p connection
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

// cleanupPeerResources cleans up resources
func (n *Node) cleanupPeerResources(peer peer.ID) {
	n.RemoveOperator(peer)	
	n.host.Peerstore().ClearAddrs(peer)
	log.Printf("[Node] Cleaned up resources for peer %s", peer.String())
}
