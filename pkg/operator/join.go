package operator

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
)

const (
	// Protocol message types
	MsgTypeAuthRequest  byte = 0x01
	MsgTypeAuthResponse byte = 0x02
	AuthTimeout              = 30 * time.Second
	RegistryProtocolID       = protocol.ID("/spotted/registry/1.0.0")
)

// Uint64ToBytes converts uint64 to bytes
func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	return b
}

// AuthHandler handles authentication related operations
type AuthHandler struct {
	node       *Node
	address    common.Address
	signer     OperatorSigner
	registryID peer.ID
	
	// 认证状态缓存
	registryInfo   *peer.AddrInfo
	registryInfoMu sync.RWMutex
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(node *Node, address common.Address, signer OperatorSigner, registryID peer.ID) *AuthHandler {
	return &AuthHandler{
		node:       node,
		address:    address,
		signer:     signer,
		registryID: registryID,
	}
}

// AuthToRegistry authenticates with the registry node
func (ah *AuthHandler) AuthToRegistry(ctx context.Context) error {
	// 创建stream
	stream, err := ah.node.host.NewStream(ctx, ah.registryID, RegistryProtocolID)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	// 准备认证请求
	timestamp := uint64(time.Now().Unix())
	message := crypto.Keccak256(
		[]byte(ah.address.Hex()),
		Uint64ToBytes(timestamp),
	)

	// 签名消息
	signature, err := ah.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// 创建请求
	req := &pb.AuthRequest{
		Address:   ah.address.Hex(),
		Timestamp: timestamp,
		Signature: hex.EncodeToString(signature),
	}

	// 发送认证请求
	if err := ah.sendAuthRequest(stream, req); err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}

	// 处理认证响应
	resp, err := ah.handleAuthResponse(stream)
	if err != nil {
		return fmt.Errorf("failed to handle auth response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("auth failed: %s", resp.Message)
	}

	// 处理registry连接
	if err := ah.handleRegistryConnection(resp.ActiveOperators); err != nil {
		return fmt.Errorf("failed to handle registry connection: %w", err)
	}

	log.Printf("[AUTH] Successfully authenticated with registry")
	return nil
}

// sendAuthRequest sends the auth request
func (ah *AuthHandler) sendAuthRequest(stream network.Stream, req *pb.AuthRequest) error {
	// 序列化请求
	reqData, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送消息类型
	if _, err := stream.Write([]byte{MsgTypeAuthRequest}); err != nil {
		return fmt.Errorf("failed to write message type: %w", err)
	}

	// 发送请求数据
	if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, reqData, AuthTimeout); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// handleAuthResponse handles the auth response
func (ah *AuthHandler) handleAuthResponse(stream network.Stream) (*pb.AuthResponse, error) {
	// 读取响应类型
	msgType := make([]byte, 1)
	if _, err := io.ReadFull(stream, msgType); err != nil {
		return nil, fmt.Errorf("failed to read response type: %w", err)
	}

	if msgType[0] != MsgTypeAuthResponse {
		return nil, fmt.Errorf("invalid response type: %d", msgType[0])
	}

	// 读取响应数据
	respData, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, AuthTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var resp pb.AuthResponse
	if err := proto.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// handleRegistryConnection handles the connection with registry
func (ah *AuthHandler) handleRegistryConnection(activeOperators []*pb.ActiveOperator) error {
	// 查找registry的信息
	var registryInfo *pb.ActiveOperator
	for _, op := range activeOperators {
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Printf("[WARN] Failed to decode peer ID %s: %v", op.PeerId, err)
			continue
		}
		if peerID == ah.registryID {
			registryInfo = op
			break
		}
	}

	if registryInfo == nil {
		return fmt.Errorf("registry info not found in active operators")
	}

	// 解析registry的地址
	addrs := make([]multiaddr.Multiaddr, 0, len(registryInfo.Multiaddrs))
	for _, addr := range registryInfo.Multiaddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.Printf("[WARN] Failed to parse multiaddr %s: %v", addr, err)
			continue
		}
		addrs = append(addrs, maddr)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("no valid addresses found for registry")
	}

	// 保存registry信息
	ah.registryInfoMu.Lock()
	ah.registryInfo = &peer.AddrInfo{
		ID:    ah.registryID,
		Addrs: addrs,
	}
	ah.registryInfoMu.Unlock()

	// 连接到其他active operators
	log.Printf("[AUTH] Connecting to other active operators...")
	for _, op := range activeOperators {
		// 跳过自己和registry
		peerID, err := peer.Decode(op.PeerId)
		if err != nil {
			log.Printf("[WARN] Failed to decode peer ID %s: %v", op.PeerId, err)
			continue
		}
		if peerID == ah.node.host.ID() || peerID == ah.registryID {
			continue
		}

		// 解析operator地址
		addrs := make([]multiaddr.Multiaddr, 0, len(op.Multiaddrs))
		for _, addr := range op.Multiaddrs {
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				log.Printf("[WARN] Failed to parse multiaddr %s: %v", addr, err)
				continue
			}
			addrs = append(addrs, maddr)
		}

		if len(addrs) == 0 {
			log.Printf("[WARN] No valid addresses found for operator %s", op.PeerId)
			continue
		}

		// 连接到operator
		peerInfo := &peer.AddrInfo{
			ID:    peerID,
			Addrs: addrs,
		}
		if err := ah.node.host.Connect(context.Background(), *peerInfo); err != nil {
			log.Printf("[WARN] Failed to connect to operator %s: %v", op.PeerId, err)
			continue
		}
		log.Printf("[AUTH] Successfully connected to operator %s", op.PeerId)
	}

	log.Printf("[AUTH] Successfully connected to registry (peer %s)", ah.registryID)
	return nil
}



