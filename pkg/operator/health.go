package operator

import (
	"context"
	"log"
	"time"
)

func (n *Node) healthCheck() {
	// Ping registry node
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := <-n.pingService.Ping(ctx, n.registryID)
	if result.Error != nil {
		log.Printf("[HealthCheck] Failed to ping registry: %v\n", result.Error)
		return
	}
	log.Printf("[HealthCheck] Successfully pinged registry (RTT: %v)\n", result.RTT)

	// Ping all known operators
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()

	for operatorID, peerInfo := range n.knownOperators {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to connect if not connected
		if err := n.host.Connect(ctx, *peerInfo); err != nil {
			log.Printf("[HealthCheck] Failed to connect to operator %s: %v\n", operatorID, err)
			continue
		}

		// Ping the operator
		result := <-n.pingService.Ping(ctx, operatorID)
		if result.Error != nil {
			log.Printf("[HealthCheck] Failed to ping operator %s: %v\n", operatorID, result.Error)
			continue
		}
		log.Printf("[HealthCheck] Successfully pinged operator %s (RTT: %v)\n", operatorID, result.RTT)
	}
} 