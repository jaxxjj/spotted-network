package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/galxe/spotted-network/internal/avs/config"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/operator"
)

func main() {
	configPath := flag.String("config", "config/operator.yaml", "Path to config file")
	registryAddr := flag.String("registry", "", "Registry node address")
	keystorePath := flag.String("keystore", "", "Path to keystore file")
	password := flag.String("password", "", "Password for keystore")
	message := flag.String("message", "", "Message to sign")
	getRegistryID := flag.Bool("get-registry-id", false, "Get registry ID")
	join := flag.Bool("join", false, "Join the network")
	databaseURL := flag.String("database-url", "", "Database URL")

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
	s, err := signer.NewLocalSigner(*keystorePath, *password)
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

		addr := s.GetAddress()
		sig, err := s.Sign([]byte(*message))
		if err != nil {
			log.Fatal("Failed to sign message:", err)
		}

		success, err := client.Join(ctx, addr.Hex(), *message, hex.EncodeToString(sig), s.GetSigningKey())
		if err != nil {
			log.Fatal(err)
		}
		if !success {
			log.Fatal("Join request failed")
		}
		return
	}

	if *message != "" {
		// Sign message
		sig, err := s.Sign([]byte(*message))
		if err != nil {
			log.Fatal("Failed to sign message:", err)
		}
		fmt.Print(hex.EncodeToString(sig))
		return
	}

	if *registryAddr == "" {
		// Get address
		addr := s.GetAddress()
		fmt.Print(addr.Hex())
		return
	}

	// Load config
	cfg, err := operator.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Print database connection info for debugging
	log.Printf("Config file database URL: %s", cfg.Database.URL)

	// Override database URL if provided
	if *databaseURL != "" {
		log.Printf("Overriding database URL from command line")
		cfg.Database.URL = *databaseURL
	}

	log.Printf("Final database URL: %s", cfg.Database.URL)

	// Initialize chain clients with environment variables
	chainConfig := &config.Config{
		Chains: map[string]*config.ChainConfig{
			"ethereum": {
				RPC: os.Getenv("CHAIN_RPC_URL"),
				Contracts: config.ContractConfig{
					Registry:     os.Getenv("REGISTRY_ADDRESS"),
					EpochManager: os.Getenv("EPOCH_MANAGER_ADDRESS"),
					StateManager: os.Getenv("STATE_MANAGER_ADDRESS"),
				},
			},
			"31337": {
				RPC: os.Getenv("CHAIN_RPC_URL"),
				Contracts: config.ContractConfig{
					Registry:     os.Getenv("REGISTRY_ADDRESS"),
					EpochManager: os.Getenv("EPOCH_MANAGER_ADDRESS"),
					StateManager: os.Getenv("STATE_MANAGER_ADDRESS"),
				},
			},
		},
	}

	chainClients, err := ethereum.NewChainClients(chainConfig)
	if err != nil {
		log.Fatal("Failed to initialize chain clients:", err)
	}
	defer chainClients.Close()

	// Start operator node
	n, err := operator.NewNode(*registryAddr, s, cfg, chainClients)
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}

	if err := n.Start(); err != nil {
		log.Fatal("Failed to start node:", err)
	}

	// Wait forever
	select {}
} 