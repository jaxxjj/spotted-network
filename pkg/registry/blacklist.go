package registry

import (
	"github.com/libp2p/go-libp2p/core/peer"
)

// BlacklistPeer adds a peer to the blacklist
func (n *Node) blacklistPeer(pid peer.ID) {
		n.pubsub.BlacklistPeer(pid)
}

