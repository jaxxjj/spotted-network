package operator

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	pb "github.com/galxe/spotted-network/proto"
)

type Node struct {
	host           host.Host
	registryID     peer.ID
	registryAddr   string
	signer         signer.Signer
	knownOperators map[peer.ID]*peer.AddrInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	
	// State sync related fields
	operatorStates map[string]*pb.OperatorState  // address -> state
	statesMu       sync.RWMutex
}

func NewNode(registryAddr string, s signer.Signer) (*Node, error) {
	// Parse the registry multiaddr
	maddr, err := multiaddr.NewMultiaddr(registryAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry address: %w", err)
	}

	// Extract peer ID from multiaddr
	addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry peer info: %w", err)
	}

	// Create a new host with default configuration
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Create ping service
	pingService := ping.NewPingService(host)

	return &Node{
		host:           host,
		registryID:     addrInfo.ID,
		registryAddr:   registryAddr,
		signer:         s,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operatorsMu:    sync.RWMutex{},
		pingService:    pingService,
		operatorStates: make(map[string]*pb.OperatorState),
		statesMu:      sync.RWMutex{},
	}, nil
}

func (n *Node) Start() error {
	log.Printf("Operator node started with ID: %s", n.host.ID())

	log.Printf("Connecting to registry...")
	if err := n.connectToRegistry(); err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	log.Printf("Successfully connected to registry")

	// Start message handler and health check
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleMessages)
	go n.healthCheck()
	log.Printf("Message handler and health check started")

	// Subscribe to state updates
	if err := n.subscribeToStateUpdates(); err != nil {
		return fmt.Errorf("failed to subscribe to state updates: %w", err)
	}
	log.Printf("Subscribed to state updates")

	// Get initial state
	if err := n.getFullState(); err != nil {
		return fmt.Errorf("failed to get initial state: %w", err)
	}
	log.Printf("Got initial state")

	// Announce to registry
	log.Printf("Announcing to registry...")
	if err := n.announceToRegistry(); err != nil {
		return fmt.Errorf("failed to announce to registry: %w", err)
	}
	log.Printf("Successfully announced to registry")

	return nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

func (n *Node) connectToRegistry() error {
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
	if err := n.host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to registry: %v", err)
	}

	// Save the registry ID
	n.registryID = peerInfo.ID

	log.Printf("Successfully connected to Registry Node with ID: %s\n", n.registryID)
	return nil
} 