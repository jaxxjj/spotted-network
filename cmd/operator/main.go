package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/operator"
)

func main() {
	registryAddr := flag.String("registry", "", "Registry node address")
	keystorePath := flag.String("keystore", "", "Path to keystore file")
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

		success, err := client.Join(ctx, addr.Hex(), *message, hex.EncodeToString(sig))
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

	// Start operator node
	n, err := operator.NewNode(*registryAddr, s)
	if err != nil {
		log.Fatal("Failed to create node:", err)
	}

	if err := n.Start(); err != nil {
		log.Fatal("Failed to start node:", err)
	}

	// Wait forever
	select {}
} 