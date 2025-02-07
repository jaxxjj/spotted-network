package node

import (
	"log"

	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) PrintConnectedPeers() {
	peers := n.host.Network().Peers()
	log.Printf("[Node] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := n.host.Network().Peerstore().Addrs(peer)
		log.Printf("[Node] - Peer %s at %v", peer.String(), addrs)
	}
}

func (n *Node) GetConnectedPeers() []peer.ID {
	peers := n.host.Network().Peers()
	return peers
}


