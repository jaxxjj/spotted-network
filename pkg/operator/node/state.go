package node

import (
	"fmt"
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

// GetPeerAgentVersion 获取指定peer的AgentVersion
func (n *node) GetPeerAgentVersion(peerID peer.ID) (string, error) {
	// 从peerstore获取AgentVersion
	av, err := n.host.Peerstore().Get(peerID, "AgentVersion")
	if err != nil {
		log.Printf("[Node] Failed to get agent version for peer %s: %v", peerID, err)
		return "", fmt.Errorf("failed to get agent version: %w", err)
	}

	// 类型断言
	version, ok := av.(string)
	if !ok {
		log.Printf("[Node] Invalid agent version type for peer %s", peerID)
		return "", fmt.Errorf("invalid agent version type")
	}

	return version, nil
}
