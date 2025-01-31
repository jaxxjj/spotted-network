package registry

import (
	"context"
	"log"
	"time"

	commonTypes "github.com/galxe/spotted-network/pkg/common/types"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(n.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.checkOperators(ctx)
		}
	}
}

func (n *Node) checkOperators(ctx context.Context) {
	n.operators.mu.Lock()
	defer n.operators.mu.Unlock()

	// Create a list of operators to remove
	var toRemove []peer.ID

	for id, state := range n.operators.active {
		// Ping the operator
		if err := n.PingPeer(ctx, id); err != nil {
			log.Printf("[Health] Operator %s is unreachable: %v", id, err)
			state.Status = string(commonTypes.OperatorStatusInactive)
			toRemove = append(toRemove, id)
		} else {
			state.LastSeen = time.Now()
			state.Status = string(commonTypes.OperatorStatusActive)
			log.Printf("[Health] Operator %s is healthy", id)
		}
	}

	// Remove unreachable operators
	for _, id := range toRemove {
		delete(n.operators.active, id)
		log.Printf("[Health] Removed unreachable operator %s from active operators", id)
	}
} 