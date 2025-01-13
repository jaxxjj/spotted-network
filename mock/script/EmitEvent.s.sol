// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockRegistry.sol";

contract EmitEventScript is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address registryAddress = vm.envAddress("REGISTRY_ADDRESS");
        address avs = makeAddr("avs");
        
        vm.startBroadcast(deployerPrivateKey);

        // Get contract instance
        MockRegistry registry = MockRegistry(registryAddress);
        
        // Create 5 different operators with unique signing keys
        for (uint i = 0; i < 5; i++) {
            // Generate operator and signing key using different seeds
            address operator = makeAddr(string.concat("operator", vm.toString(i)));
            address signingKey = makeAddr(string.concat("signingKey", vm.toString(i)));

            console.log("Emitting OperatorRegistered event for operator", i + 1);
            console.log("Operator:", operator);
            console.log("Signing Key:", signingKey);
            console.log("AVS:", avs);
            console.log("-------------------");

            // Emit event
            registry.emitOperatorRegistered(
                operator,
                signingKey,
                avs
            );
        }

        vm.stopBroadcast();
        console.log("All events emitted successfully!");
    }
} 