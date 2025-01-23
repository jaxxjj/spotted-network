package common

import (
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
)

// NumericToString converts pgtype.Numeric to string with proper handling of exponent
func NumericToString(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	weight := new(big.Int).Set(n.Int)
	if n.Exp > 0 {
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)
		weight.Mul(weight, multiplier)
	} else if n.Exp < 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)
		weight.Div(weight, divisor)
	}
	return weight.String()
}
