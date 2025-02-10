package node

import (
	"log"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *node) PrintConnectedPeers() {
	peers := n.host.Network().Peers()
	log.Printf("[Node] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := n.host.Network().Peerstore().Addrs(peer)
		log.Printf("[Node] - Peer %s at %v", peer.String(), addrs)
	}
}

// PrintPeerInfo prints the node's peer information
func (n *node) PrintPeerInfo() {
	// Get node's addresses
	addrs := n.host.Addrs()
	addrStrings := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addrStrings = append(addrStrings, addr.String())
	}

	log.Printf("Node Peer Info:")
	log.Printf("PeerID: %s", n.host.ID().String())
	log.Printf("Addresses: \n%s", strings.Join(addrStrings, "\n"))
}

func (n *node) GetConnectedPeers() []peer.ID {
	peers := n.host.Network().Peers()
	return peers
}
