package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// SignatureWithSaltAndExpiry represents a signature with salt and expiry
type SignatureWithSaltAndExpiry struct {
	Signature []byte
	Salt      [32]byte
	Expiry    *big.Int
}

// QuorumConfig represents a quorum configuration
type QuorumConfig struct {
	Strategies []Strategy
}

// Strategy represents parameters for a strategy
type Strategy struct {
	Strategy   common.Address
	Multiplier *big.Int
} 


