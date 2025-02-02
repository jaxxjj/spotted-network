package registry

import (
	"context"
	"fmt"
	"log"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	AuthTimeout = 30 * time.Second
	RegistryProtocolID = protocol.ID("/spotted/registry/1.0.0")
)



// RegistryHandler handles registry protocol operations
type RegistryHandler struct {
	node *Node
	opQuerier RegistryHandlerQuerier
}

type RegistryHandlerQuerier interface {
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
}

// NewRegistryHandler creates a new registry handler
func NewRegistryHandler(node *Node, opQuerier RegistryHandlerQuerier) *RegistryHandler {
	return &RegistryHandler{
		node: node,
		opQuerier: opQuerier,
	}
}

// HandleStream handles incoming registry streams
func (rh *RegistryHandler) HandleStream(stream network.Stream) {
	defer stream.Close()

	peerID := stream.Conn().RemotePeer()
	log.Printf("[Registry] New request from peer: %s", peerID.String())

	// Read and parse message with size limit from state sync
	var msg pb.RegistryMessage
	if err := utils.ReadStreamMessage(stream, &msg, maxMessageSize); err != nil {
		log.Printf("[Registry] Failed to read message: %v", err)
		stream.Reset()
		return
	}

	switch msg.Type {
	case pb.RegistryMessage_REGISTER:
		rh.handleRegister(stream, peerID, msg.GetRegister())
	case pb.RegistryMessage_DISCONNECT:
		rh.handleDisconnect(stream, peerID)
	default:
		log.Printf("[Registry] Unknown message type from %s: %v", peerID, msg.Type)
		stream.Reset()
	}
}

func (rh *RegistryHandler) handleRegister(stream network.Stream, peerID peer.ID, req *pb.RegisterMessage) {
	// verify the request
	success, msg := rh.verifyAuthRequest(context.Background(), req)

	// prepare the response
	activeOperators := []*pb.OperatorPeerState{}
	multiaddrs, err := utils.StringsToMultiaddrs(req.Multiaddrs)
	if err != nil {
		log.Printf("[Registry] Failed to convert multiaddrs to strings: %v", err)
		stream.Reset()
		return
	}
	
	if success {
		// create the new operator state
		now := time.Now()
		state := &OperatorPeerInfo{
			Address:    req.Address,
			PeerID:     peerID,
			Multiaddrs: multiaddrs,
			LastSeen:   now,
		}
		
		log.Printf("[Registry] Creating new operator state for %s (peer %s)", req.Address, peerID)
		
		// update the operator state
		rh.node.UpdateOperatorState(peerID, state)
		
		// get the active operators for the response
		activeOperators = rh.getActiveOperators()
		
		log.Printf("[Registry] Operator %s (peer %s) registered successfully", req.Address, peerID)
	}

	resp := &pb.RegistryResponse{
		Success: success,
		Message: msg,
		ActiveOperators: activeOperators,
	}

	// Send response
	if err := utils.WriteStreamMessage(stream, resp); err != nil {
		log.Printf("[Registry] Failed to write response: %v", err)
		stream.Reset()
		return
	}

	if !success {
		log.Printf("[Registry] Registration failed for peer %s: %s", peerID.String(), msg)
		stream.Reset()
		return
	}

	// update the operator state
	rh.node.UpdateOperatorState(peerID, state)
}

func (rh *RegistryHandler) handleDisconnect(stream network.Stream, peerID peer.ID) {
	// Remove from active operators
	rh.node.RemoveOperator(peerID)
	
	resp := &pb.RegistryResponse{
		Success: true,
		Message: "Disconnected successfully",
	}

	if err := utils.WriteStreamMessage(stream, resp); err != nil {
		log.Printf("[Registry] Failed to write disconnect response: %v", err)
		stream.Reset()
		return
	}

	log.Printf("[Registry] Operator (peer %s) disconnected", peerID)
}

// verifyAuthRequest verifies the auth request
func (rh *RegistryHandler) verifyAuthRequest(ctx context.Context, req *pb.RegisterMessage) (bool, string) {
	// verify the timestamp
	if time.Since(time.Unix(int64(req.Timestamp), 0)) > AuthTimeout {
		return false, "auth request expired"
	}

	// verify the operator status
	operator, err := rh.opQuerier.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		return false, fmt.Sprintf("failed to get operator: %v", err)
	}

	if operator == nil {
		return false, "operator not found"
	}

	if operator.Status != "active" {
		return false, fmt.Sprintf("operator is not active, current status: %s", operator.Status)
	}

	// verify the signature
	message := crypto.Keccak256(
		[]byte(req.Address),
		utils.Uint64ToBytes(req.Timestamp),
	)
	messageHash := crypto.Keccak256Hash(message)

	sigBytes := ethcommon.FromHex(req.Signature)
	pubKey, err := crypto.SigToPub(messageHash.Bytes(), sigBytes)
	if err != nil {
		return false, "invalid signature"
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	if recoveredAddr.Hex() != req.Address {
		return false, "signature verification failed"
	}

	return true, "auth successful"
}

// getActiveOperators returns list of active operators
func (rh *RegistryHandler) getActiveOperators() []*pb.OperatorPeerState {
	// get the registry node info
	registryInfo := peer.AddrInfo{
		ID:    rh.node.host.ID(),
		Addrs: rh.node.host.Addrs(),
	}

	// create the active operators list, include the registry
	activeOperators := make([]*pb.OperatorPeerState, 0)
	
	// add the registry info
	registryOp := &pb.OperatorPeerState{
		PeerId:     registryInfo.ID.String(),
		Multiaddrs: utils.MultiaddrsToStrings(registryInfo.Addrs),
	}
	activeOperators = append(activeOperators, registryOp)

	// add all active operators
	rh.node.activeOperators.mu.RLock()
	for _, state := range rh.node.activeOperators.active {
		op := &pb.OperatorPeerState{
			PeerId:     state.PeerID.String(),
			Multiaddrs: utils.MultiaddrsToStrings(state.Multiaddrs),
			Address:    state.Address,
		}
		activeOperators = append(activeOperators, op)
	}
	rh.node.activeOperators.mu.RUnlock()

	return activeOperators
}



