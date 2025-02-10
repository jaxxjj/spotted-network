package main

import (
	"context"
	"flag"
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
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/redis/go-redis/v9"
	"github.com/stumble/dcache"
	"github.com/stumble/wpgx"
)

// Repos holds all repository instances
type Repos struct {
	OperatorRepo          *operators.Queries
	BlacklistRepo         *blacklist.Queries
	TaskRepo              *tasks.Queries
	ConsensusResponseRepo *consensus_responses.Queries
}

// App holds all the dependencies
type App struct {
	signingKeyPath *string
	p2pKey64       *string
	password       *string

	ctx    context.Context
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

func main() {
	startTime := time.Now()
	defer func() {
		metric.RecordRequestDuration("main", "startup", time.Since(startTime))
	}()

	ctx := context.Background()
	app := &App{ctx: ctx}

	app.parseFlags()
	if err := app.initConfig(); err != nil {
		log.Fatal("Failed to initialize config:", err)
	}

	if err := app.initMetrics(); err != nil {
		log.Fatal("Failed to initialize metrics:", err)
	}

	if err := app.initSigner(); err != nil {
		log.Fatal("Failed to initialize signer:", err)
	}

	if err := app.initDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer app.cleanup()

	if err := app.initChainManagerAndMainnetClient(); err != nil {
		log.Fatal("Failed to initialize chain manager and mainnet client:", err)
	}

	if err := app.initEventListener(); err != nil {
		log.Fatal("Failed to initialize event listener:", err)
	}

	if err := app.initGater(); err != nil {
		log.Fatal("Failed to initialize gater:", err)
	}

	if err := app.initNode(); err != nil {
		log.Fatal("Failed to initialize node:", err)
	}

	if err := app.initEpochUpdator(); err != nil {
		log.Fatal("Failed to initialize epoch updator:", err)
	}

	if err := app.initTaskProcessor(); err != nil {
		log.Fatal("Failed to initialize task processor:", err)
	}

	if err := app.initAPI(); err != nil {
		log.Fatal("Failed to initialize API:", err)
	}

	if err := app.initHealthChecker(); err != nil {
		log.Fatal("Failed to initialize health checker:", err)
	}

	metric.RecordRequest("operator", "startup_complete")
	log.Printf("Operator node started successfully")

	// Wait forever
	select {}
}

func (a *App) parseFlags() {
	signingKeyPath := flag.String("signing-key", "", "Path to signing keystore file")
	p2pKey64 := flag.String("p2p-key-64", "", "p2p key in base64")
	password := flag.String("password", "", "Password for keystore")
	flag.Parse()

	a.signingKeyPath = signingKeyPath
	a.p2pKey64 = p2pKey64
	a.password = password
}

func (a *App) initConfig() error {
	var err error
	a.cfg, err = config.LoadConfig("config/operator.yaml")
	if err != nil {
		metric.RecordError("config_load_failed")
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func (a *App) initSigner() error {
	signer, err := signer.NewLocalSigner(&signer.Config{
		SigningKeyPath: *a.signingKeyPath,
		Password:       *a.password,
	})
	if err != nil {
		metric.RecordError("signer_init_failed")
		return fmt.Errorf("failed to initialize signer: %w", err)
	}
	a.signer = signer
	return nil
}

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

func (a *App) initNode() error {
	privKey, err := signer.Base64ToPrivKey(*a.p2pKey64)
	if err != nil {
		metric.RecordError("p2p_key_load_failed")
		return fmt.Errorf("failed to load P2P key: %w", err)
	}

	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", a.cfg.P2P.Port),
			fmt.Sprintf("/ip4/%s/tcp/%d", a.cfg.P2P.ExternalIP, a.cfg.P2P.Port),
		),
		libp2p.ConnectionGater(a.gater),
		libp2p.Identity(privKey),
		libp2p.Security("/noise", noise.New),
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
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
		ResponseSubscription:  responseSubscription,
	})
	if err != nil {
		return fmt.Errorf("failed to create task processor: %w", err)
	}
	a.taskProcessor = taskProcessor
	return nil
}

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

func (a *App) initHealthChecker() error {
	pingService := ping.NewPingService(a.host)
	healthService, err := health.NewHealthChecker(a.ctx, a.node, pingService)
	if err != nil {
		metric.RecordError("health_checker_creation_failed")
		return fmt.Errorf("failed to create health checker: %w", err)
	}
	defer healthService.Stop()

	// 获取健康状态
	status := healthService.GetStatus()
	log.Printf("[Main] Health status: %v", status)

	healthService.SetCheckInterval(30 * time.Second)

	// 手动触发检查
	healthService.TriggerCheck(a.ctx)
	return nil
}

func (a *App) cleanup() {
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
}
