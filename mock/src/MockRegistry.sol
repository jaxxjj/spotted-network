// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

contract MockRegistry {
    event OperatorRegistered(
        address indexed _operator,
        uint256 indexed blockNumber,
        address indexed _signingKey,
        uint256 timestamp,
        address _avs
    );

    // Mock function to emit operator registered event
    function emitOperatorRegistered(
        address operator,
        address signingKey,
        address avs
    ) external {
        emit OperatorRegistered(
            operator,
            block.number,
            signingKey,
            block.timestamp,
            avs
        );
    }

    // Mock function to check if operator is registered
    function operatorRegistered(address operator) external pure returns (bool) {
        return true; // Always return true for testing
    }

    // Mock function to get operator weight
    function getOperatorWeight(address operator) external pure returns (uint256) {
        return 100; // Return fixed weight for testing
    }

    // Mock function to get total weight
    function getTotalWeight() external pure returns (uint256) {
        return 1000; // Return fixed total weight for testing
    }

    // Mock function to get minimum weight
    function minimumWeight() external pure returns (uint256) {
        return 10; // Return fixed minimum weight for testing
    }
} 