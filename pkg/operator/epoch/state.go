package epoch

import (
	"math/big"
)

func (e *EpochUpdator) GetEpochState() EpochState {
	return e.currentEpochState
}

func (e *EpochUpdator) GetThresholdWeight() *big.Int {
	return e.currentEpochState.ThresholdWeight
}

func (e *EpochUpdator) GetCurrentEpochNumber() uint32 {
	return e.currentEpochState.EpochNumber
}
