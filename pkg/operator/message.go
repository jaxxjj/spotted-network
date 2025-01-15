package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"

	pb "github.com/galxe/spotted-network/proto"
)

func (n *Node) handleMessages(stream network.Stream) {
	defer stream.Close()
	
	log.Printf("New stream opened from: %s", stream.Conn().RemotePeer())

	// Read the entire message
	data, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("Error reading from stream: %v", err)
		return
	}

	// Parse protobuf message
	var req pb.JoinRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("Error parsing protobuf message: %v", err)
		return
	}

	log.Printf("Received join request from operator: %s", req.Address)

	// Send success response
	resp := &pb.JoinResponse{
		Success: true,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	if _, err := stream.Write(respData); err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	log.Printf("Successfully processed join request from: %s", req.Address)
}

func (n *Node) announceToRegistry() error {
	stream, err := n.host.NewStream(context.Background(), peer.ID(n.registryID), "/spotted/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open stream to registry: %w", err)
	}
	defer stream.Close()

	// Create join request message
	message := []byte("join_request")
	signature, err := n.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// Create protobuf request
	req := &pb.JoinRequest{
		Address:    n.signer.GetAddress().Hex(),
		Message:    string(message),
		Signature:  hex.EncodeToString(signature),
		SigningKey: n.signer.GetSigningKey(),
	}

	// Marshal and send request
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
} 