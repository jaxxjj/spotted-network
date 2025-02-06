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

type Node interface {
	DisconnectPeer(peer.ID) error
	GetConnectedPeers() []peer.ID
	PrintConnectedPeers()
}

type HealthChecker struct {
	node           Node
	pingService    PingService
}

// NewHealthChecker creates and starts a new health checker
func NewHealthChecker(ctx context.Context, node Node, pingService PingService) (*HealthChecker, error) {
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
			hc.checkPeers(ctx)
			hc.node.PrintConnectedPeers()
		}
	}
}

// checkPeers checks the health of all connected peers
func (hc *HealthChecker) checkPeers(ctx context.Context) {
	peers := hc.node.GetConnectedPeers()
	for _, peerId := range peers {
		if err := hc.pingPeer(ctx, peerId); err != nil {
			log.Printf("[Health] Peer %s failed health check: %v", peerId, err)
			hc.node.DisconnectPeer(peerId)
		}
	}
}

// pingPeer pings a specific peer
func (hc *HealthChecker) pingPeer(ctx context.Context, p peer.ID) error {
	// Add ping timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := <-hc.pingService.Ping(ctx, p)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("[Health] Successfully pinged peer %s (RTT: %v)", p, result.RTT)
	return nil
}

