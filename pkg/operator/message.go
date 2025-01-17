package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"

	pb "github.com/galxe/spotted-network/proto"
	"github.com/multiformats/go-multiaddr"
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

	// Read response
	respData, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Unmarshal response
	resp := &pb.JoinResponse{}
	if err := proto.Unmarshal(respData, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Handle join response
	if err := n.handleJoinResponse(resp); err != nil {
		return fmt.Errorf("failed to handle join response: %w", err)
	}

	log.Printf("Successfully joined registry")
	return nil
}

func (n *Node) handleJoinResponse(resp *pb.JoinResponse) error {
	if !resp.Success {
		return fmt.Errorf("join request failed: %s", resp.Error)
	}

	log.Printf("Processing join response with %d active operators", len(resp.ActiveOperators))

	// Connect to each active operator
	for _, op := range resp.ActiveOperators {
		// Parse peer ID
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Printf("Failed to decode peer ID %s: %v", op.PeerId, err)
			continue
		}

		// Skip self
		if peerID == n.host.ID() {
			continue
		}

		// Parse multiaddrs
		addrs := make([]multiaddr.Multiaddr, 0, len(op.Multiaddrs))
		for _, addr := range op.Multiaddrs {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Printf("Failed to parse multiaddr %s: %v", addr, err)
				continue
			}
			addrs = append(addrs, maddr)
		}

		// Store operator info
		n.operatorsMu.Lock()
		n.operators[peerID] = &OperatorInfo{
			ID:       peerID,
			Addrs:    addrs,
			LastSeen: time.Now(),
			Status:   string(OperatorStatusActive),
		}
		n.operatorsMu.Unlock()

		// Connect to operator
		if err := n.host.Connect(context.Background(), peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}); err != nil {
			log.Printf("Failed to connect to operator %s: %v", peerID, err)
			continue
		}

		log.Printf("Successfully connected to operator %s", peerID)
	}

	log.Printf("Successfully processed join response and connected to %d operators", len(resp.ActiveOperators))
	return nil
} 