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
)

type ChainManager interface {
	// GetMainnetClient returns the mainnet client
	GetMainnetClient() (*ethereum.ChainClient, error)
	// GetClientByChainId returns the appropriate client for a given chain ID
	GetClientByChainId(chainID uint32) (*ethereum.ChainClient, error)
}


// TaskProcessor handles task processing and consensus
type TaskProcessor struct {
	node           *Node
	signer         OperatorSigner
	task    TaskQuerier
	taskResponse             TaskResponseQuerier
	consensusResponse   ConsensusResponseQuerier
	epochState EpochStateQuerier
	chainManager   ChainManager
	responseTopic  *pubsub.Topic
	responsesMutex sync.RWMutex
	responses      map[string]map[string]*task_responses.TaskResponses // taskID -> operatorAddr -> response
	weightsMutex   sync.RWMutex
	taskWeights    map[string]map[string]*big.Int // taskID -> operatorAddr -> weight
} 

// NewTaskProcessor creates a new task processor
func NewTaskProcessor(node *Node, signer OperatorSigner, task TaskQuerier, taskResponse TaskResponseQuerier, consensusResponse ConsensusResponseQuerier, epochState EpochStateQuerier, chainManager ChainManager) (*TaskProcessor, error) {
	// Create response topic
	responseTopic, err := node.PubSub.Join(TaskResponseTopic)
	if err != nil {
		return nil, fmt.Errorf("[TaskProcessor] failed to join response topic: %w", err)
	}
	log.Printf("[TaskProcessor] Joined response topic: %s", TaskResponseTopic)

	tp := &TaskProcessor{
		node:          node,
		signer:        signer,
		task:   task,
		taskResponse:   taskResponse,
		consensusResponse:   consensusResponse,
		epochState: epochState,
		chainManager: chainManager,
		responseTopic: responseTopic,
		responses:     make(map[string]map[string]*task_responses.TaskResponses),
		taskWeights:   make(map[string]map[string]*big.Int),
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


// checkP2PStatus checks the status of P2P connections
func (tp *TaskProcessor) checkP2PStatus() {
	peers := tp.node.host.Network().Peers()
	log.Printf("[P2P] Connected to %d peers:", len(peers))
	for _, peer := range peers {
		addrs := tp.node.host.Network().Peerstore().Addrs(peer)
		log.Printf("[P2P] - Peer %s at %v", peer.String(), addrs)
	}

	// Check pubsub topic
	peers = tp.responseTopic.ListPeers()
	log.Printf("[P2P] %d peers subscribed to response topic:", len(peers))
	for _, peer := range peers {
		log.Printf("[P2P] - Subscribed peer: %s", peer.String())
	}
}