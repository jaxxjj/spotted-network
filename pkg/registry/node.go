package registry

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type MainnetClient interface {
	GetEffectiveEpochForBlock(ctx context.Context, blockNumber uint64) (uint32, error)
	WatchOperatorRegistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorRegisteredEvent) (event.Subscription, error)
	WatchOperatorDeregistered(filterOpts *bind.FilterOpts, sink chan<- *ethereum.OperatorDeregisteredEvent) (event.Subscription, error) 
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
}

// NodeConfig contains all the dependencies needed by Registry Node
type NodeConfig struct {
	Host           host.Host
	Operators      OperatorsQuerier
	MainnetClient  MainnetClient
}

type Node struct {
	host P2PHost
	
	// Connected operators
	operatorsInfo map[peer.ID]*OperatorInfo
	operatorsInfoMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	operators OperatorsQuerier

	// Event listener
	eventListener *EventListener
	
	// Epoch updator
	epochUpdator *EpochUpdator

	// Chain clients manager
	mainnetClient MainnetClient

	// State sync subscribers
	subscribers map[peer.ID]network.Stream
	subscribersMu sync.RWMutex

	// Ping service for health checks
	pingService *ping.PingService
}

func NewNode(cfg *NodeConfig) (*Node, error) {
	// Validate required dependencies
	if cfg.Host == nil {
		return nil, fmt.Errorf("host is required")
	}
	if cfg.Operators == nil {
		return nil, fmt.Errorf("operators querier is required")
	}
	if cfg.MainnetClient == nil {
		return nil, fmt.Errorf("mainnet client is required")
	}

	// Create ping service
	pingService := ping.NewPingService(cfg.Host)

	node := &Node{
		host:               cfg.Host,
		operatorsInfo:          make(map[peer.ID]*OperatorInfo),
		healthCheckInterval: 100 * time.Second,
		operators:                cfg.Operators,
		mainnetClient:      cfg.MainnetClient,
		subscribers:       make(map[peer.ID]network.Stream),
		pingService:      pingService,
	}

	// Create and initialize event listener
	node.eventListener = NewEventListener(node, cfg.MainnetClient, cfg.Operators)

	// Create and initialize epoch updator
	node.epochUpdator = NewEpochUpdator(node)

	// Start listening for chain events
	if err := node.eventListener.StartListening(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start event listener: %w", err)
	}

	// Start health check
	go node.startHealthCheck(context.Background())

	log.Printf("Registry Node started. ID: %s, Addrs: %v\n", cfg.Host.ID(), cfg.Host.Addrs())
	return node, nil
}

func (n *Node) Start(ctx context.Context) error {
	// Validate required components
	if n.host == nil {
		return fmt.Errorf("[Registry] host not initialized")
	}
	if n.operators == nil {
		return fmt.Errorf("[Registry] operators database not initialized") 
	}
	if n.epochUpdator == nil {
		return fmt.Errorf("[Registry] epoch updator not initialized")
	}
	if n.eventListener == nil {
		return fmt.Errorf("[Registry] event listener not initialized")
	}

	// Set up protocol handlers
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleStream)
	log.Printf("[Registry] Join request handler set up")

	// Start state sync service
	if err := n.startStateSync(); err != nil {
		return fmt.Errorf("failed to start state sync service: %w", err)
	}
	log.Printf("[Registry] State sync service started")

	// Start epoch monitoring
	go n.epochUpdator.Start(ctx)
	log.Printf("[Registry] Epoch monitoring started")

	return nil
}

func (n *Node) Stop() error {
	return n.host.Close()
}

// GetOperatorInfo returns information about a connected operator
func (n *Node) GetOperatorInfo(id peer.ID) *OperatorInfo {
	n.operatorsInfoMu.RLock()
	defer n.operatorsInfoMu.RUnlock()
	return n.operatorsInfo[id]
}

func (n *Node) GetConnectedOperators() []peer.ID {
	n.operatorsInfoMu.RLock()
	defer n.operatorsInfoMu.RUnlock()

	operators := make([]peer.ID, 0, len(n.operatorsInfo))
	for id := range n.operatorsInfo {
		operators = append(operators, id)
	}
	return operators
}

// GetHostID returns the node's libp2p host ID
func (n *Node) GetHostID() string {
	return n.host.ID().String()
}

// PingPeer implements ping functionality for the node
func (n *Node) PingPeer(ctx context.Context, p peer.ID) error {
	// Add ping timeout
	ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	result := <-n.pingService.Ping(ctx, p)
	if result.Error != nil {
		return fmt.Errorf("ping failed: %v", result.Error)
	}

	log.Printf("Successfully pinged peer: %s (RTT: %v)", p, result.RTT)
	return nil
}