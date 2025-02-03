package operator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/rs/zerolog/log"

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
	log.Debug().
		Str("component", "consensus").
		Str("task_id", taskID).
		Msg("Starting consensus check")
	
	// Get responses and weights from memory maps
	tp.responsesMutex.RLock()
	responses := tp.responses[taskID]
	tp.responsesMutex.RUnlock()

	tp.weightsMutex.RLock()
	weights := tp.taskWeights[taskID]
	tp.weightsMutex.RUnlock()

	log.Debug().
		Str("component", "consensus").
		Str("task_id", taskID).
		Int("responses_count", len(responses)).
		Int("weights_count", len(weights)).
		Msg("Found responses and weights")

	if len(responses) == 0 {
		log.Debug().
			Str("component", "consensus").
			Str("task_id", taskID).
			Msg("No responses found")
		return nil
	}

	// Calculate total weight and collect signatures
	totalWeight := new(big.Int)
	operatorSigs := make(map[string]map[string]interface{})
	signatures := make(map[string][]byte) // For signature aggregation
	
	for addr, resp := range responses {
		weight := weights[addr]
		if weight == nil {
			log.Debug().
				Str("component", "consensus").
				Str("operator", addr).
				Msg("No weight found for operator")
			continue
		}

		totalWeight.Add(totalWeight, weight)
		operatorSigs[addr] = map[string]interface{}{
			"signature": hex.EncodeToString(resp.Signature),
			"weight":   weight.String(),
		}
		// Collect signature for aggregation
		signatures[addr] = resp.Signature
		log.Debug().
			Str("component", "consensus").
			Str("operator", addr).
			Str("weight", weight.String()).
			Str("total_weight", totalWeight.String()).
			Msg("Added operator weight")
	}

	// Check if threshold is reached
	threshold, err := tp.getConsensusThreshold(context.Background())
	if err != nil {
		log.Error().
			Str("component", "consensus").
			Err(err).
			Msg("Failed to get consensus threshold")
		return err
	}

	log.Debug().
		Str("component", "consensus").
		Str("total_weight", totalWeight.String()).
		Str("threshold", threshold.String()).
		Msg("Checking threshold")

	if totalWeight.Cmp(threshold) < 0 {
		log.Debug().
			Str("component", "consensus").
			Str("task_id", taskID).
			Str("total_weight", totalWeight.String()).
			Str("threshold", threshold.String()).
			Msg("Threshold not reached")
		return nil
	}

	log.Info().
		Str("component", "consensus").
		Str("task_id", taskID).
		Msg("Threshold reached! Creating consensus response")

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

	log.Debug().
		Str("component", "consensus").
		Str("task_id", taskID).
		Int("signatures_count", len(signatures)).
		Msg("Aggregated signatures")

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
		log.Error().
			Str("component", "consensus").
			Str("task_id", taskID).
			Err(err).
			Msg("Failed to store consensus")
		return fmt.Errorf("failed to store consensus: %w", err)
	}

	log.Info().
		Str("component", "consensus").
		Str("task_id", taskID).
		Msg("Successfully stored consensus response")

	// If we have the task locally, mark it as completed
	_, err = tp.tasks.GetTaskByID(context.Background(), taskID)
	if err == nil {
		if _, err = tp.tasks.UpdateTaskStatus(context.Background(), tasks.UpdateTaskStatusParams{
			TaskID: taskID,
			Status: "completed",
		}); err != nil {
			log.Error().
				Str("component", "consensus").
				Str("task_id", taskID).
				Err(err).
				Msg("Failed to update task status")
		} else {
			log.Info().
				Str("component", "consensus").
				Str("task_id", taskID).
				Msg("Updated task status to completed")
		}
	}

	// Clean up local maps
	tp.cleanupTask(taskID)
	log.Debug().
		Str("component", "consensus").
		Str("task_id", taskID).
		Msg("Cleaned up local maps")

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