package operator

import (
	"context"
	"time"

	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	// Task processor constants
	TaskResponseProtocol = "/spotted/task-response/1.0.0"
	TaskResponseTopic    = "/spotted/task-response"
)

// OperatorInfo represents information about a connected operator
type OperatorInfo struct {
	ID       peer.ID
	Addrs    []multiaddr.Multiaddr
	LastSeen time.Time
	Status   string
}

// TaskQuerier defines the interface for task database operations
type TaskQuerier interface {
	CleanupOldTasks(ctx context.Context) error
	CreateTask(ctx context.Context, arg tasks.CreateTaskParams) (tasks.Tasks, error)
	DeleteTaskByID(ctx context.Context, taskID string) error
	GetTaskByID(ctx context.Context, taskID string) (tasks.Tasks, error)
	IncrementRetryCount(ctx context.Context, taskID string) (tasks.Tasks, error)
	ListAllTasks(ctx context.Context) ([]tasks.Tasks, error)
	ListConfirmingTasks(ctx context.Context) ([]tasks.Tasks, error)
	ListPendingTasks(ctx context.Context) ([]tasks.Tasks, error)
	UpdateTaskCompleted(ctx context.Context, taskID string) error
	UpdateTaskStatus(ctx context.Context, arg tasks.UpdateTaskStatusParams) (tasks.Tasks, error)
	UpdateTaskToPending(ctx context.Context, taskID string) error
	UpdateTaskValue(ctx context.Context, arg tasks.UpdateTaskValueParams) (tasks.Tasks, error)
}



