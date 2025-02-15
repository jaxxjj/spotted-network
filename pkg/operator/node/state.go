package node

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	manet "github.com/multiformats/go-multiaddr/net"
)

func (n *node) PrintConnectedPeers() {
	peers := n.host.Network().Peers()
	log.Printf("[Node] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := n.host.Network().Peerstore().Addrs(peer)
		log.Printf("[Node] - Peer %s at %v", peer.String(), addrs)
	}
}

func (n *node) PrintPeerInfo() {
	addrs := n.host.Addrs()

	var localAddrs, externalAddrs []string

	for _, addr := range addrs {
		ip, err := manet.ToIP(addr)
		if err != nil {
			continue
		}

		if ip.IsLoopback() || ip.IsPrivate() {
			localAddrs = append(localAddrs, addr.String())
		} else {
			externalAddrs = append(externalAddrs, addr.String())
		}
	}

	log.Printf("Node Peer Info:")
	log.Printf("PeerID: %s", n.host.ID().String())
	log.Printf("Local Addresses:")
	for _, addr := range localAddrs {
		log.Printf("  %s", addr)
	}
	log.Printf("External Addresses:")
	for _, addr := range externalAddrs {
		log.Printf("  %s", addr)
	}

	go func() {
		time.Sleep(5 * time.Second)

		newAddrs := n.host.Addrs()
		var newExternalAddrs []string

		for _, addr := range newAddrs {
			ip, err := manet.ToIP(addr)
			if err != nil {
				continue
			}
			if !ip.IsLoopback() && !ip.IsPrivate() {
				newExternalAddrs = append(newExternalAddrs, addr.String())
			}
		}

		if len(newExternalAddrs) > 0 {
			log.Printf("Discovered External Addresses:")
			for _, addr := range newExternalAddrs {
				log.Printf("  %s", addr)
			}
		}
	}()
}

func (n *node) GetConnectedPeers() []peer.ID {
	peers := n.host.Network().Peers()
	return peers
}

func (n *node) startPrintPeerInfo(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)

	n.PrintPeerInfo()

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("[Node] Stopping peer info printer")
				return
			case <-ticker.C:
				n.PrintPeerInfo()
			}
		}
	}()
}
