package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/registry"
)

func main() {
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
			"/ip4/0.0.0.0/tcp/9000",
		},
	}

	// Create and start registry node
	node, err := registry.NewNode(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	// Wait for interrupt signal
	select {
	case <-sigChan:
		log.Println("Received interrupt signal, shutting down...")
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}
} 