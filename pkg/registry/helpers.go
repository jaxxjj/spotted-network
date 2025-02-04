package registry

import (
	"fmt"

	"github.com/galxe/spotted-network/pkg/common/types"
)

// DetermineOperatorStatus determines operator status based on current block number and epochs
func determineOperatorStatus(currentBlock uint64, activeEpoch uint32, exitEpoch uint32) (types.OperatorStatus, string) {
	// Calculate epoch block numbers
	activeEpochStartBlock := GenesisBlock + uint64(activeEpoch) * EpochPeriod
	exitEpochStartBlock := GenesisBlock + uint64(exitEpoch) * EpochPeriod

	var status types.OperatorStatus
	var logMsg string

	if exitEpoch == 4294967295 {
		// Case 1: Only has active epoch, no exit epoch
		if currentBlock >= activeEpochStartBlock {
			status = types.OperatorStatusActive
			logMsg = fmt.Sprintf("Operator marked as active (block %d >= active epoch %d start block %d)", 
				currentBlock, activeEpoch, activeEpochStartBlock)
		} else {
			status = types.OperatorStatusInactive
			logMsg = fmt.Sprintf("Operator marked as inactive (block %d < active epoch %d start block %d)", 
				currentBlock, activeEpoch, activeEpochStartBlock)
		}
	} else if exitEpoch > activeEpoch {
		// Case 2: Has exit epoch and exit epoch > active epoch
		if currentBlock >= activeEpochStartBlock && currentBlock < exitEpochStartBlock {
			status = types.OperatorStatusActive
			logMsg = fmt.Sprintf("Operator marked as active (block %d in [%d, %d))", 
				currentBlock, activeEpochStartBlock, exitEpochStartBlock)
		} else {
			status = types.OperatorStatusInactive
			logMsg = fmt.Sprintf("Operator marked as inactive (block %d outside [%d, %d))", 
				currentBlock, activeEpochStartBlock, exitEpochStartBlock)
		}
	} else {
		// Case 3: Has exit epoch and exit epoch <= active epoch (deregistered then re-registered)
		if currentBlock < exitEpochStartBlock {
			status = types.OperatorStatusActive
			logMsg = fmt.Sprintf("Operator marked as active (block %d < exit epoch %d start block %d)", 
				currentBlock, exitEpoch, exitEpochStartBlock)
		} else if currentBlock >= exitEpochStartBlock && currentBlock <= activeEpochStartBlock {
			// if currentBlock == exitEpochStartBlock, it should be inactive
			status = types.OperatorStatusInactive
			logMsg = fmt.Sprintf("Operator marked as inactive (block %d in [%d, %d])", 
				currentBlock, exitEpochStartBlock, activeEpochStartBlock)
		} else {
			status = types.OperatorStatusActive
			logMsg = fmt.Sprintf("Operator marked as active (block %d > active epoch %d start block %d)", 
				currentBlock, activeEpoch, activeEpochStartBlock)
		}
	}

	return status, logMsg
}
