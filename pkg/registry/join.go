package registry

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

const (
	// Protocol message types
	MsgTypeJoinRequest byte = 0x01
	MsgTypeJoinResponse byte = 0x02
)

type OperatorsQuerier interface {
	UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams, getOperatorByAddress *string) (*operators.Operators, error)
	ListAllOperators(ctx context.Context) ([]operators.Operators, error)
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
	UpsertOperator(ctx context.Context, arg operators.UpsertOperatorParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams, getOperatorByAddress *string) (*operators.Operators, error)
}

// handleStream handles incoming p2p streams
func (n *Node) handleStream(stream network.Stream) {
	defer stream.Close()

	// Get peer ID from the connection
	peerID := stream.Conn().RemotePeer()
	log.Printf("New p2p connection from peer: %s", peerID.String())

	// Read message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("Failed to read message type from peer %s: %v", peerID.String(), err)
		return
	}

	// Verify message type
	if msgType[0] != MsgTypeJoinRequest {
		log.Printf("Unexpected message type from peer %s: %d", peerID.String(), msgType[0])
		return
	}

	// Read request with deadline
	data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 10*time.Second)
	if err != nil {
		log.Printf("Failed to read request from peer %s: %v", peerID.String(), err)
		return
	}

	// Parse protobuf message
	var req pb.JoinRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("Failed to parse join request from peer %s: %v", peerID.String(), err)
		return
	}
	log.Printf("Successfully read join request from peer %s", peerID.String())

	// Add peer to p2p network
	log.Printf("Adding peer %s to p2p network", peerID.String())
	n.operatorsInfoMu.Lock()
	n.operatorsInfo[peerID] = &OperatorInfo{
		ID:       peerID,
		Addrs:    n.host.Peerstore().Addrs(peerID),
		LastSeen: time.Now(),
		Status:   "active", // All connected peers are considered active
	}
	n.operatorsInfoMu.Unlock()
	log.Printf("Added new peer to p2p network: %s", peerID.String())

	// Get active peers for response
	log.Printf("Getting active peers for response to peer %s", peerID.String())
	n.operatorsInfoMu.RLock()
	activePeers := make([]*pb.ActiveOperator, 0, len(n.operatorsInfo))
	for id, info := range n.operatorsInfo {
		// Skip registry and new peer
		if id == n.host.ID() || id == peerID {
			continue
		}
		// Include all connected peers
		addrs := make([]string, len(info.Addrs))
		for i, addr := range info.Addrs {
			addrs[i] = addr.String()
		}
		activePeers = append(activePeers, &pb.ActiveOperator{
			PeerId:     id.String(),
			Multiaddrs: addrs,
		})
	}
	n.operatorsInfoMu.RUnlock()
	log.Printf("Found %d connected peers to send to peer %s", len(activePeers), peerID.String())

	// Send success response with active peers
	log.Printf("Sending success response with %d active peers to peer %s", len(activePeers), peerID.String())
	
	// Marshal response
	resp := &pb.JoinResponse{
		Success:         true,
		ActiveOperators: activePeers,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response for peer %s: %v", peerID.String(), err)
		return
	}

	// Write message type
	if _, err := stream.Write([]byte{MsgTypeJoinResponse}); err != nil {
		log.Printf("Failed to write response type to peer %s: %v", peerID.String(), err)
		return
	}

	// Write response with deadline
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, respData, 10*time.Second); err != nil {
		log.Printf("Failed to write response to peer %s: %v", peerID.String(), err)
		return
	}

	log.Printf("Successfully processed join request from peer: %s", peerID.String())
}

// HandleJoinRequest handles a join request from an operator
func (n *Node) HandleJoinRequest(ctx context.Context, req *pb.JoinRequest) error {
	log.Printf("Handling join request from operator %s", req.Address)

	// Check if operator is registered on chain
	log.Printf("Checking if operator %s is registered on chain", req.Address)
	isRegistered, err := n.mainnetClient.IsOperatorRegistered(ctx, ethcommon.HexToAddress(req.Address))
	if err != nil {
		log.Printf("Failed to check operator %s registration: %v", req.Address, err)
		return fmt.Errorf("failed to check operator registration: %w", err)
	}

	if !isRegistered {
		log.Printf("Operator %s is not registered on chain", req.Address)
		return fmt.Errorf("operator not registered on chain")
	}
	log.Printf("Operator %s is registered on chain", req.Address)

	// Get operator from database
	log.Printf("Getting operator %s from database", req.Address)
	operator, err := n.operators.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		log.Printf("Failed to get operator %s from database: %v", req.Address, err)
		return fmt.Errorf("failed to get operator: %w", err)
	}
	log.Printf("Found operator %s in database", req.Address)

	// Verify signing key matches
	log.Printf("Verifying signing key for operator %s", req.Address)
	if operator.SigningKey != req.SigningKey {
		log.Printf("Signing key mismatch for operator %s", req.Address)
		return fmt.Errorf("signing key mismatch")
	}
	log.Printf("Signing key verified for operator %s", req.Address)

	// Update operator status
	log.Printf("Updating status for operator %s", req.Address)
	if err := n.updateStatusAfterOperations(ctx, req.Address); err != nil {
		log.Printf("Failed to update operator status: %v", err)
		return fmt.Errorf("failed to update operator status: %w", err)
	}
	log.Printf("Successfully updated operator status")

	// Broadcast state update
	log.Printf("Broadcasting state update for operator %s", req.Address)
	stateUpdate := []*pb.OperatorState{
		{
			Address:                 operator.Address,
			SigningKey:             operator.SigningKey,
			RegisteredAtBlockNumber: operator.RegisteredAtBlockNumber,
			RegisteredAtTimestamp:   operator.RegisteredAtTimestamp,
			ActiveEpoch:            operator.ActiveEpoch,
			ExitEpoch:              &operator.ExitEpoch,
			Status:                 string(operator.Status),
			Weight:                 commonHelpers.NumericToString(operator.Weight),
		},
	}
	n.BroadcastStateUpdate(stateUpdate, "DELTA")
	log.Printf("Successfully broadcast state update for operator %s", req.Address)

	log.Printf("Operator %s joined successfully", req.Address)
	return nil
}

