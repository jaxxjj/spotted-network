// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

/// @title Mock Epoch Manager
/// @notice A simplified version of EpochManager for testing
contract MockEpochManager {
    /// @notice Genesis block number when contract was deployed
    uint64 public immutable GENESIS_BLOCK;

    /// @notice Length of each epoch in blocks (12 blocks for testing)
    uint64 public immutable EPOCH_LENGTH = 12;

    /// @notice Grace period duration in blocks (3 blocks for testing)
    uint64 public immutable GRACE_PERIOD = 3;

    constructor() {
        GENESIS_BLOCK = uint64(block.number);
    }

    /// @notice Gets current epoch based on block number
    function getCurrentEpoch() public view returns (uint32) {
        uint64 blocksSinceGenesis = uint64(block.number) - GENESIS_BLOCK;
        return uint32(blocksSinceGenesis / EPOCH_LENGTH);
    }

    /// @notice Gets epoch interval details
    function getEpochInterval(uint32 epoch)
        external
        view
        returns (uint64 startBlock, uint64 graceBlock, uint64 endBlock)
    {
        startBlock = GENESIS_BLOCK + (epoch * EPOCH_LENGTH);
        endBlock = startBlock + EPOCH_LENGTH;
        graceBlock = endBlock - GRACE_PERIOD;
        return (startBlock, graceBlock, endBlock);
    }

    /// @notice Checks if currently in grace period
    function isInGracePeriod() public view returns (bool) {
        uint256 currentBlock = block.number;
        uint256 epochEndBlock = getNextEpochBlock();
        return currentBlock >= epochEndBlock - GRACE_PERIOD;
    }

    /// @notice Gets remaining blocks until next epoch
    function blocksUntilNextEpoch() external view returns (uint64) {
        if (block.number >= getNextEpochBlock()) return 0;
        return getNextEpochBlock() - uint64(block.number);
    }

    /// @notice Gets the effective epoch for a given block number
    function getEffectiveEpochForBlock(uint64 blockNumber) public view returns (uint32) {
        uint64 blocksSinceGenesis = blockNumber - GENESIS_BLOCK;
        uint32 absoluteEpoch = uint32(blocksSinceGenesis / EPOCH_LENGTH);

        uint64 epochStartBlock = GENESIS_BLOCK + (uint64(absoluteEpoch) * EPOCH_LENGTH);
        uint64 epochEndBlock = epochStartBlock + EPOCH_LENGTH;

        bool isGracePeriod = blockNumber >= (epochEndBlock - GRACE_PERIOD);

        return absoluteEpoch + (isGracePeriod ? 2 : 1);
    }

    /// @notice Gets the start block of the current epoch
    function getCurrentEpochBlock() public view returns (uint64) {
        return GENESIS_BLOCK + (getCurrentEpoch() * EPOCH_LENGTH);
    }

    /// @notice Gets the next epoch block
    function getNextEpochBlock() public view returns (uint64) {
        return getCurrentEpochBlock() + EPOCH_LENGTH;
    }
}
