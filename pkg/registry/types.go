package registry

import (
	"sync"
	"time"

	eth "github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/p2p"
	"github.com/galxe/spotted-network/pkg/repos/registry/operators"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	host *p2p.Host
	
	// Connected operators
	operators map[peer.ID]*OperatorInfo
	operatorsMu sync.RWMutex
	
	// Health check interval
	healthCheckInterval time.Duration

	// Database connection
	db *operators.Queries

	// Event listener
	eventListener *EventListener

	// Chain clients manager
	chainClients *eth.ChainClients

	// State sync subscribers
	subscribers map[peer.ID]network.Stream
	subscribersMu sync.RWMutex
}

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

// StateSyncService handles operator state synchronization
type StateSyncService struct {
	host host.Host
	db   *operators.Queries
	
	// Subscribers for state updates
	subscribers   map[peer.ID]network.Stream
	subscribersMu sync.RWMutex
}

type EventListener struct {
	chainClients *eth.ChainClients     // Chain clients manager
	operatorQueries *operators.Queries     // Database queries
}

const (
	// OperatorStatusActive is the status when operator is active
	OperatorStatusActive OperatorStatus = "active"
	// OperatorStatusInactive is the status when operator is inactive
	OperatorStatusInactive OperatorStatus = "inactive"
)

const (
	// Epoch-related constants
	GenesisBlock = 0     
	EpochPeriod  = 12    
)