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
    address public constant OPERATOR_1 = address(0xCf593639B34CaE0ea3217dA27014ab5FbBAc8342);
    address public constant OPERATOR_2 = address(0xCCE3B4EC7681B4EcF5fD5b50e562A88a33E5137B);
    address public constant OPERATOR_3 = address(0xFE6B5379E861C79dB03eb3a01F3F1892FC4141D5);
    address public constant SIGNING_KEY_1 = address(0x9F0D8BAC11C5693a290527f09434b86651c66Bf2);
    address public constant SIGNING_KEY_2 = address(0xeBBAce05Db3D717A5BA82EAB8AdE712dFb151b13);
    address public constant SIGNING_KEY_3 = address(0x083739b681B85cc2c9e394471486321D6446b25b);

    event OperatorDeregistered(address indexed _operator, uint256 indexed blockNumber, address indexed _avs);

    // Mock function to emit operator registered event
    function emitOperatorRegistered(address operator, address signingKey, address avs) external {
        emit OperatorRegistered(operator, block.number, signingKey, block.timestamp, avs);
    }

    function emitOperatorDeregistered(address operator, address avs) external {
        emit OperatorDeregistered(operator, block.number, avs);
    }

    // Mock function to check if operator is registered
    function operatorRegistered(address operator) external pure returns (bool) {
        return true; // Always return true for testing
    }

    // Mock function to get operator weight
    function getOperatorWeight(address operator) external pure returns (uint256) {
        if (operator == OPERATOR_1) {
            return 333; // Return fixed weight for testing
        }
        if (operator == OPERATOR_2) {
            return 444;
        }
        if (operator == OPERATOR_3) {
            return 555;
        }
        return 0;
    }

    function getOperatorWeightAtEpoch(address _operator, uint32 _epochNumber) external view returns (uint256) {
        if (_epochNumber == 1) {
            if (_operator == OPERATOR_1) {
                return 333; // Return fixed weight for testing
            }
            if (_operator == OPERATOR_2) {
                return 444;
            }
            if (_operator == OPERATOR_3) {
                return 555;
            }
        }

        if (_epochNumber == 2) {
            if (_operator == OPERATOR_1) {
                return 345; // Return fixed weight for testing
            }
            if (_operator == OPERATOR_2) {
                return 456;
            }
            if (_operator == OPERATOR_3) {
                return 567;
            }
        }

        if (_epochNumber == 3) {
            if (_operator == OPERATOR_1) {
                return 300; // Return fixed weight for testing
            }
            if (_operator == OPERATOR_2) {
                return 400;
            }
            if (_operator == OPERATOR_3) {
                return 500;
            }
        }

        return 100;
    }

    // Mock function to get total weight
    function getTotalWeightAtEpoch(uint32 _epochNumber) external pure returns (uint256) {
        return 1332;
    }

    function getThresholdWeightAtEpoch(uint32 _epochNumber) external view returns (uint256) {
        if (_epochNumber == 0) {
            return 0;
        }
        if (_epochNumber == 1) {
            return 666;
        }

        if (_epochNumber == 2) {
            return 999;
        }
        if (_epochNumber == 3) {
            return 1234;
        }
        return 0;
    }
    // Mock function to get minimum weight

    function minimumWeight() external pure returns (uint256) {
        return 10; // Return fixed minimum weight for testing
    }

    function getOperatorSigningKeyAtEpoch(
        address _operator,
        uint32 _referenceEpoch
    ) public view returns (address) {
        if(_operator == OPERATOR_1) {
            return SIGNING_KEY_1;
        }
        if(_operator == OPERATOR_2) {
            return SIGNING_KEY_2;
        }
        return SIGNING_KEY_3;
    }
}
