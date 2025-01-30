package operator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/operator/api"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pb "github.com/galxe/spotted-network/proto"
)

// P2PHost defines the minimal interface required for p2p functionality
type P2PHost interface {
	// Core functionality
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
	Connect(ctx context.Context, pi peer.AddrInfo) error
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(pid protocol.ID, handler network.StreamHandler)
	Network() network.Network
	Close() error
}

// APIServer defines the interface for the API server
type APIServer interface {
	Start() error
	Stop(ctx context.Context) error
}

type OperatorInfo struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Status   string
}


// NodeConfig contains all the dependencies needed by Node
type NodeConfig struct {
	Host            host.Host
	ChainManager    ChainManager
	Signer          signer.Signer
	PubSub          *pubsub.PubSub
	TasksQuerier     *tasks.Queries
	TaskResponseQuerier *task_responses.Queries
	ConsensusResponseQuerier *consensus_responses.Queries
	EpochStateQuerier *epoch_states.Queries
	RegistryAddress string
	Config          *config.Config
}

// Node represents an operator node in the network
type Node struct {
	host           P2PHost
	registryID     peer.ID
	registryAddr   string
	signer         OperatorSigner
	chainManager   ChainManager
	knownOperators map[peer.ID]*peer.AddrInfo
	operators      map[peer.ID]*OperatorInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	operatorStates map[string]*pb.OperatorState
	statesMu       sync.RWMutex
	apiServer   APIServer
	taskProcessor *TaskProcessor
	tasksQuerier  TasksQuerier
	taskResponseQuerier TaskResponseQuerier
	consensusResponseQuerier ConsensusResponseQuerier
	epochStateQuerier EpochStateQuerier
	config      *config.Config
}

