// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./SystemContract.sol";

contract UserContract is SystemContract {
    uint sum = 0;
    uint product = 1;
    struct Person {
        string name;
        uint age;
    }
    uint[3] public staticValues;
    uint[] public dynamicValues = [1, 2];
    mapping(uint => Person) public personMap;
    uint[] public keys;
    
    constructor() {
        staticValues[0] = 1;
        staticValues[1] = 2;
        staticValues[2] = 3;
        dynamicValues.push(3);
        personMap[1] = Person("Alice", 20);
        keys.push(1);
        personMap[2] = Person("Bob", 30);
        keys.push(2);
    }

    function add(uint a) public returns (uint) {
        sum += a;
        return sum;
    }

    function multiply(uint a) public returns (uint) {
        product *= a;
        return product;
    }

    function getStates() external override view returns (bytes memory) {
        uint[] memory keyArray;
        Person[] memory valueArray;
        (keyArray, valueArray) = processMapping();
        return abi.encode(sum, product, staticValues, dynamicValues, keyArray, valueArray);
    }

    function setStates(bytes memory data) external override {
        uint[] memory keyArray;
        Person[] memory valueArray;
        (sum, product, staticValues, dynamicValues, keyArray, valueArray) = abi.decode(data, (uint, uint, uint[3], uint[], uint[], Person[]));

        for (uint i = 0; i < keyArray.length; i++) {
            personMap[keyArray[i]] = valueArray[i];
        }
        keys = keyArray;
    }

    function processMapping() public view returns (uint[] memory key, Person[] memory valueArray) {
        uint len = keys.length;
        valueArray = new Person[](len);

        for (uint i = 0; i < len; i++) {
            valueArray[i] = personMap[keys[i]];
        }
        return (keys, valueArray);
    }
}