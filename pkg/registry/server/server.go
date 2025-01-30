package server

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/types"
	registrynode "github.com/galxe/spotted-network/pkg/registry"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
)

type OperatorQuerier interface {
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
	UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams, getOperatorByAddress *string) (*operators.Operators, error)
}

// RegistryServer implements the Registry gRPC service
type RegistryServer struct {
	pb.UnimplementedRegistryServer
	node *registrynode.Node
	operators OperatorQuerier
}

// NewRegistryServer creates a new instance of RegistryServer
func NewRegistryServer(node *registrynode.Node, operators OperatorQuerier) *RegistryServer {
	return &RegistryServer{
		node: node,
		operators: operators,
	}
}

// GetRegistryID returns the registry node's host ID
func (s *RegistryServer) GetRegistryID(ctx context.Context, req *pb.GetRegistryIDRequest) (*pb.GetRegistryIDResponse, error) {
	return &pb.GetRegistryIDResponse{
		RegistryId: s.node.GetHostID(),
	}, nil
}

// Join handles operator join requests
func (s *RegistryServer) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	// Get operator status from database
	operator, err := s.operators.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		return &pb.JoinResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get operator: %v", err),
		}, nil
	}

	// Verify operator is in active status
	if operator.Status != types.OperatorStatusActive {
		return &pb.JoinResponse{
			Success: false,
			Error:   "Operator not in active status",
		}, nil
	}

	// Verify signing key matches
	if operator.SigningKey != req.SigningKey {
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
				PeerId:     peerID.String(),
				Multiaddrs: addrs,
			})
		}
	}

	return &pb.JoinResponse{
		Success:         true,
		ActiveOperators: activeOperators,
	}, nil
} 