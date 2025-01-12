package operator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host *p2p.Host
	
	// Registry node info
	registryID peer.ID
	registryAddr string
	
	// Health check interval
	healthCheckInterval time.Duration
}

func NewNode(ctx context.Context, cfg *p2p.Config, registryAddr string) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		registryAddr:       registryAddr,
		healthCheckInterval: 5 * time.Second,
	}

	// Connect to registry node
	if err := node.connectToRegistry(ctx); err != nil {
		return nil, err
	}

	// Start health check
	go node.startHealthCheck(ctx)

	log.Printf("Operator Node started. %s\n", host.GetHostInfo())
	return node, nil
}

func (n *Node) connectToRegistry(ctx context.Context) error {
	log.Printf("Attempting to connect to Registry Node at address: %s\n", n.registryAddr)

	// Create multiaddr from the provided address
	addr, err := multiaddr.NewMultiaddr(n.registryAddr)
	if err != nil {
		return fmt.Errorf("invalid registry address: %s", n.registryAddr)
	}

	// Parse peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to parse peer info from address: %v", err)
	}

	// Connect to registry
	if err := n.host.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to registry: %v", err)
	}

	// Save the registry ID
	n.registryID = peerInfo.ID

	log.Printf("Successfully connected to Registry Node with ID: %s\n", n.registryID)
	return nil
}

func (n *Node) startHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(n.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Ping registry node with its ID
			if err := n.host.PingPeer(ctx, n.registryID); err != nil {
				log.Printf("Registry Node is unreachable: %v\n", err)
			}
		}
	}
}

func (n *Node) Stop() error {
	return n.host.Close()
} 