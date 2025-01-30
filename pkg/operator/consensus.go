package operator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	"github.com/jackc/pgx/v5/pgtype"
)

type EpochStateQuerier interface {
    GetLatestEpochState(ctx context.Context) (*epoch_states.EpochState, error)
	UpsertEpochState(ctx context.Context, arg epoch_states.UpsertEpochStateParams, listEpochStates *int32) (*epoch_states.EpochState, error)
}

type ConsensusResponseQuerier interface {
	CreateConsensusResponse(ctx context.Context, arg consensus_responses.CreateConsensusResponseParams, getConsensusResponseByRequest *consensus_responses.GetConsensusResponseByRequestParams) (*consensus_responses.ConsensusResponse, error)
	GetConsensusResponseByTaskId(ctx context.Context, taskID string) (*consensus_responses.ConsensusResponse, error)
}

// checkConsensus checks if consensus has been re.ched for a task
func (tp *TaskProcessor) checkConsensus(taskID string) error {
	log.Printf("[Consensus] Starting consensus check for task %s", taskID)
	
	// Get responses and weights from memory maps
	tp.responsesMutex.RLock()
	responses := tp.responses[taskID]
	tp.responsesMutex.RUnlock()

	tp.weightsMutex.RLock()
	weights := tp.taskWeights[taskID]
	tp.weightsMutex.RUnlock()

	log.Printf("[Consensus] Found %d responses and %d weights for task %s", len(responses), len(weights), taskID)

	if len(responses) == 0 {
		log.Printf("[Consensus] No responses found for task %s", taskID)
		return nil
	}

	// Calculate total weight and collect signatures
	totalWeight := new(big.Int)
	operatorSigs := make(map[string]map[string]interface{})
	signatures := make(map[string][]byte) // For signature aggregation
	
	for addr, resp := range responses {
		weight := weights[addr]
		if weight == nil {
			log.Printf("[Consensus] No weight found for operator %s", addr)
			continue
		}

		totalWeight.Add(totalWeight, weight)
		operatorSigs[addr] = map[string]interface{}{
			"signature": hex.EncodeToString(resp.Signature),
			"weight":   weight.String(),
		}
		// Collect signature for aggregation
		signatures[addr] = resp.Signature
		log.Printf("[Consensus] Added weight %s from operator %s, total weight now: %s", weight.String(), addr, totalWeight.String())
	}

	// Check if threshold is reached
	threshold, err := tp.getConsensusThreshold(context.Background())
	if err != nil {
		log.Printf("[Consensus] Failed to get consensus threshold: %v", err)
		return err
	}
	log.Printf("[Consensus] Checking if total weight %s reaches threshold %s", totalWeight.String(), threshold.String())
	if totalWeight.Cmp(threshold) < 0 {
		log.Printf("[Consensus] Threshold not reached for task %s (total weight: %s, threshold: %s)", taskID, totalWeight.String(), threshold.String())
		return nil
	}

	log.Printf("[Consensus] Threshold reached for task %s! Creating consensus response", taskID)

	// Get a sample response for task details
	var sampleResp *task_responses.TaskResponses
	for _, resp := range responses {
		sampleResp = resp
		break
	}

	// Convert operator signatures to JSONB
	operatorSigsJSON, err := json.Marshal(operatorSigs)
	if err != nil {
		return fmt.Errorf("failed to marshal operator signatures: %w", err)
	}

	// Aggregate signatures using signer package
	aggregatedSigs := tp.signer.AggregateSignatures(signatures)
	log.Printf("[Consensus] Aggregated %d signatures for task %s", len(signatures), taskID)

	now := time.Now()
	// Create consensus response
	consensus := consensus_responses.CreateConsensusResponseParams{
		TaskID:              taskID,
		Epoch:              sampleResp.Epoch,
		Value:              sampleResp.Value,
		BlockNumber:      sampleResp.BlockNumber,
		ChainID:           sampleResp.ChainID,
		TargetAddress:     sampleResp.TargetAddress,
		Key:             sampleResp.Key,
		AggregatedSignatures: aggregatedSigs,
		OperatorSignatures:  operatorSigsJSON,
		TotalWeight:        pgtype.Numeric{Int: totalWeight, Exp: 0, Valid: true},
		ConsensusReachedAt:  &now,
	}

	// Store consensus in database
	if err = tp.storeConsensus(context.Background(), consensus); err != nil {
		return fmt.Errorf("[Consensus] failed to store consensus: %w", err)
	}
	log.Printf("[Consensus] Successfully stored consensus response for task %s", taskID)

	// If we have the task locally, mark it as completed
	_, err = tp.tasks.GetTaskByID(context.Background(), taskID)
	if err == nil {
		if _, err = tp.tasks.UpdateTaskStatus(context.Background(), tasks.UpdateTaskStatusParams{
			TaskID: taskID,
			Status: "completed",
		}); err != nil {
			log.Printf("[Consensus] Failed to update task status: %v", err)
		} else {
			log.Printf("[Consensus] Updated task %s status to completed", taskID)
		}
	}

	// Clean up local maps
	tp.cleanupTask(taskID)
	log.Printf("[Consensus] Cleaned up local maps for task %s", taskID)

	return nil
}

// getConsensusThreshold returns the threshold weight for consensus from the latest epoch state
func (tp *TaskProcessor) getConsensusThreshold(ctx context.Context) (*big.Int, error) {
	latestEpoch, err := tp.epochState.GetLatestEpochState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest epoch state: %w", err)
	}
	
	if !latestEpoch.ThresholdWeight.Valid || latestEpoch.ThresholdWeight.Int == nil {
		return nil, fmt.Errorf("invalid threshold weight in epoch state")
	}
	
	return latestEpoch.ThresholdWeight.Int, nil
}

// storeConsensus stores a consensus response in the database
func (tp *TaskProcessor) storeConsensus(ctx context.Context, consensus consensus_responses.CreateConsensusResponseParams) error {
	// Create GetConsensusResponseByRequestParams for cache invalidation
	getConsensusResponseByRequest := &consensus_responses.GetConsensusResponseByRequestParams{
		TargetAddress: consensus.TargetAddress,
		ChainID:      consensus.ChainID,
		BlockNumber:  consensus.BlockNumber,
		Key:         consensus.Key,
	}

	_, err := tp.consensusResponse.CreateConsensusResponse(ctx, consensus, getConsensusResponseByRequest)
	return err
}

// cleanupTask removes task data from local maps
func (tp *TaskProcessor) cleanupTask(taskID string) {
	tp.responsesMutex.Lock()
	delete(tp.responses, taskID)
	tp.responsesMutex.Unlock()

	tp.weightsMutex.Lock()
	delete(tp.taskWeights, taskID)
	tp.weightsMutex.Unlock()
}