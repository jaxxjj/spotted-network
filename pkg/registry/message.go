package registry

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

// handleStream handles incoming p2p streams
func (n *Node) handleStream(stream network.Stream) {
	defer stream.Close()

	// Get peer ID from the connection
	peerID := stream.Conn().RemotePeer()
	log.Printf("New p2p connection from peer: %s", peerID.String())

	// Read join request
	req, err := p2p.ReadJoinRequest(stream)
	if err != nil {
		log.Printf("Failed to read join request: %v", err)
		p2p.SendError(stream, fmt.Errorf("failed to read request: %v", err))
		return
	}

	// Handle join request
	err = n.HandleJoinRequest(context.Background(), req)
	if err != nil {
		log.Printf("Failed to handle join request: %v", err)
		p2p.SendError(stream, err)
		return
	}

	// Add operator to p2p network
	n.operatorsMu.Lock()
	n.operators[peerID] = &OperatorInfo{
		ID: peerID,
		Addrs: n.host.Peerstore().Addrs(peerID),
		LastSeen: time.Now(),
		Status: string(OperatorStatusWaitingActive),
	}
	n.operatorsMu.Unlock()
	log.Printf("Added new operator to p2p network: %s", peerID.String())

	// Get active operators for response
	n.operatorsMu.RLock()
	activeOperators := make([]*pb.ActiveOperator, 0, len(n.operators))
	for id, info := range n.operators {
		// Skip registry and new operator
		if id == n.host.ID() || id == peerID {
			continue
		}
		// Only include active operators
		if info.Status == string(OperatorStatusActive) {
			addrs := make([]string, len(info.Addrs))
			for i, addr := range info.Addrs {
				addrs[i] = addr.String()
			}
			activeOperators = append(activeOperators, &pb.ActiveOperator{
				PeerId: id.String(),
				Multiaddrs: addrs,
			})
		}
	}
	n.operatorsMu.RUnlock()

	// Send success response with active operators
	resp := &pb.JoinResponse{
		Success: true,
		ActiveOperators: activeOperators,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		p2p.SendError(stream, fmt.Errorf("failed to marshal response: %v", err))
		return
	}

	if _, err := stream.Write(respData); err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	// Broadcast new operator to other peers
	n.BroadcastStateUpdate([]*pb.OperatorState{{
		Address: req.Address,
		Status: string(OperatorStatusWaitingActive),
	}}, "UPDATE")

	log.Printf("Successfully processed join request from: %s", req.Address)
}

// HandleJoinRequest handles a join request from an operator
func (n *Node) HandleJoinRequest(ctx context.Context, req *pb.JoinRequest) error {
	log.Printf("Handling join request from operator %s", req.Address)

	// Check if operator is registered on chain
	mainnetClient := n.chainClients.GetMainnetClient()
	isRegistered, err := mainnetClient.IsOperatorRegistered(ctx, common.HexToAddress(req.Address))
	if err != nil {
		return fmt.Errorf("failed to check operator registration: %w", err)
	}

	if !isRegistered {
		return fmt.Errorf("operator not registered on chain")
	}

	// Get operator from database
	op, err := n.db.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		return fmt.Errorf("failed to get operator: %w", err)
	}

	// Verify signing key matches
	if op.SigningKey != req.SigningKey {
		return fmt.Errorf("signing key mismatch")
	}

	// Update status to waitingActive
	updatedOp, err := n.db.UpdateOperatorStatus(ctx, registry.UpdateOperatorStatusParams{
		Address: req.Address,
		Status:  string(OperatorStatusWaitingActive),
	})
	if err != nil {
		return fmt.Errorf("failed to update operator status: %w", err)
	}

	// Broadcast state update
	var exitEpoch *int32
	if updatedOp.ExitEpoch.Valid {
		val := updatedOp.ExitEpoch.Int32
		exitEpoch = &val
	}

	stateUpdate := []*pb.OperatorState{
		{
			Address:                 updatedOp.Address,
			SigningKey:             updatedOp.SigningKey,
			RegisteredAtBlockNumber: updatedOp.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   updatedOp.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(updatedOp.ActiveEpoch),
			ExitEpoch:              exitEpoch,
			Status:                 updatedOp.Status,
			Weight:                 updatedOp.Weight.Int.String(),
			Missing:                int32(updatedOp.Missing.Int32),
			SuccessfulResponseCount: int32(updatedOp.SuccessfulResponseCount.Int32),
		},
	}
	n.BroadcastStateUpdate(stateUpdate, "DELTA")

	log.Printf("Operator %s joined successfully", req.Address)
	return nil
}

