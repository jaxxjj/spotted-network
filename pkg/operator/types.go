package operator

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// OperatorStatus represents the status of an operator
type OperatorStatus string

const (
	// OperatorStatusWaitingJoin indicates the operator is waiting to join
	OperatorStatusWaitingJoin OperatorStatus = "waitingJoin"
	// OperatorStatusWaitingActive indicates the operator has joined but is not yet active
	OperatorStatusWaitingActive OperatorStatus = "waitingActive"
	// OperatorStatusActive indicates the operator is active and participating in consensus
	OperatorStatusActive OperatorStatus = "active"
	// OperatorStatusInactive indicates the operator is no longer active
	OperatorStatusInactive OperatorStatus = "inactive"
)

// OperatorInfo represents information about a connected operator
type OperatorInfo struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Status   string
} 