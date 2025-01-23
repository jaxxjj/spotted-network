package operator

import (
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

	// Task processor constants
	TaskResponseProtocol = "/spotted/task-response/1.0.0"
	TaskResponseTopic    = "/spotted/task-response"
)

// OperatorInfo represents information about a connected operator
type OperatorInfo struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Status   string
}

// TaskProcessor handles task processing and consensus
type TaskProcessor struct {
	node           *Node
	signer         signer.Signer
	taskQueries    *tasks.Queries
	db             *task_responses.Queries
	consensusDB    *consensus_responses.Queries
	responseTopic  *pubsub.Topic
	responsesMutex sync.RWMutex
	responses      map[string]map[string]*types.TaskResponse // taskID -> operatorAddr -> response
	weightsMutex   sync.RWMutex
	taskWeights    map[string]map[string]*big.Int // taskID -> operatorAddr -> weight
	logger         *log.Logger
} 