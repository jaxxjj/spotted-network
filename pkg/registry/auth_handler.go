package registry

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"google.golang.org/protobuf/proto"
)

const (
	MsgTypeAuthRequest  byte = 0x01
	MsgTypeAuthResponse byte = 0x02
	AuthTimeout              = 30 * time.Second
	RegistryProtocolID       = protocol.ID("/spotted/registry/1.0.0")
)

type OperatorsQuerier interface {
	UpdateOperatorStatus(ctx context.Context, arg operators.UpdateOperatorStatusParams, getOperatorByAddress *string) (*operators.Operators, error)
	ListAllOperators(ctx context.Context) ([]operators.Operators, error)
	GetOperatorByAddress(ctx context.Context, address string) (*operators.Operators, error)
	UpsertOperator(ctx context.Context, arg operators.UpsertOperatorParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorState(ctx context.Context, arg operators.UpdateOperatorStateParams, getOperatorByAddress *string) (*operators.Operators, error)
	UpdateOperatorExitEpoch(ctx context.Context, arg operators.UpdateOperatorExitEpochParams, getOperatorByAddress *string) (*operators.Operators, error)
}

// AuthHandler handles authentication related operations
type AuthHandler struct {
	node      *Node
	operators OperatorsQuerier
	
	// 认证状态缓存
	authedPeers   map[peer.ID]string  // peer.ID -> operator address
	authedPeersMu sync.RWMutex
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(node *Node, operators OperatorsQuerier) *AuthHandler {
	return &AuthHandler{
		node:        node,
		operators:   operators,
		authedPeers: make(map[peer.ID]string),
	}
}

// HandleStream handles incoming auth streams
func (ah *AuthHandler) HandleStream(stream network.Stream) {
	defer stream.Close()

	peerID := stream.Conn().RemotePeer()
	log.Printf("[AUTH] New auth request from peer: %s", peerID.String())

	// 读取消息类型
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		log.Printf("[AUTH] Failed to read message type: %v", err)
		stream.Reset()
		return
	}

	if msgType[0] != MsgTypeAuthRequest {
		log.Printf("[AUTH] Invalid message type: %d", msgType[0])
		stream.Reset()
		return
	}

	// 读取认证请求
	data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, AuthTimeout)
	if err != nil {
		log.Printf("[AUTH] Failed to read auth request: %v", err)
		stream.Reset()
		return
	}

	// 解析请求
	var req pb.AuthRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("[AUTH] Failed to unmarshal request: %v", err)
		stream.Reset()
		return
	}

	// 验证请求
	success, msg := ah.verifyAuthRequest(context.Background(), &req)

	// 准备响应
	activeOperators := []*pb.ActiveOperator{}
	if success {
		// 创建新的operator状态
		addrs := ah.node.host.Peerstore().Addrs(peerID)
		state := &OperatorState{
			Address:    req.Address,
			PeerID:     peerID,
			Multiaddrs: addrs,
			LastSeen:   time.Now(),
			Status:     "active",
		}
		
		log.Printf("[AUTH] Creating new operator state for %s (peer %s)", req.Address, peerID)
		
		// 更新operator状态
		ah.node.UpdateOperatorState(peerID, state)
		
		// 获取活跃operator列表用于响应
		activeOperators = ah.getActiveOperators()
		
		// 广播新operator加入
		update := &pb.OperatorStateUpdate{
			Type: "operator_joined",
			Operators: []*pb.OperatorState{
				{
					Address: req.Address,
					Status:  "active",
				},
			},
		}
		
		// 使用统一的广播机制
		ah.node.BroadcastStateUpdate(update.Operators, update.Type)
		
		// 记录认证成功
		ah.authedPeersMu.Lock()
		ah.authedPeers[peerID] = req.Address
		ah.authedPeersMu.Unlock()
		
		log.Printf("[AUTH] Operator %s (peer %s) added to authenticated peers", req.Address, peerID)
	}

	resp := &pb.AuthResponse{
		Success: success,
		Message: msg,
		ActiveOperators: activeOperators,
	}

	// 发送响应
	if err := ah.sendAuthResponse(stream, resp); err != nil {
		log.Printf("[AUTH] Failed to send response: %v", err)
		stream.Reset()
		return
	}

	if !success {
		log.Printf("[AUTH] Auth failed for peer %s: %s", peerID.String(), msg)
		stream.Reset()
		return
	}

	log.Printf("[AUTH] Successfully completed authentication for operator %s (peer %s)", req.Address, peerID.String())
}

