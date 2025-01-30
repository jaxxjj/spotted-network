package registry

import (
	"context"
	"log"
	"time"

	commonTypes "github.com/galxe/spotted-network/pkg/common/types"
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

	for id, info := range n.operatorsInfo {
		// Ping the operator
		if err := n.PingPeer(ctx, id); err != nil {
			log.Printf("Operator %s is unreachable: %v\n", id, err)
			info.Status = string(commonTypes.OperatorStatusInactive)
		} else {
			info.LastSeen = time.Now()
			info.Status = string(commonTypes.OperatorStatusActive)
			log.Printf("Operator %s is healthy\n", id)
		}
	}
} 