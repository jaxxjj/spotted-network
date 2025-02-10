package epoch

import (
	"log"

	"github.com/galxe/spotted-network/pkg/operator/constants"
)

// DetermineOperatorStatus determines operator status based on current block number and epochs
func IsOperatorActive(currentBlock uint64, activeEpoch uint32, exitEpoch uint32) bool {
	// Calculate epoch block numbers
	activeEpochStartBlock := constants.GenesisBlock + uint64(activeEpoch)*constants.EpochPeriod
	exitEpochStartBlock := constants.GenesisBlock + uint64(exitEpoch)*constants.EpochPeriod

	// Special case: if activeEpoch equals exitEpoch, operator is inactive
	if activeEpoch == exitEpoch {
		log.Printf("Operator marked as inactive (active epoch %d equals exit epoch)", activeEpoch)
		return false
	}
	if exitEpoch == 4294967295 {
		// Case 1: Only has active epoch, no exit epoch
		if currentBlock >= activeEpochStartBlock {
			log.Printf("Operator marked as active (block %d >= active epoch %d start block %d)",
				currentBlock, activeEpoch, activeEpochStartBlock)
			return true
		} else {
			log.Printf("Operator marked as inactive (block %d < active epoch %d start block %d)",
				currentBlock, activeEpoch, activeEpochStartBlock)
			return false
		}
	} else if exitEpoch > activeEpoch {
		// Case 2: Has exit epoch and exit epoch > active epoch
		if currentBlock >= activeEpochStartBlock && currentBlock < exitEpochStartBlock {
			log.Printf("Operator marked as active (block %d in [%d, %d))",
				currentBlock, activeEpochStartBlock, exitEpochStartBlock)
			return true
		} else {
			log.Printf("Operator marked as inactive (block %d outside [%d, %d))",
				currentBlock, activeEpochStartBlock, exitEpochStartBlock)
			return false
		}
	} else {
		// Case 3: Has exit epoch and exit epoch <= active epoch (deregistered then re-registered)
		if currentBlock < exitEpochStartBlock {
			log.Printf("Operator marked as active (block %d < exit epoch %d start block %d)",
				currentBlock, exitEpoch, exitEpochStartBlock)
			return true
		} else if currentBlock >= exitEpochStartBlock && currentBlock <= activeEpochStartBlock {
			// if currentBlock == exitEpochStartBlock, it should be inactive
			log.Printf("Operator marked as inactive (block %d in [%d, %d])",
				currentBlock, exitEpochStartBlock, activeEpochStartBlock)
			return false
		} else {
			log.Printf("Operator marked as active (block %d > active epoch %d start block %d)",
				currentBlock, activeEpoch, activeEpochStartBlock)
			return true
		}
	}
}
