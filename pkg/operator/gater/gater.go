package gater

import (
	"context"
	"errors"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

var _ connmgr.ConnectionGater = (*ConnectionGater)(nil)

type BlacklistRepo interface {
	IsBlocked(ctx context.Context, peerID string) (*bool, error)
}

type OperatorRepo interface {
	IsActiveOperator(ctx context.Context, p2pKey string) (*bool, error)
}

type Config struct {
	BlacklistRepo BlacklistRepo
    OperatorRepo OperatorRepo
}

type ConnectionGater struct {
    blacklistRepo BlacklistRepo
    operatorRepo OperatorRepo
}


func NewConnectionGater(cfg *Config) (*ConnectionGater, error) {
	if cfg == nil {
		return nil, errors.New("[ConnectionGater] config is nil")
	}
	if cfg.BlacklistRepo == nil {
		return nil, errors.New("[ConnectionGater] blacklist repo is nil")
	}
    if cfg.OperatorRepo == nil {
        return nil, errors.New("[ConnectionGater] operator repo is nil")
    }
    return &ConnectionGater{
        blacklistRepo: cfg.BlacklistRepo,
        operatorRepo: cfg.OperatorRepo,
    }, nil
}

// InterceptPeerDial tests whether we're allowed to Dial the specified peer
func (g *ConnectionGater) InterceptPeerDial(peerID peer.ID) bool {
    // check if the peer is blocked
    blocked, err := g.isBlocked(peerID)
    if err != nil {
        log.Error().Err(err).Str("peer", peerID.String()).Msg("failed to check if peer is blocked")
        return false
    }
    if blocked {
        log.Debug().Str("peer", peerID.String()).Msg("peer is blocked")
        return false
    }

    // check if the peer is active operator
    isActive, err := g.isActiveOperator(peerID)
    if err != nil {
        log.Error().Err(err).Str("peer", peerID.String()).Msg("failed to check if peer is active operator")
        return false
    }
    if !isActive {
        log.Debug().Str("peer", peerID.String()).Msg("peer is not an active operator")
        return false
    }

    return true
}

// InterceptAddrDial tests whether we're allowed to dial the specified peer at the specified addr
func (g *ConnectionGater) InterceptAddrDial(peerID peer.ID, addr ma.Multiaddr) bool {
    // check if the peer is blocked
    blocked, err := g.isBlocked(peerID)
    if err != nil {
        log.Error().Err(err).Str("peer", peerID.String()).Msg("failed to check if peer is blocked")
        return false
    }
    if blocked {
        log.Debug().Str("peer", peerID.String()).Msg("peer is blocked")
        return false
    }

    // check if the peer is active operator
    isActive, err := g.isActiveOperator(peerID)
    if err != nil {
        log.Error().Err(err).Str("peer", peerID.String()).Msg("failed to check if peer is active operator")
        return false
    }
    if !isActive {
        log.Debug().Str("peer", peerID.String()).Msg("peer is not an active operator")
        return false
    }

    return true
}

// InterceptAccept tests whether an incoming connection is allowed
func (g *ConnectionGater) InterceptAccept(addrs network.ConnMultiaddrs) bool {
    return true // Accept all incoming connections initially, peer ID check happens in InterceptSecured
}

// InterceptSecured tests whether a secured connection is allowed
func (g *ConnectionGater) InterceptSecured(direction network.Direction, peerID peer.ID, addrs network.ConnMultiaddrs) bool {
    // check if the peer is blocked
    blocked, err := g.isBlocked(peerID)
    if err != nil {
        log.Error().Err(err).Str("peer", peerID.String()).Msg("failed to check if peer is blocked")
        return false
    }
    if blocked {
        log.Debug().Str("peer", peerID.String()).Msg("peer is blocked")
        return false
    }

    // check if the peer is active operator
    isActive, err := g.isActiveOperator(peerID)
    if err != nil {
        log.Error().Err(err).Str("peer", peerID.String()).Msg("failed to check if peer is active operator")
        return false
    }
    if !isActive {
        log.Debug().Str("peer", peerID.String()).Msg("peer is not an active operator")
        return false
    }

    return true
}

// InterceptUpgraded tests whether an upgraded connection is allowed
func (g *ConnectionGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
    return true, 0
}