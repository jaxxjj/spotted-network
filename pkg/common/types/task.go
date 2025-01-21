package types

import (
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Task represents a state query task
type Task struct {
	ID            string
	ChainID       uint64
	TargetAddress string
	Key           *big.Int
	Value         *big.Int
	BlockNumber   *big.Int
	Timestamp     *big.Int
	Epoch         uint32
	CreatedAt     time.Time
}

// NumericFromBigInt converts a big.Int to pgtype.Numeric
func NumericFromBigInt(n *big.Int) pgtype.Numeric {
	if n == nil {
		return pgtype.Numeric{}
	}
	var num pgtype.Numeric
	num.Scan(n.String())
	return num
} 