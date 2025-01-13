package main

import (
	"context"
	"encoding/hex"
	"log"
	"net"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/registry"
	pb "github.com/galxe/spotted-network/proto"
	"google.golang.org/grpc"
)

type registryServer struct {
	pb.UnimplementedRegistryServer
	node *registry.Node
}

func (s *registryServer) GetRegistryID(ctx context.Context, req *pb.GetRegistryIDRequest) (*pb.GetRegistryIDResponse, error) {
	return &pb.GetRegistryIDResponse{
		RegistryId: s.node.GetHostID(),
	}, nil
}

func (s *registryServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	// 解码签名
	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Invalid signature format",
		}, nil
	}

	// 验证签名
	addr := common.HexToAddress(req.Address)
	message := []byte(req.Message)
	if !signer.VerifySignature(addr, message, sigBytes) {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Invalid signature",
		}, nil
	}

	// 签名验证成功
	return &pb.JoinResponse{
		Success: true,
	}, nil
}

func main() {
	// 创建p2p host配置
	cfg := &p2p.Config{
		ListenAddrs: []string{"/ip4/0.0.0.0/tcp/9000"},
	}
	
	// 创建Registry Node
	node, err := registry.NewNode(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer node.Stop()

	// 启动gRPC服务器
	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServer(grpcServer, &registryServer{node: node})

	log.Printf("Registry gRPC server listening on :8000")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
} 