// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./SystemContract.sol";

contract ERC20 is SystemContract {
    string public name;
    string public symbol;
    uint8 public decimals;
    uint256 public totalSupply;

    mapping(address => uint256) public balances;
    mapping(address => mapping(address => uint256)) public allowance;

    address[] private holders; // 持有者数组
    address[] private allowanceOwners; // 记录授权的用户
    mapping(address => address[]) private spenders; // 记录某用户所有授权的spender

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);

    constructor(string memory _name, string memory _symbol, uint8 _decimals, uint256 _initialSupply) {
        name = _name;
        symbol = _symbol;
        decimals = _decimals;
        totalSupply = _initialSupply;

        balances[msg.sender] = _initialSupply;
        holders.push(msg.sender);
        emit Transfer(address(0), msg.sender, _initialSupply);
    }

    function transfer(address recipient, uint256 amount) public returns (bool) {
        require(balances[msg.sender] >= amount, "Insufficient balance");
        
        balances[msg.sender] -= amount;
        balances[recipient] += amount;
        
        if (!isHolder(recipient)) {
            holders.push(recipient);
        }

        emit Transfer(msg.sender, recipient, amount);
        return true;
    }

    function approve(address spender, uint256 amount) public returns (bool) {
        allowance[msg.sender][spender] = amount;
        
        if (!isAllowanceOwner(msg.sender)) {
            allowanceOwners.push(msg.sender);
        }
        
        if (!isSpender(msg.sender, spender)) {
            spenders[msg.sender].push(spender);
        }

        emit Approval(msg.sender, spender, amount);
        return true;
    }

    function transferFrom(address sender, address recipient, uint256 amount) public returns (bool) {
        require(balances[sender] >= amount, "Insufficient balance");
        require(allowance[sender][msg.sender] >= amount, "Allowance exceeded");

        balances[sender] -= amount;
        balances[recipient] += amount;
        allowance[sender][msg.sender] -= amount;

        if (!isHolder(recipient)) {
            holders.push(recipient);
        }

        emit Transfer(sender, recipient, amount);
        return true;
    }

    function balanceOf(address account) public view returns (uint256) {
        return balances[account];
    }

    function isHolder(address holder) internal view returns (bool) {
        for (uint256 i = 0; i < holders.length; i++) {
            if (holders[i] == holder) {
                return true;
            }
        }
        return false;
    }

    function isAllowanceOwner(address owner) internal view returns (bool) {
        for (uint256 i = 0; i < allowanceOwners.length; i++) {
            if (allowanceOwners[i] == owner) {
                return true;
            }
        }
        return false;
    }
    

    function isSpender(address owner, address spender) internal view returns (bool) {
        for (uint256 i = 0; i < spenders[owner].length; i++) {
            if (spenders[owner][i] == spender) {
                return true;
            }
        }
        return false;
    }

    function getStates() external view override returns (bytes memory)  {
        uint256[] memory balancesArray = new uint256[](holders.length);
        for (uint256 i = 0; i < holders.length; i++) {
            balancesArray[i] = balances[holders[i]];
        }

        address[][] memory spenderArray = new address[][](allowanceOwners.length);
        uint256[][] memory allowanceArray = new uint256[][](allowanceOwners.length);
        for (uint256 i = 0; i < allowanceOwners.length; i++) {
            address owner = allowanceOwners[i];
            spenderArray[i] = spenders[owner];
            allowanceArray[i] = new uint256[](spenders[owner].length);
            for (uint256 j = 0; j < spenders[owner].length; j++) {
                allowanceArray[i][j] = allowance[owner][spenders[owner][j]];
            }
        }

        return abi.encode(name, symbol, decimals, totalSupply, holders, balancesArray, allowanceOwners, spenderArray, allowanceArray);
    }

    function setStates(bytes memory data) external override {
        (string memory _name, string memory _symbol, uint8 _decimals, uint256 _totalSupply, address[] memory _holders, uint256[] memory _balancesArray, address[] memory _allowanceOwners, address[][] memory _spenderArray, uint256[][] memory _allowanceArray) =
            abi.decode(data, (string, string, uint8, uint256, address[], uint256[], address[], address[][], uint256[][]));

        name = _name;
        symbol = _symbol;
        decimals = _decimals;
        totalSupply = _totalSupply;

        for (uint256 i = 0; i < _holders.length; i++) {
            balances[_holders[i]] = _balancesArray[i];
        }
        holders = _holders;

        for (uint256 i = 0; i < _allowanceOwners.length; i++) {
            address owner = _allowanceOwners[i];
            for (uint256 j = 0; j < _spenderArray[i].length; j++) {
                address spender = _spenderArray[i][j];
                allowance[owner][spender] = _allowanceArray[i][j];
                spenders[owner].push(spender);
            }
        }
        allowanceOwners = _allowanceOwners;
    }
}
