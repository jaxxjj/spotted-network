package registry

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry"
	pb "github.com/galxe/spotted-network/proto"
)

type Node struct {
	host *p2p.Host
	
	// Connected operators
	operators map[peer.ID]*OperatorInfo
	operatorsMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	db *registry.Queries

	// Event listener
	eventListener *EventListener

	// Chain clients manager
	chainClients *ethereum.ChainClients

	// State sync subscribers
	subscribers map[peer.ID]network.Stream
	subscribersMu sync.RWMutex
}

type OperatorInfo struct {
	ID peer.ID
	Addrs []multiaddr.Multiaddr
	LastSeen time.Time
	Status string
}

func NewNode(ctx context.Context, cfg *p2p.Config, db *registry.Queries, chainClients *ethereum.ChainClients) (*Node, error) {
	host, err := p2p.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	node := &Node{
		host:               host,
		operators:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 5 * time.Second,
		db:                db,
		chainClients:      chainClients,
		subscribers:       make(map[peer.ID]network.Stream),
	}

	// Create and initialize event listener
	node.eventListener = NewEventListener(chainClients, db)

	// Start listening for chain events
	if err := node.eventListener.StartListening(ctx); err != nil {
		return nil, fmt.Errorf("failed to start event listener: %w", err)
	}

	// Set stream handler for operator connections
	node.SetupProtocols()

	// Start health check
	go node.startHealthCheck(ctx)

	log.Printf("Registry Node started. %s\n", host.GetHostInfo())
	return node, nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

// SetupProtocols sets up the protocol handlers for the registry node
func (n *Node) SetupProtocols() {
	// Set up join request handler
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleStream)
	
	// Set up state sync handler
	n.host.SetStreamHandler("/state-sync/1.0.0", n.handleStateSync)
	
	log.Printf("Protocol handlers set up")
}

// BroadcastStateUpdate sends a state update to all subscribers
func (n *Node) BroadcastStateUpdate(operators []*pb.OperatorState, updateType string) {
	update := &pb.OperatorStateUpdate{
		Type:      updateType,
		Operators: operators,
	}

	data, err := proto.Marshal(update)
	if err != nil {
		log.Printf("Error marshaling state update: %v", err)
		return
	}

	n.subscribersMu.RLock()
	defer n.subscribersMu.RUnlock()

	for peer, stream := range n.subscribers {
		if _, err := stream.Write(data); err != nil {
			log.Printf("Error sending update to %s: %v", peer, err)
			stream.Reset()
			delete(n.subscribers, peer)
		} else {
			log.Printf("Sent state update to %s - Type: %s, Operators: %d", 
				peer, updateType, len(operators))
		}
	}
}

// GetOperatorInfo returns information about a connected operator
func (n *Node) GetOperatorInfo(id peer.ID) *OperatorInfo {
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()
	return n.operators[id]
}