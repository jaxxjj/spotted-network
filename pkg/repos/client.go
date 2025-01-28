package repos

import (
	"github.com/galxe/spotted-network/pkg/repos/operator/consensus_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/epoch_states"
	"github.com/galxe/spotted-network/pkg/repos/operator/task_responses"
	"github.com/galxe/spotted-network/pkg/repos/operator/tasks"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
    db *pgxpool.Pool

    Tasks              *tasks.Queries
    TaskResponses      *task_responses.Queries
    ConsensusResponses *consensus_responses.Queries
    EpochStates       *epoch_states.Queries
}

func NewClient(db *pgxpool.Pool) *Client {
    return &Client{
        db: db,
        Tasks:              tasks.New(db),
        TaskResponses:      task_responses.New(db),
        ConsensusResponses: consensus_responses.New(db),
        EpochStates:       epoch_states.New(db),
    }
}