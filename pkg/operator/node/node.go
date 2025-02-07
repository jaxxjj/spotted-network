package node

import (
	"context"
	"fmt"
	"log"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"

	"github.com/galxe/spotted-network/internal/metric"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
)

// Config contains all the dependencies needed by Node
type Config struct {
	Host             host.Host
	BlacklistRepo    *blacklist.Queries
	BootstrapPeers   []peer.AddrInfo
	RendezvousString string
}

// Node represents an operator node in the network
type Node struct {
	ctx              context.Context
	cancel           context.CancelFunc
	host             host.Host
	blacklistRepo    BlacklistRepo
	dht              *dht.IpfsDHT
	routingDiscovery *routing.RoutingDiscovery
}

// NewNode creates a new operator node with the given dependencies
func NewNode(ctx context.Context, cfg *Config) (*Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Host == nil {
		return nil, fmt.Errorf("host is required")
	}
	if cfg.BlacklistRepo == nil {
		return nil, fmt.Errorf("blacklist repo is required")
	}

	ctx, cancel := context.WithCancel(ctx)
	node := &Node{
		ctx:           ctx,
		cancel:        cancel,
		host:          cfg.Host,
		blacklistRepo: cfg.BlacklistRepo,
	}

	// Initialize DHT
	if err := node.initDHT(cfg.BootstrapPeers); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize DHT: %w", err)
	}

	// Start peer discovery
	if err := node.startPeerDiscovery(cfg.RendezvousString); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start peer discovery: %w", err)
	}

	return node, nil
}

func (n *Node) initDHT(bootstrapPeers []peer.AddrInfo) error {
	var err error
	n.dht, err = dht.New(n.ctx, n.host,
		dht.Mode(dht.ModeServer),
		dht.BootstrapPeers(bootstrapPeers...),
		dht.ProtocolPrefix("/spotted"),
	)
	if err != nil {
		metric.RecordError("dht_creation_failed")
		return fmt.Errorf("failed to create DHT: %w", err)
	}

	if err = n.dht.Bootstrap(n.ctx); err != nil {
		metric.RecordError("dht_bootstrap_failed")
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	n.routingDiscovery = routing.NewRoutingDiscovery(n.dht)
	return nil
}

func (n *Node) startPeerDiscovery(rendezvous string) error {
	// Advertise our presence
	ttl, err := n.routingDiscovery.Advertise(n.ctx, rendezvous)
	if err != nil {
		metric.RecordError("peer_advertise_failed")
		return fmt.Errorf("failed to advertise: %w", err)
	}
	log.Printf("Advertising this node with TTL: %s", ttl)

	// Start peer discovery
	go n.discoverPeers(rendezvous)

	// Start routing table maintenance
	go n.maintainRoutingTable()

	return nil
}

func (n *Node) discoverPeers(rendezvous string) {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			peerChan, err := n.routingDiscovery.FindPeers(n.ctx, rendezvous)
			if err != nil {
				log.Printf("Peer discovery error: %v", err)
				metric.RecordError("peer_discovery_failed")
				time.Sleep(time.Minute)
				continue
			}

			for peer := range peerChan {
				if peer.ID == n.host.ID() {
					continue // Skip self
				}

				if err := n.host.Connect(n.ctx, peer); err != nil {
					log.Printf("Failed to connect to peer %s: %v", peer.ID, err)
					metric.RecordError("peer_connection_failed")
					continue
				}

				metric.RecordRequest("peer_connected", peer.ID.String())
				log.Printf("Connected to peer: %s", peer.ID)
			}
		}
	}
}

func (n *Node) maintainRoutingTable() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			if err := n.dht.RefreshRoutingTable(); err != nil {
				log.Printf("Failed to refresh routing table: %v", err)
				metric.RecordError("routing_table_refresh_failed")
			}
		}
	}
}

func (n *Node) Stop(ctx context.Context) error {
	n.cancel()
	if n.dht != nil {
		if err := n.dht.Close(); err != nil {
			log.Printf("Failed to close DHT: %v", err)
		}
	}
	return n.host.Close()
}
