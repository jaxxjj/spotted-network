package operator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/operator/api"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/libp2p/go-libp2p/core/host"
)

type Node struct {
	host           host.Host
	registryID     peer.ID
	registryAddr   string
	signer         OperatorSigner
	knownOperators map[peer.ID]*peer.AddrInfo
	operators      map[peer.ID]*OperatorInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	chainManager   ChainManager
	operatorStates map[string]*pb.OperatorState
	statesMu       sync.RWMutex
	
	// Database connection
	db          *pgxpool.Pool
	epochState EpochStateQuerier
	
	// API server
	apiServer *api.Server

	// P2P pubsub
	PubSub *pubsub.PubSub

	// Task processor
	taskProcessor *TaskProcessor
}

func NewNode(registryAddr string, cfg *config.Config, chainManager ChainManager, signer OperatorSigner) (*Node, error) {
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
	taskQuerier := tasks.New(db)
	taskResponseQuerier := task_responses.New(db)
	consensusResponseQuerier := consensus_responses.New(db)
	epochStatesQuerier := epoch_states.New(db)

	// Create ping service
	pingService := ping.NewPingService(host)

	// Create the node instance first
	node := &Node{
		host:           host,
		registryID:     addrInfo.ID,
		registryAddr:   registryAddr,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operators:      make(map[peer.ID]*OperatorInfo),
		operatorsMu:    sync.RWMutex{},
		pingService:    pingService,
		operatorStates: make(map[string]*pb.OperatorState),
		statesMu:       sync.RWMutex{},
		db:            db,
		PubSub:        ps,
	}

	// Initialize task processor
	taskProcessor, err := NewTaskProcessor(node, signer, taskQuerier, taskResponseQuerier, consensusResponseQuerier, epochStatesQuerier)
	if err != nil {
		return nil, fmt.Errorf("failed to create task processor: %w", err)
	}
	node.taskProcessor = taskProcessor

	// Initialize API handler and server with task processor
	apiHandler := api.NewHandler(taskQuerier, chainManager, consensusResponseQuerier, taskProcessor, cfg)
	apiServer := api.NewServer(apiHandler, cfg.HTTP.Port)
	node.apiServer = apiServer
	
	// Start API server
	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	return node, nil
}

func (n *Node) Start(ctx context.Context) error {
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
	
	// Close database connection
	n.db.Close()
	
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

func initDatabase(ctx context.Context, db *pgxpool.Pool) error {
	// Check if database is accessible
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("[Node] failed to ping database: %w", err)
	}
	log.Printf("[Node] Successfully connected to database")
	return nil
}


