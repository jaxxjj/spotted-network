// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package tasks

import (
	"context"
)

type Querier interface {
	CleanupOldTasks(ctx context.Context) error
	CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error)
	DeleteTaskByID(ctx context.Context, taskID string) error
	GetTaskByID(ctx context.Context, taskID string) (Task, error)
	IncrementRetryCount(ctx context.Context, taskID string) (Task, error)
	ListAllTasks(ctx context.Context) ([]Task, error)
	ListConfirmingTasks(ctx context.Context) ([]Task, error)
	ListPendingTasks(ctx context.Context) ([]Task, error)
	UpdateTaskCompleted(ctx context.Context, taskID string) error
	UpdateTaskStatus(ctx context.Context, arg UpdateTaskStatusParams) (Task, error)
	UpdateTaskToPending(ctx context.Context, taskID string) error
	UpdateTaskValue(ctx context.Context, arg UpdateTaskValueParams) (Task, error)
}

var _ Querier = (*Queries)(nil)
