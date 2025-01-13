package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/registry"
)

type JoinRequest struct {
	Address   string `json:"address"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type JoinResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	HostID  string `json:"host_id,omitempty"`
}

func main() {
	// Create context that listens for the interrupt signal from the OS
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channel for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start HTTP server first for host ID endpoint
	httpServer := &http.Server{
		Addr: ":8000",
	}
	var node *registry.Node

	// Add HTTP endpoint to expose host ID
	http.HandleFunc("/p2p/id", func(w http.ResponseWriter, r *http.Request) {
		if node != nil {
			fmt.Fprintf(w, "%s", node.GetHostID())
		}
	})

	// Add endpoint for join requests
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse request
		var req JoinRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		// Decode signature
		signature, err := hex.DecodeString(req.Signature)
		if err != nil {
			http.Error(w, "Invalid signature format", http.StatusBadRequest)
			return
		}

		// Verify signature
		address := common.HexToAddress(req.Address)
		message := []byte(req.Message)
		if !signer.VerifySignature(message, signature, address) {
			response := JoinResponse{
				Success: false,
				Error:   "Invalid signature",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return success with host ID
		response := JoinResponse{
			Success: true,
			HostID:  node.GetHostID(),
		}
		json.NewEncoder(w).Encode(response)
	})

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Create p2p config
	cfg := &p2p.Config{
		ListenAddrs: []string{
			"/ip4/0.0.0.0/tcp/9000",
		},
	}

	// Create and start registry node
	var err error
	node, err = registry.NewNode(ctx, cfg)
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

	// Graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}
} 