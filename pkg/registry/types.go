package registry

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// JoinRequest represents a request from an operator to join the network
type JoinRequest struct {
	OperatorAddress string
	SigningKey      string
	Message         string
	Signature       []byte
}

type OperatorInfo struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Status   string
}

// OperatorStatus represents the status of an operator
type OperatorStatus string

const (
	// OperatorStatusActive is the status when operator is active
	OperatorStatusActive OperatorStatus = "active"
	// OperatorStatusInactive is the status when operator is inactive
	OperatorStatusInactive OperatorStatus = "inactive"
)

const (
	GenesisBlock = 0
	EpochPeriod  = 12
)