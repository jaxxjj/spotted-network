package registry

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)
const (
	HealthCheckInterval = 20 * time.Second
)
type PingService interface {
	Ping(ctx context.Context, p peer.ID) <-chan ping.Result
}

type HealthChecker struct {
	node           *Node
	pingService    PingService
}

// NewHealthChecker creates and starts a new health checker
func NewHealthChecker(ctx context.Context, node *Node, pingService PingService) (*HealthChecker, error) {
	hc := &HealthChecker{
		node:          node,
		pingService:   pingService,
	}
	
	// Start health check service
	go hc.start(ctx)
	log.Printf("[Health] Health check service started with interval %v", HealthCheckInterval)
	
	return hc, nil
}

// start is now private as it's called internally by NewHealthChecker
func (hc *HealthChecker) start(ctx context.Context) {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkOperators(ctx)
		}
	}
}

// checkOperators checks the health of all connected operators
func (hc *HealthChecker) checkOperators(ctx context.Context) {
	operators := hc.node.GetConnectedOperators()
	for _, id := range operators {
		if err := hc.pingOperator(ctx, id); err != nil {
			log.Printf("[Health] Operator %s failed health check: %v", id, err)
			hc.node.RemoveOperator(id)
		}
	}
}

// pingOperator pings a specific operator
func (hc *HealthChecker) pingOperator(ctx context.Context, p peer.ID) error {
	// Add ping timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := <-hc.pingService.Ping(ctx, p)
	if result.Error != nil {
		return result.Error
	}

	// Update last seen time for the operator
	if state := hc.node.GetOperatorState(p); state != nil {
		state.LastSeen = time.Now()
		hc.node.UpdateOperatorState(p, state)
	}

	log.Printf("[Health] Successfully pinged operator %s (RTT: %v)", p, result.RTT)
	return nil
}