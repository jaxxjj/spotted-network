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
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
)

const (
	// Protocol message types
	MsgTypeJoinRequest byte = 0x01
	MsgTypeJoinResponse byte = 0x02
)

func (n *Node) handleMessages(stream network.Stream) {
	defer stream.Close()
	
	log.Printf("[Messages] New stream opened from: %s", stream.Conn().RemotePeer())

	// Read message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("[Messages] Error reading message type: %v", err)
		return
	}

	// Verify message type
	if msgType[0] != MsgTypeJoinRequest {
		log.Printf("[Messages] Unexpected message type: %d", msgType[0])
		return
	}

	// Read request with deadline
	data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 5*time.Second)
	if err != nil {
		log.Printf("[Messages] Error reading request: %v", err)
		return
	}

	// Parse protobuf message
	var req pb.JoinRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("[Messages] Error parsing protobuf message: %v", err)
		return
	}

	log.Printf("[Messages] Received join request from operator: %s", req.Address)

	// Get our peer info
	peerInfo := peer.AddrInfo{
		ID:    n.host.ID(),
		Addrs: n.host.Addrs(),
	}

	// Create active operators list with our info
	activeOperators := []*pb.ActiveOperator{
		{
			PeerId:     peerInfo.ID.String(),
			Multiaddrs: make([]string, len(peerInfo.Addrs)),
		},
	}
	for i, addr := range peerInfo.Addrs {
		activeOperators[0].Multiaddrs[i] = addr.String()
	}

	// Send success response with our info
	resp := &pb.JoinResponse{
		Success:         true,
		ActiveOperators: activeOperators,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("[Messages] Error marshaling response: %v", err)
		return
	}

	// Write message type
	if _, err := stream.Write([]byte{MsgTypeJoinResponse}); err != nil {
		log.Printf("[Messages] Error writing response type: %v", err)
		return
	}

	// Write response with deadline
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, respData, 5*time.Second); err != nil {
		log.Printf("[Messages] Error writing response: %v", err)
		return
	}

	log.Printf("[Messages] Successfully processed join request from: %s", req.Address)
}

func (n *Node) announceToRegistry() ([]*peer.AddrInfo, error) {
	// Validate signer
	if n.signer == nil {
		return nil, fmt.Errorf("[Announce] signer not initialized")
	}

	log.Printf("[Announce] Starting to announce to registry %s", n.registryID)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Open stream with timeout context
	stream, err := n.host.NewStream(ctx, peer.ID(n.registryID), "/spotted/1.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to open stream to registry (timeout 30s): %w", err)
	}
	defer stream.Close()
	log.Printf("[Announce] Successfully opened stream to registry")

	// Create join request message
	message := []byte("join_request")
	signature, err := n.signer.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	log.Printf("[Announce] Created and signed join request message")

	// Send join request
	return n.sendJoinRequest(stream, message, signature)
}

func (n *Node) sendJoinRequest(stream network.Stream, message []byte, signature []byte) ([]*peer.AddrInfo, error) {
	// Create protobuf request
	req := &pb.JoinRequest{
		Address:    n.signer.GetOperatorAddress().Hex(),
		Message:    string(message),
		Signature:  hex.EncodeToString(signature),
		SigningKey: n.signer.GetSigningAddress().Hex(),
	}
	log.Printf("[Announce] Created join request for address: %s", req.Address)

	// Marshal request
	data, err := proto.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write message type
	if _, err := stream.Write([]byte{MsgTypeJoinRequest}); err != nil {
		return nil, fmt.Errorf("failed to write message type: %w", err)
	}

	// Write request with deadline
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, data, 30*time.Second); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	log.Printf("[Announce] Successfully sent join request to registry")

	// Read response message type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		return nil, fmt.Errorf("[Announce] failed to read response type: %w", err)
	}

	if msgType[0] != MsgTypeJoinResponse {
		return nil, fmt.Errorf("[Announce] unexpected response type: 0x%02x", msgType[0])
	}

	// Read response with deadline
	respData, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("[Announce] failed to read response: %w", err)
	}

	// Unmarshal response
	resp := &pb.JoinResponse{}
	if err := proto.Unmarshal(respData, resp); err != nil {
		return nil, fmt.Errorf("[Announce] failed to unmarshal response: %w", err)
	}

	log.Printf("[Announce] Successfully unmarshaled response: success=%v, error=%s, active_operators=%d",
		resp.Success, resp.Error, len(resp.ActiveOperators))

	if !resp.Success {
		return nil, fmt.Errorf("[Announce] join request failed: %s", resp.Error)
	}

	// Convert active operators to AddrInfo
	activeOperators := make([]*peer.AddrInfo, 0, len(resp.ActiveOperators))
	for _, op := range resp.ActiveOperators {
		// Parse peer ID
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Printf("[Announce] Failed to decode peer ID %s: %v", op.PeerId, err)
			continue
		}

		// Parse multiaddrs
		addrs := make([]multiaddr.Multiaddr, 0, len(op.Multiaddrs))
		for _, addr := range op.Multiaddrs {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Printf("[Announce] Failed to parse multiaddr %s: %v", addr, err)
				continue
			}
			addrs = append(addrs, maddr)
		}

		// Create peer info
		addrInfo := &peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}
		activeOperators = append(activeOperators, addrInfo)
	}

	log.Printf("[Announce] Successfully processed %d active operators", len(activeOperators))
	return activeOperators, nil
}

