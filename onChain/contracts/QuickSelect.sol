// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./SystemContract.sol";

contract QuickSelect is SystemContract {
    function quickSelect(int256[] memory arr, uint256 k) public returns (int256) {
        require(k > 0 && k <= arr.length, "Invalid k");
        uint256 index = arr.length - k; 
        return _quickSelect(arr, 0, int256(arr.length - 1), int256(index));
    }

    function _quickSelect(int256[] memory arr, int256 left, int256 right, int256 k) private pure returns (int256) {
        while (left <= right) {
            int256 pivotIndex = _partition(arr, left, right);
            if (pivotIndex == k) {
                return arr[uint256(pivotIndex)];
            } else if (pivotIndex < k) {
                left = pivotIndex + 1;
            } else {
                right = pivotIndex - 1;
            }
        }
        return -1;
    }

    function _partition(int256[] memory arr, int256 left, int256 right) private pure returns (int256) {
        int256 pivot = arr[uint256(right)];
        int256 i = left - 1;
        for (int256 j = left; j < right; j++) {
            if (arr[uint256(j)] <= pivot) {
                i++;
                (arr[uint256(i)], arr[uint256(j)]) = (arr[uint256(j)], arr[uint256(i)]);
            }
        }
        (arr[uint256(i + 1)], arr[uint256(right)]) = (arr[uint256(right)], arr[uint256(i + 1)]);
        return i + 1;
    }

    function getStates() external pure override returns (bytes memory) {
        return bytes("");
    }

    function setStates(bytes memory _states) external pure override {
    }
}
