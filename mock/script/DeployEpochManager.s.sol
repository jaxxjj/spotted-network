// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockEpochManager.sol";

contract DeployEpochManagerScript is Script {
    function run() external {
        // Get deployer's private key from env
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        
        // Start broadcasting transactions
        vm.startBroadcast(deployerPrivateKey);

        // Deploy MockEpochManager
        MockEpochManager epochManager = new MockEpochManager();
        
        // Stop broadcasting transactions
        vm.stopBroadcast();

        // Log the deployed contract address
        console.log("MockEpochManager deployed at:", address(epochManager));
    }
} 