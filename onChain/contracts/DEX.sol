// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./SystemContract.sol";

interface IERC20 {
    function transfer(address recipient, uint256 amount) external returns (bool);
    function transferFrom(address sender, address recipient, uint256 amount) external returns (bool);
    function balanceOf(address account) external view returns (uint256);
}

contract DEX is SystemContract {
    IERC20 public tokenA;
    IERC20 public tokenB;
    uint256 public totalLiquidity;
    mapping(address => uint256) public liquidity;
    address[] public holders;

    constructor(address _tokenA, address _tokenB) {
        tokenA = IERC20(_tokenA);
        tokenB = IERC20(_tokenB);
    }

    function addLiquidity(uint256 tokenAAmount, uint256 tokenBAmount) external returns (uint256) {
        require(tokenAAmount > 0 && tokenBAmount > 0, "Must add both TokenA and TokenB");
        
        tokenA.transferFrom(msg.sender, address(this), tokenAAmount);
        tokenB.transferFrom(msg.sender, address(this), tokenBAmount);
        
        uint256 liquidityMinted = tokenAAmount;
        liquidity[msg.sender] += liquidityMinted;
        totalLiquidity += liquidityMinted;

        if(!isHolder(msg.sender)) {
            holders.push(msg.sender);
        }
        return liquidityMinted;
    }

    // exchange TokenA -> TokenB
    function swap(uint256 tokenAAmount) external {
        require(tokenAAmount > 0, "Must send TokenA");
        uint256 tokenBReserve = tokenB.balanceOf(address(this));
        uint256 tokenBAmount = getSwapAmount(tokenAAmount, tokenA.balanceOf(address(this)), tokenBReserve);
        
        tokenA.transferFrom(msg.sender, address(this), tokenAAmount);
        tokenB.transfer(msg.sender, tokenBAmount);
    }

    // exchange TokenB -> TokenA
    function tokenBToTokenASwap(uint256 tokenBAmount) external {
        require(tokenBAmount > 0, "Must send TokenB");
        uint256 tokenAReserve = tokenA.balanceOf(address(this));
        uint256 tokenAAmount = getSwapAmount(tokenBAmount, tokenB.balanceOf(address(this)), tokenAReserve);
        
        tokenB.transferFrom(msg.sender, address(this), tokenBAmount);
        tokenA.transfer(msg.sender, tokenAAmount);
    }

    function removeLiquidity(uint256 amount) external returns (uint256, uint256) {
        require(liquidity[msg.sender] >= amount, "Not enough liquidity");
        uint256 tokenAAmount = (amount * tokenA.balanceOf(address(this))) / totalLiquidity;
        uint256 tokenBAmount = (amount * tokenB.balanceOf(address(this))) / totalLiquidity;
        
        liquidity[msg.sender] -= amount;
        totalLiquidity -= amount;
        
        tokenA.transfer(msg.sender, tokenAAmount);
        tokenB.transfer(msg.sender, tokenBAmount);
        return (tokenAAmount, tokenBAmount);
    }

    // compute output amount based on (x * y = k)
    function getSwapAmount(uint256 inputAmount, uint256 inputReserve, uint256 outputReserve) public pure returns (uint256) {
        uint256 inputAmountWithFee = inputAmount * 997; // 0.3% exchange fee
        uint256 numerator = inputAmountWithFee * outputReserve;
        uint256 denominator = (inputReserve * 1000) + inputAmountWithFee;
        return numerator / denominator;
    }

    function isHolder(address account) public view returns (bool) {
        for (uint256 i = 0; i < holders.length; i++) {
            if (holders[i] == account) {
                return true;
            }
        }
        return false;
    }

    function getStates() external view override returns(bytes memory) {
        uint256[] memory liquidityAmount = new uint256[](holders.length);
        for (uint256 i = 0; i < holders.length; i++) {
            liquidityAmount[i] = liquidity[holders[i]];
        }
        return abi.encode(tokenA, tokenB, totalLiquidity, holders, liquidityAmount);
    }

    function setStates(bytes memory _data) external override {
        (IERC20 _tokenA, IERC20 _tokenB, uint256 _totalLiquidity, address[] memory _holders, uint256[] memory liquidityAmount) = abi.decode(_data, (IERC20, IERC20, uint256, address[], uint256[]));

        tokenA = _tokenA;
        tokenB = _tokenB;
        totalLiquidity = _totalLiquidity;
        holders = _holders;

        for (uint256 i = 0; i < holders.length; i++) {
            liquidity[holders[i]] = liquidityAmount[i];
        }
    }

    function getInteractContracts() external view virtual override returns (address[] memory) {
        address[] memory contracts = new address[](2);
        contracts[0] = address(tokenA);
        contracts[1] = address(tokenB);
        return contracts;
    }
}
