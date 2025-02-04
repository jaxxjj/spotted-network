package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/galxe/spotted-network/internal/database/cache"
	dbwpgx "github.com/galxe/spotted-network/internal/database/wpgx"
	"github.com/galxe/spotted-network/internal/metric"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/operator"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
)

func main() {
	startTime := time.Now()
	defer func() {
		metric.RecordRequestDuration("main", "startup", time.Since(startTime))
	}()

	registryAddress := flag.String("registry", "", "Registry node address")
	operatorKeyPath := flag.String("operator-key", "", "Path to operator keystore file")
	signingKeyPath := flag.String("signing-key", "", "Path to signing keystore file")
	password := flag.String("password", "", "Password for keystore")
	message := flag.String("message", "", "Message to sign")

	flag.Parse()

	// Create signer
	signer, err := signer.NewLocalSigner(*operatorKeyPath, *signingKeyPath, *password)
	if err != nil {
		metric.RecordError("signer_creation_failed")
		log.Fatal("Failed to create signer:", err)
	}

	if *message != "" {
		// Sign message with signing key
		sig, err := signer.Sign([]byte(*message))
		if err != nil {
			metric.RecordError("message_signing_failed")
			log.Fatal("Failed to sign message:", err)
		}
		fmt.Print(hex.EncodeToString(sig))
		return
	}

	if *registryAddress == "" {
		metric.RecordError("missing_registry_address")
		log.Fatal("Registry address is required")
	}

	// Get config path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/operator.yaml"  // default value
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
	tasksQuerier := tasks.New(db.WConn(), dCache)
	taskResponseQuerier := task_responses.New(db.WConn(), dCache)
	consensusResponseQuerier := consensus_responses.New(db.WConn(), dCache)
	epochStatesQuerier := epoch_states.New(db.WConn(), dCache)

	// Start operator node
	node, err := operator.NewNode(ctx, &operator.NodeConfig{
		ChainManager:    chainManager,
		Signer:          signer,
		TasksQuerier:     tasksQuerier,
		TaskResponseQuerier: taskResponseQuerier,
		ConsensusResponseQuerier: consensusResponseQuerier,
		EpochStateQuerier: epochStatesQuerier,
		RegistryAddress: *registryAddress,
		Config:          cfg,
	})
	defer node.Stop(ctx)
	if err != nil {
		metric.RecordError("node_creation_failed")
		log.Fatal("Failed to create operator node:", err)
	}

	// Start the node
	if err := node.Start(ctx); err != nil {
		metric.RecordError("node_start_failed")
		log.Fatal("Failed to start operator node:", err)
	}

	metric.RecordRequest("operator", "startup_complete")
	log.Printf("Operator node started successfully")

	// Wait forever
	select {}
}

