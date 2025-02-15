package node

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"

	"github.com/galxe/spotted-network/internal/metric"
)

const (
	maxRetries        = 3
	retryInterval     = 5 * time.Second
	bootstrapInterval = 20 * time.Second
)

// Config contains all the dependencies needed by Node
type Config struct {
	Host             host.Host
	BootstrapPeers   []peer.AddrInfo
	RendezvousString string
}

// node represents an operator node in the network
type node struct {
	ctx              context.Context
	cancel           context.CancelFunc
	host             host.Host
	dht              *dht.IpfsDHT
	routingDiscovery *routing.RoutingDiscovery
}

// NewNode creates a new operator node with the given dependencies
func NewNode(ctx context.Context, cfg *Config) (Node, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Host == nil {
		return nil, fmt.Errorf("host is required")
	}
	if cfg.RendezvousString == "" {
		return nil, fmt.Errorf("rendezvous string is required")
	}

	ctx, cancel := context.WithCancel(ctx)
	node := &node{
		ctx:    ctx,
		cancel: cancel,
		host:   cfg.Host,
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

	// Print peer info before returning
	node.PrintPeerInfo()
	node.startPrintPeerInfo(ctx)
	return node, nil
}

func (n *node) Stop(ctx context.Context) error {
	n.cancel()
	if n.dht != nil {
		if err := n.dht.Close(); err != nil {
			log.Printf("Failed to close DHT: %v", err)
		}
	}
	return n.host.Close()
}

func (n *node) initDHT(bootstrapPeers []peer.AddrInfo) error {
	log.Printf("Initializing DHT with bootstrap peers: %+v", bootstrapPeers)

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

	log.Printf("Starting DHT bootstrap...")
	if err = n.dht.Bootstrap(n.ctx); err != nil {
		metric.RecordError("dht_bootstrap_failed")
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}
	log.Printf("DHT bootstrap completed")

	n.routingDiscovery = routing.NewRoutingDiscovery(n.dht)

	go n.maintainBootstrapConnections(bootstrapPeers)

	return nil
}

func (n *node) maintainBootstrapConnections(bootstrapPeers []peer.AddrInfo) {
	ticker := time.NewTicker(bootstrapInterval)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			for _, peer := range bootstrapPeers {
				if len(n.host.Network().ConnsToPeer(peer.ID)) > 0 {
					continue
				}

				log.Printf("Attempting to reconnect to bootstrap peer: %s", peer.ID)

				for retry := 0; retry < maxRetries; retry++ {
					err := n.connectWithTimeout(peer)
					if err == nil {
						log.Printf("Successfully reconnected to bootstrap peer: %s", peer.ID)
						break
					}

					if retry < maxRetries-1 {
						log.Printf("Failed to connect to bootstrap peer %s (attempt %d/%d): %v",
							peer.ID, retry+1, maxRetries, err)
						time.Sleep(retryInterval)
					} else {
						log.Printf("Failed to connect to bootstrap peer %s after %d attempts: %v",
							peer.ID, maxRetries, err)
						metric.RecordError("bootstrap_peer_connection_failed")
					}
				}
			}
		}
	}
}

func (n *node) connectWithTimeout(peer peer.AddrInfo) error {
	ctx, cancel := context.WithTimeout(n.ctx, 10*time.Second)
	defer cancel()

	if err := n.host.Connect(ctx, peer); err != nil {
		if strings.Contains(err.Error(), "failed to dial") {
			return fmt.Errorf("failed to dial: %w", err)
		}
		return err
	}

	for retry := 0; retry < 5; retry++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(n.host.Network().ConnsToPeer(peer.ID)) > 0 {
				return nil
			}
			time.Sleep(time.Second)
		}
	}

	return fmt.Errorf("connection not established after timeout")
}

func (n *node) startPeerDiscovery(rendezvous string) error {
	// Advertise our presence
	ttl, err := n.routingDiscovery.Advertise(n.ctx, rendezvous)
	if err != nil {
		// if it's first node, acceptable
		log.Printf("Warning: Failed to advertise (this is normal for the first node): %v", err)
		metric.RecordRequest("first_node_advertise", err.Error())
	} else {
		log.Printf("Advertising this node with TTL: %s", ttl)
	}

	// Start peer discovery
	go n.discoverPeers(rendezvous)

	// Start routing table maintenance
	go n.maintainRoutingTable()

	return nil
}

func (n *node) discoverPeers(rendezvous string) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			log.Printf("Starting peer discovery cycle with rendezvous: %s", rendezvous)
			peerChan, err := n.routingDiscovery.FindPeers(n.ctx, rendezvous)
			if err != nil {
				log.Printf("Peer discovery error: %v", err)
				metric.RecordRequest("peer_discovery_attempt", err.Error())
				continue
			}

			peers := n.host.Network().Peers()
			log.Printf("[Node] Connected to %d peers:", len(peers))
			for _, p := range peers {
				log.Printf("  - %s", p.String())
			}

			for peer := range peerChan {
				if peer.ID == n.host.ID() {
					continue // Skip self
				}

				log.Printf("Found peer: %s, attempting connection...", peer.ID)
				if err := n.host.Connect(n.ctx, peer); err != nil {
					log.Printf("Failed to connect to peer %s: %v", peer.ID, err)
					metric.RecordError("peer_connection_failed")
					continue
				}

				metric.RecordRequest("peer_connected", peer.ID.String())
				log.Printf("Successfully connected to peer: %s", peer.ID)
			}
		}
	}
}

func (n *node) maintainRoutingTable() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			log.Printf("Starting routing table refresh...")

			// get current router table
			peers := n.dht.RoutingTable().ListPeers()
			log.Printf("Current routing table has %d peers", len(peers))

			if err := n.dht.RefreshRoutingTable(); err != nil {
				log.Printf("No new peers to refresh routing table: %v (peers: %d)", err, len(peers))
				for _, p := range peers {
					conns := n.host.Network().ConnsToPeer(p)
					log.Printf("Peer %s connections: %d", p.String(), len(conns))
				}
				metric.RecordError("routing_table_refresh_failed")
			} else {
				newPeers := n.dht.RoutingTable().ListPeers()
				log.Printf("Routing table refreshed successfully. Peers: %d -> %d",
					len(peers), len(newPeers))
			}
		}
	}
}
