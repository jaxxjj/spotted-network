package p2p

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type Host struct {
	host.Host
	pingService *ping.PingService
}

type Config struct {
	ListenAddrs []string
	BootstrapPeers []string
}

func NewHost(ctx context.Context, cfg *Config) (*Host, error) {
	// Create listen addresses
	listenAddrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range cfg.ListenAddrs {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid listen addr: %s", addr)
		}
		listenAddrs = append(listenAddrs, ma)
	}

	// Create libp2p host with more options
	opts := []libp2p.Option{
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.NATPortMap(),
		// Add security transport
		libp2p.DefaultSecurity,
		// Add multiplexer
		libp2p.DefaultMuxers,
		// Add transport
		libp2p.DefaultTransports,
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// Create ping service with timeout
	ps := ping.NewPingService(h)

	// Log the host addresses
	fmt.Printf("Host created. ID: %s\n", h.ID())
	for _, addr := range h.Addrs() {
		fmt.Printf("Listening on: %s/p2p/%s\n", addr, h.ID())
	}

	return &Host{
		Host:        h,
		pingService: ps,
	}, nil
}

// Connect to a peer with timeout
func (h *Host) ConnectPeer(ctx context.Context, peerAddr string) error {
	// Add connection timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	fmt.Printf("Connecting to peer: %s\n", peerAddr)

	// Parse the multiaddr
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %s", peerAddr)
	}

	// Extract peer info including ID
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("failed to parse peer address: %v", err)
	}

	// Verify we have a peer ID
	if peerInfo.ID == "" {
		return fmt.Errorf("peer address must include peer ID: %s", peerAddr)
	}

	// Connect to the peer
	if err := h.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("failed to connect to peer: %v", err)
	}

	fmt.Printf("Successfully connected to peer: %s\n", peerAddr)
	return nil
}

// Ping a peer with timeout
func (h *Host) PingPeer(ctx context.Context, p peer.ID) error {
	// Add ping timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fmt.Printf("Pinging peer: %s\n", p)

	result := <-h.pingService.Ping(ctx, p)
	if result.Error != nil {
		return fmt.Errorf("ping failed: %v", result.Error)
	}

	fmt.Printf("Successfully pinged peer: %s (RTT: %v)\n", p, result.RTT)
	return nil
}

// Get host info for logging
func (h *Host) GetHostInfo() string {
	addrs := h.Addrs()
	id := h.ID()
	return fmt.Sprintf("Host ID: %s\nAddresses: %v", id, addrs)
}

// SendOperatorInfo sends operator information to a peer
func (h *Host) SendOperatorInfo(ctx context.Context, to peer.ID, operatorID peer.ID, addrs []multiaddr.Multiaddr) error {
	// Ensure we're connected to the peer
	if err := h.Connect(ctx, peer.AddrInfo{ID: to}); err != nil {
		return fmt.Errorf("failed to connect to peer %s: %w", to, err)
	}

	// Format addresses into strings
	addrStrs := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrs[i] = addr.String()
	}

	// Format message with operator ID and addresses
	// Format: NEW_OPERATOR:<operatorID>:<addr1>,<addr2>,...
	msg := fmt.Sprintf("NEW_OPERATOR:%s:%s", operatorID.String(), strings.Join(addrStrs, ","))
	
	return h.SendMessage(ctx, to, msg)
}

// SendMessage sends a message to a peer
func (h *Host) SendMessage(ctx context.Context, to peer.ID, msg string) error {
	s, err := h.NewStream(ctx, to, "/spotted/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to create stream: %v", err)
	}
	defer s.Close()

	if _, err := s.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
} 