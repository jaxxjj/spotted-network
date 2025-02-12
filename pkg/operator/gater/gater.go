package gater

import (
	"context"
	"errors"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var _ connmgr.ConnectionGater = (*connectionGater)(nil)

type Node interface {
	GetPeerAgentVersion(peerID peer.ID) (string, error)
}

type ConnectionGater interface {
	connmgr.ConnectionGater
	IncrementViolationCount(ctx context.Context, params ViolationParams) error
	UnblockNode(ctx context.Context, peerID peer.ID) error
	SetNode(node Node)
}

type Config struct {
	Version       string
	BlacklistRepo BlacklistRepo
	OperatorRepo  OperatorRepo
	Node          Node
}

type connectionGater struct {
	node          Node
	version       string
	blacklistRepo BlacklistRepo
	operatorRepo  OperatorRepo
}

func NewConnectionGater(cfg *Config) (ConnectionGater, error) {
	if cfg == nil {
		return nil, errors.New("[ConnectionGater] config is nil")
	}
	if cfg.BlacklistRepo == nil {
		return nil, errors.New("[ConnectionGater] blacklist repo is nil")
	}
	if cfg.OperatorRepo == nil {
		return nil, errors.New("[ConnectionGater] operator repo is nil")
	}
	return &connectionGater{
		node:          cfg.Node,
		version:       cfg.Version,
		blacklistRepo: cfg.BlacklistRepo,
		operatorRepo:  cfg.OperatorRepo,
	}, nil
}

// InterceptPeerDial tests whether we're allowed to Dial the specified peer
func (g *connectionGater) InterceptPeerDial(peerID peer.ID) bool {
	return g.checkPeerPermission(peerID)
}

// InterceptAddrDial tests whether we're allowed to dial the specified peer at the specified addr
func (g *connectionGater) InterceptAddrDial(peerID peer.ID, addr ma.Multiaddr) bool {
	return g.checkPeerPermission(peerID)
}

// InterceptAccept tests whether an incoming connection is allowed
func (g *connectionGater) InterceptAccept(addrs network.ConnMultiaddrs) bool {
	return validateBasicAddr(addrs.RemoteMultiaddr())
}

// InterceptSecured tests whether a secured connection is allowed
func (g *connectionGater) InterceptSecured(direction network.Direction, peerID peer.ID, addrs network.ConnMultiaddrs) bool {
	if !g.checkPeerPermission(peerID) {
		return false
	}

	return g.checkPeerPermission(peerID)
}

// InterceptUpgraded tests whether an upgraded connection is allowed
func (g *connectionGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}

func (g *connectionGater) SetNode(node Node) {
	g.node = node
}
