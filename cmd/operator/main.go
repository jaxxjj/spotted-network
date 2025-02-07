package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/galxe/spotted-network/internal/database/cache"
	dbwpgx "github.com/galxe/spotted-network/internal/database/wpgx"
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
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

func main() {
	startTime := time.Now()
	defer func() {
		metric.RecordRequestDuration("main", "startup", time.Since(startTime))
	}()

	signingKeyPath := flag.String("signing-key", "", "Path to signing keystore file")
	password := flag.String("password", "", "Password for keystore")

	flag.Parse()

	// Create signer
	signer, err := signer.NewLocalSigner(*signingKeyPath, *password)
	if err != nil {
		metric.RecordError("signer_creation_failed")
		log.Fatal("Failed to create signer:", err)
	}

	// Get config path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/operator.yaml" // default value
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)

	if err != nil {
		metric.RecordError("config_load_failed")
		log.Fatal("Failed to load config:", err)
	}

	// Initialize metrics server
	metricServer := metric.New(&metric.Config{
		Port: cfg.Metric.Port,
	})
	go func() {
		if err := metricServer.Start(); err != nil {
			metric.RecordError("metric_server_start_failed")
			log.Printf("Failed to start metric server: %v", err)
		}
	}()

	// Initialize chain manager with application config
	chainManager, err := ethereum.NewManager(cfg)
	if err != nil {
		metric.RecordError("chain_manager_init_failed")
		log.Fatal("Failed to initialize chain manager:", err)
	}
	defer chainManager.Close()

	// Initialize database connection
	ctx := context.Background()
	db, err := dbwpgx.NewWPGXPool(ctx, "POSTGRES")
	if err != nil {
		metric.RecordError("database_connection_failed")
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis and DCache
	redisConn, dCache, err := cache.InitCache("operator")
	if err != nil {
		metric.RecordError("cache_init_failed")
		log.Fatal("Failed to initialize cache:", err)
	}
	defer redisConn.Close()

	// Initialize database queries
	operatorRepo := operators.New(db.WConn(), dCache)
	tasksRepo := tasks.New(db.WConn(), dCache)
	consensusResponseRepo := consensus_responses.New(db.WConn(), dCache)
	blacklistRepo := blacklist.New(db.WConn(), dCache)

	mainnetClient, err := chainManager.GetMainnetClient()
	if err != nil {
		metric.RecordError("mainnet_client_creation_failed")
		log.Fatal("Failed to create mainnet client:", err)
	}
	_, err = event.NewEventListener(ctx, &event.Config{
		MainnetClient: mainnetClient,
		OperatorRepo:  operatorRepo,
	})

	if err != nil {
		metric.RecordError("event_listener_creation_failed")
		log.Fatal("Failed to create event listener:", err)
	}
	// Create gate
	gater, err := gater.NewConnectionGater(&gater.Config{
		BlacklistRepo: blacklistRepo,
		OperatorRepo:  operatorRepo,
	})
	if err != nil {
		metric.RecordError("gater_creation_failed")
		log.Fatal("Failed to create gater:", err)
	}

	// Use gater to create host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/%s/tcp/%d", cfg.P2P.ExternalIP, cfg.P2P.Port),
		),
		libp2p.ConnectionGater(gater),
	)
	if err != nil {
		metric.RecordError("host_creation_failed")
		log.Fatal("Failed to create host:", err)
	}
	// Start operator node
	node, err := node.NewNode(ctx, &node.Config{
		Host:          host,
		BlacklistRepo: blacklistRepo,
	})
	defer node.Stop(ctx)
	if err != nil {
		metric.RecordError("node_creation_failed")
		log.Fatal("Failed to create operator node:", err)
	}

	epochUpdator, err := epoch.NewEpochUpdator(ctx, &epoch.Config{
		OperatorRepo:  operatorRepo,
		MainnetClient: mainnetClient,
		TxManager:     db,
	})
	if err != nil {
		metric.RecordError("epoch_updator_creation_failed")
		log.Fatal("Failed to create epoch updator:", err)
	}
	pubsub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		metric.RecordError("pubsub_creation_failed")
		log.Fatal("Failed to create pubsub:", err)
	}

	taskProcessor, err := task.NewTaskProcessor(&task.TaskProcessorConfig{
		ChainManager:          chainManager,
		Signer:                signer,
		ConsensusResponseRepo: consensusResponseRepo,
		EpochStateQuerier:     epochUpdator,
		BlacklistRepo:         blacklistRepo,
		TaskRepo:              tasksRepo,
		OperatorRepo:          operatorRepo,
		PubSub:                pubsub,
	})

	if err != nil {
		metric.RecordError("task_processor_creation_failed")
		log.Fatal("Failed to create task processor:", err)
	}

	pingService := ping.NewPingService(host)

	_, err = health.NewHealthChecker(ctx, node, pingService)
	if err != nil {
		metric.RecordError("health_checker_creation_failed")
		log.Fatal("Failed to create health checker:", err)
	}
	// Create API handler and server
	apiHandler, err := api.NewHandler(api.Config{
		TaskRepo:              tasksRepo,
		ChainManager:          chainManager,
		ConsensusResponseRepo: consensusResponseRepo,
		TaskProcessor:         taskProcessor,
		Config:                cfg,
	})
	if err != nil {
		metric.RecordError("api_handler_creation_failed")
		log.Fatal("Failed to create API handler:", err)
	}
	apiServer := api.NewServer(apiHandler, cfg.HTTP.Port)
	if err != nil {
		metric.RecordError("api_server_creation_failed")
		log.Fatal("Failed to create API server:", err)
	}
	metric.RecordRequest("operator", "startup_complete")
	log.Printf("Operator node started successfully")

	// Start API server
	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	// Wait forever
	select {}
}
