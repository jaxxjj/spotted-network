package operator

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	pb "github.com/galxe/spotted-network/proto"
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
	
	log.Printf("New stream opened from: %s", stream.Conn().RemotePeer())

	// Read the entire message
	data, err := io.ReadAll(stream)
	if err != nil {
		log.Printf("Error reading from stream: %v", err)
		return
	}

	// Parse protobuf message
	var req pb.JoinRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		log.Printf("Error parsing protobuf message: %v", err)
		return
	}

	log.Printf("Received join request from operator: %s", req.Address)

	// Send success response
	resp := &pb.JoinResponse{
		Success: true,
	}
	respData, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}

	if _, err := stream.Write(respData); err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	log.Printf("Successfully processed join request from: %s", req.Address)
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

	// Create join request message
	message := []byte("join_request")
	signature, err := n.signer.Sign(message)
	if err != nil {
		return fmt.Errorf("failed to sign message: %w", err)
	}

	// Create protobuf request
	req := &pb.JoinRequest{
		Address:    n.signer.GetAddress().Hex(),
		Message:    string(message),
		Signature:  hex.EncodeToString(signature),
		SigningKey: n.signer.GetSigningKey(),
	}

	// Marshal and send request
	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if _, err := stream.Write(data); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
} 