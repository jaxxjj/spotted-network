package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/p2p"
	registrynode "github.com/galxe/spotted-network/pkg/registry"
	"github.com/galxe/spotted-network/pkg/registry/server"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type ChainManager interface {
	GetMainnetClient() (*ethereum.ChainClient, error)
}

func main() {
	ctx := context.Background()

	// Get config path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/registry.yaml"  // default value
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database connection
	dbConn, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Create database queries
	operatorsQuerier := operators.New(dbConn)

	chainManager, err := ethereum.NewManager(cfg)
	if err != nil {
		log.Fatal("Failed to initialize chain manager:", err)
	}
	defer chainManager.Close()
	mainnetClient, err := chainManager.GetMainnetClient()
	if err != nil {
		log.Fatal("Failed to initialize mainnet client:", err)
	}


	// Create p2p host configuration
	p2pConfig := &p2p.Config{
		ListenAddrs: []string{fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.P2P.Port)},
	}
	
	// Create Registry Node
	node, err := registrynode.NewNode(ctx, p2pConfig, operatorsQuerier, mainnetClient)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	// Start the node
	if err := node.Start(ctx); err != nil {
		log.Fatalf("Failed to start registry node: %v", err)
	}

	// Create Registry Server
	registryServer := server.NewRegistryServer(node, operatorsQuerier)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HTTP.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServer(grpcServer, registryServer)

	log.Printf("Registry gRPC server listening on :%d", cfg.HTTP.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
} 