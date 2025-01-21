package types

import "time"

type TaskFinalResponse struct {
    TaskID              string            `json:"task_id"`
    Epoch               uint32            `json:"epoch"`
    Status             string            `json:"status"`
    Value              string            `json:"value"`
    BlockNumber        uint64            `json:"block_number"`
    ChainID            uint64            `json:"chain_id"`
    TargetAddress      string            `json:"target_address"`
    Key                string            `json:"key"`
    OperatorSignatures map[string][]byte `json:"operator_signatures"`
    TotalWeight        string            `json:"total_weight"`
    ConsensusReachedAt time.Time         `json:"consensus_reached_at"`
} 