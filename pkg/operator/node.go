package operator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host        *p2p.Host
	registryID  peer.ID
	registryAddr string
	signer      signer.Signer

	// Known operators map stores operator IDs and their addresses
	knownOperators map[peer.ID]*peer.AddrInfo
	operatorsMu    sync.RWMutex
}

func NewNode(registryAddr string, s signer.Signer) (*Node, error) {
	// Parse the registry multiaddr
	maddr, err := multiaddr.NewMultiaddr(registryAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry address: %w", err)
	}

	// Extract peer ID from multiaddr
	addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registry peer info: %w", err)
	}

	// Create a new host with default configuration
	cfg := &p2p.Config{
		ListenAddrs: []string{"/ip4/0.0.0.0/tcp/0"},
	}
	
	host, err := p2p.NewHost(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	return &Node{
		host:           host,
		registryID:     addrInfo.ID,
		registryAddr:   registryAddr,
		signer:         s,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operatorsMu:    sync.RWMutex{},
	}, nil
}

func (n *Node) Start() error {
	// Connect to registry
	if err := n.connectToRegistry(); err != nil {
		return fmt.Errorf("failed to connect to registry: %v", err)
	}

	// Start message handler first to ensure we don't miss any operator announcements
	go n.handleMessages()

	// Start health check
	go n.healthCheck()

	// Announce ourselves to the registry by opening a stream
	s, err := n.host.NewStream(context.Background(), n.registryID, "/spotted/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to announce to registry: %v", err)
	}
	s.Close()

	log.Printf("Operator node started with ID: %s\n", n.host.ID())
	return nil
}

func (n *Node) handleMessages() {
	n.host.SetStreamHandler("/spotted/1.0.0", func(s network.Stream) {
		defer s.Close()

		// Read the message
		buf := make([]byte, 1024)
		readN, err := s.Read(buf)
		if err != nil {
			log.Printf("Error reading from stream: %v\n", err)
			return
		}

		msg := string(buf[:readN])
		parts := strings.Split(msg, ":")
		if len(parts) < 3 || parts[0] != "NEW_OPERATOR" {
			log.Printf("Invalid message format: %s\n", msg)
			return
		}

		// Parse operator ID
		operatorID, err := peer.Decode(parts[1])
		if err != nil {
			log.Printf("Invalid operator ID in message: %v\n", err)
			return
		}

		// Skip if this is our own ID
		if operatorID == n.host.ID() {
			return
		}

		// Parse addresses
		addrStrs := strings.Split(parts[2], ",")
		addrs := make([]multiaddr.Multiaddr, 0, len(addrStrs))
		for _, addrStr := range addrStrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Printf("Invalid address format %s: %v\n", addrStr, err)
				continue
			}
			addrs = append(addrs, addr)
		}

		// Add operator to known operators
		n.operatorsMu.Lock()
		n.knownOperators[operatorID] = &peer.AddrInfo{
			ID: operatorID,
			Addrs: addrs,
		}
		n.operatorsMu.Unlock()

		log.Printf("Added new operator %s with addresses %v\n", operatorID, addrs)

		// Try to connect to the new operator
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := n.host.Connect(ctx, peer.AddrInfo{ID: operatorID, Addrs: addrs}); err != nil {
			log.Printf("Failed to connect to operator %s: %v\n", operatorID, err)
			return
		}

		log.Printf("Successfully connected to operator %s\n", operatorID)
	})
}

func (n *Node) healthCheck() {
	// Ping registry node
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := n.host.PingPeer(ctx, n.registryID); err != nil {
		log.Printf("Failed to ping registry: %v\n", err)
		return
	}
	log.Printf("Successfully pinged registry\n")

	// Ping all known operators
	n.operatorsMu.RLock()
	defer n.operatorsMu.RUnlock()

	for operatorID, peerInfo := range n.knownOperators {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to connect if not connected
		if err := n.host.Connect(ctx, *peerInfo); err != nil {
			log.Printf("Failed to connect to operator %s: %v\n", operatorID, err)
			continue
		}

		// Ping the operator
		if err := n.host.PingPeer(ctx, operatorID); err != nil {
			log.Printf("Failed to ping operator %s: %v\n", operatorID, err)
			continue
		}
		log.Printf("Successfully pinged operator %s\n", operatorID)
	}
}

func (n *Node) connectToRegistry() error {
	log.Printf("Attempting to connect to Registry Node at address: %s\n", n.registryAddr)

	// Create multiaddr from the provided address
	addr, err := multiaddr.NewMultiaddr(n.registryAddr)
	if err != nil {
		return fmt.Errorf("invalid registry address: %s", n.registryAddr)
	}

	// Parse peer info from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to parse peer info from address: %v", err)
	}

	// Connect to registry
	if err := n.host.Connect(context.Background(), *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to registry: %v", err)
	}

	// Save the registry ID
	n.registryID = peerInfo.ID

	log.Printf("Successfully connected to Registry Node with ID: %s\n", n.registryID)
	return nil
}

func (n *Node) Stop() error {
	return n.host.Close()
} 