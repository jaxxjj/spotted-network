package registry

// JoinRequest represents a request from an operator to join the network
type JoinRequest struct {
	OperatorAddress string
	SigningKey      string
	Message         string
	Signature       []byte
}

// OperatorStatus represents the status of an operator
type OperatorStatus string

const (
	// OperatorStatusWaitingJoin is the status when operator has registered on chain but not joined p2p network
	OperatorStatusWaitingJoin OperatorStatus = "waitingJoin"
	// OperatorStatusWaitingActive is the status when operator has joined p2p network but not yet active
	OperatorStatusWaitingActive OperatorStatus = "waitingActive"
	// OperatorStatusActive is the status when operator is active and participating
	OperatorStatusActive OperatorStatus = "active"
	// OperatorStatusInactive is the status when operator has been deregistered
	OperatorStatusInactive OperatorStatus = "inactive"
	// OperatorStatusSuspended is the status when operator has been suspended due to poor performance
	OperatorStatusSuspended OperatorStatus = "suspended"
) 