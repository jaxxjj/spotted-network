// SPDX-License-Identifier: MIT
pragma solidity ^0.8.12;

import "forge-std/Script.sol";
import "../src/MockStateManager.sol";
import "../src/IStateManager.sol";

contract DeployMockStateManager is Script {
    // Test addresses
    address constant ALICE = address(0x1111);
    address constant BOB = address(0x2222);
    address constant CHARLIE = address(0x3333);

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

        // Set test data for ALICE
        // Set balance history
        stateManager.setValue(ALICE, BALANCE_KEY, 100 ether);
        vm.roll(block.number + 1);
        stateManager.setValue(ALICE, BALANCE_KEY, 150 ether);
        vm.roll(block.number + 1);
        stateManager.setValue(ALICE, BALANCE_KEY, 200 ether);

        // Set stake history
        stateManager.setValue(ALICE, STAKE_KEY, 50 ether);
        vm.roll(block.number + 1);
        stateManager.setValue(ALICE, STAKE_KEY, 75 ether);

        // Set reward
        stateManager.setValue(ALICE, REWARD_KEY, 5 ether);

        // Set test data for BOB
        stateManager.setValue(BOB, BALANCE_KEY, 500 ether);
        vm.roll(block.number + 1);
        stateManager.setValue(BOB, STAKE_KEY, 200 ether);

        // Set test data for CHARLIE
        stateManager.setValue(CHARLIE, BALANCE_KEY, 1000 ether);
        vm.roll(block.number + 1);
        stateManager.setValue(CHARLIE, STAKE_KEY, 400 ether);
        vm.roll(block.number + 1);
        stateManager.setValue(CHARLIE, REWARD_KEY, 20 ether);

        // Test batch set values
        IStateManager.SetValueParams[] memory params = new IStateManager.SetValueParams[](2);
        params[0] = IStateManager.SetValueParams({key: BALANCE_KEY, value: 1200 ether});
        params[1] = IStateManager.SetValueParams({key: STAKE_KEY, value: 500 ether});
        stateManager.batchSetValues(CHARLIE, params);

        vm.stopBroadcast();

        // Log some test data for verification
        console.log("Test data set:");
        console.log("ALICE balance:", stateManager.getCurrentValue(ALICE, BALANCE_KEY));
        console.log("ALICE stake:", stateManager.getCurrentValue(ALICE, STAKE_KEY));
        console.log("ALICE reward:", stateManager.getCurrentValue(ALICE, REWARD_KEY));
        console.log("BOB balance:", stateManager.getCurrentValue(BOB, BALANCE_KEY));
        console.log("BOB stake:", stateManager.getCurrentValue(BOB, STAKE_KEY));
        console.log("CHARLIE balance:", stateManager.getCurrentValue(CHARLIE, BALANCE_KEY));
        console.log("CHARLIE stake:", stateManager.getCurrentValue(CHARLIE, STAKE_KEY));
        console.log("CHARLIE reward:", stateManager.getCurrentValue(CHARLIE, REWARD_KEY));
    }
}
