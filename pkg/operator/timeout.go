package operator

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
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
			tasks, err := tp.tasks.ListPendingTasks(ctx)
			if err != nil {
				log.Printf("[Timeout] Failed to list pending tasks: %v", err)
				continue
			}

			for _, task := range tasks {
				// Add nil check for task
				if task.TaskID == "" {
					log.Printf("[Timeout] Invalid task with empty ID")
					continue
				}

				log.Printf("[Timeout] Processing pending task %s (retry count: %d)", task.TaskID, task.RetryCount)
				
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
					log.Printf("[Timeout] Failed to increment retry count: %v", err)
					continue
				}
				log.Printf("[Timeout] Incremented retry count for task %s to %d", task.TaskID, newRetryCount.RetryCount)

				// If retry count reaches max, delete the task
				if newRetryCount.RetryCount >= maxRetryCount {
					log.Printf("[Timeout] Task %s reached max retries, deleting", task.TaskID)
					if err := tp.tasks.DeleteTaskByID(ctx, task.TaskID); err != nil {
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
				err = tp.ProcessTask(ctx, &task)
				if err != nil {
					log.Printf("[Timeout] Failed to process task %s: %v", task.TaskID, err)
				} else {
					// Check if task was already processed
					_, err := tp.taskResponse.GetTaskResponse(ctx, task_responses.GetTaskResponseParams{
						TaskID: task.TaskID,
						OperatorAddress: tp.signer.GetOperatorAddress().Hex(),
					})
					if err == nil {
						log.Printf("[Timeout] Task %s was already processed, skipping", task.TaskID)
					} else {
						log.Printf("[Timeout] Successfully processed pending task %s", task.TaskID)
					}
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
			tasks, err := tp.tasks.ListConfirmingTasks(ctx)
			if err != nil {
				log.Printf("[Confirmation] Failed to list confirming tasks: %v", err)
				continue
			}
			log.Printf("[Confirmation] Found %d tasks in confirming status", len(tasks))

			for _, task := range tasks {
				log.Printf("[Confirmation] Processing task %s", task.TaskID)
				
				// Get state client for the chain
				stateClient, err := tp.chainManager.GetClientByChainId(task.ChainID)
				if err != nil {
					log.Printf("[Confirmation] Failed to get state client for chain %d: %v", task.ChainID, err)
					continue
				}

				// Get latest block number
				latestBlock, err := stateClient.BlockNumber(ctx)
				if err != nil {
					log.Printf("[Confirmation] Failed to get latest block number: %v", err)
					continue
				}
				log.Printf("[Confirmation] Latest block: %d", latestBlock)

				// Get target block number
				if task.BlockNumber == 0 {
					log.Printf("[Confirmation] Invalid block number for task %s", task.TaskID)
					continue
				}
				log.Printf("[Confirmation] Raw block number from task: %v", task.BlockNumber)
				
				// Convert block number properly considering exponent
				targetBlock := task.BlockNumber + uint64(task.RequiredConfirmations)

				log.Printf("[Confirmation] Target block: %d (from user request)", targetBlock)

				if latestBlock >= targetBlock {
					log.Printf("[Confirmation] Task %s has reached required confirmations (latest: %d >= target+confirmations: %d), changing status to pending", 
						task.TaskID, latestBlock, targetBlock)
					
					// Change task status to pending
					err = tp.tasks.UpdateTaskToPending(ctx, task.TaskID)
					if err != nil {
						log.Printf("[Confirmation] Failed to update task status to pending: %v", err)
						continue
					}
					log.Printf("[Confirmation] Successfully changed task %s status to pending", task.TaskID)

					// Process task immediately
					if err := tp.ProcessTask(ctx, &task); err != nil {
						log.Printf("[Confirmation] Failed to process task immediately: %v", err)
						continue
					}
					log.Printf("[Confirmation] Successfully processed task %s immediately", task.TaskID)
				} else {
					log.Printf("[Confirmation] Task %s needs more confirmations (latest: %d, target+confirmations: %d)", 
						task.TaskID, latestBlock, targetBlock)
				}
			}
		}
	}
}

// periodicCleanup runs cleanup of all task maps periodically
func (tp *TaskProcessor) periodicCleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[TaskProcessor] Starting periodic cleanup...")
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

	log.Printf("[TaskProcessor] Cleaned up all local task maps")
}

