package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/registry"
	"github.com/galxe/spotted-network/pkg/registry/gater"
	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/rs/zerolog/log"

	"github.com/galxe/spotted-network/internal/database/cache"
	dbwpgx "github.com/galxe/spotted-network/internal/database/wpgx"
	"github.com/galxe/spotted-network/internal/metric"
	zlog "github.com/galxe/spotted-network/internal/zerolog"
)

type ChainManager interface {
	GetMainnetClient() (*ethereum.ChainClient, error)
}

func main() {
	startTime := time.Now()
	defer func() {
		metric.RecordRequestDuration("main", "startup", time.Since(startTime))
	}()

	// Get config path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/registry.yaml"  // default value
	}

	// Initialize logger first
	isDebug := os.Getenv("LOG_LEVEL") == "debug"
	zlog.InitLogger(isDebug)
	
	// Create context with logger
	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)
	
	log.Info().Str("component", "registry").Msg("Registry node starting...")

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		metric.RecordError("config_load_failed")
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to load config")
	}

	// Initialize metrics server
	metricServer := metric.New(&metric.Config{
		Port: cfg.Metric.Port,
	})
	go func() {
		if err := metricServer.Start(); err != nil {
			log.Error().Err(err).Str("component", "registry").Msg("Failed to start metric server")
		}
	}()

	// Initialize database connection
	db, err := dbwpgx.NewWPGXPool(ctx, "POSTGRES")
	if err != nil {
		metric.RecordError("database_connection_failed")
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to connect to database")
	}
	defer db.Close()

	// Initialize Redis and DCache
	redisConn, dCache, err := cache.InitCache("registry")
	if err != nil {
		metric.RecordError("cache_init_failed")
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to initialize cache")
	}
	defer redisConn.Close()

	// Create database queries
	operatorsQuerier := operators.New(db.WConn(), dCache)
	blacklistQuerier := blacklist.New(db.WConn(), dCache)

	gater := gater.NewOperatorGater(ctx, blacklistQuerier)
	// Create P2P host
	host, err := libp2p.New(
		libp2p.ConnectionGater(gater),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/%s/tcp/%d", cfg.P2P.ExternalIP, cfg.P2P.Port),
		),
	)
	if err != nil {
		metric.RecordError("p2p_host_creation_failed")
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to create P2P host")
	}
	defer host.Close()

	// Initialize PubSub
	pubsub, err := pubsub.NewGossipSub(ctx, host,
		pubsub.WithBlacklist(pubsub.NewMapBlacklist()),
	)
	if err != nil {
		metric.RecordError("pubsub_creation_failed")
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to create pubsub")
	}



	chainManager, err := ethereum.NewManager(cfg)
	if err != nil {
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to initialize chain manager")
	}
	defer chainManager.Close()
	mainnetClient, err := chainManager.GetMainnetClient()
	if err != nil {
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to initialize mainnet client")
	}

	// Create Registry Node
	node, err := registry.NewNode(ctx, &registry.NodeConfig{
		Host:          host,
		OperatorsQuerier:     operatorsQuerier,
		MainnetClient: mainnetClient,
		PubSub:        pubsub,
		TxManager:     db,
	})
	if err != nil {
		log.Fatal().Err(err).Str("component", "registry").Msg("Failed to create node")
	}
	defer node.Stop()

	metric.RecordRequest("registry", "startup_complete")
	log.Info().Str("component", "registry").Msg("Registry node started successfully")

	// Wait forever
	select {}
} 