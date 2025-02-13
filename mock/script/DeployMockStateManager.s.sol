// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockStateManager.sol";
import "../src/IStateManager.sol";

contract DeployMockStateManager is Script {
    // Test addresses
    address constant ALICE = address(0x1111);
    address constant BOB = address(0x2222);

    // Test keys
    uint256 constant BALANCE_KEY = 1;
    uint256 constant STAKE_KEY = 2;
    uint256 constant REWARD_KEY = 3;

    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);

        // Deploy MockStateManager
        StateManager stateManager = new StateManager();
        console.log("MockStateManager deployed at:", address(stateManager));

        // Set test data for ALICE using batch operation
        IStateManager.SetValueParams[] memory aliceParams = new IStateManager.SetValueParams[](3);
        aliceParams[0] = IStateManager.SetValueParams({key: BALANCE_KEY, value: 200 ether});
        aliceParams[1] = IStateManager.SetValueParams({key: STAKE_KEY, value: 75 ether});
        aliceParams[2] = IStateManager.SetValueParams({key: REWARD_KEY, value: 5 ether});
        stateManager.batchSetValues(ALICE, aliceParams);

        // Set test data for BOB using batch operation
        IStateManager.SetValueParams[] memory bobParams = new IStateManager.SetValueParams[](2);
        bobParams[0] = IStateManager.SetValueParams({key: BALANCE_KEY, value: 500 ether});
        bobParams[1] = IStateManager.SetValueParams({key: STAKE_KEY, value: 200 ether});
        stateManager.batchSetValues(BOB, bobParams);


        vm.stopBroadcast();

        // Log test data for verification
        console.log("Test data set:");
        console.log("ALICE balance:", stateManager.getCurrentValue(ALICE, BALANCE_KEY));
        console.log("ALICE stake:", stateManager.getCurrentValue(ALICE, STAKE_KEY));
        console.log("ALICE reward:", stateManager.getCurrentValue(ALICE, REWARD_KEY));
        console.log("BOB balance:", stateManager.getCurrentValue(BOB, BALANCE_KEY));
        console.log("BOB stake:", stateManager.getCurrentValue(BOB, STAKE_KEY));

    }
}
