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

type Network interface {
	Peers() []peer.ID
	Peerstore() Peerstore
	Connect(ctx context.Context, pi peer.AddrInfo) error
	ClosePeer(peer.ID) error
	Connectedness(peer.ID) network.Connectedness
}

type Peerstore interface {
	Addrs(peer.ID) []multiaddr.Multiaddr
	RemovePeer(peer.ID)
	ClearAddrs(peer.ID)
}

// P2PHost defines the minimal interface required for p2p functionality
type P2PHost interface {
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
	Connect(ctx context.Context, pi peer.AddrInfo) error
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(peerId protocol.ID, handler network.StreamHandler)
	RemoveStreamHandler(peerId protocol.ID)
	Network() Network
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


func (n *Node) getHostID() peer.ID {
	return n.host.ID()
}