// Uint64ToBytes converts uint64 to bytes
func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// verifyAuthRequest verifies the auth request
func (ah *AuthHandler) verifyAuthRequest(ctx context.Context, req *pb.AuthRequest) (bool, string) {
	// 验证时间戳
	if time.Since(time.Unix(int64(req.Timestamp), 0)) > AuthTimeout {
		return false, "auth request expired"
	}

	// 验证operator状态
	operator, err := ah.operators.GetOperatorByAddress(ctx, req.Address)
	if err != nil {
		return false, fmt.Sprintf("failed to get operator: %v", err)
	}

	if operator == nil {
		return false, "operator not found"
	}

	if operator.Status != "active" {
		return false, fmt.Sprintf("operator is not active, current status: %s", operator.Status)
	}

	// 验证签名
	message := crypto.Keccak256(
		[]byte(req.Address),
		Uint64ToBytes(req.Timestamp),
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

// sendAuthResponse sends the auth response
func (ah *AuthHandler) sendAuthResponse(stream network.Stream, resp *pb.AuthResponse) error {
	// 序列化响应
	respData, err := proto.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// 发送响应类型
	if _, err := stream.Write([]byte{MsgTypeAuthResponse}); err != nil {
		return fmt.Errorf("failed to write response type: %w", err)
	}

	// 发送响应数据
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, respData, AuthTimeout); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}

// getActiveOperators returns list of active operators
func (ah *AuthHandler) getActiveOperators() []*pb.ActiveOperator {
	// 获取registry节点信息
	registryInfo := peer.AddrInfo{
		ID:    ah.node.host.ID(),
		Addrs: ah.node.host.Addrs(),
	}

	// 创建活跃operators列表，包含registry
	activeOperators := make([]*pb.ActiveOperator, 0)
	
	// 添加registry信息
	registryOp := &pb.ActiveOperator{
		PeerId:     registryInfo.ID.String(),
		Multiaddrs: make([]string, len(registryInfo.Addrs)),
	}
	for i, addr := range registryInfo.Addrs {
		registryOp.Multiaddrs[i] = addr.String()
	}
	activeOperators = append(activeOperators, registryOp)

	// 添加所有活跃的operators
	ah.node.operators.mu.RLock()
	for _, state := range ah.node.operators.active {
		op := &pb.ActiveOperator{
			PeerId:     state.PeerID.String(),
			Multiaddrs: make([]string, len(state.Multiaddrs)),
		}
		for i, addr := range state.Multiaddrs {
			op.Multiaddrs[i] = addr.String()
		}
		activeOperators = append(activeOperators, op)
	}
	ah.node.operators.mu.RUnlock()

	return activeOperators
}

// IsAuthenticated checks if a peer is authenticated
func (ah *AuthHandler) IsAuthenticated(peerID peer.ID) bool {
	ah.authedPeersMu.RLock()
	defer ah.authedPeersMu.RUnlock()
	_, exists := ah.authedPeers[peerID]
	return exists
}

// GetOperatorAddress gets the operator address for an authenticated peer
func (ah *AuthHandler) GetOperatorAddress(peerID peer.ID) (string, bool) {
	ah.authedPeersMu.RLock()
	defer ah.authedPeersMu.RUnlock()
	addr, exists := ah.authedPeers[peerID]
	return addr, exists
}

// RemoveAuthPeer removes a peer from authenticated list
func (ah *AuthHandler) RemoveAuthPeer(peerID peer.ID) {
	ah.authedPeersMu.Lock()
	delete(ah.authedPeers, peerID)
	ah.authedPeersMu.Unlock()
}

