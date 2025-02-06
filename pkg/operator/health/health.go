package operator

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
func newHealthChecker(ctx context.Context, node *Node, pingService PingService) (*HealthChecker, error) {
	if node == nil {
		log.Fatal("node is nil")
	}
	if pingService == nil {
		log.Fatal("pingService is nil")
	}
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
			hc.printConnectedPeers()
		}
	}
}

// checkOperators checks the health of all connected operators
func (hc *HealthChecker) checkOperators(ctx context.Context) {
	operators := hc.node.getActivePeerIDs()
	for _, id := range operators {
		if id == hc.node.host.ID() {
			continue
		}
		if err := hc.pingOperator(ctx, id); err != nil {
			log.Printf("[Health] Operator %s failed health check: %v", id, err)
			hc.node.disconnectPeer(id)
			hc.node.sp.verifyStateWithRegistry(ctx)
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

	log.Printf("[Health] Successfully pinged operator %s (RTT: %v)", p, result.RTT)
	return nil
}

func (hc *HealthChecker) printConnectedPeers() {
	peers := hc.node.host.Network().Peers()
	log.Printf("[Node] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := hc.node.host.Network().Peerstore().Addrs(peer)
		log.Printf("[Node] - Peer %s at %v", peer.String(), addrs)
	}
}
