package gater

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/operators"
	"github.com/galxe/spotted-network/pkg/version"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

type OperatorRepo interface {
	IsActiveOperator(ctx context.Context, lower string) (*bool, error)
	GetOperatorByP2PKey(ctx context.Context, lower string) (*operators.Operators, error)
	Dump(ctx context.Context, beforeDump ...operators.BeforeDump) ([]byte, error)
}

// isBlocked checks if the peer is in the active operators map
func (g *connectionGater) isBlocked(peerID peer.ID) (bool, error) {
	blocked, err := g.blacklistRepo.IsBlocked(context.Background(), peerID.String())
	if err != nil {
		log.Printf("[Gater] Error checking if peer %s is blocked: %v", peerID, err)
		return false, err
	}
	if blocked == nil {
		log.Printf("[Gater] Peer %s not found in blacklist", peerID)
		return false, nil
	}
	return *blocked, nil
}

// isActiveOperator checks if the peer is an active operator
func (g *connectionGater) isActiveOperator(peerID peer.ID) (bool, error) {
	p2pKey, err := utils.PeerIDToP2PKey(peerID)
	if err != nil {
		log.Printf("[Gater] Error converting peerID to p2p key for %s: %v", peerID, err)
		return false, fmt.Errorf("failed to convert peerID to p2p key: %w", err)
	}

	// Try direct lookup
	operator, err := g.operatorRepo.GetOperatorByP2PKey(context.Background(), p2pKey)
	if err != nil {
		log.Printf("[Gater] Error getting operator by p2p key %s: %v", p2pKey, err)
		return false, err
	}

	if operator == nil {
		log.Printf("[Gater] No operator found for p2p key: %s", p2pKey)
		return false, nil
	}

	log.Printf("[Gater] Found operator: address=%s, p2p_key=%s, is_active=%v, active_epoch=%d, exit_epoch=%d",
		operator.Address,
		operator.P2pKey,
		operator.IsActive,
		operator.ActiveEpoch,
		operator.ExitEpoch,
	)

	return operator.IsActive, nil
}

// extractVersionFromAgent 从UserAgent字符串中提取version
func extractVersionFromAgent(agent string) (string, error) {
	if agent == "" {
		log.Debug().Msg("Empty agent string")
		return "", fmt.Errorf("empty agent string")
	}

	parts := strings.Split(agent, "/")
	if len(parts) != 2 || parts[0] != "spotted-network" {
		log.Debug().Str("agent", agent).Msg("Invalid agent format")
		return "", fmt.Errorf("invalid agent format: %s", agent)
	}

	version := parts[1]
	if _, err := semver.NewVersion(version); err != nil {
		log.Debug().Str("version", version).Msg("Invalid version format")
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	return version, nil
}

// checkVersionCompatibility 检查两个version是否兼容
func checkVersionCompatibility(v1, v2 string) bool {
	ver1, err1 := semver.NewVersion(v1)
	ver2, err2 := semver.NewVersion(v2)
	if err1 != nil || err2 != nil {
		log.Error().
			Err(err1).
			Str("version1", v1).
			Str("version2", v2).
			Msg("Failed to parse versions")
		return false
	}

	// 只检查major version
	if ver1.Major() != ver2.Major() {
		log.Debug().
			Str("version1", v1).
			Str("version2", v2).
			Msg("Major versions mismatch")
		return false
	}

	return true
}

// checkPeerVersion 检查peer的version是否兼容
func (g *connectionGater) checkPeerVersion(peerID peer.ID) bool {
	// 检查node是否已初始化
	if g.node == nil {
		log.Debug().
			Str("peer", peerID.String()).
			Msg("Node not initialized yet, rejecting version check")
		return false
	}

	// 获取对方的AgentVersion
	av, err := g.node.GetPeerAgentVersion(peerID)
	if err != nil {
		log.Error().
			Err(err).
			Str("peer", peerID.String()).
			Msg("Failed to get peer agent version")
		return false
	}

	// 获取我们的version
	ourAgent := fmt.Sprintf("spotted-network/%s", version.Version)

	// 提取并检查version
	theirVersion, err := extractVersionFromAgent(av)
	if err != nil {
		log.Error().
			Err(err).
			Str("peer", peerID.String()).
			Msg("Failed to extract peer version")
		return false
	}

	ourVersion, err := extractVersionFromAgent(ourAgent)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to extract our version")
		return false
	}

	return checkVersionCompatibility(ourVersion, theirVersion)
}

// checkPeerPermission checks if a peer is allowed to connect
func (g *connectionGater) checkPeerPermission(peerID peer.ID) bool {
	log.Debug().Str("peer", peerID.String()).Msg("Checking peer permission")

	// check if the peer is blocked
	blocked, err := g.isBlocked(peerID)
	if err != nil {
		log.Error().Err(err).Str("peer", peerID.String()).Msg("Error checking if peer is blocked")
		return false
	}
	if blocked {
		log.Debug().Str("peer", peerID.String()).Msg("Peer is blocked")
		return false
	}

	// check if the peer is active operator
	isActive, err := g.isActiveOperator(peerID)
	if err != nil {
		log.Error().Err(err).Str("peer", peerID.String()).Msg("Error checking if peer is active operator")
		return false
	}
	if !isActive {
		log.Debug().Str("peer", peerID.String()).Msg("Peer is not an active operator")
		return false
	}

	log.Debug().Str("peer", peerID.String()).Msg("Peer permission check passed")
	return true
}

func validateBasicAddr(addr ma.Multiaddr) bool {
	if addr == nil {
		log.Printf("[Gater] Error: multiaddr is nil")
		return false
	}

	// 检查必需的协议
	required := []int{ma.P_IP4, ma.P_TCP}
	for _, proto := range required {
		value, err := addr.ValueForProtocol(proto)
		if err != nil {
			log.Printf("[Gater] Error: missing protocol %d in addr: %s", proto, addr)
			return false
		}

		switch proto {
		case ma.P_IP4:
			if net.ParseIP(value) == nil {
				log.Printf("[Gater] Error: invalid IP address: %s", value)
				return false
			}
		case ma.P_TCP:
			port, err := strconv.Atoi(value)
			if err != nil || port < 1 || port > 65535 {
				log.Printf("[Gater] Error: invalid port number: %s", value)
				return false
			}
		}
	}

	log.Printf("[Gater] Address validation successful: %s", addr)
	return true
}
