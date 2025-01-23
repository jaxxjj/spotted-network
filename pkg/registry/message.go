package registry

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/p2p"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

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
	n.operatorsMu.Lock()
	n.operators[peerID] = &OperatorInfo{
		ID: peerID,
		Addrs: n.host.Peerstore().Addrs(peerID),
		LastSeen: time.Now(),
		Status: "active", // All connected peers are considered active
	}
	n.operatorsMu.Unlock()
	log.Printf("Added new peer to p2p network: %s", peerID.String())

	// Get active peers for response
	log.Printf("Getting active peers for response to peer %s", peerID.String())
	n.operatorsMu.RLock()
	activePeers := make([]*pb.ActiveOperator, 0, len(n.operators))
	for id, info := range n.operators {
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
	n.operatorsMu.RUnlock()
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
	mainnetClient := n.chainClients.GetMainnetClient()
	isRegistered, err := mainnetClient.IsOperatorRegistered(ctx, ethcommon.HexToAddress(req.Address))
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
	op, err := n.db.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		log.Printf("Failed to get operator %s from database: %v", req.Address, err)
		return fmt.Errorf("failed to get operator: %w", err)
	}
	log.Printf("Found operator %s in database", req.Address)

	// Verify signing key matches
	log.Printf("Verifying signing key for operator %s", req.Address)
	if op.SigningKey != req.SigningKey {
		log.Printf("Signing key mismatch for operator %s", req.Address)
		return fmt.Errorf("signing key mismatch")
	}
	log.Printf("Signing key verified for operator %s", req.Address)

	// Update operator status
	log.Printf("Updating status for operator %s", req.Address)
	if err := n.eventListener.updateStatusAfterOperations(ctx, req.Address); err != nil {
		log.Printf("Failed to update operator status: %v", err)
		return fmt.Errorf("failed to update operator status: %w", err)
	}
	log.Printf("Successfully updated operator status")

	// Broadcast state update
	log.Printf("Broadcasting state update for operator %s", req.Address)
	var exitEpoch *int32
	defaultExitEpoch := pgtype.Numeric{
		Int:    new(big.Int).SetUint64(4294967295),
		Valid:  true,
		Exp:    0,
	}
	if op.ExitEpoch.Int.Cmp(defaultExitEpoch.Int) != 0 { // Check if not default max value
		val := int32(op.ExitEpoch.Int.Int64())
		exitEpoch = &val
	}

	stateUpdate := []*pb.OperatorState{
		{
			Address:                 op.Address,
			SigningKey:             op.SigningKey,
			RegisteredAtBlockNumber: op.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   op.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(op.ActiveEpoch.Int.Int64()),
			ExitEpoch:              exitEpoch,
			Status:                 op.Status,
			Weight:                 common.NumericToString(op.Weight),
		},
	}
	n.BroadcastStateUpdate(stateUpdate, "DELTA")
	log.Printf("Successfully broadcast state update for operator %s", req.Address)

	log.Printf("Operator %s joined successfully", req.Address)
	return nil
}

