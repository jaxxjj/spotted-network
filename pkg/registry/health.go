package registry

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)
const (
	HealthCheckInterval = 1 * time.Second
)
type PingService interface {
	Ping(ctx context.Context, p peer.ID) <-chan ping.Result
}
type OperatorController interface {
	getActivePeerIDs() []peer.ID
	GetOperatorState(p peer.ID) *OperatorPeerInfo
	UpdateOperatorState(p peer.ID, info *OperatorPeerInfo)
	disconnectPeer(p peer.ID) error
}

type HealthChecker struct {
	node         OperatorController
	pingService  PingService
}

// NewHealthChecker creates and starts a new health checker
func newHealthChecker(ctx context.Context, node OperatorController, pingService PingService) (*HealthChecker, error) {
	if node == nil {
		log.Fatal("[Health] node is nil")
	}
	if pingService == nil {
		log.Fatal("[Health] pingService is nil")
	}

	hc := &HealthChecker{
		node:        node,
		pingService: pingService,
	}
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Health] Recovered from panic: %v", r)
			}
		}()
		
		if err := hc.start(ctx); err != nil && err != context.Canceled {
			log.Printf("[Health] Health checker stopped with error: %v", err)
		}
	}()
	
	log.Printf("[Health] Health check service started with interval %v", HealthCheckInterval)
	return hc, nil
}

func (hc *HealthChecker) start(ctx context.Context) error {
	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := hc.checkOperators(ctx); err != nil {
				log.Printf("[Health] Error checking operators: %v", err)
				// continue to check other nodes, don't return error
			}
		}
	}
}

// checkOperators checks the health of all connected operators
func (hc *HealthChecker) checkOperators(ctx context.Context) error {
	operators := hc.node.getActivePeerIDs()
	for _, id := range operators {
		if err := hc.pingOperator(ctx, id); err != nil {
			log.Printf("[Health] Operator %s failed health check: %v", id, err)
			if err := hc.node.disconnectPeer(id); err != nil {
				log.Printf("[Health] Error disconnecting operator %s: %v", id, err)
			}
			// continue to check other nodes, don't return error
		}
	}
	return nil
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