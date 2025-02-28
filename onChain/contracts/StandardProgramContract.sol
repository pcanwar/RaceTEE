pragma solidity ^0.8.28;

import "./ManagementContract.sol";

abstract contract StandardProgramContract {
    address private immutable MCAdress;
    ManagementContract private immutable MC; // Management contract instance

    // The encryption key is transactionPubKey from the Management contract
    constructor(bytes memory encryptedCode,
        bytes memory encryptedConfig, 
        bytes memory transactionKey,
        address _MCAddress) payable {
        MCAdress = _MCAddress;
        MC = ManagementContract(_MCAddress);
        MC.deploy{value: msg.value}(encryptedCode, encryptedConfig, transactionKey, msg.sender);
    }

	function execution(bytes calldata encrytedInput, bytes calldata encryptedResultKey, bytes calldata transactionKey) external payable {
        MC.execution{value: msg.value}(encrytedInput, encryptedResultKey, transactionKey, msg.sender);
	}
	function changeACL(bytes calldata encrytedInput, bytes calldata transactionKey) external payable {
        MC.changeACL{value: msg.value}(encrytedInput, transactionKey, msg.sender);
	}
	
	// Set code and states, called only by the Management contract
    // Ensure only the Management contract can call this function
    modifier onlyManagement() {
        require(msg.sender == MCAdress, "Only the Management contract can call this function");
        _;
    }
	// Emit result so the client can retrieve and decrypt
	event Result (bytes encryptedResult, bytes encryptedResultKey);
	function setResult(bytes calldata encryptedResult, bytes calldata encryptedResultKey) external onlyManagement {
		emit Result (encryptedResult, encryptedResultKey);
	}
}