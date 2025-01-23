package operator

import (
	"context"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/common"
)

// checkTimeouts periodically checks for pending tasks and retries them
func (tp *TaskProcessor) checkTimeouts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tp.logger.Printf("[Timeout] Starting pending tasks check...")
			
			// Get all pending tasks
			tasks, err := tp.taskQueries.ListPendingTasks(ctx)
			if err != nil {
				tp.logger.Printf("[Timeout] Failed to list pending tasks: %v", err)
				continue
			}

			for _, task := range tasks {
				tp.logger.Printf("[Timeout] Processing pending task %s (retry count: %d)", task.TaskID, task.RetryCount)
				
				// Try to process the task
				if err := tp.ProcessPendingTask(ctx, &task); err != nil {
					tp.logger.Printf("[Timeout] Failed to process task %s: %v", task.TaskID, err)
					
					// Increment retry count
					if _, err := tp.taskQueries.IncrementRetryCount(ctx, task.TaskID); err != nil {
						tp.logger.Printf("[Timeout] Failed to increment retry count: %v", err)
						continue
					}
					
					// If retry count reaches 3, delete the task
					if task.RetryCount >= 2 { // Check for 2 since we just incremented
						tp.logger.Printf("[Timeout] Task %s reached max retries, deleting", task.TaskID)
						
						if err := tp.taskQueries.DeleteTasksByRetryCount(ctx, 3); err != nil {
							tp.logger.Printf("[Timeout] Failed to delete task: %v", err)
							continue
						}
						
						// Clean up memory
						tp.responsesMutex.Lock()
						delete(tp.responses, task.TaskID)
						tp.responsesMutex.Unlock()

						tp.weightsMutex.Lock()
						delete(tp.taskWeights, task.TaskID)
						tp.weightsMutex.Unlock()
					}
				} else {
					tp.logger.Printf("[Timeout] Successfully processed pending task %s", task.TaskID)
				}
			}
		}
	}
}


// checkConfirmations periodically checks block confirmations for tasks in confirming status
func (tp *TaskProcessor) checkConfirmations(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tp.logger.Printf("[Confirmation] Starting confirmation check...")
			
			// Get all tasks in confirming status
			tasks, err := tp.taskQueries.ListConfirmingTasks(ctx)
			if err != nil {
				tp.logger.Printf("[Confirmation] Failed to list confirming tasks: %v", err)
				continue
			}
			tp.logger.Printf("[Confirmation] Found %d tasks in confirming status", len(tasks))

			for _, task := range tasks {
				tp.logger.Printf("[Confirmation] Processing task %s", task.TaskID)
				
				// Get state client for the chain
				stateClient, err := tp.node.chainClient.GetStateClient(int64(task.ChainID))
				if err != nil {
					tp.logger.Printf("[Confirmation] Failed to get state client for chain %d: %v", task.ChainID, err)
					continue
				}

				// Get latest block number
				latestBlock, err := stateClient.GetLatestBlockNumber(ctx)
				if err != nil {
					tp.logger.Printf("[Confirmation] Failed to get latest block number: %v", err)
					continue
				}
				tp.logger.Printf("[Confirmation] Latest block: %d", latestBlock)

				// Get target block number
				if !task.BlockNumber.Valid || task.BlockNumber.Int == nil {
					tp.logger.Printf("[Confirmation] Invalid block number for task %s", task.TaskID)
					continue
				}
				tp.logger.Printf("[Confirmation] Raw block number from task: %v", task.BlockNumber)
				
				// Convert block number properly considering exponent
				blockNumStr := common.NumericToString(task.BlockNumber)
				targetBlock, _ := new(big.Int).SetString(blockNumStr, 10)
				if targetBlock == nil {
					tp.logger.Printf("[Confirmation] Failed to parse block number for task %s", task.TaskID)
					continue
				}
				tp.logger.Printf("[Confirmation] Target block: %d (from user request)", targetBlock.Uint64())

				// Get required confirmations
				if !task.RequiredConfirmations.Valid {
					tp.logger.Printf("[Confirmation] Invalid required confirmations for task %s", task.TaskID)
					continue
				}
				requiredConfirmations := task.RequiredConfirmations.Int32
				tp.logger.Printf("[Confirmation] Required confirmations: %d", requiredConfirmations)

				// Check if we have enough confirmations
				targetBlockUint64 := targetBlock.Uint64()
				requiredTarget := targetBlockUint64 + uint64(requiredConfirmations)
				if latestBlock >= requiredTarget {
					tp.logger.Printf("[Confirmation] Task %s has reached required confirmations (latest: %d >= target+confirmations: %d), changing status to pending", 
						task.TaskID, latestBlock, requiredTarget)
					
					// Change task status to pending
					err = tp.taskQueries.UpdateTaskToPending(ctx, task.TaskID)
					if err != nil {
						tp.logger.Printf("[Confirmation] Failed to update task status to pending: %v", err)
						continue
					}
					tp.logger.Printf("[Confirmation] Successfully changed task %s status to pending", task.TaskID)

					// Process task immediately
					if err := tp.ProcessPendingTask(ctx, &task); err != nil {
						tp.logger.Printf("[Confirmation] Failed to process task immediately: %v", err)
						continue
					}
					tp.logger.Printf("[Confirmation] Successfully processed task %s immediately", task.TaskID)
				} else {
					tp.logger.Printf("[Confirmation] Task %s needs more confirmations (latest: %d, target+confirmations: %d)", 
						task.TaskID, latestBlock, requiredTarget)
				}
			}
		}
	}
}