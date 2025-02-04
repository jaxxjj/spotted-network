package registry

import (
	"fmt"
	"testing"

	"github.com/galxe/spotted-network/pkg/common/types"
	"github.com/stretchr/testify/assert"
)

func TestDetermineOperatorStatus(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		currentBlock uint64
		activeEpoch  uint32
		exitEpoch    uint32
		wantStatus   types.OperatorStatus
		wantLogMsg   string
	}{
		// Case 1: Only active epoch (exitEpoch == 4294967295)
		{
			name:         "active_epoch_only_before_active",
			description:  "Block before active epoch start, should be inactive",
			currentBlock: GenesisBlock + EpochPeriod/2,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d < active epoch 1 start block %d)", 
				GenesisBlock+EpochPeriod/2, GenesisBlock+EpochPeriod),
		},
		{
			name:         "active_epoch_only_after_active",
			description:  "Block after active epoch start, should be active",
			currentBlock: GenesisBlock + EpochPeriod*3/2,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d >= active epoch 1 start block %d)", 
				GenesisBlock+EpochPeriod*3/2, GenesisBlock+EpochPeriod),
		},

		// Case 2: Has exit epoch and exit epoch > active epoch
		{
			name:         "exit_after_active_before_active",
			description:  "Block before active epoch, should be inactive",
			currentBlock: GenesisBlock + EpochPeriod/4,
			activeEpoch:  1,
			exitEpoch:    2,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d outside [%d, %d))", 
				GenesisBlock+EpochPeriod/4, GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod*2),
		},
		{
			name:         "exit_after_active_during_active",
			description:  "Block between active and exit epochs, should be active",
			currentBlock: GenesisBlock + EpochPeriod*3/2,
			activeEpoch:  1,
			exitEpoch:    2,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d in [%d, %d))", 
				GenesisBlock+EpochPeriod*3/2, GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod*2),
		},
		{
			name:         "exit_after_active_after_exit",
			description:  "Block after exit epoch, should be inactive",
			currentBlock: GenesisBlock + EpochPeriod*5/2,
			activeEpoch:  1,
			exitEpoch:    2,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d outside [%d, %d))", 
				GenesisBlock+EpochPeriod*5/2, GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod*2),
		},

		// Case 3: Has exit epoch and exit epoch <= active epoch
		{
			name:         "exit_before_active_before_exit",
			description:  "Block before exit epoch, should be active",
			currentBlock: GenesisBlock + EpochPeriod/4,
			activeEpoch:  2,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d < exit epoch 1 start block %d)", 
				GenesisBlock+EpochPeriod/4, GenesisBlock+EpochPeriod),
		},
		{
			name:         "exit_before_active_between_epochs",
			description:  "Block between exit and active epochs, should be inactive",
			currentBlock: GenesisBlock + EpochPeriod*3/2,
			activeEpoch:  2,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d in [%d, %d])", 
				GenesisBlock+EpochPeriod*3/2, GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod*2),
		},
		{
			name:         "exit_before_active_after_active",
			description:  "Block after active epoch, should be active",
			currentBlock: GenesisBlock + EpochPeriod*5/2,
			activeEpoch:  2,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d > active epoch 2 start block %d)", 
				GenesisBlock+EpochPeriod*5/2, GenesisBlock+EpochPeriod*2),
		},
		{
			name:         "exit_equals_active_at_epoch",
			description:  "Block at exit/active epoch boundary, should be inactive",
			currentBlock: GenesisBlock + EpochPeriod,
			activeEpoch:  1,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d in [%d, %d])", 
				GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotLogMsg := determineOperatorStatus(tt.currentBlock, tt.activeEpoch, tt.exitEpoch)
			
			assert.Equal(t, tt.wantStatus, gotStatus, "status mismatch")
			assert.Equal(t, tt.wantLogMsg, gotLogMsg, "log message mismatch")
		})
	}
}

func TestDetermineOperatorStatusEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		currentBlock uint64
		activeEpoch  uint32
		exitEpoch    uint32
		wantStatus   types.OperatorStatus
		wantLogMsg   string
	}{
		{
			name:         "genesis_block",
			description:  "Current block is genesis block",
			currentBlock: GenesisBlock,
			activeEpoch:  1,
			exitEpoch:    4294967295,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d < active epoch 1 start block %d)", 
				GenesisBlock, GenesisBlock+EpochPeriod),
		},
		{
			name:         "active_epoch_zero",
			description:  "Active epoch is 0",
			currentBlock: GenesisBlock + EpochPeriod/2,
			activeEpoch:  0,
			exitEpoch:    4294967295,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d >= active epoch 0 start block %d)", 
				GenesisBlock+EpochPeriod/2, GenesisBlock),
		},
		{
			name:         "exit_epoch_zero",
			description:  "Exit epoch is 0",
			currentBlock: GenesisBlock + EpochPeriod/2,
			activeEpoch:  1,
			exitEpoch:    0,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d in [%d, %d])", 
				GenesisBlock+EpochPeriod/2, GenesisBlock, GenesisBlock+EpochPeriod),
		},
		{
			name:         "active_epoch_max",
			description:  "Active epoch is max uint32",
			currentBlock: GenesisBlock + EpochPeriod/2,
			activeEpoch:  4294967295,
			exitEpoch:    4294967295,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d < active epoch 4294967295 start block %d)", 
				GenesisBlock+EpochPeriod/2, GenesisBlock+uint64(4294967295)*EpochPeriod),
		},
		{
			name:         "equal_epochs_at_boundary",
			description:  "Active and exit epochs are equal, block at epoch boundary",
			currentBlock: GenesisBlock + EpochPeriod,
			activeEpoch:  1,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusInactive,
			wantLogMsg:   fmt.Sprintf("Operator marked as inactive (block %d in [%d, %d])", 
				GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod, GenesisBlock+EpochPeriod),
		},
		{
			name:         "equal_epochs_before_boundary",
			description:  "Active and exit epochs are equal, block before epoch",
			currentBlock: GenesisBlock + EpochPeriod/2,
			activeEpoch:  1,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d < exit epoch 1 start block %d)", 
				GenesisBlock+EpochPeriod/2, GenesisBlock+EpochPeriod),
		},
		{
			name:         "equal_epochs_after_boundary",
			description:  "Active and exit epochs are equal, block after epoch",
			currentBlock: GenesisBlock + EpochPeriod*3/2,
			activeEpoch:  1,
			exitEpoch:    1,
			wantStatus:   types.OperatorStatusActive,
			wantLogMsg:   fmt.Sprintf("Operator marked as active (block %d > active epoch 1 start block %d)", 
				GenesisBlock+EpochPeriod*3/2, GenesisBlock+EpochPeriod),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotLogMsg := determineOperatorStatus(tt.currentBlock, tt.activeEpoch, tt.exitEpoch)
			
			assert.Equal(t, tt.wantStatus, gotStatus, "status mismatch")
			assert.Equal(t, tt.wantLogMsg, gotLogMsg, "log message mismatch")
		})
	}
} 