// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package task_responses

import (
	"context"
)

type Querier interface {
	CreateTaskResponse(ctx context.Context, arg CreateTaskResponseParams) (TaskResponses, error)
	DeleteTaskResponse(ctx context.Context, arg DeleteTaskResponseParams) error
	GetTaskResponse(ctx context.Context, arg GetTaskResponseParams) (TaskResponses, error)
	ListOperatorResponses(ctx context.Context, arg ListOperatorResponsesParams) ([]TaskResponses, error)
	ListTaskResponses(ctx context.Context, taskID string) ([]TaskResponses, error)
}

var _ Querier = (*Queries)(nil)
