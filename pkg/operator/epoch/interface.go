package epoch

import (
	"math/big"
)

// EpochStateQuerier defines the interface for querying epoch state
type EpochStateQuerier interface {
	// GetEpochState returns the current epoch state
	GetEpochState() EpochState
	// GetThresholdWeight returns the current threshold weight
	GetThresholdWeight() *big.Int
	// GetCurrentEpochNumber returns the current epoch number
	GetCurrentEpochNumber() uint32
}

func (e *epochUpdator) GetEpochState() EpochState {
	return e.currentEpochState
}

func (e *epochUpdator) GetThresholdWeight() *big.Int {
	return e.currentEpochState.ThresholdWeight
}

func (e *epochUpdator) GetCurrentEpochNumber() uint32 {
	return e.currentEpochState.EpochNumber
}
