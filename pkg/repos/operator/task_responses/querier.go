// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package task_responses

import (
	"context"
)

type Querier interface {
	CreateTaskResponse(ctx context.Context, arg CreateTaskResponseParams) (TaskResponse, error)
	DeleteTaskResponse(ctx context.Context, arg DeleteTaskResponseParams) error
	GetTaskResponse(ctx context.Context, arg GetTaskResponseParams) (TaskResponse, error)
	ListOperatorResponses(ctx context.Context, arg ListOperatorResponsesParams) ([]TaskResponse, error)
	ListTaskResponses(ctx context.Context, taskID string) ([]TaskResponse, error)
}

var _ Querier = (*Queries)(nil)
