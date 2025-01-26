// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockRegistry.sol";

contract EmitDeregisterEvent is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address registryAddress = vm.envAddress("REGISTRY_ADDRESS");
        address avs = makeAddr("avs");

        vm.startBroadcast(deployerPrivateKey);

        // Get contract instance
        MockRegistry registry = MockRegistry(registryAddress);

        // Emit events for our 3 operators with their actual addresses
        address[3] memory operators = [
            0xCf593639B34CaE0ea3217dA27014ab5FbBAc8342, // operator1
            0xFE6B5379E861C79dB03eb3a01F3F1892FC4141D5, // operator2
            0xCCE3B4EC7681B4EcF5fD5b50e562A88a33E5137B // operator3
        ];

        for (uint256 i = 0; i < operators.length; i++) {
            console.log("Emitting OperatorDeregistered event for operator", i + 1);
            console.log("Operator:", operators[i]);
            console.log("AVS:", avs);
            console.log("-------------------");

            // Emit event
            registry.emitOperatorDeregistered(operators[i], avs);
        }

        vm.stopBroadcast();
    }
}
