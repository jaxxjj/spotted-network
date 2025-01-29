package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/operator"
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
	s, err := signer.NewLocalSigner(*operatorKeyPath, *signingKeyPath, *password)
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

		addr := s.GetOperatorAddress()
		sig, err := s.SignJoinRequest([]byte(*message))
		if err != nil {
			log.Fatal("Failed to sign message:", err)
		}

		success, err := client.Join(ctx, addr.Hex(), *message, hex.EncodeToString(sig), s.GetSigningAddress().Hex())
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
		sig, err := s.Sign([]byte(*message))
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

	// Start operator node

	n, err := operator.NewNode(*registryAddr, cfg, chainManager, s)
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}

	if err := n.Start(context.Background()); err != nil {
		log.Fatal("Failed to start node:", err)
	}

	// Wait forever
	select {}
} 