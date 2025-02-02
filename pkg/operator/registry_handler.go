package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
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
	"google.golang.org/protobuf/proto"
)

const (
	// Protocol message types
	MsgTypeAuthRequest  byte = 0x01
	MsgTypeAuthResponse byte = 0x02
	AuthTimeout              = 30 * time.Second
	RegistryProtocolID       = protocol.ID("/spotted/registry/1.0.0")
)


// AuthHandler handles authentication related operations
type RegistryHandler struct {
	node       *Node
	address    common.Address
	signer     OperatorSigner
	registryID peer.ID
	
	// auth status cache
	registryInfo   *peer.AddrInfo
	registryInfoMu sync.RWMutex
}

// NewAuthHandler creates a new auth handler
func NewRegistryHandler(node *Node, address common.Address, signer OperatorSigner, registryID peer.ID) *RegistryHandler {
	return &RegistryHandler{
		node:       node,
		address:    address,
		signer:     signer,
		registryID: registryID,
	}
}

// AuthToRegistry authenticates with the registry node
func (rh *RegistryHandler) AuthToRegistry(ctx context.Context) error {
	// create the stream
	stream, err := rh.node.host.NewStream(ctx, rh.registryID, RegistryProtocolID)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// prepare the auth request
	timestamp := uint64(time.Now().Unix())
	message := crypto.Keccak256(
		[]byte(rh.address.Hex()),
		utils.Uint64ToBytes(timestamp),
	)

	// sign the message
	signature, err := rh.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// create the request
	req := &pb.AuthRequest{
		Address:   rh.address.Hex(),
		Timestamp: timestamp,
		Signature: hex.EncodeToString(signature),
	}

	// send the auth request
	if err := rh.sendAuthRequest(stream, req); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// handle the auth response
	resp, err := rh.handleAuthResponse(stream)
	if err != nil {
		return fmt.Errorf("failed to handle auth response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("auth failed: %s", resp.Message)
	}

	// handle the registry connection
	if err := rh.handleRegistryConnection(resp.ActiveOperators); err != nil {
		return fmt.Errorf("failed to handle registry connection: %w", err)
	}

	log.Info().
		Str("component", "auth").
		Msg("Successfully authenticated with registry")
	return nil
}

// sendAuthRequest sends the auth request
func (rh *RegistryHandler) sendAuthRequest(stream network.Stream, req *pb.AuthRequest) error {
	reqData, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// send the message type
	if _, err := stream.Write([]byte{MsgTypeAuthRequest}); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// send the request data
	if err := utils.WriteLengthPrefixedDataWithDeadline(stream, reqData, AuthTimeout); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// handleAuthResponse handles the auth response
func (rh *RegistryHandler) handleRegistryStream(stream network.Stream) (*pb.AuthResponse, error) {
	// read the response type
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		return nil, fmt.Errorf("failed to read response type: %w", err)
	}

	if msgType[0] != MsgTypeAuthResponse {
		return nil, fmt.Errorf("invalid response type: %d", msgType[0])
	}

	// read the response data
	respData, err := utils.ReadLengthPrefixedDataWithDeadline(stream, AuthTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// parse the response
	var resp pb.AuthResponse
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// handleRegistryConnection handles the connection with registry
func (rh *RegistryHandler) handleRegistryConnection(activeOperators []*pb.ActiveOperator) error {
	// find the registry info
	var registryInfo *pb.ActiveOperator
	for _, op := range activeOperators {
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Warn().
				Str("component", "auth").
				Str("peer_id", op.PeerId).
				Err(err).
				Msg("Failed to decode peer ID")
			continue
		}
		if peerID == rh.registryID {
			registryInfo = op
			break
		}
	}

	if registryInfo == nil {
		return fmt.Errorf("registry info not found in active operators")
	}

	// parse the registry addresses
	addrs := make([]multiaddr.Multiaddr, 0, len(registryInfo.Multiaddrs))
	for _, addr := range registryInfo.Multiaddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Warn().
				Str("component", "auth").
				Str("multiaddr", addr).
				Err(err).
				Msg("Failed to parse multiaddr")
			continue
		}
		addrs = append(addrs, maddr)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("no valid addresses found for registry")
	}

	// save the registry info
	rh.registryInfoMu.Lock()
	rh.registryInfo = &peer.AddrInfo{
		ID:    rh.registryID,
		Addrs: addrs,
	}
	rh.registryInfoMu.Unlock()

	// connect to other active operators
	log.Info().
		Str("component", "auth").
		Msg("Connecting to other active operators...")
	
	for _, op := range activeOperators {
		// skip the self and registry
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Warn().
				Str("component", "auth").
				Str("peer_id", op.PeerId).
				Err(err).
				Msg("Failed to decode peer ID")
			continue
		}
		if peerID == rh.node.host.ID() || peerID == rh.registryID {
			continue
		}

		// parse the operator addresses
		addrs := make([]multiaddr.Multiaddr, 0, len(op.Multiaddrs))
		for _, addr := range op.Multiaddrs {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Warn().
					Str("component", "auth").
					Str("multiaddr", addr).
					Err(err).
					Msg("Failed to parse multiaddr")
				continue
			}
			addrs = append(addrs, maddr)
		}

		if len(addrs) == 0 {
			log.Warn().
				Str("component", "auth").
				Str("peer_id", op.PeerId).
				Msg("No valid addresses found for operator")
			continue
		}

		// connect to the operator
		peerInfo := &peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}
		if err := rh.node.host.Connect(context.Background(), *peerInfo); err != nil {
			log.Warn().
				Str("component", "auth").
				Str("peer_id", op.PeerId).
				Err(err).
				Msg("Failed to connect to operator")
			continue
		}
		log.Info().
			Str("component", "auth").
			Str("peer_id", op.PeerId).
			Msg("Successfully connected to operator")
	}

	log.Info().
		Str("component", "auth").
		Str("registry_id", rh.registryID.String()).
		Msg("Successfully connected to registry")
	return nil
}



