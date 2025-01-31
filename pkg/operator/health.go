package operator

import (
	"context"
	"log"
	"time"
)

// healthCheck periodically checks the health of connected operators
func (n *Node) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Get current active operators
		operators := n.GetActiveOperators()
		
		for _, op := range operators {
			// Skip if no peer ID (not connected yet)
			if op.PeerID == "" {
				continue
			}

			// Create context with timeout for health check
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			// Ping the operator
			err := n.PingPeer(ctx, op.PeerID)
			
			if err != nil {
				log.Printf("[Health] Failed to ping operator %s (peer ID: %s): %v", 
					op.Address, op.PeerID, err)
				
				// Update operator status to inactive
				op.Status = "inactive"
				op.LastSeen = time.Now()
				n.UpdateOperatorState(op.Address, op.PeerID, op)
				
			} else {
				// Update last seen time on successful ping
				op.LastSeen = time.Now()
				op.Status = "active" 
				n.UpdateOperatorState(op.Address, op.PeerID, op)
			}

			cancel()
		}

		// Clean up stale operators
		n.cleanupStaleOperators()
	}
}

// cleanupStaleOperators removes operators that haven't been seen for too long
func (n *Node) cleanupStaleOperators() {
	staleThreshold := 5 * time.Minute
	
	n.operators.mu.RLock()
	now := time.Now()
	
	for addr, op := range n.operators.byAddress {
		if now.Sub(op.LastSeen) > staleThreshold {
			log.Printf("[Health] Removing stale operator %s (peer ID: %s)", 
				addr, op.PeerID)
			
			// Remove operator from maps
			n.RemoveOperator(addr, op.PeerID)
		}
	}
	n.operators.mu.RUnlock()
} 