package registry

import (
	"context"
	"fmt"
	"log"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
)
type OperatorsQuerier interface {
	UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams) (operators.Operators, error)
	ListAllOperators(ctx context.Context) ([]operators.Operators, error)
	GetOperatorByAddress(ctx context.Context, address string) (operators.Operators, error)
	UpsertOperator(ctx context.Context, arg operators.UpsertOperatorParams) (operators.Operators, error)
	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams) (operators.Operators, error)
	UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams) (operators.Operators, error)
}

// handleStream handles incoming p2p streams
func (n *Node) handleStream(stream network.Stream) {
	defer stream.Close()

	// Get peer ID from the connection
	peerID := stream.Conn().RemotePeer()
	log.Printf("New p2p connection from peer: %s", peerID.String())

	// Set read deadline
	if err := stream.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("Warning: Failed to set read deadline: %v", err)
	}

	// Read and verify join request message type
	log.Printf("Reading join request from peer: %s", peerID.String())
	msgType, _, err := p2p.ReadLengthPrefixed(stream)
	if err != nil {
		log.Printf("Failed to read join request from peer %s: %v", peerID.String(), err)
		p2p.SendError(stream, fmt.Errorf("failed to read request: %v", err))
		return
	}
	if msgType != p2p.MsgTypeJoinRequest {
		log.Printf("Unexpected message type from peer %s: %d", peerID.String(), msgType)
		p2p.SendError(stream, fmt.Errorf("unexpected message type"))
		return
	}
	log.Printf("Successfully read join request from peer %s", peerID.String())

	// Reset read deadline
	if err := stream.SetReadDeadline(time.Time{}); err != nil {
		log.Printf("Warning: Failed to reset read deadline: %v", err)
	}

	// Add peer to p2p network
	log.Printf("Adding peer %s to p2p network", peerID.String())
	n.operatorsInfoMu.Lock()
	n.operatorsInfo[peerID] = &OperatorInfo{
		ID: peerID,
		Addrs: n.host.Peerstore().Addrs(peerID),
		LastSeen: time.Now(),
		Status: "active", // All connected peers are considered active
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
			PeerId: id.String(),
			Multiaddrs: addrs,
		})
	}
	n.operatorsInfoMu.RUnlock()
	log.Printf("Found %d connected peers to send to peer %s", len(activePeers), peerID.String())

	// Set write deadline
	if err := stream.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("Warning: Failed to set write deadline: %v", err)
	}

	// Send success response with active peers
	log.Printf("Sending success response with %d active peers to peer %s", len(activePeers), peerID.String())
	p2p.SendSuccess(stream, activePeers)
	log.Printf("Successfully sent response to peer %s", peerID.String())

	// Reset write deadline
	if err := stream.SetWriteDeadline(time.Time{}); err != nil {
		log.Printf("Warning: Failed to reset write deadline: %v", err)
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
			Weight:                 common.NumericToString(operator.Weight),
		},
	}
	n.BroadcastStateUpdate(stateUpdate, "DELTA")
	log.Printf("Successfully broadcast state update for operator %s", req.Address)

	log.Printf("Operator %s joined successfully", req.Address)
	return nil
}