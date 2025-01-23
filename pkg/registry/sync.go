package registry

import (
	"context"
	"io"
	"log"
	"math/big"

	"github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/protobuf/proto"
)

func (n *Node) handleStateSync(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] Received state sync request from %s", remotePeer)
	
	// Read message type
	log.Printf("[StateSync] Reading message type from peer %s", remotePeer)
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("[StateSync] Error reading message type from peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("[StateSync] Received message type 0x%x from peer %s", msgType[0], remotePeer)

	// Read length prefix
	log.Printf("[StateSync] Reading message length from peer %s", remotePeer)
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(stream, lengthBytes); err != nil {
		log.Printf("[StateSync] Error reading length prefix from peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	length := uint32(lengthBytes[0])<<24 | 
		uint32(lengthBytes[1])<<16 | 
		uint32(lengthBytes[2])<<8 | 
		uint32(lengthBytes[3])
	log.Printf("[StateSync] Message length from peer %s: %d bytes", remotePeer, length)

	// Read request data
	log.Printf("[StateSync] Reading request data from peer %s", remotePeer)
	data := make([]byte, length)
	if _, err := io.ReadFull(stream, data); err != nil {
		log.Printf("[StateSync] Error reading request data from peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("[StateSync] Successfully read %d bytes of request data from peer %s", length, remotePeer)

	switch msgType[0] {
	case 0x01: // GetFullStateRequest
		log.Printf("[StateSync] Processing GetFullStateRequest from peer %s", remotePeer)
		var req pb.GetFullStateRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("[StateSync] Error unmarshaling GetFullStateRequest from peer %s: %v", remotePeer, err)
			stream.Reset()
			return
		}
		log.Printf("[StateSync] Successfully unmarshaled GetFullStateRequest from peer %s", remotePeer)
		n.handleGetFullState(stream)
		return

	case 0x02: // SubscribeRequest
		log.Printf("[StateSync] Processing SubscribeRequest from peer %s", remotePeer)
		var req pb.SubscribeRequest
		if err := proto.Unmarshal(data, &req); err != nil {
			log.Printf("[StateSync] Error unmarshaling SubscribeRequest from peer %s: %v", remotePeer, err)
			stream.Reset()
			return
		}
		log.Printf("[StateSync] Successfully unmarshaled SubscribeRequest from peer %s", remotePeer)
		n.handleSubscribe(stream)
		return

	default:
		log.Printf("[StateSync] Received unknown message type 0x%x from peer %s", msgType[0], remotePeer)
		stream.Reset()
		return
	}
}

func (n *Node) handleGetFullState(stream network.Stream) {
	remotePeer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] Handling GetFullState request from %s", remotePeer)
	
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
			Int: new(big.Int).SetUint64(4294967295),
		}
		if op.ExitEpoch.Valid && op.ExitEpoch.Int.Cmp(defaultExitEpoch.Int) != 0 { // Check if valid and not default max value
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
		log.Printf("[StateSync] Added operator to response for peer %s - Address: %s, Status: %s", remotePeer, op.Address, op.Status)
	}

	// Create response
	log.Printf("[StateSync] Creating GetFullStateResponse with %d operators for peer %s", len(protoOperators), remotePeer)
	resp := &pb.GetFullStateResponse{
		Operators: protoOperators,
	}

	// Marshal response
	log.Printf("[StateSync] Marshaling response for peer %s", remotePeer)
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("[StateSync] Error marshaling response for peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("[StateSync] Successfully marshaled response of %d bytes for peer %s", len(data), remotePeer)

	// Write length prefix
	log.Printf("[StateSync] Writing response length prefix (%d bytes) to peer %s", len(data), remotePeer)
	length := uint32(len(data))
	lengthBytes := make([]byte, 4)
	lengthBytes[0] = byte(length >> 24)
	lengthBytes[1] = byte(length >> 16)
	lengthBytes[2] = byte(length >> 8)
	lengthBytes[3] = byte(length)
	
	if _, err := stream.Write(lengthBytes); err != nil {
		log.Printf("[StateSync] Error writing response length to peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}
	log.Printf("[StateSync] Successfully wrote length prefix to peer %s", remotePeer)

	// Write response data
	log.Printf("[StateSync] Writing response data (%d bytes) to peer %s", len(data), remotePeer)
	bytesWritten, err := stream.Write(data)
	if err != nil {
		log.Printf("[StateSync] Error sending response to peer %s: %v", remotePeer, err)
		stream.Reset()
		return
	}

	log.Printf("[StateSync] Successfully sent full state with %d operators (%d bytes) to %s", 
		len(protoOperators), bytesWritten, remotePeer)
}

func (n *Node) handleSubscribe(stream network.Stream) {
	peer := stream.Conn().RemotePeer()
	log.Printf("[StateSync] Handling subscribe request from peer %s", peer)
	
	n.subscribersMu.Lock()
	n.subscribers[peer] = stream
	subscriberCount := len(n.subscribers)
	n.subscribersMu.Unlock()

	log.Printf("[StateSync] Added subscriber %s (total subscribers: %d)", peer, subscriberCount)

	// Keep stream open until closed by peer or error
	log.Printf("[StateSync] Starting read loop for subscriber %s", peer)
	buf := make([]byte, 1024)
	for {
		_, err := stream.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[StateSync] Error reading from subscriber %s: %v", peer, err)
			} else {
				log.Printf("[StateSync] Subscriber %s closed connection", peer)
			}
			n.subscribersMu.Lock()
			delete(n.subscribers, peer)
			remainingSubscribers := len(n.subscribers)
			n.subscribersMu.Unlock()
			stream.Close()
			log.Printf("[StateSync] Removed subscriber %s (remaining subscribers: %d)", peer, remainingSubscribers)
			return
		}
	}
} 