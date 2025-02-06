package operator

import (
	"context"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/rs/zerolog/log"
)

const (
	maxRetryCount = 3
	pendingTaskCheckInterval = 30 * time.Second
	confirmationCheckInterval = 10 * time.Second
	cleanupInterval = 1 * time.Hour
)

// checkTimeouts periodically checks for pending tasks and retries them
func (tp *TaskProcessor) checkTimeouts(ctx context.Context) {
	ticker := time.NewTicker(pendingTaskCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info().
				Str("component", "timeout").
				Msg("Starting pending tasks check...")
			
			// Get all pending tasks
			tasks, err := tp.tasks.ListPendingTasks(ctx)
			if err != nil {
				log.Error().
					Str("component", "timeout").
					Err(err).
					Msg("Failed to list pending tasks")
				continue
			}

			for _, task := range tasks {
				// Add nil check for task
				if task.TaskID == "" {
					log.Warn().
						Str("component", "timeout").
						Msg("Invalid task with empty ID")
					continue
				}

				log.Info().
					Str("component", "timeout").
					Str("task_id", task.TaskID).
					Uint16("retry_count", task.RetryCount).
					Msg("Processing pending task")
				
				// Clean up memory maps before retrying
				tp.responsesMutex.Lock()
				delete(tp.responses, task.TaskID)
				tp.responsesMutex.Unlock()

				tp.weightsMutex.Lock()
				delete(tp.taskWeights, task.TaskID)
				tp.weightsMutex.Unlock()

				// Increment retry count before processing
				newRetryCount, err := tp.tasks.IncrementRetryCount(ctx, task.TaskID)
				if err != nil {
					log.Error().
						Str("component", "timeout").
						Str("task_id", task.TaskID).
						Err(err).
						Msg("Failed to increment retry count")
					continue
				}
				log.Info().
					Str("component", "timeout").
					Str("task_id", task.TaskID).
					Uint16("new_retry_count", newRetryCount.RetryCount).
					Msg("Incremented retry count for task")

				// If retry count reaches max, delete the task
				if newRetryCount.RetryCount >= maxRetryCount {
					log.Info().
						Str("component", "timeout").
						Str("task_id", task.TaskID).
						Msg("Task reached max retries, deleting")
					if err := tp.tasks.DeleteTaskByID(ctx, task.TaskID); err != nil {
						log.Error().
							Str("component", "timeout").
							Str("task_id", task.TaskID).
							Err(err).
							Msg("Failed to delete task")
					}

					// Clean up memory
					tp.responsesMutex.Lock()
					delete(tp.responses, task.TaskID)
					tp.responsesMutex.Unlock()

					tp.weightsMutex.Lock()
					delete(tp.taskWeights, task.TaskID)
					tp.weightsMutex.Unlock()

					continue
				}
				
				// Try to process the task
				err = tp.ProcessTask(ctx, &task)
				if err != nil {
					log.Error().
						Str("component", "timeout").
						Str("task_id", task.TaskID).
						Err(err).
						Msg("Failed to process task")
				} else {
					// Check if task was already processed
					_, err := tp.taskResponse.GetTaskResponse(ctx, task_responses.GetTaskResponseParams{
						TaskID: task.TaskID,
						OperatorAddress: tp.signer.GetOperatorAddress().Hex(),
					})
					if err == nil {
						log.Info().
							Str("component", "timeout").
							Str("task_id", task.TaskID).
							Msg("Task was already processed, skipping")
					} else {
						log.Info().
							Str("component", "timeout").
							Str("task_id", task.TaskID).
							Msg("Successfully processed pending task")
					}
				}
			}
		}
	}
}

// checkConfirmations periodically checks block confirmations for tasks in confirming status
func (tp *TaskProcessor) checkConfirmations(ctx context.Context) {
	ticker := time.NewTicker(confirmationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info().
				Str("component", "confirmation").
				Msg("Starting confirmation check...")
			
			// Get all tasks in confirming status
			tasks, err := tp.tasks.ListConfirmingTasks(ctx)
			if err != nil {
				log.Error().
					Str("component", "confirmation").
					Err(err).
					Msg("Failed to list confirming tasks")
				continue
			}
			log.Info().
				Str("component", "confirmation").
				Int("task_count", len(tasks)).
				Msg("Found tasks in confirming status")

			for _, task := range tasks {
				log.Info().
					Str("component", "confirmation").
					Str("task_id", task.TaskID).
					Msg("Processing task")
				
				// Get state client for the chain
				stateClient, err := tp.chainManager.GetClientByChainId(task.ChainID)
				if err != nil {
					log.Error().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Uint32("chain_id", task.ChainID).
						Err(err).
						Msg("Failed to get state client for chain")
					continue
				}

				// Get latest block number
				latestBlock, err := stateClient.BlockNumber(ctx)
				if err != nil {
					log.Error().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Err(err).
						Msg("Failed to get latest block number")
					continue
				}
				log.Info().
					Str("component", "confirmation").
					Uint64("latest_block", latestBlock).
					Msg("Latest block")

				// Get target block number
				if task.BlockNumber == 0 {
					log.Warn().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Msg("Invalid block number for task")
					continue
				}
				log.Debug().
					Str("component", "confirmation").
					Uint64("block_number", task.BlockNumber).
					Msg("Raw block number from task")
				
				// Convert block number properly considering exponent
				targetBlock := task.BlockNumber + uint64(task.RequiredConfirmations)

				log.Info().
					Str("component", "confirmation").
					Uint64("target_block", targetBlock).
					Msg("Target block (from user request)")

				if latestBlock >= targetBlock {
					log.Info().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Uint64("latest_block", latestBlock).
						Uint64("target_block", targetBlock).
						Msg("Task has reached required confirmations, changing status to pending")
					
					// Change task status to pending
					err = tp.tasks.UpdateTaskToPending(ctx, task.TaskID)
					if err != nil {
						log.Error().
							Str("component", "confirmation").
							Str("task_id", task.TaskID).
							Err(err).
							Msg("Failed to update task status to pending")
						continue
					}
					log.Info().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Msg("Successfully changed task status to pending")

					// Process task immediately
					if err := tp.ProcessTask(ctx, &task); err != nil {
						log.Error().
							Str("component", "confirmation").
							Str("task_id", task.TaskID).
							Err(err).
							Msg("Failed to process task immediately")
						continue
					}
					log.Info().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Msg("Successfully processed task immediately")
				} else {
					log.Info().
						Str("component", "confirmation").
						Str("task_id", task.TaskID).
						Uint64("latest_block", latestBlock).
						Uint64("target_block", targetBlock).
						Msg("Task needs more confirmations")
				}
			}
		}
	}
}

// periodicCleanup runs cleanup of all task maps periodically
func (tp *TaskProcessor) periodicCleanup(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Info().
				Str("component", "task_processor").
				Msg("Starting periodic cleanup...")
			tp.cleanupAllTasks()
		}
	}
}

// cleanupAllTasks removes all task data from local maps
func (tp *TaskProcessor) cleanupAllTasks() {
	// Clean up responses map
	tp.responsesMutex.Lock()
	tp.responses = make(map[string]map[string]*task_responses.TaskResponses)
	tp.responsesMutex.Unlock()

	// Clean up weights map
	tp.weightsMutex.Lock()
	tp.taskWeights = make(map[string]map[string]*big.Int)
	tp.weightsMutex.Unlock()

	log.Info().
		Str("component", "task_processor").
		Msg("Cleaned up all local task maps")
}

