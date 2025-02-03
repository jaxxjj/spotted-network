package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	utils "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

const (
	AuthTimeout              = 30 * time.Second
	RegistryProtocol       = protocol.ID("/spotted/registry/1.0.0")
	maxMessageSize = 1 << 20 // 1MB
)

type StreamCreator interface {
	Addrs() []multiaddr.Multiaddr
	createStream(ctx context.Context, p peer.ID, pid protocol.ID) (network.Stream, error)
}

// RegistryHandler handles registry protocol operations
type RegistryHandler struct {
	node       StreamCreator
	address    common.Address
	signer     OperatorSigner
	registryID peer.ID
	
	// registry info cache
	registryInfo   *peer.AddrInfo
	registryInfoMu sync.RWMutex
}

// NewRegistryHandler creates a new registry handler
func NewRegistryHandler(node StreamCreator, address common.Address, signer OperatorSigner, registryID peer.ID) *RegistryHandler {
	return &RegistryHandler{
		node:       node,
		address:    address,
		signer:     signer,
		registryID: registryID,
	}
}

// AuthToRegistry authenticates with the registry node
func (rh *RegistryHandler) AuthToRegistry(ctx context.Context) error {
	// Create the stream
	stream, err := rh.node.createStream(ctx, rh.registryID, RegistryProtocol)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Get node's multiaddrs
	multiaddrs := rh.node.Addrs()
	multiaddrsStr := make([]string, len(multiaddrs))
	for i, addr := range multiaddrs {
		multiaddrsStr[i] = addr.String()
	}

	// Prepare the register message
	timestamp := uint64(time.Now().Unix())
	message := crypto.Keccak256(
		[]byte(rh.address.Hex()),
		utils.Uint64ToBytes(timestamp),
	)

	// Sign the message
	signature, err := rh.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// Create the register request
	req := &pb.RegistryMessage{
		Type: pb.RegistryMessage_REGISTER,
		Message: &pb.RegistryMessage_Register{
			Register: &pb.RegisterMessage{
				Address:    rh.address.Hex(),
				Timestamp: timestamp,
				Signature: hex.EncodeToString(signature),
				Multiaddrs: multiaddrsStr,
			},
		},
	}

	// Send the register request
	if err := utils.WriteStreamMessage(stream, req); err != nil {
		return fmt.Errorf("failed to write register request: %w", err)
	}

	// Read the response
	var resp pb.RegistryResponse
	if err := utils.ReadStreamMessage(stream, &resp, maxMessageSize); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	// Update registry info
	rh.registryInfoMu.Lock()
	rh.registryInfo = &peer.AddrInfo{
		ID: rh.registryID,
	}
	rh.registryInfoMu.Unlock()

	log.Info().
		Str("component", "registry").
		Str("address", rh.address.Hex()).
		Msg("Successfully registered with registry")

	return nil
}

// Disconnect sends disconnect request to registry
func (rh *RegistryHandler) Disconnect(ctx context.Context) error {
	// Create the stream
	stream, err := rh.node.createStream(ctx, rh.registryID, RegistryProtocol)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// Create the disconnect request
	req := &pb.RegistryMessage{
		Type: pb.RegistryMessage_DISCONNECT,
	}

	// Send the disconnect request
	if err := utils.WriteStreamMessage(stream, req); err != nil {
		return fmt.Errorf("failed to write disconnect request: %w", err)
	}

	// Read the response
	var resp pb.RegistryResponse
	if err := utils.ReadStreamMessage(stream, &resp, maxMessageSize); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("disconnect failed: %s", resp.Message)
	}

	log.Info().
		Str("component", "registry").
		Str("address", rh.address.Hex()).
		Msg("Successfully disconnected from registry")

	return nil
}



