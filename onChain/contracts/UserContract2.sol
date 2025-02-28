pragma solidity ^0.8.28;

import "./SystemContract.sol";
import "./UserContract.sol";

contract UserContract2 is SystemContract {
    uint sum = 0;

    UserContract public userContract;
    constructor(address userContractAdr){
        userContract = UserContract(userContractAdr);
    }

    function add(uint a) public returns (uint) {
        sum += a;
        userContract.multiply(a);
        return sum;
    }

    function getStates() external override view returns (bytes memory) {
        return abi.encode(sum, userContract);
    }

    function setStates(bytes memory data) external override {
        (sum, userContract) = abi.decode(data, (uint, UserContract));
    }

    function getInteractContracts () external view override returns (address[] memory) {
        address[] memory contracts = new address[](1);
        contracts[0] = address(userContract);
        return contracts;
    }
}