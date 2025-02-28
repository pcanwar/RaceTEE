// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./SystemContract.sol";

contract Calculate is SystemContract {
    function cal(uint256 n) external returns (int256) {
        int256 sum = 0;
        for (uint256 i = 1; i <= n; i++) {
            if (i % 2 == 1) {
                sum += int256(i);
            } else {
                sum -= int256(i);
            }
        }
        return sum;
    }

    function getStates() external pure override returns (bytes memory) {
        return bytes("");
    }

    function setStates(bytes memory _states) external pure override {
    }
}
