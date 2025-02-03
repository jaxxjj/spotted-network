package operator

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/common/contracts/ethereum"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	TaskResponseTopic    = "/spotted/task-response"
	p2pStatusCheckInterval = 30 * time.Second
)


type ChainManager interface {
	// GetMainnetClient returns the mainnet client
	GetMainnetClient() (*ethereum.ChainClient, error)
	// GetClientByChainId returns the appropriate client for a given chain ID
	GetClientByChainId(chainID uint32) (*ethereum.ChainClient, error)
}

// PubSubService defines the interface for pubsub functionality needed by TaskProcessor
type PubSubService interface {
	Join(topic string, opts ...pubsub.TopicOpt) (*pubsub.Topic, error)
}

// ResponseTopic defines the interface for response topic functionality
type ResponseTopic interface {
	// Subscribe returns a new subscription for the topic
	Subscribe(opts ...pubsub.SubOpt) (*pubsub.Subscription, error)
	// Publish publishes data to the topic
	Publish(ctx context.Context, data []byte, opts ...pubsub.PubOpt) error
	// ListPeers returns the peer IDs of peers in the topic
	ListPeers() []peer.ID
	// String returns the string representation of the topic
	String() string
}
type TasksQuerier interface {
	CleanupOldTasks(ctx context.Context) error
	DeleteTaskByID(ctx context.Context, taskID string) error 
	GetTaskByID(ctx context.Context, taskID string) (*tasks.Tasks, error)
	IncrementRetryCount(ctx context.Context, taskID string) (*tasks.Tasks, error) 
	ListConfirmingTasks(ctx context.Context) ([]tasks.Tasks, error)
	ListPendingTasks(ctx context.Context) ([]tasks.Tasks, error)
	UpdateTaskCompleted(ctx context.Context, taskID string) error
	UpdateTaskStatus(ctx context.Context, arg tasks.UpdateTaskStatusParams) (*tasks.Tasks, error)
	UpdateTaskToPending(ctx context.Context, taskID string) error
}
// TaskProcessorConfig contains all dependencies needed by TaskProcessor
type TaskProcessorConfig struct {
	Node                *Node
	Signer              OperatorSigner
	Tasks                TasksQuerier
	TaskResponse        TaskResponseQuerier
	ConsensusResponse   ConsensusResponseQuerier
	EpochState          EpochStateQuerier
	ChainManager        ChainManager
	PubSub              PubSubService
}

// TaskProcessor handles task processing and consensus
type TaskProcessor struct {
	node               *Node
	signer             OperatorSigner
	tasks               TasksQuerier
	taskResponse       TaskResponseQuerier
	consensusResponse  ConsensusResponseQuerier
	epochState         EpochStateQuerier
	chainManager       ChainManager
	responseTopic     ResponseTopic
	responsesMutex     sync.RWMutex
	responses          map[string]map[string]*task_responses.TaskResponses // taskID -> operatorAddr -> response
	weightsMutex       sync.RWMutex
	taskWeights        map[string]map[string]*big.Int // taskID -> operatorAddr -> weight
} 

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(cfg *TaskProcessorConfig) (*TaskProcessor, error) {
	if cfg == nil {
		log.Fatal("[TaskProcessor] config is nil")
	}
	if cfg.Node == nil {
		log.Fatal("[TaskProcessor] node is nil")
	}
	if cfg.Signer == nil {
		log.Fatal("[TaskProcessor] signer is nil")
	}
	if cfg.Tasks == nil {
		log.Fatal("[TaskProcessor] task querier is nil")
	}
	if cfg.TaskResponse == nil {
		log.Fatal("[TaskProcessor] task response querier is nil")
	}
	if cfg.ConsensusResponse == nil {
		log.Fatal("[TaskProcessor] consensus response querier is nil")
	}
	if cfg.EpochState == nil {
		log.Fatal("[TaskProcessor] epoch state querier is nil")
	}
	if cfg.ChainManager == nil {
		log.Fatal("[TaskProcessor] chain manager is nil")
	}
	if cfg.PubSub == nil {
		log.Fatal("[TaskProcessor] pubsub is nil")
	}

	// Create response topic
	responseTopic, err := cfg.PubSub.Join(TaskResponseTopic)
	if err != nil {
		return nil, fmt.Errorf("[TaskProcessor] failed to join response topic: %w", err)
	}
	log.Printf("[TaskProcessor] Joined response topic: %s", TaskResponseTopic)

	tp := &TaskProcessor{
		node:              cfg.Node,
		signer:            cfg.Signer,
		tasks:              cfg.Tasks,
		taskResponse:      cfg.TaskResponse,
		consensusResponse: cfg.ConsensusResponse,
		epochState:        cfg.EpochState,
		chainManager:      cfg.ChainManager,
		responseTopic:     responseTopic,
		responses:         make(map[string]map[string]*task_responses.TaskResponses),
		taskWeights:       make(map[string]map[string]*big.Int),
	}

	// Subscribe to response topic
	sub, err := responseTopic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("[TaskProcessor] failed to subscribe to response topic: %w", err)
	}
	log.Printf("[TaskProcessor] Subscribed to response topic")
	time.Sleep(5 * time.Second)
	// Check initial topic subscription status
	peers := responseTopic.ListPeers()
	log.Printf("[TaskProcessor] Initial topic subscription: %d peers", len(peers))
	for _, peer := range peers {
		log.Printf("[TaskProcessor] - Subscribed peer: %s", peer.String())
	}

	// Start goroutines
	ctx := context.Background()
	go tp.handleResponses(sub)
	go tp.checkTimeouts(ctx)
	go tp.checkConfirmations(ctx)
	go tp.periodicCleanup(ctx)  // Start periodic cleanup

	return tp, nil
}

// Stop gracefully stops the task processor
func (tp *TaskProcessor) Stop() {
	log.Printf("[TaskProcessor] Stopping task processor")
	
	// Clean up responses map
	tp.responsesMutex.Lock()
	tp.responses = make(map[string]map[string]*task_responses.TaskResponses)
	tp.responsesMutex.Unlock()
	
	log.Printf("[TaskProcessor] Task processor stopped")
}


