// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockEpochManager.sol";

contract DeployEpochManagerScript is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);

        MockEpochManager epochManager = new MockEpochManager();

        vm.stopBroadcast();

        console.log("MockEpochManager deployed at:", address(epochManager));
    }
}
