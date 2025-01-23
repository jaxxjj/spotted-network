package operator

import (
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

// NumericFromBigInt converts a big.Int to pgtype.Numeric
func NumericFromBigInt(n *big.Int) pgtype.Numeric {
	if n == nil {
		return pgtype.Numeric{}
	}
	var num pgtype.Numeric
	num.Scan(n.String())
	return num
} 