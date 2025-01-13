package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/operator"
)

func main() {
	// Define flags
	registryAddr := flag.String("registry", "", "Registry node address")
	keystorePath := flag.String("keystore", "", "Path to keystore file")
	password := flag.String("password", "", "Password for keystore")
	message := flag.String("message", "", "Message to sign")
	sign := flag.Bool("sign", false, "Sign a message")
	getAddress := flag.Bool("address", false, "Get address from keystore")

	flag.Parse()

	if *sign {
		if *keystorePath == "" || *password == "" || *message == "" {
			log.Fatal("keystore, password and message are required for signing")
		}

		s, err := signer.NewLocalSigner(*keystorePath, *password)
		if err != nil {
			log.Fatalf("failed to create signer: %v", err)
		}

		sig, err := s.Sign([]byte(*message))
		if err != nil {
			log.Fatalf("failed to sign message: %v", err)
		}

		fmt.Print(hex.EncodeToString(sig))
		return
	}

	if *getAddress {
		if *keystorePath == "" || *password == "" {
			log.Fatal("keystore and password are required for getting address")
		}

		s, err := signer.NewLocalSigner(*keystorePath, *password)
		if err != nil {
			log.Fatalf("failed to create signer: %v", err)
		}

		addr := s.GetAddress()
		fmt.Print(addr.Hex())
		return
	}

	if *registryAddr == "" {
		log.Fatal("registry address is required")
	}

	if *keystorePath == "" || *password == "" {
		log.Fatal("keystore and password are required")
	}

	s, err := signer.NewLocalSigner(*keystorePath, *password)
	if err != nil {
		log.Fatalf("failed to create signer: %v", err)
	}

	n, err := operator.NewNode(*registryAddr, s)
	if err != nil {
		log.Fatalf("failed to create node: %v", err)
	}

	if err := n.Start(); err != nil {
		log.Fatalf("failed to start node: %v", err)
	}

	select {}
} 