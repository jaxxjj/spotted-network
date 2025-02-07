package task

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	TaskResponseTopic      = "/spotted/task-response"
	p2pStatusCheckInterval = 30 * time.Second
)

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
	Signer                OperatorSigner
	EpochStateQuerier     EpochStateQuerier
	ConsensusResponseRepo ConsensusResponseRepo
	BlacklistRepo         BlacklistRepo
	TaskRepo              TaskRepo
	OperatorRepo          OperatorRepo
	ChainManager          ChainManager
	PubSub                PubSubService
}

type taskResponse struct {
	taskID        string
	signature     []byte
	epoch         uint32
	chainID       uint32
	targetAddress string
	key           string
	value         string
	blockNumber   uint64
}

// TaskResponseTracker tracks task responses and their corresponding weights
type TaskResponseTrack struct {
	mu        sync.RWMutex                       // single mutex for both maps
	responses map[string]map[string]taskResponse // taskID -> operatorID -> response
	weights   map[string]map[string]*big.Int     // taskID -> operatorID -> weight
}

// TaskProcessor handles task processing and consensus
type TaskProcessor struct {
	signer            OperatorSigner
	epochStateQuerier EpochStateQuerier
	taskRepo          TaskRepo
	consensusRepo     ConsensusResponseRepo
	blacklistRepo     BlacklistRepo
	operatorRepo      OperatorRepo
	chainManager      ChainManager
	responseTopic     ResponseTopic

	taskResponseTrack TaskResponseTrack
	subscription      *pubsub.Subscription
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(cfg *TaskProcessorConfig) (*TaskProcessor, error) {
	if cfg == nil {
		return nil, fmt.Errorf("[TaskProcessor] config is nil")
	}

	if cfg.Signer == nil {
		return nil, fmt.Errorf("[TaskProcessor] signer is nil")
	}

	if cfg.EpochStateQuerier == nil {
		return nil, fmt.Errorf("[TaskProcessor] epoch state querier is nil")
	}

	if cfg.TaskRepo == nil {
		return nil, fmt.Errorf("[TaskProcessor] task repo is nil")
	}

	if cfg.OperatorRepo == nil {
		return nil, fmt.Errorf("[TaskProcessor] operator repo is nil")
	}

	if cfg.ConsensusResponseRepo == nil {
		return nil, fmt.Errorf("[TaskProcessor] consensus response repo is nil")
	}

	if cfg.ChainManager == nil {
		return nil, fmt.Errorf("[TaskProcessor] chain manager is nil")
	}

	if cfg.PubSub == nil {
		return nil, fmt.Errorf("[TaskProcessor] pubsub is nil")
	}

	// Create response topic
	responseTopic, err := cfg.PubSub.Join(TaskResponseTopic)
	if err != nil {
		return nil, fmt.Errorf("[TaskProcessor] failed to join response topic: %w", err)
	}
	log.Printf("[TaskProcessor] Joined response topic: %s", TaskResponseTopic)

	ctx, cancel := context.WithCancel(context.Background())

	tp := &TaskProcessor{
		signer:            cfg.Signer,
		epochStateQuerier: cfg.EpochStateQuerier,
		taskRepo:          cfg.TaskRepo,
		consensusRepo:     cfg.ConsensusResponseRepo,
		operatorRepo:      cfg.OperatorRepo,
		chainManager:      cfg.ChainManager,
		responseTopic:     responseTopic,
		cancel:            cancel,
		taskResponseTrack: TaskResponseTrack{
			responses: make(map[string]map[string]taskResponse),
			weights:   make(map[string]map[string]*big.Int),
		},
	}

	// Subscribe to response topic
	sub, err := responseTopic.Subscribe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("[TaskProcessor] failed to subscribe to response topic: %w", err)
	}
	tp.subscription = sub
	log.Printf("[TaskProcessor] Subscribed to response topic")

	// Start goroutines
	tp.wg.Add(4)
	go func() {
		defer tp.wg.Done()
		tp.handleResponses(ctx, sub)
	}()
	go func() {
		defer tp.wg.Done()
		tp.checkTimeouts(ctx)
	}()
	go func() {
		defer tp.wg.Done()
		tp.checkConfirmations(ctx)
	}()
	go func() {
		defer tp.wg.Done()
		tp.periodicCleanup(ctx)
	}()

	return tp, nil
}

// Stop gracefully stops the task processor
func (tp *TaskProcessor) Stop() {
	log.Printf("[TaskProcessor] Stopping task processor...")

	// 1. Cancel context to stop all goroutines
	tp.cancel()

	// 2. Wait for all goroutines to finish
	tp.wg.Wait()

	tp.chainManager.Close()
	// 3. Cancel subscription
	if tp.subscription != nil {
		tp.subscription.Cancel()
		tp.subscription = nil
	}

	// 4. Clean up topic
	if tp.responseTopic != nil {
		tp.responseTopic = nil
	}

	// 5. Clean up maps
	tp.taskResponseTrack.mu.Lock()
	tp.taskResponseTrack.responses = make(map[string]map[string]taskResponse)
	tp.taskResponseTrack.weights = make(map[string]map[string]*big.Int)
	tp.taskResponseTrack.mu.Unlock()

	log.Printf("[TaskProcessor] Task processor stopped")
}