func (n *Node) handleStateSync(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("Received state sync request from %s", remotePeer)
	
	// Read message type
	log.Printf("Reading message type from peer %s", remotePeer)
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("Error reading message type from peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("Received message type 0x%x from peer %s", msgType[0], remotePeer)

	// Read length prefix
	log.Printf("Reading message length from peer %s", remotePeer)
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(stream, lengthBytes); err != nil {
		log.Printf("Error reading length prefix from peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	length := uint32(lengthBytes[0])<<24 | 
		uint32(lengthBytes[1])<<16 | 
		uint32(lengthBytes[2])<<8 | 
		uint32(lengthBytes[3])
	log.Printf("Message length from peer %s: %d bytes", remotePeer, length)

	// Read request data
	log.Printf("Reading request data from peer %s", remotePeer)
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Printf("Error reading request data from peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("Successfully read %d bytes of request data from peer %s", length, remotePeer)

	switch msgType[0] {
	case 0x01: // GetFullStateRequest
		log.Printf("Processing GetFullStateRequest from peer %s", remotePeer)
		var req pb.GetFullStateRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("Error unmarshaling GetFullStateRequest from peer %s: %v", remotePeer, err)
			stream.Reset()
			return
		}
		log.Printf("Successfully unmarshaled GetFullStateRequest from peer %s", remotePeer)
		n.handleGetFullState(stream)
		return

	case 0x02: // SubscribeRequest
		log.Printf("Processing SubscribeRequest from peer %s", remotePeer)
		var req pb.SubscribeRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("Error unmarshaling SubscribeRequest from peer %s: %v", remotePeer, err)
			stream.Reset()
			return
		}
		log.Printf("Successfully unmarshaled SubscribeRequest from peer %s", remotePeer)
		n.handleSubscribe(stream)
		return

	default:
		log.Printf("Received unknown message type 0x%x from peer %s", msgType[0], remotePeer)
		stream.Reset()
		return
	}
}

func (n *Node) handleGetFullState(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("Handling GetFullState request from %s", remotePeer)
	
	// Get all operators from database regardless of status
	log.Printf("Querying database for all operators for peer %s", remotePeer)
	operators, err := n.db.ListAllOperators(context.Background())
	if err != nil {
		log.Printf("Error getting operators from database for peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}

	log.Printf("Found %d total operators in database for peer %s", len(operators), remotePeer)
	if len(operators) == 0 {
		log.Printf("Warning: No operators found in database for peer %s", remotePeer)
	}

	// Convert to proto message
	log.Printf("Converting %d operators to proto message for peer %s", len(operators), remotePeer)
	protoOperators := make([]*pb.OperatorState, 0, len(operators))
	for _, op := range operators {
		var exitEpoch *int32
		defaultExitEpoch := pgtype.Numeric{
			Int:    new(big.Int).SetUint64(4294967295),
			Valid:  true,
			Exp:    0,
		}
		if op.ExitEpoch.Int.Cmp(defaultExitEpoch.Int) != 0 { // Check if not default max value
			val := int32(op.ExitEpoch.Int.Int64())
			exitEpoch = &val
		}

		// Log operator details before conversion
		log.Printf("Converting operator for peer %s - Address: %s, Status: %s, BlockNumber: %v, Timestamp: %v, Weight: %v",
			remotePeer, op.Address, op.Status, op.RegisteredAtBlockNumber, op.RegisteredAtTimestamp, op.Weight)

		protoOperators = append(protoOperators, &pb.OperatorState{
			Address:                 op.Address,
			SigningKey:             op.SigningKey,
			RegisteredAtBlockNumber: op.RegisteredAtBlockNumber.Int.Int64(),
			RegisteredAtTimestamp:   op.RegisteredAtTimestamp.Int.Int64(),
			ActiveEpoch:            int32(op.ActiveEpoch.Int.Int64()),
			ExitEpoch:              exitEpoch,
			Status:                 op.Status,
			Weight:                 common.NumericToString(op.Weight),
			Missing:                0, // These fields are no longer in the table
			SuccessfulResponseCount: 0, // These fields are no longer in the table
		})
		log.Printf("Added operator to response for peer %s - Address: %s, Status: %s", remotePeer, op.Address, op.Status)
	}

	// Create response
	log.Printf("Creating GetFullStateResponse with %d operators for peer %s", len(protoOperators), remotePeer)
	resp := &pb.GetFullStateResponse{
		Operators: protoOperators,
	}

	// Marshal response
	log.Printf("Marshaling response for peer %s", remotePeer)
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response for peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("Successfully marshaled response of %d bytes for peer %s", len(data), remotePeer)

	// Write length prefix
	log.Printf("Writing response length prefix (%d bytes) to peer %s", len(data), remotePeer)
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		log.Printf("Error writing response length to peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("Successfully wrote length prefix to peer %s", remotePeer)

	// Write response data
	log.Printf("Writing response data (%d bytes) to peer %s", len(data), remotePeer)
	bytesWritten, err := stream.Write(data)
	if err != nil {
		log.Printf("Error sending response to peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}

	log.Printf("Successfully sent full state with %d operators (%d bytes) to %s", 
		len(protoOperators), bytesWritten, remotePeer)
}

func (n *Node) handleSubscribe(stream network.Stream) {
	peer := stream.Conn().RemotePeer()
	log.Printf("Handling subscribe request from peer %s", peer)
	
	n.subscribersMu.Lock()
	n.subscribers[peer] = stream
	subscriberCount := len(n.subscribers)
	n.subscribersMu.Unlock()

	log.Printf("Added subscriber %s (total subscribers: %d)", peer, subscriberCount)

	// Keep stream open until closed by peer or error
	log.Printf("Starting read loop for subscriber %s", peer)
	buf := make([]byte, 1024)
	for {
		_, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from subscriber %s: %v", peer, err)
			} else {
				log.Printf("Subscriber %s closed connection", peer)
			}
			n.subscribersMu.Lock()
			delete(n.subscribers, peer)
			remainingSubscribers := len(n.subscribers)
			n.subscribersMu.Unlock()
			stream.Close()
			log.Printf("Removed subscriber %s (remaining subscribers: %d)", peer, remainingSubscribers)
			return
		}
	}
} 