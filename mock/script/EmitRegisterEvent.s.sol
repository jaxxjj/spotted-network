// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockRegistry.sol";

contract EmitRegisterEvent is Script {
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
        address[3] memory signingKeys = [
            0x9F0D8BAC11C5693a290527f09434b86651c66Bf2, // operator1
            0xeBBAce05Db3D717A5BA82EAB8AdE712dFb151b13, // operator2
            0x083739b681B85cc2c9e394471486321D6446b25b // operator3
        ];

        address[3] memory p2pKeys = [
            0x310C8425b620980DCFcf756e46572bb6ac80Eb07, // operator1
            0x01078ffBf1De436d6f429f5Ce6Be8Fd9D6E16165, // operator2
            0x67aa23adde2459a1620BE2Ea28982310597521b0 // operator3
        ];

        for (uint256 i = 0; i < operators.length; i++) {
            console.log("Emitting OperatorRegistered event for operator", i + 1);
            console.log("Operator:", operators[i]);
            console.log("Signing Key:", operators[i]); // Using same address as signing key
            console.log("P2P Key:", p2pKeys[i]);
            console.log("AVS:", avs);
            console.log("-------------------");

            // Emit event
            registry.emitOperatorRegistered(
                operators[i],
                signingKeys[i], // Using same address as signing key
                p2pKeys[i],
                avs
            );
        }

        vm.stopBroadcast();
    }
}
