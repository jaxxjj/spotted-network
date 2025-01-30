package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"

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
	getRegistryID := flag.Bool("get-registry-id", false, "Get registry ID")
	join := flag.Bool("join", false, "Join the network")

	flag.Parse()

	if *getRegistryID {
		client, err := operator.NewRegistryClient("registry:8000")
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := client.GetRegistryID(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(id)
		return
	}

	// Create signer
	signer, err := signer.NewLocalSigner(*operatorKeyPath, *signingKeyPath, *password)
	if err != nil {
		log.Fatal("Failed to create signer:", err)
	}

	if *join {
		client, err := operator.NewRegistryClient("registry:8000")
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		addr := signer.GetOperatorAddress()
		sig, err := signer.SignJoinRequest([]byte(*message))
		if err != nil {
			log.Fatal("Failed to sign message:", err)
		}

		success, err := client.Join(ctx, addr.Hex(), *message, hex.EncodeToString(sig), signer.GetSigningAddress().Hex())
		if err != nil {
			log.Fatal(err)
		}
		if !success {
			log.Fatal("Join request failed")
		}
		return
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

	// Set database environment variables for wpgx
	os.Setenv("POSTGRES_CONN_STRING", cfg.Database.URL)
	os.Setenv("POSTGRES_MAX_CONNS", fmt.Sprintf("%d", cfg.Database.MaxOpenConns))
	os.Setenv("POSTGRES_MIN_CONNS", fmt.Sprintf("%d", cfg.Database.MaxIdleConns))
	os.Setenv("POSTGRES_MAX_CONN_LIFETIME", cfg.Database.ConnMaxLifetime.String())

	// Initialize database connection
	ctx := context.Background()
	db, err := dbwpgx.NewWPGXPool(ctx, "POSTGRES")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize database tables
	if err := initDatabase(ctx, db); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Initialize database queries
	taskQuerier := tasks.New(db)
	taskResponseQuerier := task_responses.New(db)
	consensusResponseQuerier := consensus_responses.New(db)
	epochStatesQuerier := epoch_states.New(db)

	// Start operator node
	n, err := operator.NewNode(&operator.NodeConfig{
		Host:            host,
		DB:              db,
		ChainManager:    chainManager,
		Signer:          signer,
		TaskQuerier:     taskQuerier,
		TaskResponseQuerier: taskResponseQuerier,
		ConsensusResponseQuerier: consensusResponseQuerier,
		EpochState:      epochStatesQuerier,
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

func initDatabase(ctx context.Context, db *dbwpgx.Pool) error {
	// Check if database is accessible
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("[Node] failed to ping database: %w", err)
	}
	log.Printf("[Node] Successfully connected to database")
	return nil
}