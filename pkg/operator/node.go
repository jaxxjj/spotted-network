package operator

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
)

type Node struct {
	host           host.Host
	registryID     peer.ID
	registryAddr   string
	signer         signer.Signer
	knownOperators map[peer.ID]*peer.AddrInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
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
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Create ping service
	pingService := ping.NewPingService(host)

	return &Node{
		host:           host,
		registryID:     addrInfo.ID,
		registryAddr:   registryAddr,
		signer:         s,
		knownOperators: make(map[peer.ID]*peer.AddrInfo),
		operatorsMu:    sync.RWMutex{},
		pingService:    pingService,
	}, nil
}

func (n *Node) Start() error {
	log.Printf("Operator node started with ID: %s", n.host.ID())

	log.Printf("Connecting to registry...")
	if err := n.connectToRegistry(); err != nil {
		return fmt.Errorf("failed to connect to registry: %w", err)
	}
	log.Printf("Successfully connected to registry")

	// Start message handler and health check
	n.host.SetStreamHandler("/spotted/1.0.0", n.handleMessages)
	go n.healthCheck()
	log.Printf("Message handler and health check started")

	// Announce to registry
	log.Printf("Announcing to registry...")
	if err := n.announceToRegistry(); err != nil {
		return fmt.Errorf("failed to announce to registry: %w", err)
	}
	log.Printf("Successfully announced to registry")

	return nil
}

func (n *Node) handleMessages(stream network.Stream) {
	defer stream.Close()
	
	reader := bufio.NewReader(stream)
	log.Printf("New stream opened from: %s", stream.Conn().RemotePeer())

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from stream: %v", err)
			}
			return
		}

		message = strings.TrimSpace(message)
		log.Printf("Received message: %s", message)

		if !strings.HasPrefix(message, "ANNOUNCE ") {
			log.Printf("Invalid message format, expected ANNOUNCE prefix: %s", message)
			continue
		}

		parts := strings.Split(strings.TrimPrefix(message, "ANNOUNCE "), " ")
		if len(parts) != 2 {
			log.Printf("Invalid message format: expected 2 parts, got %d", len(parts))
			continue
		}

		operatorID, err := peer.Decode(parts[0])
		if err != nil {
			log.Printf("Invalid operator ID %s: %v", parts[0], err)
			continue
		}

		addrStrs := strings.Split(parts[1], ",")
		var addrs []multiaddr.Multiaddr
		for _, addrStr := range addrStrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				log.Printf("Invalid address format %s: %v", addrStr, err)
				continue
			}
			addrs = append(addrs, addr)
		}

		if len(addrs) == 0 {
			log.Printf("No valid addresses found for operator %s", operatorID)
			continue
		}

		log.Printf("New operator announced: %s with %d addresses", operatorID, len(addrs))

		peerInfo := &peer.AddrInfo{
			ID:    operatorID,
			Addrs: addrs,
		}

		n.operatorsMu.Lock()
		n.knownOperators[operatorID] = peerInfo
		n.operatorsMu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := n.host.Connect(ctx, *peerInfo); err != nil {
			log.Printf("Failed to connect to operator %s: %v", operatorID, err)
		} else {
			log.Printf("Successfully connected to operator %s", operatorID)
		}
		cancel()
	}
}

func (n *Node) healthCheck() {
	// Ping registry node
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := <-n.pingService.Ping(ctx, n.registryID)
	if result.Error != nil {
		log.Printf("Failed to ping registry: %v\n", result.Error)
		return
	}
	log.Printf("Successfully pinged registry (RTT: %v)\n", result.RTT)

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
		result := <-n.pingService.Ping(ctx, operatorID)
		if result.Error != nil {
			log.Printf("Failed to ping operator %s: %v\n", operatorID, result.Error)
			continue
		}
		log.Printf("Successfully pinged operator %s (RTT: %v)\n", operatorID, result.RTT)
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

func (n *Node) announceToRegistry() error {
	stream, err := n.host.NewStream(context.Background(), peer.ID(n.registryID), "/spotted/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open stream to registry: %w", err)
	}
	defer stream.Close()

	// Construct announcement message
	addrStrs := make([]string, 0, len(n.host.Addrs()))
	for _, addr := range n.host.Addrs() {
		addrStrs = append(addrStrs, addr.String())
	}
	msg := fmt.Sprintf("ANNOUNCE %s %s\n", n.host.ID(), strings.Join(addrStrs, ","))
	
	// Write message to stream
	if _, err := stream.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write announcement: %w", err)
	}

	return nil
} 