package node

import (
	"context"
	"fmt"
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
	// 获取所有地址
	addrs := n.host.Addrs()

	// 分类存储地址
	var localAddrs, externalAddrs []string

	for _, addr := range addrs {
		// 解析地址
		ip, err := manet.ToIP(addr)
		if err != nil {
			continue
		}

		// 判断地址类型
		if ip.IsLoopback() || ip.IsPrivate() {
			localAddrs = append(localAddrs, addr.String())
		} else {
			externalAddrs = append(externalAddrs, addr.String())
		}
	}

	// 打印信息
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

	// 等待外部地址发现
	go func() {
		// 等待一段时间让 NAT 发现完成
		time.Sleep(5 * time.Second)

		// 重新获取地址并打印新发现的外部地址
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

// 添加新的方法
func (n *node) startPrintPeerInfo(ctx context.Context) {
	// 创建一个30秒的定时器
	ticker := time.NewTicker(30 * time.Second)

	// 立即打印一次初始信息
	n.PrintPeerInfo()

	// 启动定时打印循环
	go func() {
		// 将defer移到goroutine内部,确保在goroutine结束时才停止ticker
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
