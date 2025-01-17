package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/operator/api"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Node struct {
	host           host.Host
	registryID     peer.ID
	registryAddr   string
	signer         signer.Signer
	knownOperators map[peer.ID]*peer.AddrInfo
	operators      map[peer.ID]*OperatorInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	
	// State sync related fields
	operatorStates map[string]*pb.OperatorState
	statesMu       sync.RWMutex
	
	// Database connection
	db          *pgxpool.Pool
	taskQueries *tasks.Queries
	responseQueries *task_responses.Queries
	consensusQueries *consensus_responses.Queries
	
	// Chain clients
	chainClient *ethereum.ChainClients
	
	// API server
	apiServer *api.Server

	// P2P pubsub
	PubSub *pubsub.PubSub

	// Task processor
	taskProcessor *TaskProcessor
}

func NewNode(registryAddr string, s signer.Signer, cfg *Config, chainClients *ethereum.ChainClients) (*Node, error) {
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

	// Create a new host with P2P configuration
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", cfg.P2P.ExternalIP, cfg.P2P.Port)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Initialize database connection
	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure database pool
	db.Config().MaxConns = int32(cfg.Database.MaxOpenConns)
	db.Config().MinConns = int32(cfg.Database.MaxIdleConns)

	// Initialize database tables
	if err := initDatabase(context.Background(), db); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize pubsub
	ps, err := pubsub.NewGossipSub(context.Background(), host)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// Initialize database queries
	taskQueries := tasks.New(db)
	responseQueries := task_responses.New(db)
	consensusQueries := consensus_responses.New(db)

	// Create ping service
	pingService := ping.NewPingService(host)

	// Initialize API handler and server
	apiHandler := api.NewHandler(taskQueries, chainClients, consensusQueries)
	apiServer := api.NewServer(apiHandler, cfg.HTTP.Port)
	
	// Start API server
	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	return &Node{
		host:           host,
		registryID:     addrInfo.ID,
		registryAddr:   registryAddr,
		signer:         s,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operators:      make(map[peer.ID]*OperatorInfo),
		operatorsMu:    sync.RWMutex{},
		pingService:    pingService,
		operatorStates: make(map[string]*pb.OperatorState),
		statesMu:       sync.RWMutex{},
		db:            db,
		taskQueries:   taskQueries,
		responseQueries: responseQueries,
		consensusQueries: consensusQueries,
		chainClient:   chainClients,
		apiServer:     apiServer,
		PubSub:        ps,
	}, nil
}

func (n *Node) Start() error {
	log.Printf("[Node] Starting operator node with ID: %s", n.host.ID())
	log.Printf("[Node] Listening addresses: %v", n.host.Addrs())

	log.Printf("[Node] Connecting to registry...")
	if err := n.connectToRegistry(); err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	log.Printf("[Node] Successfully connected to registry")

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

	// Announce to registry
	log.Printf("[Node] Announcing to registry...")
	if err := n.announceToRegistry(); err != nil {
		return fmt.Errorf("failed to announce to registry: %w", err)
	}
	log.Printf("[Node] Successfully announced to registry")

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
	log.Printf("[Node] Connected to 2 peers:")
	for _, peer := range peers {
		addrs := n.host.Network().Peerstore().Addrs(peer)
		log.Printf("[Node] - Peer %s at %v", peer.String(), addrs)
	}

	// Initialize and start task processor after P2P connections are established
	var err error
	n.taskProcessor, err = NewTaskProcessor(n, n.taskQueries, n.responseQueries, n.consensusQueries)
	if err != nil {
		return fmt.Errorf("failed to create task processor: %w", err)
	}
	log.Printf("[Node] Task processor initialized")

	// Start message handler and health check
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleMessages)
	go n.healthCheck()
	log.Printf("[Node] Message handler and health check started")

	return nil
}

func (n *Node) Stop() error {
	// Stop API server
	if err := n.apiServer.Stop(context.Background()); err != nil {
		log.Printf("Error stopping API server: %v", err)
	}
	
	// Close database connection
	n.db.Close()
	
	// Close chain clients
	if err := n.chainClient.Close(); err != nil {
		log.Printf("Error closing chain clients: %v", err)
	}

	// Clean up task processor resources
	if n.taskProcessor != nil {
		n.taskProcessor.Stop()
	}
	
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

func initDatabase(ctx context.Context, db *pgxpool.Pool) error {
	// Check if database is accessible
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	log.Printf("Successfully connected to database")
	return nil
}

// getOperatorWeight returns the weight of an operator
func (n *Node) getOperatorWeight(addr string) *big.Int {
	n.statesMu.RLock()
	defer n.statesMu.RUnlock()
	
	if state, exists := n.operatorStates[addr]; exists {
		weight := new(big.Int)
		weight.SetString(state.Weight, 10)
		return weight
	}
	return big.NewInt(0)
}

// getConsensusThreshold returns the required weight threshold for consensus
func (n *Node) getConsensusThreshold() *big.Int {
	n.statesMu.RLock()
	defer n.statesMu.RUnlock()
	
	// Calculate total weight
	totalWeight := big.NewInt(0)
	for _, state := range n.operatorStates {
		weight := new(big.Int)
		weight.SetString(state.Weight, 10)
		totalWeight.Add(totalWeight, weight)
	}
	
	// Threshold is 2/3 of total weight
	threshold := new(big.Int).Mul(totalWeight, big.NewInt(2))
	threshold.Div(threshold, big.NewInt(3))
	
	return threshold
}

// CalculateTotalWeight calculates the total weight of operator signatures
func (n *Node) CalculateTotalWeight(sigs []byte) string {
	var operatorSigs map[string][]byte
	if err := json.Unmarshal(sigs, &operatorSigs); err != nil {
		return "0"
	}

	totalWeight := new(big.Int)
	n.statesMu.RLock()
	defer n.statesMu.RUnlock()

	for addr := range operatorSigs {
		if state, ok := n.operatorStates[addr]; ok {
			weight := new(big.Int)
			weight.SetString(state.Weight, 10)
			totalWeight.Add(totalWeight, weight)
		}
	}

	return totalWeight.String()
} 