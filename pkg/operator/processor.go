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
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
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
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Node == nil {
		return nil, fmt.Errorf("node is required")
	}
	if cfg.Signer == nil {
		return nil, fmt.Errorf("signer is required")
	}
	if cfg.Tasks == nil {
		return nil, fmt.Errorf("task querier is required")
	}
	if cfg.TaskResponse == nil {
		return nil, fmt.Errorf("task response querier is required")
	}
	if cfg.ConsensusResponse == nil {
		return nil, fmt.Errorf("consensus response querier is required")
	}
	if cfg.EpochState == nil {
		return nil, fmt.Errorf("epoch state querier is required")
	}
	if cfg.ChainManager == nil {
		return nil, fmt.Errorf("chain manager is required")
	}
	if cfg.PubSub == nil {
		return nil, fmt.Errorf("pubsub is required")
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

	// Wait for topic subscription to propagate
	log.Printf("[TaskProcessor] Waiting for topic subscription to propagate...")
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

	// Start periodic P2P status check
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			tp.checkP2PStatus()
			<-ticker.C
		}
	}()

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


