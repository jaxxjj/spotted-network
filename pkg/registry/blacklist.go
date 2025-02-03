package registry

import (
	"log"

	"github.com/libp2p/go-libp2p/core/peer"
)

// BlacklistPeer adds a peer to the blacklist
func (n *Node) blacklistPeer(pid peer.ID, reason string) {
	if n.blacklist != nil {
		log.Printf("[Blacklist] Adding peer %s to blacklist: %s", pid, reason)
		n.blacklist.Add(pid)
		n.pubsub.BlacklistPeer(pid)
	}
}

