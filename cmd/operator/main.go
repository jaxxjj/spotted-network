package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/libp2p/go-libp2p"

	"github.com/galxe/spotted-network/internal/database/cache"
	dbwpgx "github.com/galxe/spotted-network/internal/database/wpgx"
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
	registryAddr := flag.String("registry", "", "Registry node address")
	operatorKeyPath := flag.String("operator-key", "", "Path to operator keystore file")
	signingKeyPath := flag.String("signing-key", "", "Path to signing keystore file")
	password := flag.String("password", "", "Password for keystore")
	message := flag.String("message", "", "Message to sign")

	flag.Parse()

	// Create signer
	signer, err := signer.NewLocalSigner(*operatorKeyPath, *signingKeyPath, *password)
	if err != nil {
		log.Fatal("Failed to create signer:", err)
	}

	if *message != "" {
		// Sign message with signing key
		sig, err := signer.Sign([]byte(*message))
		if err != nil {
			log.Fatal("Failed to sign message:", err)
		}
		fmt.Print(hex.EncodeToString(sig))
		return
	}

	if *registryAddr == "" {
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
		log.Fatal("Failed to load config:", err)
	}

	// Initialize chain manager with application config
	chainManager, err := ethereum.NewManager(cfg)
	if err != nil {
		log.Fatal("Failed to initialize chain manager:", err)
	}
	defer chainManager.Close()

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
	redisConn, dCache, err := cache.InitCache("operator")
	if err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}
	defer redisConn.Close()

	// Initialize database queries
	tasksQuerier := tasks.New(db.WConn(), dCache)
	taskResponseQuerier := task_responses.New(db.WConn(), dCache)
	consensusResponseQuerier := consensus_responses.New(db.WConn(), dCache)
	epochStatesQuerier := epoch_states.New(db.WConn(), dCache)

	// Start operator node
	n, err := operator.NewNode(&operator.NodeConfig{
		Host:            host,
		ChainManager:    chainManager,
		Signer:          signer,
		TasksQuerier:     tasksQuerier,
		TaskResponseQuerier: taskResponseQuerier,
		ConsensusResponseQuerier: consensusResponseQuerier,
		EpochStateQuerier: epochStatesQuerier,
		RegistryAddress: *registryAddr,
		Config:          cfg,
	})
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}

	if err := n.Start(context.Background()); err != nil {
		log.Fatal("Failed to start node:", err)
	}

	// Wait forever
	select {}
}

