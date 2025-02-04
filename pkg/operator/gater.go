package operator

import (
	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

// operatorGater implements connection gating for operator nodes
type operatorGater struct {
    node *Node 
}


var _ connmgr.ConnectionGater = (*operatorGater)(nil)


func newOperatorGater(node *Node) *operatorGater {
    return &operatorGater{
        node: node,
    }
}

// isAllowed checks if the peer is in the active operators map
func (g *operatorGater) isAllowed(peerID peer.ID) bool {
	if peerID == g.node.registryID {
		return true
	}
    g.node.activeOperators.mu.RLock()
    _, exists := g.node.activeOperators.active[peerID]
    g.node.activeOperators.mu.RUnlock()
    return exists
}

// InterceptPeerDial tests whether we're allowed to Dial the specified peer
func (g *operatorGater) InterceptPeerDial(peerID peer.ID) bool {
    allowed := g.isAllowed(peerID)
    if !allowed {
        log.Debug().Str("peer", peerID.String()).Msg("peer not in active operators list")
    }
    return allowed
}

// InterceptAddrDial tests whether we're allowed to dial the specified peer at the specified addr
func (g *operatorGater) InterceptAddrDial(peerID peer.ID, addr ma.Multiaddr) bool {
    return g.isAllowed(peerID)
}

// InterceptAccept tests whether an incoming connection is allowed
func (g *operatorGater) InterceptAccept(addrs network.ConnMultiaddrs) bool {
    return true // Accept all incoming connections initially, peer ID check happens in InterceptSecured
}

// InterceptSecured tests whether a secured connection is allowed
func (g *operatorGater) InterceptSecured(direction network.Direction, peerID peer.ID, addrs network.ConnMultiaddrs) bool {
    return g.isAllowed(peerID)
}

// InterceptUpgraded tests whether an upgraded connection is allowed
func (g *operatorGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
    return true, 0
}