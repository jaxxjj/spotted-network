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

type ConnectionGater interface {
	connmgr.ConnectionGater
	IncrementViolationCount(ctx context.Context, params ViolationParams) error
	UnblockNode(ctx context.Context, peerID peer.ID) error
}

type Config struct {
	Version       string
	BlacklistRepo BlacklistRepo
	OperatorRepo  OperatorRepo
}

type connectionGater struct {
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
	return g.checkPeerPermission(peerID)
}

// InterceptUpgraded tests whether an upgraded connection is allowed
func (g *connectionGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
