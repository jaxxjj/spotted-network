package operator

import (
	"math/big"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/common/crypto/signer"
	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pb "github.com/galxe/spotted-network/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/operator/api"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Node struct {
	host           host.Host
	registryID     peer.ID
	registryAddr   string
	signer         signer.Signer
	knownOperators map[peer.ID]*peer.AddrInfo
	operators      map[peer.ID]*OperatorInfo
	operatorsMu    sync.RWMutex
	pingService    *ping.PingService
	
	// State sync related fields
	operatorStates map[string]*pb.OperatorState
	statesMu       sync.RWMutex
	
	// Database connection
	db          *pgxpool.Pool
	taskQueries *tasks.Queries
	responseQueries *task_responses.Queries
	consensusQueries *consensus_responses.Queries
	epochStates *epoch_states.Queries
	
	// Chain clients
	chainClient *ethereum.ChainClients
	
	// API server
	apiServer *api.Server

	// P2P pubsub
	PubSub *pubsub.PubSub

	// Task processor
	taskProcessor *TaskProcessor
}

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
} 