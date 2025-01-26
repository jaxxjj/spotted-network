package operator

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/common"
)

const (
	maxRetryCount = 3
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
			log.Printf("[Timeout] Starting pending tasks check...")
			
			// Get all pending tasks
			tasks, err := tp.taskQueries.ListPendingTasks(ctx)
			if err != nil {
				log.Printf("[Timeout] Failed to list pending tasks: %v", err)
				continue
			}

			for _, task := range tasks {
				log.Printf("[Timeout] Processing pending task %s (retry count: %d)", task.TaskID, task.RetryCount)
				
				// Clean up memory maps before retrying
				tp.responsesMutex.Lock()
				delete(tp.responses, task.TaskID)
				tp.responsesMutex.Unlock()

				tp.weightsMutex.Lock()
				delete(tp.taskWeights, task.TaskID)
				tp.weightsMutex.Unlock()

				// Increment retry count before processing
				newRetryCount, err := tp.taskQueries.IncrementRetryCount(ctx, task.TaskID)
				if err != nil {
					log.Printf("[Timeout] Failed to increment retry count: %v", err)
					continue
				}
				log.Printf("[Timeout] Incremented retry count for task %s to %d", task.TaskID, newRetryCount.RetryCount)

				// If retry count reaches max, delete the task
				if newRetryCount.RetryCount >= maxRetryCount {
					log.Printf("[Timeout] Task %s reached max retries, deleting", task.TaskID)
					if err := tp.taskQueries.DeleteTaskByID(ctx, task.TaskID); err != nil {
						log.Printf("[Timeout] Failed to delete task: %v", err)
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
				if err := tp.ProcessPendingTask(ctx, &task); err != nil {
					log.Printf("[Timeout] Failed to process task %s: %v", task.TaskID, err)
				} else {
					log.Printf("[Timeout] Successfully processed pending task %s", task.TaskID)
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
			log.Printf("[Confirmation] Starting confirmation check...")
			
			// Get all tasks in confirming status
			tasks, err := tp.taskQueries.ListConfirmingTasks(ctx)
			if err != nil {
				log.Printf("[Confirmation] Failed to list confirming tasks: %v", err)
				continue
			}
			log.Printf("[Confirmation] Found %d tasks in confirming status", len(tasks))

			for _, task := range tasks {
				log.Printf("[Confirmation] Processing task %s", task.TaskID)
				
				// Get state client for the chain
				stateClient, err := tp.node.chainClient.GetStateClient(int64(task.ChainID))
				if err != nil {
					log.Printf("[Confirmation] Failed to get state client for chain %d: %v", task.ChainID, err)
					continue
				}

				// Get latest block number
				latestBlock, err := stateClient.GetLatestBlockNumber(ctx)
				if err != nil {
					log.Printf("[Confirmation] Failed to get latest block number: %v", err)
					continue
				}
				log.Printf("[Confirmation] Latest block: %d", latestBlock)

				// Get target block number
				if !task.BlockNumber.Valid || task.BlockNumber.Int == nil {
					log.Printf("[Confirmation] Invalid block number for task %s", task.TaskID)
					continue
				}
				log.Printf("[Confirmation] Raw block number from task: %v", task.BlockNumber)
				
				// Convert block number properly considering exponent
				blockNumStr := common.NumericToString(task.BlockNumber)
				targetBlock, _ := new(big.Int).SetString(blockNumStr, 10)
				if targetBlock == nil {
					log.Printf("[Confirmation] Failed to parse block number for task %s", task.TaskID)
					continue
				}
				log.Printf("[Confirmation] Target block: %d (from user request)", targetBlock.Uint64())

				// Get required confirmations
				if !task.RequiredConfirmations.Valid {
					log.Printf("[Confirmation] Invalid required confirmations for task %s", task.TaskID)
					continue
				}
				requiredConfirmations := task.RequiredConfirmations.Int32
				log.Printf("[Confirmation] Required confirmations: %d", requiredConfirmations)

				// Check if we have enough confirmations
				targetBlockUint64 := targetBlock.Uint64()
				requiredTarget := targetBlockUint64 + uint64(requiredConfirmations)
				if latestBlock >= requiredTarget {
					log.Printf("[Confirmation] Task %s has reached required confirmations (latest: %d >= target+confirmations: %d), changing status to pending", 
						task.TaskID, latestBlock, requiredTarget)
					
					// Change task status to pending
					err = tp.taskQueries.UpdateTaskToPending(ctx, task.TaskID)
					if err != nil {
						log.Printf("[Confirmation] Failed to update task status to pending: %v", err)
						continue
					}
					log.Printf("[Confirmation] Successfully changed task %s status to pending", task.TaskID)

					// Process task immediately
					if err := tp.ProcessPendingTask(ctx, &task); err != nil {
						log.Printf("[Confirmation] Failed to process task immediately: %v", err)
						continue
					}
					log.Printf("[Confirmation] Successfully processed task %s immediately", task.TaskID)
				} else {
					log.Printf("[Confirmation] Task %s needs more confirmations (latest: %d, target+confirmations: %d)", 
						task.TaskID, latestBlock, requiredTarget)
				}
			}
		}
	}
}