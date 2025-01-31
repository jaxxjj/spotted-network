package operator

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// operatorGater implements the connection gating interface for operator nodes
type operatorGater struct {
    mu sync.RWMutex
    allowedPeers map[peer.ID]struct{}
    registryID peer.ID
    node *Node // reference to Node for accessing operator info
}

// Ensure operatorGater implements connmgr.ConnectionGater
var _ connmgr.ConnectionGater = (*operatorGater)(nil)

// newOperatorGater creates a new connection gater for the operator
func newOperatorGater(node *Node, registryID peer.ID) *operatorGater {
    gater := &operatorGater{
        allowedPeers: make(map[peer.ID]struct{}),
        registryID:   registryID,
        node:        node,
    }
    // 初始化时允许registry连接
    gater.allowedPeers[registryID] = struct{}{}
    return gater
}

// InterceptPeerDial tests whether we're allowed to Dial the specified peer
func (g *operatorGater) InterceptPeerDial(p peer.ID) bool {
    g.mu.RLock()
    defer g.mu.RUnlock()
    _, ok := g.allowedPeers[p]
    return ok
}

// InterceptAddrDial tests whether we're allowed to dial the specified peer at the specified addr
func (g *operatorGater) InterceptAddrDial(p peer.ID, addr ma.Multiaddr) bool {
    return g.InterceptPeerDial(p)
}

// InterceptAccept tests whether an incoming connection is allowed
func (g *operatorGater) InterceptAccept(connMultiaddrs network.ConnMultiaddrs) bool {
    return true // allow incoming connections, verify during handshake
}

// InterceptSecured tests whether a secured connection is allowed
func (g *operatorGater) InterceptSecured(direction network.Direction, p peer.ID, connMultiaddrs network.ConnMultiaddrs) bool {
    return g.InterceptPeerDial(p)
}

// InterceptUpgraded tests whether an upgraded connection is allowed
func (g *operatorGater) InterceptUpgraded(conn network.Conn) (allow bool, reason control.DisconnectReason) {
    return true, 0 // allow all upgraded connections
}

// AllowPeer adds a peer to the allowed list
func (g *operatorGater) AllowPeer(p peer.ID) {
    g.mu.Lock()
    g.allowedPeers[p] = struct{}{}
    g.mu.Unlock()
}

// RemovePeer removes a peer from the allowed list
func (g *operatorGater) RemovePeer(p peer.ID) {
    g.mu.Lock()
    delete(g.allowedPeers, p)
    g.mu.Unlock()
}

// IsRegistryAllowed checks if the registry is allowed
func (g *operatorGater) IsRegistryAllowed() bool {
    return g.InterceptPeerDial(g.registryID)
}

// AllowOperator adds another operator to the allowed list
func (g *operatorGater) AllowOperator(operatorID peer.ID) {
    if operatorID != g.node.host.ID() { 
        g.AllowPeer(operatorID)
    }
}