package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	utils "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/rs/zerolog/log"
)

const (
	AuthTimeout              = 30 * time.Second
	RegistryProtocol       = protocol.ID("/spotted/registry/1.0.0")
	maxMessageSize = 1024 * 1024 // 1MB
	// Current protocol version
	CurrentVersion = "1.0.0"
)


type RegistryHandler struct {
	node       *Node
	signer     OperatorSigner
}

// NewRegistryHandler creates a new registry handler
func NewRegistryHandler(node *Node, signer OperatorSigner, registryID peer.ID) *RegistryHandler {
	rh := &RegistryHandler{
		node:       node,
		signer:     signer,
	}
	return rh
}

// handleRegistryResponse processes a registry response on a stream
func (rh *RegistryHandler) handleRegistryResponse(stream network.Stream) error {
	log.Printf("[Registry] Handling incoming registry response...")

	// Read the response
	var resp pb.RegistryResponse
	if err := utils.ReadStreamMessage(stream, &resp, maxMessageSize); err != nil {
		log.Printf("[Registry] Failed to read registry response: %v", err)
		return fmt.Errorf("failed to read registry response: %w", err)
	}

	if !resp.Success {
		log.Printf("[Registry] Registration failed: %s", resp.Message)
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	// Convert and update operator states
	log.Printf("[Registry] Processing %d operator states from response", len(resp.ActiveOperators))
	states := make([]*OperatorState, 0, len(resp.ActiveOperators))
	for _, opState := range resp.ActiveOperators {
		state, err := rh.node.convertToOperatorState(opState)
		if err != nil {
			log.Printf("[Registry] Failed to convert operator state for peer %s: %v", opState.PeerId, err)
			continue
		}
		states = append(states, state)
	}

	// Update states using the common function
	rh.node.updateOperatorStates(context.Background(), states)
	return nil
}

// AuthToRegistry authenticates with the registry node
func (rh *RegistryHandler) AuthToRegistry(ctx context.Context) error {
	log.Printf("[Registry] Starting authentication with registry node...")
	
	// Create the stream
	stream, err := rh.node.createStreamToRegistry(ctx, RegistryProtocol)
	if err != nil {
		log.Printf("[Registry] Failed to create stream to registry: %v", err)
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Get node's multiaddrs
	multiaddrs := rh.node.host.Addrs()
	multiaddrsStr := make([]string, len(multiaddrs))
	for i, addr := range multiaddrs {
		multiaddrsStr[i] = addr.String()
	}
	log.Printf("[Registry] Using multiaddrs: %v", multiaddrsStr)

	// Prepare the register message
	timestamp := uint64(time.Now().Unix())
	message := crypto.Keccak256(
		[]byte(rh.signer.GetOperatorAddress().Hex()),
		utils.Uint64ToBytes(timestamp),
		[]byte(CurrentVersion),
	)

	// Sign the message
	signature, err := rh.signer.Sign(message)
	if err != nil {
		log.Printf("[Registry] Failed to sign auth message: %v", err)
		return fmt.Errorf("failed to sign message: %w", err)
	}
	log.Printf("[Registry] Successfully signed auth message")

	// Create the register request
	req := &pb.RegistryMessage{
		Type: pb.RegistryMessage_REGISTER,
		Message: &pb.RegistryMessage_Register{
			Register: &pb.RegisterMessage{
				Version:    CurrentVersion,
				Address:    rh.signer.GetOperatorAddress().Hex(),
				Timestamp: timestamp,
				Signature: hex.EncodeToString(signature),
				Multiaddrs: multiaddrsStr,
			},
		},
	}

	// Send the register request
	log.Printf("[Registry] Sending register request to registry...")
	if err := utils.WriteStreamMessage(stream, req); err != nil {
		log.Printf("[Registry] Failed to write register request: %v", err)
		return fmt.Errorf("failed to write register request: %w", err)
	}

	// Handle the response
	if err := rh.handleRegistryResponse(stream); err != nil {
		return err
	}

	log.Printf("[Registry] Successfully authenticated with registry node")
	return nil
}

// Disconnect sends disconnect request to registry
func (rh *RegistryHandler) Disconnect(ctx context.Context) error {
	log.Printf("[Registry] Initiating disconnect from registry node...")
	
	// Create the stream
	stream, err := rh.node.createStreamToRegistry(ctx, RegistryProtocol)
	if err != nil {
		log.Printf("[Registry] Failed to create stream for disconnect: %v", err)
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Create the disconnect request
	req := &pb.RegistryMessage{
		Type: pb.RegistryMessage_DISCONNECT,
	}

	// Send the disconnect request
	log.Printf("[Registry] Sending disconnect request...")
	if err := utils.WriteStreamMessage(stream, req); err != nil {
		log.Printf("[Registry] Failed to write disconnect request: %v", err)
		return fmt.Errorf("failed to write disconnect request: %w", err)
	}

	log.Printf("[Registry] Successfully sent disconnect request")
	return nil
}



