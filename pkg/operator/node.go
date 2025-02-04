package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
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
	"github.com/libp2p/go-libp2p"
)

// P2PHost defines the minimal interface required for p2p functionality
type P2PHost interface {
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
	Connect(ctx context.Context, pi peer.AddrInfo) error
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(peerId protocol.ID, handler network.StreamHandler)
	RemoveStreamHandler(peerId protocol.ID)
	Network() network.Network
	Peerstore() peerstore.Peerstore
	Close() error
}

// APIServer defines the interface for the API server
type APIServer interface {
	Start() error
	Stop(ctx context.Context) error
}

// OperatorState stores business state synced from registry
type OperatorState struct {
	// Peer Info
	PeerID     peer.ID
	Multiaddrs []multiaddr.Multiaddr

	// Business State
	Address                 string
	SigningKey             string
	Weight                 *big.Int
}

type ActivePeerStates struct {
	active map[peer.ID]*OperatorState
	stateRoot []byte
	mu sync.RWMutex
}

// NodeConfig contains all the dependencies needed by Node
type NodeConfig struct {
	ChainManager    ChainManager
	Signer          signer.Signer
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

	signer         OperatorSigner
	chainManager   ChainManager
	apiServer      APIServer
	tasksQuerier   TasksQuerier
	taskResponseQuerier TaskResponseQuerier
	consensusResponseQuerier ConsensusResponseQuerier
	epochStateQuerier EpochStateQuerier
	
	tp  *TaskProcessor
	rh *RegistryHandler
	sp *StateSyncProcessor

	config          *config.Config

	registryID     peer.ID
	registryAddress   string
	activeOperators ActivePeerStates

}

// NewNode creates a new operator node with the given dependencies
func NewNode(ctx context.Context, cfg *NodeConfig) (*Node, error) {
	// Validate required dependencies
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

	// Create initial incomplete node instance (without host)
	node := &Node{
		registryID:     addrInfo.ID,
		registryAddress:   cfg.RegistryAddress,
		signer:         cfg.Signer,
		chainManager:   cfg.ChainManager,
		tasksQuerier:   cfg.TasksQuerier,
		taskResponseQuerier: cfg.TaskResponseQuerier,
		consensusResponseQuerier: cfg.ConsensusResponseQuerier,
		epochStateQuerier: cfg.EpochStateQuerier,
		config:        cfg.Config,
		activeOperators: ActivePeerStates{
			active: make(map[peer.ID]*OperatorState),
		},
	}

	// Create gater
	gater := newOperatorGater(node)

	// Use gater to create host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.Config.P2P.Port),
			fmt.Sprintf("/ip4/%s/tcp/%d", cfg.Config.P2P.ExternalIP, cfg.Config.P2P.Port),
		),
		libp2p.ConnectionGater(gater),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Set host to node
	node.host = host

	// Initialize pubsub
	pubsub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
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
		PubSub:            pubsub,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create task processor: %w", err)
	}
	node.tp = taskProcessor
	node.sp, err = NewStateSyncProcessor(ctx, node, pubsub)
	if err != nil {
		return nil, fmt.Errorf("failed to create state sync processor: %w", err)
	}
	registryId, err := peer.Decode(cfg.Config.RegistryID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode registry ID: %w", err)
	}

	// Initialize ping service
	pingService := ping.NewPingService(host)
	_, err = newHealthChecker(ctx, node, pingService)
	if err != nil {
		return nil, fmt.Errorf("failed to create health checker: %w", err)
	}
	node.rh = NewRegistryHandler(node, cfg.Signer, registryId)

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
	if n.tp == nil {
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

	// Step 1: Update epoch state
	currentEpoch, err := n.currentEpoch()
	if err != nil {
		return fmt.Errorf("failed to get current epoch: %w", err)
	}
	if err := n.updateEpochState(ctx, currentEpoch); err != nil {
		return fmt.Errorf("failed to update epoch state: %w", err)
	}

	// Step 2: Connect to registry
	log.Printf("[Node] Connecting to registry...")
	if err := n.connectToRegistry(ctx); err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	log.Printf("[Node] Successfully connected to registry")

	// Step 3: Authenticate with registry
	log.Printf("[Node] Authenticating with registry...")
	if err := n.rh.AuthToRegistry(ctx); err != nil {
		return fmt.Errorf("failed to authenticate with registry: %w", err)
	}
	log.Printf("[Node] Successfully authenticated with registry")

	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	// Stop API server
	if err := n.apiServer.Stop(ctx); err != nil {
		log.Printf("[Node] Error stopping API server: %v", err)
	}
	n.rh.Disconnect(ctx)
	n.sp.Stop()
	n.tp.Stop()
	return n.host.Close()
}

func (n *Node) createStreamToRegistry(ctx context.Context, peerId protocol.ID) (network.Stream, error) {
	return n.host.NewStream(ctx, n.registryID, peerId)
}

func (n *Node) connectToRegistry(ctx context.Context) error {
	log.Printf("[Node] Attempting to connect to Registry Node at address: %s\n", n.registryAddress)

	// Create multiaddr from the provided address
	addr, err := multiaddr.NewMultiaddr(n.registryAddress)
	if err != nil {
		return fmt.Errorf("[Node] invalid registry address: %s", n.registryAddress)
	}

	// Parse peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("[Node] failed to parse peer info from address: %v", err)
	}

	// Connect to registry
	if err := n.host.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("[Node] failed to connect to registry: %v", err)
	}

	// Save the registry ID
	n.registryID = peerInfo.ID

	log.Printf("[Node] Successfully connected to Registry Node with ID: %s\n", n.registryID)
	return nil
}

func (n *Node) getHostID() peer.ID {
	return n.host.ID()
}

// GetOperatorState returns the current operator state
func (n *Node) GetOperatorState(peerID peer.ID) *OperatorState {
	n.activeOperators.mu.RLock()
	defer n.activeOperators.mu.RUnlock()
	return n.activeOperators.active[peerID]
}

// RemoveOperator removes an operator from the active set
func (n *Node) RemoveOperator(id peer.ID) {
	n.activeOperators.mu.Lock()
	defer n.activeOperators.mu.Unlock()
	delete(n.activeOperators.active, id)
}

func (n *Node) currentEpoch() (uint32, error) {
	mainnetClient, err := n.chainManager.GetMainnetClient()
	if err != nil {
		return 0, fmt.Errorf("failed to get mainnet client: %w", err)
	}
	blockNumber, err := mainnetClient.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get block number: %w", err)
	}
	currentEpoch := (blockNumber - GenesisBlock) / EpochPeriod
	return uint32(currentEpoch), nil
}



