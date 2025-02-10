package task

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/tasks"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	TaskResponseTopic      = "/spotted/task-response"
	p2pStatusCheckInterval = 30 * time.Second
)

type TaskProcessor interface {
	Start(ctx context.Context, responseSubscription *pubsub.Subscription) error
	Stop() error
	ProcessTask(ctx context.Context, task *tasks.Tasks) error
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
type Config struct {
	Signer                OperatorSigner
	EpochStateQuerier     EpochStateQuerier
	ConsensusResponseRepo ConsensusResponseRepo
	BlacklistRepo         BlacklistRepo
	TaskRepo              TaskRepo
	OperatorRepo          OperatorRepo
	ChainManager          ChainManager
	ResponseTopic         ResponseTopic
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
type taskProcessor struct {
	signer            OperatorSigner
	epochStateQuerier EpochStateQuerier
	taskRepo          TaskRepo
	consensusRepo     ConsensusResponseRepo
	blacklistRepo     BlacklistRepo
	operatorRepo      OperatorRepo
	chainManager      ChainManager
	responseTopic     ResponseTopic

	taskResponseTrack TaskResponseTrack

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(cfg *Config) (TaskProcessor, error) {
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

	if cfg.ResponseTopic == nil {
		return nil, fmt.Errorf("[TaskProcessor] response topic is nil")
	}

	tp := &taskProcessor{
		signer:            cfg.Signer,
		epochStateQuerier: cfg.EpochStateQuerier,
		taskRepo:          cfg.TaskRepo,
		consensusRepo:     cfg.ConsensusResponseRepo,
		blacklistRepo:     cfg.BlacklistRepo,
		operatorRepo:      cfg.OperatorRepo,
		chainManager:      cfg.ChainManager,
		responseTopic:     cfg.ResponseTopic,
		taskResponseTrack: TaskResponseTrack{
			responses: make(map[string]map[string]taskResponse),
			weights:   make(map[string]map[string]*big.Int),
		},
	}
	return tp, nil
}

func (tp *taskProcessor) Start(ctx context.Context, responseSubscription *pubsub.Subscription) error {
	ctx, cancel := context.WithCancel(ctx)
	tp.cancel = cancel

	// Start goroutines
	tp.wg.Add(4)
	go func() {
		defer tp.wg.Done()
		tp.handleResponses(ctx, responseSubscription)
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

	return nil
}

// Stop gracefully stops the task processor
func (tp *taskProcessor) Stop() error {
	log.Printf("[TaskProcessor] Stopping task processor...")

	// 1. Cancel context to stop all goroutines
	tp.cancel()

	// 2. Wait for all goroutines to finish
	tp.wg.Wait()

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
	return nil
}