func (n *Node) handleStateSync(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("Received state sync request from %s", remotePeer)
	
	// Read message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("Error reading message type: %v", err)
		stream.Reset()
		return
	}
	log.Printf("Received message type: 0x%x", msgType[0])

	// Read length prefix
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(stream, lengthBytes); err != nil {
		log.Printf("Error reading length prefix: %v", err)
		stream.Reset()
		return
	}
	length := uint32(lengthBytes[0])<<24 | 
		uint32(lengthBytes[1])<<16 | 
		uint32(lengthBytes[2])<<8 | 
		uint32(lengthBytes[3])
	log.Printf("Message length: %d bytes", length)

	// Read request data
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Printf("Error reading request data: %v", err)
		stream.Reset()
		return
	}

	switch msgType[0] {
	case 0x01: // GetFullStateRequest
		var req pb.GetFullStateRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("Error unmarshaling GetFullStateRequest: %v", err)
			stream.Reset()
			return
		}
		log.Printf("Successfully identified request as GetFullStateRequest")
		n.handleGetFullState(stream)
		return

	case 0x02: // SubscribeRequest
		var req pb.SubscribeRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("Error unmarshaling SubscribeRequest: %v", err)
			stream.Reset()
			return
		}
		log.Printf("Successfully identified request as SubscribeRequest")
		n.handleSubscribe(stream)
		return

	default:
		log.Printf("Unknown message type: 0x%x", msgType[0])
		stream.Reset()
		return
	}
}

func (n *Node) handleGetFullState(stream network.Stream) {
	log.Printf("Handling GetFullState request from %s", stream.Conn().RemotePeer())
	
	// Get all operators from database regardless of status
	operators, err := n.db.ListAllOperators(context.Background())
	if err != nil {
		log.Printf("Error getting operators from database: %v", err)
		stream.Reset()
		return
	}

	log.Printf("Found %d total operators in database", len(operators))
	if len(operators) == 0 {
		log.Printf("Warning: No operators found in database")
	}

	// Convert to proto message
	protoOperators := make([]*pb.OperatorState, 0, len(operators))
	for _, op := range operators {
		var exitEpoch *int32
		if op.ExitEpoch.Valid {
			val := op.ExitEpoch.Int32
			exitEpoch = &val
		}

		// Log operator details before conversion
		log.Printf("Converting operator - Address: %s, Status: %s, BlockNumber: %v, Timestamp: %v, Weight: %v",
			op.Address, op.Status, op.RegisteredAtBlockNumber, op.RegisteredAtTimestamp, op.Weight)

		protoOperators = append(protoOperators, &pb.OperatorState{
			Address:                 op.Address,
			SigningKey:             op.SigningKey,
			RegisteredAtBlockNumber: op.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   op.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(op.ActiveEpoch),
			ExitEpoch:              exitEpoch,
			Status:                 op.Status,
			Weight:                 op.Weight.Int.String(),
			Missing:                int32(op.Missing.Int32),
			SuccessfulResponseCount: int32(op.SuccessfulResponseCount.Int32),
		})
		log.Printf("Added operator to response - Address: %s, Status: %s", op.Address, op.Status)
	}

	// Create response
	resp := &pb.GetFullStateResponse{
		Operators: protoOperators,
	}

	// Marshal response
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		stream.Reset()
		return
	}

	// Write length prefix
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		log.Printf("Error writing response length: %v", err)
		stream.Reset()
		return
	}

	// Write response data
	bytesWritten, err := stream.Write(data)
	if err != nil {
		log.Printf("Error sending response: %v", err)
		stream.Reset()
		return
	}

	log.Printf("Successfully sent full state with %d operators (%d bytes) to %s", 
		len(protoOperators), bytesWritten, stream.Conn().RemotePeer())
}

func (n *Node) handleSubscribe(stream network.Stream) {
	peer := stream.Conn().RemotePeer()
	
	n.subscribersMu.Lock()
	n.subscribers[peer] = stream
	n.subscribersMu.Unlock()

	log.Printf("Added subscriber %s", peer)

	// Keep stream open until closed by peer or error
	buf := make([]byte, 1024)
	for {
		_, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from subscriber %s: %v", peer, err)
			}
			n.subscribersMu.Lock()
			delete(n.subscribers, peer)
			n.subscribersMu.Unlock()
			stream.Close()
			log.Printf("Removed subscriber %s", peer)
			return
		}
	}
} 