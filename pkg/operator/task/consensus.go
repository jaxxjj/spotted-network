package task

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sort"

	utils "github.com/galxe/spotted-network/pkg/common"
	"github.com/galxe/spotted-network/pkg/repos/consensus_responses"
)

type EpochStateQuerier interface {
	GetThresholdWeight() *big.Int
	GetCurrentEpochNumber() uint32
}

type ConsensusResponseRepo interface {
	CreateConsensusResponse(ctx context.Context, arg consensus_responses.CreateConsensusResponseParams) (*consensus_responses.ConsensusResponse, error)
	GetConsensusResponseByTaskId(ctx context.Context, taskID string) (*consensus_responses.ConsensusResponse, error)
}

// checkConsensus checks if consensus has been reached for a task
func (tp *taskProcessor) checkConsensus(ctx context.Context, response taskResponse) error {
	taskID := response.taskID
	log.Printf("[Consensus] Starting consensus check for task %s", taskID)

	// Get responses and weights from memory maps
	tp.taskResponseTrack.mu.RLock()
	responses := tp.taskResponseTrack.responses[taskID]
	weights := tp.taskResponseTrack.weights[taskID]
	tp.taskResponseTrack.mu.RUnlock()

	log.Printf("[Consensus] Found %d responses and %d weights for task %s",
		len(responses), len(weights), taskID)

	if len(responses) == 0 {
		log.Printf("[Consensus] No responses found for task %s", taskID)
		return nil
	}

	// Calculate total weight first
	totalWeight := tp.calculateTotalWeight(weights)

	// Check threshold
	threshold := tp.epochStateQuerier.GetThresholdWeight()
	if totalWeight.Cmp(threshold) < 0 {
		log.Printf("[Consensus] Threshold not reached - total weight: %s, threshold: %s",
			totalWeight.String(), threshold.String())
		return nil
	}

	log.Printf("[Consensus] Threshold reached for task %s! Creating consensus response", taskID)

	// Only collect signatures and addresses if threshold is reached
	signatures, addresses, err := tp.collectSignaturesAndAddresses(context.Background(), responses)
	if err != nil {
		return fmt.Errorf("failed to collect signatures and addresses: %w", err)
	}

	log.Printf("[Consensus] Aggregated %d signatures for task %s",
		len(signatures), taskID)

	// Create consensus response
	consensus := consensus_responses.CreateConsensusResponseParams{
		TaskID:               taskID,
		Epoch:                response.epoch,
		Value:                utils.StringToNumeric(response.value),
		BlockNumber:          response.blockNumber,
		ChainID:              response.chainID,
		TargetAddress:        response.targetAddress,
		Key:                  utils.StringToNumeric(response.key),
		AggregatedSignatures: signatures,
		OperatorAddresses:    addresses,
	}

	_, err = tp.consensusRepo.CreateConsensusResponse(ctx, consensus)
	if err != nil {
		return fmt.Errorf("failed to create consensus response: %w", err)
	}

	if err := tp.markTaskAsCompleted(ctx, taskID); err != nil {
		return fmt.Errorf("failed to mark task as completed: %w", err)
	}
	log.Printf("[Consensus] Successfully stored consensus response for task %s", taskID)
	return nil
}

// cleanupTask removes task data from local maps
func (tp *taskProcessor) cleanupTask(taskID string) {
	tp.taskResponseTrack.mu.Lock()
	delete(tp.taskResponseTrack.responses, taskID)
	delete(tp.taskResponseTrack.weights, taskID)
	tp.taskResponseTrack.mu.Unlock()

}

// calculateTotalWeight sums up all weights
func (tp *taskProcessor) calculateTotalWeight(weights map[string]*big.Int) *big.Int {
	totalWeight := new(big.Int)
	for _, weight := range weights {
		totalWeight.Add(totalWeight, weight)
	}
	return totalWeight
}

func (tp *taskProcessor) markTaskAsCompleted(ctx context.Context, taskID string) error {
	// If we have the task locally, mark it as completed
	_, err := tp.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		log.Printf("[Consensus] Failed to get task %s: %v", taskID, err)
		return fmt.Errorf("failed to get task %s: %w", taskID, err)
	}
	if err := tp.taskRepo.UpdateTaskToCompleted(ctx, taskID); err != nil {
		log.Printf("[Consensus] Failed to update task %s status: %v", taskID, err)
		return fmt.Errorf("failed to update task %s status: %w", taskID, err)
	}
	log.Printf("[Consensus] Updated task %s status to completed", taskID)

	return nil
}

// collectSignaturesAndAddresses collects signatures and addresses ordered by signing keys
func (tp *taskProcessor) collectSignaturesAndAddresses(
	ctx context.Context,
	responses map[string]taskResponse,
) ([][]byte, []string, error) {
	// Get sorted signing keys
	signingKeys := make([]string, 0, len(responses))
	for key := range responses {
		signingKeys = append(signingKeys, key)
	}
	sort.Strings(signingKeys)

	// Initialize result
	signatures := make([][]byte, 0, len(responses))
	addresses := make([]string, 0, len(responses))

	// Collect signatures and addresses in sorted order
	for _, signingKey := range signingKeys {
		response := responses[signingKey]

		// Get operator address
		operator, err := tp.operatorRepo.GetOperatorBySigningKey(ctx, signingKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get operator for signing key %s: %w", signingKey, err)
		}
		if operator == nil {
			return nil, nil, fmt.Errorf("operator not found for signing key %s", signingKey)
		}

		signatures = append(signatures, response.signature)
		addresses = append(addresses, operator.Address)

		log.Printf("[Consensus] Added operator %s (address: %s)", signingKey, operator.Address)
	}

	return signatures, addresses, nil
}
