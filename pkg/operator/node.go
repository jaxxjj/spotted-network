package operator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/operator/api"
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
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	
	// State sync related fields
	operatorStates map[string]*pb.OperatorState
	statesMu       sync.RWMutex
	
	// Database connection
	db          *pgxpool.Pool
	taskQueries *tasks.Queries
	
	// Chain clients
	chainClient *ethereum.ChainClients
	
	// API server
	apiServer *api.Server
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

	// Create queries
	taskQueries := tasks.New(db)

	// Create ping service
	pingService := ping.NewPingService(host)

	// Initialize API handler and server
	apiHandler := api.NewHandler(taskQueries, chainClients)
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
		operatorsMu:    sync.RWMutex{},
		pingService:    pingService,
		operatorStates: make(map[string]*pb.OperatorState),
		statesMu:       sync.RWMutex{},
		db:            db,
		taskQueries:   taskQueries,
		chainClient:   chainClients,
		apiServer:     apiServer,
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
	// Create tasks table
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS tasks (
			task_id TEXT PRIMARY KEY,
			target_address TEXT NOT NULL,
			chain_id INT NOT NULL,
			block_number NUMERIC(78),
			timestamp NUMERIC(78),
			epoch INT NOT NULL,
			key NUMERIC(78) NOT NULL,
			value NUMERIC(78),
			expire_time TIMESTAMP NOT NULL,
			retries INT DEFAULT 0,
			status TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'expired', 'failed')),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
		CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
		CREATE INDEX IF NOT EXISTS idx_tasks_expire_time ON tasks(expire_time);
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	return nil
} 