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

	"github.com/galxe/spotted-network/pkg/p2p"
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
	log.Printf("[Announce] Starting to announce to registry %s", n.registryID)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Open stream with timeout context
	stream, err := n.host.NewStream(ctx, peer.ID(n.registryID), "/spotted/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open stream to registry (timeout 30s): %w", err)
	}
	defer stream.Close()
	log.Printf("[Announce] Successfully opened stream to registry")

	// Create join request message
	message := []byte("join_request")
	signature, err := n.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}
	log.Printf("[Announce] Created and signed join request message")

	// Create protobuf request
	req := &pb.JoinRequest{
		Address:    n.signer.GetAddress().Hex(),
		Message:    string(message),
		Signature:  hex.EncodeToString(signature),
		SigningKey: n.signer.GetSigningKey(),
	}
	log.Printf("[Announce] Created join request for address: %s", req.Address)

	// Marshal request
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write request with length prefix
	if err := p2p.WriteLengthPrefixed(stream, p2p.MsgTypeJoinRequest, data); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}
	log.Printf("[Announce] Successfully sent join request to registry")

	// Set read deadline
	if err := stream.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		log.Printf("[Announce] Warning: Failed to set read deadline: %v", err)
	}

	// Read response with length prefix
	msgType, respData, err := p2p.ReadLengthPrefixed(stream)
	if err != nil {
		return fmt.Errorf("failed to read response (timeout 30s): %w", err)
	}
	log.Printf("[Announce] Received response data of length: %d bytes", len(respData))

	// Verify message type
	if msgType != p2p.MsgTypeJoinResponse {
		return fmt.Errorf("unexpected response message type: %d", msgType)
	}

	// Reset read deadline
	if err := stream.SetReadDeadline(time.Time{}); err != nil {
		log.Printf("[Announce] Warning: Failed to reset read deadline: %v", err)
	}

	// Unmarshal response
	resp := &pb.JoinResponse{}
	if err := proto.Unmarshal(respData, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	log.Printf("[Announce] Successfully unmarshaled response: success=%v, error=%s, active_operators=%d",
		resp.Success, resp.Error, len(resp.ActiveOperators))

	// Handle join response
	if err := n.handleJoinResponse(resp); err != nil {
		return fmt.Errorf("failed to handle join response: %w", err)
	}

	log.Printf("[Announce] Successfully completed announcement to registry")
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