// NewNode creates a new operator node with the given dependencies
func NewNode(cfg *NodeConfig) (*Node, error) {
	// Validate required dependencies
	if cfg.Host == nil {
		return nil, fmt.Errorf("host is required")
	}
	if cfg.ChainManager == nil {
		return nil, fmt.Errorf("chain manager is required")
	}
	if cfg.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if cfg.TasksQuerier == nil {
		return nil, fmt.Errorf("task querier is required")
	}
	if cfg.TaskResponseQuerier == nil {
		return nil, fmt.Errorf("task response querier is required")
	}
	if cfg.ConsensusResponseQuerier == nil {
		return nil, fmt.Errorf("consensus response querier is required")
	}
	if cfg.EpochStateQuerier == nil {
		return nil, fmt.Errorf("epoch state querier is required")
	}
	if cfg.RegistryAddress == "" {
		return nil, fmt.Errorf("registry address is required")
	}
	if cfg.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Parse registry address
	maddr, err := multiaddr.NewMultiaddr(cfg.RegistryAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry address: %w", err)
	}

	// Extract peer ID from multiaddr
	addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry peer info: %w", err)
	}

	// Create ping service
	pingService := ping.NewPingService(cfg.Host)

	node := &Node{
		host:           cfg.Host,
		registryID:     addrInfo.ID,
		registryAddr:   cfg.RegistryAddress,
		signer:         cfg.Signer,
		chainManager:   cfg.ChainManager,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operators:      make(map[peer.ID]*OperatorInfo),
		operatorsMu:    sync.RWMutex{},
		operatorStates: make(map[string]*pb.OperatorState),
		statesMu:       sync.RWMutex{},
		tasksQuerier:   cfg.TasksQuerier,
		taskResponseQuerier: cfg.TaskResponseQuerier,
		consensusResponseQuerier: cfg.ConsensusResponseQuerier,
		epochStateQuerier: cfg.EpochStateQuerier,
		config:        cfg.Config,
		pingService:   pingService,
	}
	// Initialize pubsub
	ps, err := pubsub.NewGossipSub(context.Background(), cfg.Host)
	if err != nil {
		log.Fatal("Failed to create pubsub:", err)
	}

	// Initialize task processor
	taskProcessor, err := NewTaskProcessor(&TaskProcessorConfig{
		Node:               node,
		Signer:            cfg.Signer,
		Tasks:              cfg.TasksQuerier,
		TaskResponse:      cfg.TaskResponseQuerier,
		ConsensusResponse: cfg.ConsensusResponseQuerier,
		EpochState:        cfg.EpochStateQuerier,
		ChainManager:      cfg.ChainManager,
		PubSub:            ps,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task processor: %w", err)
	}
	node.taskProcessor = taskProcessor
	// Create API handler and server
	apiHandler := api.NewHandler(cfg.TasksQuerier, cfg.ChainManager, cfg.ConsensusResponseQuerier, taskProcessor, cfg.Config)
	node.apiServer = api.NewServer(apiHandler, cfg.Config.HTTP.Port)
	
	// Start API server
	go func() {
		if err := node.apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	return node, nil
}

func (n *Node) Start(ctx context.Context) error {
	// Validate required components
	if n.signer == nil {
		log.Fatal("[Node] signer not initialized")
	}
	if n.host == nil {
		log.Fatal("[Node] host not initialized")
	}
	if n.registryID == "" {
		log.Fatal("[Node] registry ID not set")
	}
	if n.taskProcessor == nil {
		log.Fatal("[Node] task processor not initialized")
	}
	if n.chainManager == nil {
		log.Fatal("[Node] chain manager not initialized")
	}
	if n.epochStateQuerier == nil {
		log.Fatal("[Node] epoch state querier not initialized")
	}

	log.Printf("[Node] Starting operator node with ID: %s", n.host.ID())
	log.Printf("[Node] Listening addresses: %v", n.host.Addrs())

	log.Printf("[Node] Connecting to registry...")
	if err := n.connectToRegistry(); err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	log.Printf("[Node] Successfully connected to registry")

	// Start message handler and health check
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleMessages)
	// Set up state sync handler
	n.host.SetStreamHandler("/state-sync/1.0.0", n.handleStateUpdates)
	go n.healthCheck()
	log.Printf("[Node] Message handler and health check started")

	// Start epoch monitoring
	go n.monitorEpochUpdates(ctx)

	// Subscribe to state updates first
	if err := n.subscribeToStateUpdates(); err != nil {
		return fmt.Errorf("failed to subscribe to state updates: %w", err)
	}
	log.Printf("[Node] Subscribed to state updates")

	// Get initial state
	if err := n.getFullState(); err != nil {
		return fmt.Errorf("failed to get initial state: %w", err)
	}
	log.Printf("[Node] Got initial state")

	// Announce to registry and get active operators
	log.Printf("[Node] Announcing to registry...")
	activeOperators, err := n.announceToRegistry()
	if err != nil {
		return fmt.Errorf("failed to announce to registry: %w", err)
	}
	log.Printf("[Node] Successfully announced to registry and received %d active operators", len(activeOperators))

	// Store active operators
	n.operatorsMu.Lock()
	for _, addrInfo := range activeOperators {
		if addrInfo.ID == n.host.ID() {
			log.Printf("[Node] Skipping self in active operators list")
			continue
		}
		n.knownOperators[addrInfo.ID] = addrInfo
		log.Printf("[Node] Added operator %s with addresses %v", addrInfo.ID, addrInfo.Addrs)
	}
	n.operatorsMu.Unlock()

	// Wait for a short time to allow other operators to be discovered
	log.Printf("[Node] Waiting for other operators to be discovered...")
	time.Sleep(5 * time.Second)

	// Connect to known operators
	n.operatorsMu.RLock()
	log.Printf("[Node] Found %d known operators", len(n.knownOperators))
	for _, addrInfo := range n.knownOperators {
		if addrInfo.ID == n.host.ID() {
			log.Printf("[Node] Skipping self connection")
			continue // Skip self
		}
		log.Printf("[Node] Connecting to operator %s at %v", addrInfo.ID, addrInfo.Addrs)
		if err := n.host.Connect(context.Background(), *addrInfo); err != nil {
			log.Printf("[Node] Failed to connect to operator %s: %v", addrInfo.ID, err)
			continue
		}
		log.Printf("[Node] Successfully connected to operator %s", addrInfo.ID)
	}
	n.operatorsMu.RUnlock()

	// Print connected peers
	peers := n.host.Network().Peers()
	log.Printf("[Node] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := n.host.Network().Peerstore().Addrs(peer)
		log.Printf("[Node] - Peer %s at %v", peer.String(), addrs)
	}

	return nil
}

func (n *Node) Stop() error {
	// Stop API server
	if err := n.apiServer.Stop(context.Background()); err != nil {
		log.Printf("[Node] Error stopping API server: %v", err)
	}
	
	return n.host.Close()
}

func (n *Node) connectToRegistry() error {
	log.Printf("[Node] Attempting to connect to Registry Node at address: %s\n", n.registryAddr)

	// Create multiaddr from the provided address
	addr, err := multiaddr.NewMultiaddr(n.registryAddr)
	if err != nil {
		return fmt.Errorf("[Node] invalid registry address: %s", n.registryAddr)
	}

	// Parse peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("[Node] failed to parse peer info from address: %v", err)
	}

	// Connect to registry
	if err := n.host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("[Node] failed to connect to registry: %v", err)
	}

	// Save the registry ID
	n.registryID = peerInfo.ID

	log.Printf("[Node] Successfully connected to Registry Node with ID: %s\n", n.registryID)
	return nil
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


