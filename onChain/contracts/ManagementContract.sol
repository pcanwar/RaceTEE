pragma solidity ^0.8.28;

import "./StandardProgramContract.sol";

struct BlockInfo {
    uint64 blockNumber;
    bytes32 blockHash;
}

contract ManagementContract {
    // Stores all registered TEE information.
    // Key: the address of the TEE host, Value: a TEEDetail struct.
    struct TEEDetail {
        bytes      key;              // TEE public key
        bytes      attestationReport;// IAS report for verifying SGX authenticity
        uint128    deposit;          // Amount of deposit
        uint64     blockNumber;      // Block number of the registration time
    }
    mapping(address => TEEDetail) public TEEList;
    address[] public TEEListArray;
    uint256 public constant deposit = 1 ether; // 1ETH
    // Stores all registered privacy programs.
    // Key: the program contract address, Value: the encrypted PrivacyProgram struct hash.
    mapping(address => bytes32) public ProgramList;
    mapping(address => bytes32) public ProgramStates;
    mapping(address => bytes32) public ProgramCodes;
    // Records the system’s progression, specifying the blocks at which it has run.
    BlockInfo public latestExecutionBlock;
    // Used for clients to encrypt transactions.
    string public transactionPubKey;

    // output Structre
    enum TransType {Execution, Deploy, Interact, ACL, Err}
    struct Output {
        address programAddress;
        bytes32 code;
        bytes32 info;
        bytes32 states;
        bytes result;
        bytes encryptedResultKey;
        TransType transType;
    }

    

    constructor(){
        latestExecutionBlock = BlockInfo({
            // blockNumber: 21768779,
            blockNumber: 0,
            blockHash: bytes32(0)
        });
        // Fixed by using pubkey from tee/key/tempTxKey.json (prototype)
        transactionPubKey = "042029597ee1100996aa76ceb36770e18e02a5addc7e5b546c1be44f142399508cbb6e88a530db365c5e7e790891e91d13d08a8893858fbbc7d994315142c2a126";
    }

    // Ensure only the TEEs can call this function
    modifier onlyTEE() {
        // check msg.sender in TEEList;
        require(TEEList[msg.sender].deposit != 0x0, "Only registered TEE can call this function");
        _;
    }
    modifier checkTEESig(BlockInfo calldata start, BlockInfo calldata end, Output[] calldata outputs, bytes calldata signature){
        bytes32 messageHash = hashOutputs(start, end, outputs);
        require(verifySignature(messageHash, signature, TEEList[msg.sender].key), "Invalid signature");
        _;
    }
    modifier checkBlock(BlockInfo calldata start, BlockInfo calldata end){
        // Ensure the input is the latest input (only not the first block)
        if (latestExecutionBlock.blockHash != bytes32(0)) {
            require(start.blockNumber == latestExecutionBlock.blockNumber, "Start block number does not match");
            require(start.blockHash == latestExecutionBlock.blockHash, "Start block hash does not match");
        }

        // Ensure the output is based on the latest input
        require(end.blockNumber > latestExecutionBlock.blockNumber, "End block number is not greater than the latest input block number");
        require(end.blockHash == blockhash(end.blockNumber), "End block hash does not match");
        _;
    }

    // Also used by outside caller to generate the hash for signing
    function hashOutputs(BlockInfo calldata start, BlockInfo calldata end, Output[] calldata outputs) pure public returns(bytes32){ 
        // must have this prefix for ecrecover get the real signer
        return keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", abi.encode(start, end, outputs))
        );
    }

    // The output function only accepts calls from the corresponding TEEs. 
    // It uses the input to verify that the computed outputs are based on the latest system state.
    function output(BlockInfo calldata start, BlockInfo calldata end, Output[] calldata outputs, bytes calldata signature) external
        checkBlock(start, end) onlyTEE checkTEESig(start, end, outputs, signature) {
        uint64 len = uint64(outputs.length);
        for (uint64 i = 0; i < len; i++) {
            Output calldata outp = outputs[i];
            address progAddr = outp.programAddress;
            // These data include all changes corresponding to each privacy program
            if (outp.transType == TransType.Execution) {
                // Update contract internal state
                ProgramList[progAddr] = outp.info;
                // Update corresponding contract states
                ProgramStates[progAddr] = outp.states;
                StandardProgramContract(progAddr).setResult(outp.result, outp.encryptedResultKey);
            }
            // These contract is called by other privacy programs
            else if(outp.transType == TransType.Interact) {
                ProgramList[progAddr] = outp.info;
                // Update corresponding contract states
                ProgramStates[progAddr] = outp.states;
                // StandardProgramContract(progAddr).setStates(outp.states);
            }
            else if (outp.transType == TransType.Deploy) {
                ProgramList[progAddr] = outp.info;
                // Update corresponding contract states
                ProgramStates[progAddr] = outp.states;
                ProgramCodes[progAddr] = outp.code;
            }
            // change ACL
            else if(outp.transType == TransType.ACL) {
                ProgramList[progAddr] = outp.info;
            }
            else if(outp.transType == TransType.Err) {
                // Update corresponding contract states
                StandardProgramContract(progAddr).setResult(outp.result, outp.encryptedResultKey);
            }
        }
        // Update the system state
        latestExecutionBlock = end;
    }

    function register(bytes calldata attestationReport, bytes calldata key) external payable{
        require(msg.value >= deposit, "Deposit is not enough");
        // TODO: check attestationReport and key is valid
        TEEList[msg.sender] = TEEDetail({
            key: key,
            attestationReport: attestationReport,
            deposit: uint128(msg.value),
            blockNumber: uint64(block.number)
        });
    }

    // Ensure only the system contracts can call 
    modifier validCall(address caller) {
        require(ProgramList[caller].length > 0, "Program address not found in ProgramList");
        _;
    }
    // The following functions are used for indirect calls from the program contract
    event Deploy(bytes encryptedCode, bytes encryptedConfig, bytes transactionKey, address caller, address programAddress);
    function deploy(bytes calldata encryptedCode, bytes calldata encryptedConfig, bytes calldata transactionKey, address caller) external payable{
        // TODO: check transaction fee is enough
        emit Deploy(encryptedCode, encryptedConfig, transactionKey, caller, msg.sender);
    }
    event Execution(bytes encryptedInput, bytes encryptedResultKey, bytes transactionKey, address caller, address programAddress);
    function execution(bytes calldata encryptedInput, bytes calldata encryptedResultKey, bytes calldata transactionKey, address caller) external payable validCall(msg.sender){
        // TODO: check transaction fee is enough
        emit Execution(encryptedInput, encryptedResultKey, transactionKey, caller, msg.sender);
    }
    event ACL(bytes encryptedInput, bytes transactionKey, address caller, address programAddress);
    function changeACL(bytes calldata encryptedInput, bytes calldata transactionKey, address caller) external payable validCall(msg.sender){
        // TODO: check transaction fee is enough
        emit ACL(encryptedInput, transactionKey, caller, msg.sender);
    }

    // verify signature is generated by corresponding public key
    function verifySignature(
        bytes32 messageHash,
        bytes memory signature,
        bytes memory expectedPublicKey
    ) internal pure returns (bool) {
        require(signature.length == 65, "Invalid signature length");
        bytes32 r;
        bytes32 s;
        uint8 v;
        // decode signature
        assembly {
            r := mload(add(signature, 0x20))
            s := mload(add(signature, 0x40))
            v := byte(0, mload(add(signature, 0x60)))
        }
        // adjust v value
        if (v < 27) {
            v += 27;
        }
        // recover address from signature
        address recovered = ecrecover(messageHash, v, r, s);
        // calculate the address corresponding to the expectedPublicKey
        bytes32 pubKeyHash = keccak256(expectedPublicKey); // calculate the public key hash
        address expected = address(uint160(uint256(pubKeyHash))); // take the last 20 bytes
        return recovered == expected;
    }
}