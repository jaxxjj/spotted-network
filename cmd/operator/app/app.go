package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/galxe/spotted-network/internal/database/cache"
	wpgxInitor "github.com/galxe/spotted-network/internal/database/wpgx"
	"github.com/galxe/spotted-network/internal/metric"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/operator/api"
	"github.com/galxe/spotted-network/pkg/operator/epoch"
	"github.com/galxe/spotted-network/pkg/operator/event"
	"github.com/galxe/spotted-network/pkg/operator/gater"
	"github.com/galxe/spotted-network/pkg/operator/health"
	"github.com/galxe/spotted-network/pkg/operator/node"
	"github.com/galxe/spotted-network/pkg/operator/task"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/redis/go-redis/v9"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx"
)

// repos holds all repository instances
type Repos struct {
	OperatorRepo          *operators.Queries
	BlacklistRepo         *blacklist.Queries
	TaskRepo              *tasks.Queries
	ConsensusResponseRepo *consensus_responses.Queries
}

// app holds all the dependencies
type App struct {
	ctx            context.Context
	isKeyPath      bool
	signingKeyPath string
	signingKeyPriv string
	p2pKey         string
	password       string

	cfg    *config.Config
	signer signer.Signer

	chainManager  ethereum.ChainManager
	mainnetClient ethereum.ChainClient
	db            *wpgx.Pool
	redisConn     redis.UniversalClient
	dCache        *dcache.DCache
	metricServer  *metric.Server
	host          host.Host
	node          node.Node
	apiServer     *api.Server
	repos         *Repos

	gater         gater.ConnectionGater
	taskProcessor task.TaskProcessor
	epochUpdator  epoch.EpochStateQuerier
}

// new creates a new application instance
func New(ctx context.Context) *App {
	return &App{
		ctx: ctx,
	}
}

