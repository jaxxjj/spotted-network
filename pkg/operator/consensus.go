package operator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/jackc/pgx/v5/pgtype"
)

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
	threshold, err := tp.node.getConsensusThreshold(context.Background())
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
	var sampleResp *types.TaskResponse
	for _, resp := range responses {
		sampleResp = resp
		break
	}

	// Convert operator signatures to JSONB
	operatorSigsJSON, err := json.Marshal(operatorSigs)
	if err != nil {
		return fmt.Errorf("failed to marshal operator signatures: %w", err)
	}

	// Aggregate all signatures
	aggregatedSigs := tp.aggregateSignatures(signatures)
	log.Printf("[Consensus] Aggregated %d signatures for task %s", len(signatures), taskID)

	// Create consensus response
	consensus := consensus_responses.CreateConsensusResponseParams{
		TaskID:              taskID,
		Epoch:              int32(sampleResp.Epoch),
		Value:              NumericFromBigInt(sampleResp.Value),
		BlockNumber:      NumericFromBigInt(sampleResp.BlockNumber),
		ChainID:           int32(sampleResp.ChainID),
		TargetAddress:     sampleResp.TargetAddress,
		Key:             NumericFromBigInt(sampleResp.Key),
		AggregatedSignatures: aggregatedSigs,
		OperatorSignatures:  operatorSigsJSON,
		TotalWeight:        pgtype.Numeric{Int: big.NewInt(0), Exp: 0, Valid: true},
		ConsensusReachedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	// Store consensus in database
	if err = tp.storeConsensus(context.Background(), consensus); err != nil {
		return fmt.Errorf("[Consensus] failed to store consensus: %w", err)
	}
	log.Printf("[Consensus] Successfully stored consensus response for task %s", taskID)

	// If we have the task locally, mark it as completed
	_, err = tp.taskQueries.GetTaskByID(context.Background(), taskID)
	if err == nil {
		if _, err = tp.taskQueries.UpdateTaskStatus(context.Background(), struct {
			TaskID string `json:"task_id"`
			Status string `json:"status"`
		}{
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

// aggregateSignatures combines multiple ECDSA signatures by concatenation
func (tp *TaskProcessor) aggregateSignatures(sigs map[string][]byte) []byte {
	// Convert map to sorted slice to ensure deterministic ordering
	sigSlice := make([][]byte, 0, len(sigs))
	for _, sig := range sigs {
		sigSlice = append(sigSlice, sig)
	}
	
	// Concatenate all signatures
	var aggregated []byte
	for _, sig := range sigSlice {
		aggregated = append(aggregated, sig...)
	}
	return aggregated
}

// getConsensusThreshold returns the threshold weight for consensus from the latest epoch state
func (n *Node) getConsensusThreshold(ctx context.Context) (*big.Int, error) {
	latestEpoch, err := n.epochStates.GetLatestEpochState(ctx)
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
	// Create consensus response
	_, err := tp.consensusDB.CreateConsensusResponse(ctx, consensus_responses.CreateConsensusResponseParams{
		TaskID:              consensus.TaskID,
		Epoch:              consensus.Epoch,
		Value:              consensus.Value,
		BlockNumber:        consensus.BlockNumber,
		ChainID:           consensus.ChainID,
		TargetAddress:     consensus.TargetAddress,
		Key:               consensus.Key,
		AggregatedSignatures: consensus.AggregatedSignatures,
		OperatorSignatures:  consensus.OperatorSignatures,
		TotalWeight:        consensus.TotalWeight,
		ConsensusReachedAt:  consensus.ConsensusReachedAt,
	})
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