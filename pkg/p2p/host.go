package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

type Host struct {
	libp2pHost host.Host
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
		libp2pHost:  h,
		pingService: ps,
	}, nil
}

// Implement host.Host interface methods to delegate to libp2pHost

func (h *Host) ID() peer.ID {
	return h.libp2pHost.ID()
}

func (h *Host) Peerstore() peerstore.Peerstore {
	return h.libp2pHost.Peerstore()
}

func (h *Host) Addrs() []multiaddr.Multiaddr {
	return h.libp2pHost.Addrs()
}

func (h *Host) Network() network.Network {
	return h.libp2pHost.Network()
}

func (h *Host) Connect(ctx context.Context, pi peer.AddrInfo) error {
	return h.libp2pHost.Connect(ctx, pi)
}

func (h *Host) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	h.libp2pHost.SetStreamHandler(pid, handler)
}

func (h *Host) SetStreamHandlerMatch(pid protocol.ID, match func(protocol.ID) bool, handler network.StreamHandler) {
	h.libp2pHost.SetStreamHandlerMatch(pid, match, handler)
}

func (h *Host) RemoveStreamHandler(pid protocol.ID) {
	h.libp2pHost.RemoveStreamHandler(pid)
}

func (h *Host) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	return h.libp2pHost.NewStream(ctx, p, pids...)
}

func (h *Host) Close() error {
	return h.libp2pHost.Close()
}



// Ping a peer with timeout
func (h *Host) PingPeer(ctx context.Context, p peer.ID) error {
	// Add ping timeout
	ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
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

// GetLibp2pHost returns the underlying libp2p host
func (h *Host) GetLibp2pHost() host.Host {
	return h.libp2pHost
} 