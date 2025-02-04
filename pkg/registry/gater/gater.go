package gater

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"

	"github.com/galxe/spotted-network/pkg/repos/registry/blacklist"
)

const (
	// Rate limit constants
	GlobalRateLimit = 100
	GlobalBurst = 200
	PerPeerRateLimit = 10
	PerPeerBurst = 20
	CleanupInterval = time.Hour
)

// operatorGater implements the connection gating interface for operator nodes
type operatorGater struct {
	blacklistRepo *blacklist.Queries
	
	// rate limiter
	globalLimiter *rate.Limiter
	peerLimiters  sync.Map // peer.ID -> *rate.Limiter
}

// Ensure operatorGater implements connmgr.ConnectionGater
var _ connmgr.ConnectionGater = (*operatorGater)(nil)

// newOperatorGater creates a new connection gater for the operator
func NewOperatorGater(ctx context.Context, blacklistRepo *blacklist.Queries) *operatorGater {
	g := &operatorGater{
		blacklistRepo:  blacklistRepo,
		globalLimiter: rate.NewLimiter(rate.Limit(GlobalRateLimit), GlobalBurst),
	}
	
	go g.startCleanup(ctx)
	
	return g
}

func (g *operatorGater) getPeerLimiter(p peer.ID) *rate.Limiter {
	if limiter, exists := g.peerLimiters.Load(p); exists {
		return limiter.(*rate.Limiter)
	}
	
	limiter := rate.NewLimiter(rate.Limit(PerPeerRateLimit), PerPeerBurst)
	g.peerLimiters.Store(p, limiter)
	return limiter
}

func (g *operatorGater) startCleanup(ctx context.Context) {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping gater cleanup")
			return
		case <-ticker.C:
			if err := g.blacklistRepo.CleanExpiredBlocks(ctx); err != nil {
				log.Error().Err(err).Msg("failed to clean expired blocks")
			}
		}
	}
}

// InterceptPeerDial tests whether we're allowed to Dial the specified peer
func (g *operatorGater) InterceptPeerDial(p peer.ID) bool {
	isBlocked, err := g.blacklistRepo.IsBlocked(context.Background(), blacklist.IsBlockedParams{
		PeerID: p.String(),
		Ip:     "",
	})
	if err != nil {
		log.Error().Err(err).Msg("error checking peer blacklist")
		return true
	}
	if *isBlocked {
		return false
	}
	
	if !g.globalLimiter.Allow() {
		log.Debug().Msg("global rate limit exceeded")
		return false
	}
	
	if !g.getPeerLimiter(p).Allow() {
		log.Debug().Str("peer", p.String()).Msg("peer rate limit exceeded")
		return false
	}
	
	return true
}

// InterceptAddrDial tests whether we're allowed to dial the specified peer at the specified addr
func (g *operatorGater) InterceptAddrDial(p peer.ID, addr ma.Multiaddr) bool {
	ip, err := manet.ToIP(addr)
	if err != nil {
		log.Warn().Err(err).Msg("error converting multiaddr to IP")
		return true
	}
	
	isBlocked, err := g.blacklistRepo.IsBlocked(context.Background(), blacklist.IsBlockedParams{
		PeerID: p.String(),
		Ip:     ip.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("error checking IP blacklist")
		return true
	}
	if *isBlocked {
		return false
	}
	
	// 检查限流
	return g.globalLimiter.Allow() && g.getPeerLimiter(p).Allow()
}

// InterceptAccept tests whether an incoming connection is allowed
func (g *operatorGater) InterceptAccept(connMultiaddrs network.ConnMultiaddrs) bool {
	if !g.globalLimiter.Allow() {
		log.Debug().Msg("global rate limit exceeded for incoming connection")
		return false
	}
	
	// get remote peer's address
	remoteAddr := connMultiaddrs.RemoteMultiaddr()
	ip, err := manet.ToIP(remoteAddr)
	if err != nil {
		log.Warn().Err(err).Msg("error converting multiaddr to IP")
		return true
	}
	
	// check blacklist
	isBlocked, err := g.blacklistRepo.IsBlocked(context.Background(), blacklist.IsBlockedParams{
		PeerID: "", // when accepting connections, the peer ID is not known yet
		Ip:     ip.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("error checking IP blacklist")
		return true
	}
	if *isBlocked {
		return false
	}
	
	return true
}

// InterceptSecured tests whether a secured connection is allowed
func (g *operatorGater) InterceptSecured(direction network.Direction, p peer.ID, connMultiaddrs network.ConnMultiaddrs) bool {
	if direction == network.DirOutbound {
		return true // Already checked in InterceptPeerDial/InterceptAddrDial
	}

	// For inbound, check if peer is blacklisted
	isBlocked, err := g.blacklistRepo.IsBlocked(context.Background(), blacklist.IsBlockedParams{
		PeerID: p.String(),
		Ip:     "", // empty IP means only check peer_id
	})
	if err != nil {
		log.Error().Err(err).Msg("error checking peer blacklist")
		return true
	}

	return !(*isBlocked)
}

// InterceptUpgraded tests whether an upgraded connection is allowed
func (g *operatorGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	return true, 0
}
