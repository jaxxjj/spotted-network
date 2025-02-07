package node

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"

	"github.com/galxe/spotted-network/pkg/config"
	"github.com/galxe/spotted-network/pkg/repos/blacklist"
)

type Network interface {
	Peers() []peer.ID
	Peerstore() Peerstore
	Connect(ctx context.Context, pi peer.AddrInfo) error
	ClosePeer(peer.ID) error
	Connectedness(peer.ID) network.Connectedness
}

type Peerstore interface {
	Addrs(peer.ID) []multiaddr.Multiaddr
	RemovePeer(peer.ID)
	ClearAddrs(peer.ID)
}

// P2PHost defines the minimal interface required for p2p functionality
type P2PHost interface {
	ID() peer.ID
	Addrs() []multiaddr.Multiaddr
	Connect(ctx context.Context, pi peer.AddrInfo) error
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(peerID protocol.ID, handler network.StreamHandler)
	RemoveStreamHandler(peerID protocol.ID)
	Network() Network
	Close() error
}

// NodeConfig contains all the dependencies needed by Node
type Config struct {
	Host          P2PHost
	BlacklistRepo *blacklist.Queries
	Config        *config.Config
}

// Node represents an operator node in the network
type Node struct {
	host          P2PHost
	blacklistRepo BlacklistRepo
	config        *config.Config
}

// NewNode creates a new operator node with the given dependencies
func NewNode(ctx context.Context, cfg Config) (*Node, error) {
	if cfg.Config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.Host == nil {
		return nil, fmt.Errorf("host is required")
	}

	if cfg.BlacklistRepo == nil {
		return nil, fmt.Errorf("blacklist repo is required")
	}

	node := &Node{
		host:          cfg.Host,
		blacklistRepo: cfg.BlacklistRepo,
	}

	return node, nil
}

func (n *Node) Stop(ctx context.Context) error {
	return n.host.Close()
}
