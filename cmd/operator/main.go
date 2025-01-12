package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/galxe/spotted-network/pkg/operator"
	"github.com/galxe/spotted-network/pkg/p2p"
)

func main() {
	// Parse command line flags
	registryAddr := flag.String("registry", "", "Registry node multiaddr")
	flag.Parse()

	if *registryAddr == "" {
		log.Fatal("Registry address is required")
	}

	// Create context that listens for the interrupt signal from the OS
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channel for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Create p2p config
	cfg := &p2p.Config{
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/0",
		},
	}

	// Create and start operator node
	node, err := operator.NewNode(ctx, cfg, *registryAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	log.Printf("Operator node started, connecting to registry at %s", *registryAddr)

	// Wait for interrupt signal
	select {
	case <-sigChan:
		log.Println("Received interrupt signal, shutting down...")
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}
} 