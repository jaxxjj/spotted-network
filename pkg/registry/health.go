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
	n.operatorsInfoMu.Lock()
	defer n.operatorsInfoMu.Unlock()

	// Create a list of operators to remove
	var toRemove []peer.ID

	for id, info := range n.operatorsInfo {
		// Ping the operator
		if err := n.PingPeer(ctx, id); err != nil {
			log.Printf("[Health] Operator %s is unreachable: %v", id, err)
			info.Status = string(commonTypes.OperatorStatusInactive)
			toRemove = append(toRemove, id)
		} else {
			info.LastSeen = time.Now()
			info.Status = string(commonTypes.OperatorStatusActive)
			log.Printf("[Health] Operator %s is healthy", id)
		}
	}

	// Remove unreachable operators
	for _, id := range toRemove {
		delete(n.operatorsInfo, id)
		log.Printf("[Health] Removed unreachable operator %s from operatorsInfo", id)
	}
} 