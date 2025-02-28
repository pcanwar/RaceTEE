pragma solidity ^0.8.28;

import "./StandardProgramContract.sol";

contract ProgramContract is StandardProgramContract {
	constructor(
        bytes memory encryptedCode,
        bytes memory encrytedConfig,
        bytes memory transactionKey,
        address MCAddress) payable
        StandardProgramContract(encryptedCode, encrytedConfig, transactionKey, MCAddress) {
	}
}