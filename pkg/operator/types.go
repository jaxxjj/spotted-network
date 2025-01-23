package operator

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// OperatorStatus represents the status of an operator
type OperatorStatus string

const (
	// OperatorStatusActive indicates the operator is active
	OperatorStatusActive OperatorStatus = "active"
	// OperatorStatusInactive indicates the operator is inactive
	OperatorStatusInactive OperatorStatus = "inactive"
)

// OperatorInfo represents information about a connected operator
type OperatorInfo struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Status   string
} 