// run starts the application with the provided configuration
func (a *App) Run(isKeyPath bool, signingKey, p2pKey, password string) error {
	// store parameters
	a.isKeyPath = isKeyPath
	if isKeyPath {
		a.signingKeyPath = signingKey
		a.password = password
	} else {
		a.signingKeyPriv = signingKey
	}
	a.p2pKey = p2pKey

	// Initialize all components
	if err := a.initConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := a.initMetrics(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	if err := a.initSigner(); err != nil {
		return fmt.Errorf("failed to initialize signer: %w", err)
	}

	if err := a.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := a.initChainManagerAndMainnetClient(); err != nil {
		return fmt.Errorf("failed to initialize chain manager and mainnet client: %w", err)
	}

	if err := a.initEventListener(); err != nil {
		return fmt.Errorf("failed to initialize event listener: %w", err)
	}

	if err := a.initGater(); err != nil {
		return fmt.Errorf("failed to initialize gater: %w", err)
	}

	if err := a.initNode(); err != nil {
		return fmt.Errorf("failed to initialize node: %w", err)
	}

	if err := a.initEpochUpdator(); err != nil {
		return fmt.Errorf("failed to initialize epoch updator: %w", err)
	}

	if err := a.initTaskProcessor(); err != nil {
		return fmt.Errorf("failed to initialize task processor: %w", err)
	}

	if err := a.initAPI(); err != nil {
		return fmt.Errorf("failed to initialize API: %w", err)
	}

	if err := a.initHealthChecker(); err != nil {
		return fmt.Errorf("failed to initialize health checker: %w", err)
	}

	metric.RecordRequest("operator", "startup_complete")
	log.Printf("Operator node started successfully")

	// Wait forever
	select {}
}

// shutdown gracefully stops the application
func (a *App) Shutdown() error {
	if a.db != nil {
		a.db.Close()
	}
	if a.redisConn != nil {
		a.redisConn.Close()
	}
	if a.node != nil {
		a.node.Stop(a.ctx)
	}
	if a.chainManager != nil {
		a.chainManager.Close()
	}
	return nil
}

// initConfig loads application configuration
func (a *App) initConfig() error {
	var err error
	a.cfg, err = config.LoadConfig("config/operator.yaml")
	if err != nil {
		metric.RecordError("config_load_failed")
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

// initSigner initializes the signing service
func (a *App) initSigner() error {
	var cfg *signer.Config
	if a.isKeyPath {
		cfg = &signer.Config{
			SigningKeyPath: a.signingKeyPath,
			Password:       a.password,
		}
	} else {
		cfg = &signer.Config{
			SigningKey: a.signingKeyPriv,
		}
	}

	var err error
	a.signer, err = signer.NewLocalSigner(cfg)
	if err != nil {
		metric.RecordError("signer_init_failed")
		return fmt.Errorf("failed to initialize signer: %w", err)
	}
	return nil
}

// initMetrics starts the metrics server
func (a *App) initMetrics() error {
	a.metricServer = metric.New(&metric.Config{
		Port: a.cfg.Metric.Port,
	})
	go func() {
		if err := a.metricServer.Start(); err != nil {
			metric.RecordError("metric_server_start_failed")
			log.Printf("Failed to start metric server: %v", err)
		}
	}()
	return nil
}

// initDatabase connects to postgres and redis
func (a *App) initDatabase() error {
	var err error
	a.db, err = wpgxInitor.NewWPGXPool(a.ctx, "POSTGRES")
	if err != nil {
		metric.RecordError("database_connection_failed")
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	a.redisConn, a.dCache, err = cache.InitCache("operator")
	if err != nil {
		metric.RecordError("cache_init_failed")
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Initialize all repos
	a.repos = &Repos{
		OperatorRepo:          operators.New(a.db.WConn(), a.dCache),
		BlacklistRepo:         blacklist.New(a.db.WConn(), a.dCache),
		TaskRepo:              tasks.New(a.db.WConn(), a.dCache),
		ConsensusResponseRepo: consensus_responses.New(a.db.WConn(), a.dCache),
	}

	return nil
}

// initChainManagerAndMainnetClient sets up blockchain connections
func (a *App) initChainManagerAndMainnetClient() error {

	chainManager, err := ethereum.NewManager(a.cfg)
	if err != nil {
		metric.RecordError("chain_manager_init_failed")
		return fmt.Errorf("failed to initialize chain manager: %w", err)
	}

	mainnetClient, err := chainManager.GetMainnetClient()
	if err != nil {
		metric.RecordError("mainnet_client_creation_failed")
		return fmt.Errorf("failed to create mainnet client: %w", err)
	}

	a.chainManager = chainManager
	a.mainnetClient = mainnetClient
	return nil
}

// initEventListener creates contract event listener
func (a *App) initEventListener() error {
	_, err := event.NewEventListener(a.ctx, &event.Config{
		MainnetClient: a.mainnetClient,
		OperatorRepo:  a.repos.OperatorRepo,
	})
	if err != nil {
		metric.RecordError("event_listener_creation_failed")
		return fmt.Errorf("failed to create event listener: %w", err)
	}
	return nil
}

// initGater creates p2p connection gater
func (a *App) initGater() error {
	gater, err := gater.NewConnectionGater(&gater.Config{
		BlacklistRepo: a.repos.BlacklistRepo,
		OperatorRepo:  a.repos.OperatorRepo,
	})
	if err != nil {
		metric.RecordError("gater_creation_failed")
		return fmt.Errorf("failed to create gater: %w", err)
	}
	a.gater = gater
	return nil
}

// initNode sets up p2p networking
func (a *App) initNode() error {
	privKey, err := signer.Base64ToPrivKey(a.p2pKey)
	if err != nil {
		metric.RecordError("p2p_key_load_failed")
		return fmt.Errorf("failed to load P2P key: %w", err)
	}

	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", a.cfg.P2P.Port),
		),
		libp2p.EnableNATService(),
		libp2p.NATPortMap(),
		libp2p.ConnectionGater(a.gater),
		libp2p.Identity(privKey),
	)
	if err != nil {
		metric.RecordError("host_creation_failed")
		return fmt.Errorf("failed to create host: %w", err)
	}
	a.host = host
	bootstrapPeers, err := a.cfg.P2P.GetBootstrapPeers()
	if err != nil {
		metric.RecordError("bootstrap_peers_creation_failed")
		return fmt.Errorf("failed to create bootstrap peers: %w", err)
	}
	a.node, err = node.NewNode(a.ctx, &node.Config{
		Host:             host,
		BootstrapPeers:   bootstrapPeers,
		RendezvousString: a.cfg.P2P.Rendezvous,
	})
	if err != nil {
		metric.RecordError("node_creation_failed")
		return fmt.Errorf("failed to create operator node: %w", err)
	}

	return nil
}

// initEpochUpdator creates epoch state manager
func (a *App) initEpochUpdator() error {

	epochUpdator, err := epoch.NewEpochUpdator(a.ctx, &epoch.Config{
		OperatorRepo:  a.repos.OperatorRepo,
		MainnetClient: a.mainnetClient,
		TxManager:     a.db,
	})
	if err != nil {
		metric.RecordError("epoch_updator_creation_failed")
		return fmt.Errorf("failed to create epoch updator: %w", err)
	}
	a.epochUpdator = epochUpdator
	return nil
}

// initTaskProcessor creates and starts task processor
func (a *App) initTaskProcessor() error {
	gossipOpts := []pubsub.Option{
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(true),
		pubsub.WithPeerExchange(true),
	}

	ps, err := pubsub.NewGossipSub(a.ctx, a.host, gossipOpts...)
	if err != nil {
		return fmt.Errorf("failed to create gossipsub: %w", err)
	}

	// Create response topic
	responseTopic, err := ps.Join(task.TaskResponseTopic)
	if err != nil {
		log.Fatalf("[Main] Failed to join response topic: %v", err)
	}
	log.Printf("[Main] Joined response topic: %s", task.TaskResponseTopic)

	// Subscribe to response topic
	responseSubscription, err := responseTopic.Subscribe()
	if err != nil {
		log.Fatalf("[Main] Failed to subscribe to response topic: %v", err)
	}
	log.Printf("[Main] Subscribed to response topic")

	// Initialize task processor
	taskProcessor, err := task.NewTaskProcessor(&task.Config{
		ChainManager:          a.chainManager,
		Signer:                a.signer,
		EpochStateQuerier:     a.epochUpdator,
		ConsensusResponseRepo: a.repos.ConsensusResponseRepo,
		BlacklistRepo:         a.repos.BlacklistRepo,
		TaskRepo:              a.repos.TaskRepo,
		OperatorRepo:          a.repos.OperatorRepo,
		ResponseTopic:         responseTopic,
	})
	if err != nil {
		return fmt.Errorf("failed to create task processor: %w", err)
	}

	// 显式启动processor
	if err := taskProcessor.Start(a.ctx, responseSubscription); err != nil {
		return fmt.Errorf("failed to start task processor: %w", err)
	}

	a.taskProcessor = taskProcessor
	return nil
}

// initAPI starts the HTTP API server
func (a *App) initAPI() error {
	apiHandler, err := api.NewHandler(api.Config{
		TaskRepo:              a.repos.TaskRepo,
		ChainManager:          a.chainManager,
		ConsensusResponseRepo: a.repos.ConsensusResponseRepo,
		TaskProcessor:         a.taskProcessor,
		Config:                a.cfg,
	})
	if err != nil {
		metric.RecordError("api_handler_creation_failed")
		return fmt.Errorf("failed to create API handler: %w", err)
	}

	a.apiServer = api.NewServer(apiHandler, a.cfg.HTTP.Port)
	go func() {
		if err := a.apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	return nil
}

// initHealthChecker creates health monitoring service
func (a *App) initHealthChecker() error {
	pingService := ping.NewPingService(a.host)
	healthService, err := health.NewHealthChecker(a.ctx, a.node, pingService)
	if err != nil {
		metric.RecordError("health_checker_creation_failed")
		return fmt.Errorf("failed to create health checker: %w", err)
	}
	defer healthService.Stop()

	status := healthService.GetStatus()
	log.Printf("[Main] Health status: %v", status)

	healthService.SetCheckInterval(30 * time.Second)

	healthService.TriggerCheck(a.ctx)
	return nil
}
