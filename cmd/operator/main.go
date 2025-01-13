package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/galxe/spotted-network/pkg/operator"
)

func main() {
	// Parse command line flags
	registryAddr := flag.String("registry", "", "Registry node multiaddr")
	flag.Parse()

	if *registryAddr == "" {
		log.Fatal("Registry address is required")
	}

	// Create and start operator node
	node, err := operator.NewNode(*registryAddr)
	if err != nil {
		log.Fatalf("Failed to create operator node: %v", err)
	}

	// Start the node
	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start operator node: %v", err)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Stop the node
	if err := node.Stop(); err != nil {
		log.Printf("Error stopping node: %v", err)
	}
} 