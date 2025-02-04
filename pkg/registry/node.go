package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
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
type OperatorsQuerier interface {
	UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams, getOperatorByAddress *string) (*operators.Operators, error)
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
	ListAllOperators(ctx context.Context) ([]operators.Operators, error)
	WithTx(tx *wpgx.WTx) *operators.Queries
	UpsertOperator(ctx context.Context, arg operators.UpsertOperatorParams, getOperatorByAddress *string) (*operators.Operators, error)
}

type MainnetClient interface {
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)
	BlockNumber(ctx context.Context) (uint64, error)
	GetOperatorWeight(ctx context.Context, operator ethcommon.Address) (*big.Int, error)
	IsOperatorRegistered(ctx context.Context, operator ethcommon.Address) (bool, error) 
	GetOperatorSigningKey(ctx context.Context, address ethcommon.Address, epoch uint32) (ethcommon.Address, error)
}

// P2PHost defines the minimal interface required for p2p functionality
type P2PHost interface {
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
	Connect(ctx context.Context, pi peer.AddrInfo) error
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(peerId protocol.ID, handler network.StreamHandler)
	Peerstore() peerstore.Peerstore
	Close() error
	Network() network.Network
}

// NodeConfig contains all the dependencies needed by Registry Node
type NodeConfig struct {
	Host           host.Host
	OperatorsQuerier      *operators.Queries
	BlacklistQuerier      *blacklist.Queries
	MainnetClient  *ethereum.ChainClient
	TxManager      *wpgx.Pool
	PubSub         *pubsub.PubSub
}

type Node struct {
	// libp2p host for p2p communication
	host P2PHost
	// Operator querier for querying the operators from the database
	opQuerier OperatorsQuerier
	// Blacklist querier for querying the blacklist from the database
	blacklistQuerier *blacklist.Queries
	// Chain clients manager
	mainnetClient MainnetClient

	// Event listener for listening to events from the mainnet (operators registration/deregistration)
	el *EventListener
	// Epoch updator for updating the epoch 
	eu *EpochUpdator
	// Registry handler for handling nodes join requests
	rh *RegistryHandler
	// State sync processor for syncing the state of the operators
	sp *StateSyncProcessor
	// Health checker for checking the health of the operator nodes
	hc *HealthChecker

	// stores the active operators in memory
	activeOperators ActiveOperatorPeers
	// Pubsub for publishing and subscribing to topics
	pubsub PubSubService
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

	node := &Node{
		host:               cfg.Host,
		opQuerier:          cfg.OperatorsQuerier,
		mainnetClient:      cfg.MainnetClient,
		pubsub:            cfg.PubSub,
	}

	// Initialize operators management
	node.activeOperators.active = make(map[peer.ID]*OperatorPeerInfo)
	
	// Create and initialize event listener
	node.el = NewEventListener(ctx, &EventListenerConfig{
		node: node,
		mainnetClient: cfg.MainnetClient,
		operators: cfg.OperatorsQuerier,
	})


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
		txManager: cfg.TxManager,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create epoch updator: %w", err)
	}
	node.eu = epochUpdator
	// Create and start health checker
	healthChecker, err := newHealthChecker(ctx, node, pingService)
	if err != nil {
		return nil, fmt.Errorf("failed to start health checker: %w", err)
	}
	node.hc = healthChecker
	log.Printf("[Registry] Health check service started")

	log.Printf("[Registry] Epoch monitoring started")

	log.Printf("Registry Node started. ID: %s, Addrs: %v\n", cfg.Host.ID(), cfg.Host.Addrs())
	return node, nil
}

func (n *Node) Stop() error {
	log.Printf("[Registry] Stopping registry node...")

	n.el.Stop()
	n.eu.Stop()
	n.sp.Stop()
	return n.host.Close()
}

// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
}

func (n *Node) startRegistryService(cfg *NodeConfig) error {
	// Create and initialize registry handler
	n.rh = NewRegistryHandler(&RegistryHandlerConfig{
		node: n,
		opQuerier: cfg.OperatorsQuerier,
	})

	// Set up protocol handler
	n.host.SetStreamHandler(RegistryProtocol, n.rh.HandleRegistryStream)
	log.Printf("[Registry] Registry handler set up for protocol: %s", RegistryProtocol)

	return nil
}

func (n *Node) startStateSync(cfg *NodeConfig) error {
    // Create state sync processor
    processor, err := NewStateSyncProcessor(&PubsubConfig{
		node: n,
		pubsub: cfg.PubSub,
	})
    if err != nil {
        return fmt.Errorf("failed to create state sync processor: %w", err)
    }
    
    // Set up protocol handler
    n.host.SetStreamHandler(StateVerifyProtocol, processor.handleStateVerifyStream)
    log.Printf("[StateSync] Stream handler set up for protocol: %s", StateVerifyProtocol)
    
    n.sp = processor
    return nil
}

// NotifyEpochUpdate handles all necessary state synchronization when an epoch is updated
func (n *Node) NotifyEpochUpdate(ctx context.Context, epoch uint32) error {
	if err := n.syncPeerInfo(ctx); err != nil {
		return fmt.Errorf("failed to sync peer info: %w", err)
	}

	if err := n.sp.broadcastStateUpdate(&epoch); err != nil {
		return fmt.Errorf("failed to broadcast state update: %w", err)
	}

	return nil
}