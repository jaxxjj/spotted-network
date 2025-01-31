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

// OperatorState represents the complete state of an operator
type OperatorState struct {
	// 基本信息
	Address    string
	PeerID     peer.ID
	Multiaddrs []multiaddr.Multiaddr
	
	// 运行时状态
	LastSeen   time.Time
	Status     string
	
	// 状态同步
	SyncStream network.Stream
}

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

type OperatorInfo struct {
	// Active operators (peer.ID -> OperatorState)
	active map[peer.ID]*OperatorState
	mu sync.RWMutex
}

// NodeConfig contains all the dependencies needed by Registry Node
type NodeConfig struct {
	Host           host.Host
	Operators      OperatorsQuerier
	MainnetClient  MainnetClient
}

type Node struct {
	host P2PHost
	
	// Operator management
	operators OperatorInfo
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	operatorsDB OperatorsQuerier

	// Event listener
	eventListener *EventListener
	
	// Epoch updator
	epochUpdator *EpochUpdator

	// Chain clients manager
	mainnetClient MainnetClient

	// Ping service for health checks
	pingService *ping.PingService

	// Auth handler
	authHandler *AuthHandler
}

func NewNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	// Validate required dependencies
	if cfg.Host == nil {
		log.Fatal("[Registry] host not initialized")
	}
	if cfg.Operators == nil {
		log.Fatal("[Registry] operators querier not initialized")
	}
	if cfg.MainnetClient == nil {
		log.Fatal("[Registry] mainnet client not initialized")
	}

	// Create ping service
	pingService := ping.NewPingService(cfg.Host)

	node := &Node{
		host:               cfg.Host,
		healthCheckInterval: 100 * time.Second,
		operatorsDB:        cfg.Operators,
		mainnetClient:      cfg.MainnetClient,
		pingService:        pingService,
	}

	// Initialize operators management
	node.operators.active = make(map[peer.ID]*OperatorState)
	
	// Create and initialize event listener
	node.eventListener = NewEventListener(node, cfg.MainnetClient, cfg.Operators)

	// Create and initialize epoch updator
	node.epochUpdator = NewEpochUpdator(node)
	if err := node.eventListener.StartListening(ctx); err != nil {
		return nil, fmt.Errorf("failed to start event listener: %w", err)
	}

	// Create and initialize auth handler
	node.authHandler = NewAuthHandler(node, cfg.Operators)

	// Set up protocol handlers
	node.host.SetStreamHandler(RegistryProtocolID, node.authHandler.HandleStream)
	log.Printf("[Registry] Auth handler set up")

	// Start state sync service
	if err := node.startStateSync(); err != nil {
		return nil, fmt.Errorf("failed to start state sync service: %w", err)
	}
	log.Printf("[Registry] State sync service started")

	go node.startHealthCheck(ctx)
	// Start epoch monitoring
	go node.epochUpdator.Start(ctx)
	log.Printf("[Registry] Epoch monitoring started")

	log.Printf("Registry Node started. ID: %s, Addrs: %v\n", cfg.Host.ID(), cfg.Host.Addrs())
	return node, nil
}

func (n *Node) Start(ctx context.Context) error {
	// Validate required components
	if n.host == nil {
		return fmt.Errorf("[Registry] host not initialized")
	}
	if n.operatorsDB == nil {
		return fmt.Errorf("[Registry] operators database not initialized") 
	}
	if n.epochUpdator == nil {
		return fmt.Errorf("[Registry] epoch updator not initialized")
	}
	if n.eventListener == nil {
		return fmt.Errorf("[Registry] event listener not initialized")
	}

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

// GetOperatorState returns the state of a connected operator
func (n *Node) GetOperatorState(id peer.ID) *OperatorState {
	n.operators.mu.RLock()
	defer n.operators.mu.RUnlock()
	return n.operators.active[id]
}

// GetConnectedOperators returns all connected operator IDs
func (n *Node) GetConnectedOperators() []peer.ID {
	n.operators.mu.RLock()
	defer n.operators.mu.RUnlock()

	operators := make([]peer.ID, 0, len(n.operators.active))
	for id := range n.operators.active {
		operators = append(operators, id)
	}
	return operators
}

// UpdateOperatorState updates an operator's state
func (n *Node) UpdateOperatorState(id peer.ID, state *OperatorState) {
	n.operators.mu.Lock()
	defer n.operators.mu.Unlock()
	n.operators.active[id] = state
}

// RemoveOperator removes an operator from the active set
func (n *Node) RemoveOperator(id peer.ID) {
	n.operators.mu.Lock()
	defer n.operators.mu.Unlock()
	delete(n.operators.active, id)
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