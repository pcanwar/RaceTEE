// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

abstract contract SystemContract {
    function getStates() external view virtual returns (bytes memory);
    function setStates(bytes memory data) external virtual; 

    function getInteractContracts() external view virtual returns (address[] memory) {
        return new address[](0);
    }
}