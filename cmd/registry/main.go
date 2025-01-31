package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/registry"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/libp2p/go-libp2p"

	"github.com/galxe/spotted-network/internal/database/cache"
	dbwpgx "github.com/galxe/spotted-network/internal/database/wpgx"
)

type ChainManager interface {
	GetMainnetClient() (*ethereum.ChainClient, error)
}

func main() {
	// Get config path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/registry.yaml"  // default value
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create P2P host
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/%s/tcp/%d", cfg.P2P.ExternalIP, cfg.P2P.Port),
		),
	)
	if err != nil {
		log.Fatal("Failed to create P2P host:", err)
	}
	defer host.Close()

	// Initialize database connection
	ctx := context.Background()
	db, err := dbwpgx.NewWPGXPool(ctx, "POSTGRES")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis and DCache
	redisConn, dCache, err := cache.InitCache("registry")
	if err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}
	defer redisConn.Close()

	// Create database queries
	operatorsQuerier := operators.New(db.WConn(), dCache)
	chainManager, err := ethereum.NewManager(cfg)
	if err != nil {
		log.Fatal("Failed to initialize chain manager:", err)
	}
	defer chainManager.Close()
	mainnetClient, err := chainManager.GetMainnetClient()
	if err != nil {
		log.Fatal("Failed to initialize mainnet client:", err)
	}

	// Create Registry Node
	node, err := registry.NewNode(ctx, &registry.NodeConfig{
		Host:          host,
		Operators:     operatorsQuerier,
		MainnetClient: mainnetClient,
	})
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}
	defer node.Stop()

	// Start the node
	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start registry node: %v", err)
	}
		
	// Wait forever
	select {}
} 