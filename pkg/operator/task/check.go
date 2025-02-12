package task

import (
	"context"
	"math/big"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/tasks"
	"github.com/rs/zerolog/log"
)

var (
	maxRetryCount             = 3
	pendingTaskCheckInterval  = 30 * time.Second
	confirmationCheckInterval = 10 * time.Second
	cleanupInterval           = 1 * time.Hour
)

type TaskRepo interface {
	GetTaskByID(ctx context.Context, taskID string) (*tasks.Tasks, error)
	UpdateTaskToCompleted(ctx context.Context, taskID string) error
	ListPendingTasks(ctx context.Context) ([]tasks.Tasks, error)
	IncrementRetryCount(ctx context.Context, taskID string) (*tasks.Tasks, error)
	DeleteTaskByID(ctx context.Context, taskID string) error
	ListConfirmingTasks(ctx context.Context) ([]tasks.Tasks, error)
	UpdateTaskToPending(ctx context.Context, taskID string) error
}

// checkTimeouts periodically checks for pending tasks and retries them
func (tp *taskProcessor) checkTimeouts(ctx context.Context) {
	ticker := time.NewTicker(pendingTaskCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[INFO] [timeout] Starting pending tasks check...")

			// Get all pending tasks
			tasks, err := tp.taskRepo.ListPendingTasks(ctx)
			if err != nil {
				log.Printf("[ERROR] [timeout] Failed to list pending tasks: error=%v", err)
				continue
			}

			for _, task := range tasks {
				log.Printf("[INFO] [timeout] Processing pending task: task_id=%s retry_count=%d", task.TaskID, task.RetryCount)

				newRetryCount, err := tp.taskRepo.IncrementRetryCount(ctx, task.TaskID)
				if err != nil {
					log.Printf("[ERROR] [timeout] Failed to increment retry count: task_id=%s error=%v", task.TaskID, err)
					continue
				}
				log.Printf("[INFO] [timeout] Incremented retry count after failed processing: task_id=%s new_retry_count=%d",
					task.TaskID, newRetryCount.RetryCount)

				// Check if already at max retries before processing
				if newRetryCount.RetryCount >= uint16(maxRetryCount) {
					log.Printf("[INFO] [timeout] Task reached max retries, deleting task_id=%s", task.TaskID)
					if err := tp.taskRepo.DeleteTaskByID(ctx, task.TaskID); err != nil {
						log.Printf("[ERROR] [timeout] Failed to delete task: task_id=%s error=%v", task.TaskID, err)
					}
					// Clean up memory
					tp.cleanupTask(task.TaskID)
					continue // Skip to next task
				}

				// Only process task if not deleted
				if err := tp.ProcessTask(ctx, &task); err != nil {
					log.Printf("[ERROR] [timeout] Failed to process task: task_id=%s error=%v", task.TaskID, err)
				}
			}
		}
	}
}

// checkConfirmations periodically checks block confirmations for tasks in confirming status
func (tp *taskProcessor) checkConfirmations(ctx context.Context) {
	ticker := time.NewTicker(confirmationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[INFO] [confirmation] Starting confirmation check...")

			// Get all tasks in confirming status
			tasks, err := tp.taskRepo.ListConfirmingTasks(ctx)
			if err != nil {
				log.Printf("[ERROR] [confirmation] Failed to list confirming tasks: error=%v", err)
				continue
			}
			log.Printf("[INFO] [confirmation] Found tasks in confirming status: task_count=%d", len(tasks))

			for _, task := range tasks {
				log.Printf("[INFO] [confirmation] Processing task: task_id=%s", task.TaskID)

				// Get state client for the chain
				chainClient, err := tp.chainManager.GetClientByChainId(task.ChainID)
				if err != nil {
					log.Printf("[ERROR] [confirmation] Failed to get state client for chain: task_id=%s chain_id=%d error=%v", task.TaskID, task.ChainID, err)
					continue
				}

				// Get latest block number
				latestBlock, err := chainClient.BlockNumber(ctx)
				if err != nil {
					log.Printf("[ERROR] [confirmation] Failed to get latest block number: task_id=%s error=%v", task.TaskID, err)
					continue
				}
				log.Printf("[INFO] [confirmation] Latest block: latest_block=%d", latestBlock)

				// Get target block number
				if task.BlockNumber == 0 {
					log.Printf("[WARN] [confirmation] Invalid block number for task: task_id=%s", task.TaskID)
					continue
				}
				log.Printf("[DEBUG] [confirmation] Raw block number from task: block_number=%d", task.BlockNumber)

				// Convert block number properly considering exponent
				targetBlock := task.BlockNumber + uint64(task.RequiredConfirmations)

				log.Printf("[INFO] [confirmation] Target block (from user request): target_block=%d", targetBlock)

				if latestBlock >= targetBlock {
					log.Printf("[INFO] [confirmation] Task has reached required confirmations: task_id=%s latest_block=%d target_block=%d", task.TaskID, latestBlock, targetBlock)

					// Change task status to pending
					err = tp.taskRepo.UpdateTaskToPending(ctx, task.TaskID)
					if err != nil {
						log.Printf("[ERROR] [confirmation] Failed to update task status to pending: task_id=%s error=%v", task.TaskID, err)
						continue
					}
					log.Printf("[INFO] [confirmation] Successfully changed task status to pending: task_id=%s", task.TaskID)

					// Process task immediately
					if err := tp.ProcessTask(ctx, &task); err != nil {
						log.Printf("[ERROR] [confirmation] Failed to process task immediately: task_id=%s error=%v", task.TaskID, err)
						continue
					}
					log.Printf("[INFO] [confirmation] Successfully processed task immediately: task_id=%s", task.TaskID)
				} else {
					log.Printf("[INFO] [confirmation] Task needs more confirmations: task_id=%s latest_block=%d target_block=%d", task.TaskID, latestBlock, targetBlock)
				}
			}
		}
	}
}

// periodicCleanup runs cleanup of all task maps periodically
func (tp *taskProcessor) periodicCleanup(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("[INFO] [task_processor] Starting periodic cleanup...")
			tp.cleanupAllTasks()
		}
	}
}

// cleanupAllTasks removes all task data from local maps
func (tp *taskProcessor) cleanupAllTasks() {
	// Clean up responses map
	tp.taskResponseTrack.mu.Lock()
	tp.taskResponseTrack.responses = make(map[string]map[string]taskResponse)
	tp.taskResponseTrack.weights = make(map[string]map[string]*big.Int)
	tp.taskResponseTrack.mu.Unlock()

	log.Printf("[INFO] [task_processor] Cleaned up all local task maps")
}
