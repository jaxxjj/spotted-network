package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/internal/avs/config"
	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/p2p"
	registrynode "github.com/galxe/spotted-network/pkg/registry"
	"github.com/galxe/spotted-network/pkg/repos/registry"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type registryServer struct {
	pb.UnimplementedRegistryServer
	node *registrynode.Node
}

func (s *registryServer) GetRegistryID(ctx context.Context, req *pb.GetRegistryIDRequest) (*pb.GetRegistryIDResponse, error) {
	return &pb.GetRegistryIDResponse{
		RegistryId: s.node.GetHostID(),
	}, nil
}

func (s *registryServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	// Get operator status from database
	op, err := s.node.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		return &pb.JoinResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get operator: %v", err),
		}, nil
	}

	// Verify operator is in waitingJoin status
	if op.Status != string(registrynode.OperatorStatusWaitingJoin) {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Operator not in waitingJoin status",
		}, nil
	}

	// Verify signing key matches
	if op.SigningKey != req.SigningKey {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Signing key mismatch",
		}, nil
	}

	// Decode signature
	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Invalid signature format",
		}, nil
	}

	// Verify signature
	addr := common.HexToAddress(req.Address)
	message := []byte(req.Message)
	if !signer.VerifySignature(addr, message, sigBytes) {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Invalid signature",
		}, nil
	}

	// Update operator status to waitingActive
	if err := s.node.UpdateOperatorStatus(ctx, req.Address, string(registrynode.OperatorStatusWaitingActive)); err != nil {
		return &pb.JoinResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update operator status: %v", err),
		}, nil
	}

	// Get list of active operators
	activeOperators := make([]*pb.ActiveOperator, 0)
	for _, peerID := range s.node.GetConnectedOperators() {
		info := s.node.GetOperatorInfo(peerID)
		if info != nil && info.Status == string(registrynode.OperatorStatusActive) {
			addrs := make([]string, len(info.Addrs))
			for i, addr := range info.Addrs {
				addrs[i] = addr.String()
			}
			activeOperators = append(activeOperators, &pb.ActiveOperator{
				PeerId: peerID.String(),
				Multiaddrs: addrs,
			})
		}
	}

	return &pb.JoinResponse{
		Success: true,
		ActiveOperators: activeOperators,
	}, nil
}

func main() {
	ctx := context.Background()

	// Get ports from environment
	p2pPort := os.Getenv("P2P_PORT")
	if p2pPort == "" {
		p2pPort = "9000"
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8000"
	}

	// Initialize database connection
	dbConn, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Create database queries
	queries := registry.New(dbConn)

	// Load chain configuration
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
		},
	}

	// Initialize chain clients
	chainClients, err := ethereum.NewChainClients(chainConfig)
	if err != nil {
		log.Fatalf("Failed to initialize chain clients: %v", err)
	}
	defer chainClients.Close()

	// Create p2p host configuration
	p2pConfig := &p2p.Config{
		ListenAddrs: []string{fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", p2pPort)},
	}
	
	// Create Registry Node
	node, err := registrynode.NewNode(ctx, p2pConfig, queries, chainClients)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", httpPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServer(grpcServer, &registryServer{node: node})

	log.Printf("Registry gRPC server listening on :%s", httpPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
} 