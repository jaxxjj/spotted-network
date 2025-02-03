package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"github.com/stumble/wpgx"
)

type ActiveOperatorPeers struct {
	// Active operators (peer.ID -> OperatorState)
	active map[peer.ID]*OperatorPeerInfo
	stateRoot []byte
	mu sync.RWMutex
}

// OperatorState represents the complete state of an operator
type OperatorPeerInfo struct {
	// Address of the operator
	Address    string
	// Peer ID of the operator
	PeerID     peer.ID
	// Multiaddrs of the operator
	Multiaddrs []multiaddr.Multiaddr
	
	// Runtime status
	LastSeen   time.Time
	
	// Active streams
	RegistryStream network.Stream
	StateSyncStream network.Stream
}
type NodeQuerier interface {
	UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams, getOperatorByAddress *string) (*operators.Operators, error)
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)

	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams, getOperatorByAddress *string) (*operators.Operators, error)
}

type MainnetClient interface {
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)
	BlockNumber(ctx context.Context) (uint64, error)
	GetOperatorWeight(ctx context.Context, operator common.Address) (*big.Int, error)
	IsOperatorRegistered(ctx context.Context, operator common.Address) (bool, error) 
}

// P2PHost defines the minimal interface required for p2p functionality
type P2PHost interface {
	// Core functionality
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
	Connect(ctx context.Context, pi peer.AddrInfo) error
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(pid protocol.ID, handler network.StreamHandler)
	Peerstore() peerstore.Peerstore
	Close() error
	Network() network.Network
}


// NodeConfig contains all the dependencies needed by Registry Node
type NodeConfig struct {
	Host           host.Host
	OperatorsQuerier      *operators.Queries
	MainnetClient  *ethereum.ChainClient
	TxManager      *wpgx.Pool
	PubSub         *pubsub.PubSub
	Blacklist      pubsub.Blacklist  // Optional blacklist implementation
}

type Node struct {
	host P2PHost
	
	// Operator management
	activeOperators ActiveOperatorPeers

	// Database connection
	opQuerier NodeQuerier

	// Event listener
	eventListener *EventListener
	
	// Epoch updator
	epochUpdator *EpochUpdator

	// Chain clients manager
	mainnetClient MainnetClient

	// Auth handler
	registryHandler *RegistryHandler

	// State sync processor
	stateSyncProcessor *StateSyncProcessor

	// Health checker
	healthChecker *HealthChecker

	// PubSub service
	pubsub *pubsub.PubSub

	// Blacklist for managing blocked peers
	blacklist pubsub.Blacklist
}



func NewNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	// Validate required dependencies
	if cfg.Host == nil {
		log.Fatal("[Registry] host not initialized")
	}
	if cfg.OperatorsQuerier == nil {
		log.Fatal("[Registry] operators querier not initialized")
	}
	if cfg.MainnetClient == nil {
		log.Fatal("[Registry] mainnet client not initialized")
	}
	if cfg.PubSub == nil {
		log.Fatal("[Registry] pubsub not initialized")
	}

	// Create ping service
	pingService := ping.NewPingService(cfg.Host)

	// Initialize blacklist if not provided
	blacklist := cfg.Blacklist
	if blacklist == nil {
		blacklist = pubsub.NewMapBlacklist()
	}

	node := &Node{
		host:               cfg.Host,
		opQuerier:          cfg.OperatorsQuerier,
		mainnetClient:      cfg.MainnetClient,
		pubsub:            cfg.PubSub,
		blacklist:         blacklist,
	}

	// Initialize operators management
	node.activeOperators.active = make(map[peer.ID]*OperatorPeerInfo)
	
	// Create and initialize event listener
	node.eventListener = NewEventListener(ctx, node, cfg.MainnetClient, cfg.OperatorsQuerier)


	// Start registry service
	if err := node.startRegistryService(cfg); err != nil {
		return nil, fmt.Errorf("failed to start registry service: %w", err)
	}
	log.Printf("[Registry] Registry service started")

	// Start state sync service
	if err := node.startStateSync(cfg); err != nil {
		return nil, fmt.Errorf("failed to start state sync service: %w", err)
	}
	log.Printf("[Registry] State sync service started")

	// Create and initialize epoch updator
	epochUpdator, err := NewEpochUpdator(ctx, &EpochUpdatorConfig{
		node: node,
		opQuerier: cfg.OperatorsQuerier,
		pubsub: cfg.PubSub,
		mainnetClient: cfg.MainnetClient,
		sp: node.stateSyncProcessor,
		txManager: cfg.TxManager,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create epoch updator: %w", err)
	}
	node.epochUpdator = epochUpdator
	// Create and start health checker
	healthChecker, err := newHealthChecker(ctx, node, pingService)
	if err != nil {
		return nil, fmt.Errorf("failed to start health checker: %w", err)
	}
	node.healthChecker = healthChecker
	log.Printf("[Registry] Health check service started")

	log.Printf("[Registry] Epoch monitoring started")

	log.Printf("Registry Node started. ID: %s, Addrs: %v\n", cfg.Host.ID(), cfg.Host.Addrs())
	return node, nil
}


func (n *Node) Stop() error {
	return n.host.Close()
}


// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
}

func (n *Node) startRegistryService(cfg *NodeConfig) error {
	// Create and initialize registry handler
	n.registryHandler = NewRegistryHandler(n, cfg.OperatorsQuerier)

	// Set up protocol handler
	n.host.SetStreamHandler(RegistryProtocol, n.registryHandler.HandleRegistryStream)
	log.Printf("[Registry] Registry handler set up for protocol: %s", RegistryProtocol)

	return nil
}

func (n *Node) startStateSync(cfg *NodeConfig) error {
    // Create state sync processor
    processor, err := NewStateSyncProcessor(n, cfg.PubSub)
    if err != nil {
        return fmt.Errorf("failed to create state sync processor: %w", err)
    }
    
    // Set up protocol handler
    n.host.SetStreamHandler(StateVerifyProtocol, processor.handleStateVerifyStream)
    log.Printf("[StateSync] Stream handler set up for protocol: %s", StateVerifyProtocol)
    
    n.stateSyncProcessor = processor
    return nil
}