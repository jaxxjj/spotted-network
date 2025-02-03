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
	UpsertEpochState(ctx context.Context, arg epoch_states.UpsertEpochStateParams) (*epoch_states.EpochState, error)
}

type ConsensusResponseQuerier interface {
	CreateConsensusResponse(ctx context.Context, arg consensus_responses.CreateConsensusResponseParams) (*consensus_responses.ConsensusResponse, error)
	GetConsensusResponseByTaskId(ctx context.Context, taskID string) (*consensus_responses.ConsensusResponse, error)
}

// checkConsensus checks if consensus has been reached for a task
func (tp *TaskProcessor) checkConsensus(taskID string) error {
	log.Printf("[Consensus] Starting consensus check for task %s", taskID)
	
	// Get responses and weights from memory maps
	tp.responsesMutex.RLock()
	responses := tp.responses[taskID]
	tp.responsesMutex.RUnlock()

	tp.weightsMutex.RLock()
	weights := tp.taskWeights[taskID]
	tp.weightsMutex.RUnlock()

	log.Printf("[Consensus] Found %d responses and %d weights for task %s", 
		len(responses), len(weights), taskID)

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
		log.Printf("[Consensus] Added operator %s weight %s, total weight now %s", 
			addr, weight.String(), totalWeight.String())
	}

	// Check if threshold is reached
	threshold, err := tp.getConsensusThreshold(context.Background())
	if err != nil {
		log.Printf("[Consensus] Failed to get consensus threshold: %v", err)
		return err
	}

	log.Printf("[Consensus] Checking threshold - total weight: %s, threshold: %s", 
		totalWeight.String(), threshold.String())

	if totalWeight.Cmp(threshold) < 0 {
		log.Printf("[Consensus] Threshold not reached for task %s - total weight: %s, threshold: %s",
			taskID, totalWeight.String(), threshold.String())
		return nil
	}

	log.Printf("[Consensus] Threshold reached for task %s! Creating consensus response", taskID)

	// Get a sample response for task details
	var sampleResp *task_responses.TaskResponses
	for _, resp := range responses {
		if resp == nil {
			continue
		}
		sampleResp = resp
		break
	}
	
	if sampleResp == nil {
		return fmt.Errorf("no valid response found for task %s", taskID)
	}

	// Convert operator signatures to JSONB
	operatorSigsJSON, err := json.Marshal(operatorSigs)
	if err != nil {
		return fmt.Errorf("failed to marshal operator signatures: %w", err)
	}

	// Aggregate signatures using signer package
	aggregatedSigs := tp.signer.AggregateSignatures(signatures)
	if aggregatedSigs == nil {
		return fmt.Errorf("failed to aggregate signatures")
	}

	log.Printf("[Consensus] Aggregated %d signatures for task %s", 
		len(signatures), taskID)

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
		log.Printf("[Consensus] Failed to store consensus for task %s: %v", taskID, err)
		return fmt.Errorf("failed to store consensus: %w", err)
	}

	log.Printf("[Consensus] Successfully stored consensus response for task %s", taskID)

	// If we have the task locally, mark it as completed
	_, err = tp.tasks.GetTaskByID(context.Background(), taskID)
	if err == nil {
		if _, err = tp.tasks.UpdateTaskStatus(context.Background(), tasks.UpdateTaskStatusParams{
			TaskID: taskID,
			Status: "completed",
		}); err != nil {
			log.Printf("[Consensus] Failed to update task %s status: %v", taskID, err)
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
	if tp.epochState == nil {
		return nil, fmt.Errorf("epoch state querier not initialized")
	}

	latestEpoch, err := tp.epochState.GetLatestEpochState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest epoch state: %w", err)
	}
	
	if latestEpoch == nil {
		return nil, fmt.Errorf("latest epoch state is nil")
	}
	
	if !latestEpoch.ThresholdWeight.Valid || latestEpoch.ThresholdWeight.Int == nil {
		return nil, fmt.Errorf("invalid threshold weight in epoch state")
	}
	
	return latestEpoch.ThresholdWeight.Int, nil
}

// storeConsensus stores a consensus response in the database
func (tp *TaskProcessor) storeConsensus(ctx context.Context, consensus consensus_responses.CreateConsensusResponseParams) error {

	response, err := tp.consensusResponse.CreateConsensusResponse(ctx, consensus)
	if err != nil {
		return fmt.Errorf("failed to create consensus response: %w", err)
	}
	if response == nil {
		return fmt.Errorf("created consensus response is nil")
	}
	return nil
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