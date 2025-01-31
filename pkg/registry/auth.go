// pkg/registry/auth.go

package registry

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	commonHelpers "github.com/galxe/spotted-network/pkg/common"
	pb "github.com/galxe/spotted-network/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/protobuf/proto"
)

// handleAuthRequest handles incoming authentication requests
func (n *Node) handleAuthRequest(stream network.Stream) {
    defer stream.Close()
    
    remotePeer := stream.Conn().RemotePeer()
    log.Printf("[Auth] New auth request from %s", remotePeer)

    // Read request with deadline
    data, err := commonHelpers.ReadLengthPrefixedDataWithDeadline(stream, 10*time.Second)
    if err != nil {
        log.Printf("[Auth] Failed to read request from %s: %v", remotePeer, err)
        stream.Reset()
        return
    }

    // Parse request
    var req pb.AuthRequest
    if err := proto.Unmarshal(data, &req); err != nil {
        log.Printf("[Auth] Failed to unmarshal request from %s: %v", remotePeer, err)
        stream.Reset()
        return
    }

    // Verify operator registration
    isRegistered, err := n.mainnetClient.IsOperatorRegistered(context.Background(), ethcommon.HexToAddress(req.Address))
    if err != nil {
        log.Printf("[Auth] Failed to verify operator %s registration: %v", req.Address, err)
        n.sendAuthResponse(stream, false, fmt.Sprintf("failed to verify registration: %v", err))
        return
    }

    if !isRegistered {
        log.Printf("[Auth] Operator %s is not registered", req.Address)
        n.sendAuthResponse(stream, false, "operator not registered")
        return
    }

    // Verify signature
    sigBytes, err := hex.DecodeString(req.Signature)
    if err != nil {
        log.Printf("[Auth] Invalid signature format from %s: %v", remotePeer, err)
        n.sendAuthResponse(stream, false, "invalid signature format")
        return
    }

    // TODO: Add signature verification logic here
    // This should match the verification logic from the existing join process

    // Add to allowlist
    n.allowlist.Add(remotePeer)
    
    // Store operator info
    n.operatorsInfoMu.Lock()
    n.operatorsInfo[remotePeer] = &OperatorInfo{
        ID:       remotePeer,
        Addrs:    n.host.Peerstore().Addrs(remotePeer),
        LastSeen: time.Now(),
        Status:   "active",
    }
    n.operatorsInfoMu.Unlock()

    // Send success response
    n.sendAuthResponse(stream, true, "")
    
    log.Printf("[Auth] Successfully authenticated operator %s", req.Address)
}

func (n *Node) sendAuthResponse(stream network.Stream, success bool, errMsg string) {
    resp := &pb.AuthResponse{
        Success: success,
        Error:   errMsg,
    }

    data, err := proto.Marshal(resp)
    if err != nil {
        log.Printf("[Auth] Failed to marshal response: %v", err)
        stream.Reset()
        return
    }

    if err := commonHelpers.WriteLengthPrefixedDataWithDeadline(stream, data, 10*time.Second); err != nil {
        log.Printf("[Auth] Failed to write response: %v", err)
        stream.Reset()
        return
    }
}

// IsAllowed checks if a peer is allowed to access the network
func (n *Node) IsAllowed(peerID peer.ID) bool {
    return n.allowlist.Allowed(peerID)
}

// RemoveFromAllowlist removes a peer from the allowlist
func (n *Node) RemoveFromAllowlist(peerID peer.ID) {
    n.allowlist.Remove(peerID)
    
    // Also remove from operators info
    n.operatorsInfoMu.Lock()
    delete(n.operatorsInfo, peerID)
    n.operatorsInfoMu.Unlock()
}