// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./SystemContract.sol";

interface IERC20 {
    function transfer(address recipient, uint256 amount) external returns (bool);
    function transferFrom(address sender, address recipient, uint256 amount) external returns (bool);
}

contract SecondPriceAuction is SystemContract {
    IERC20 public token;
    address private seller;
    uint256 private biddingEnd;
    address private highestBidder;
    uint256 private highestBid;
    uint256 private secondHighestBid;
    bool public auctionEnded;
    
    mapping(address => uint256) private bids;
    address[] private bidders;

    event AuctionFinalized(address winner, uint256 finalPrice);

    constructor(IERC20 _token, uint256 _biddingTime) {
        seller = msg.sender;
        biddingEnd = block.timestamp + _biddingTime;
        token = _token;
    }

    function placeBid(uint256 _amount) external {
        require(block.timestamp < biddingEnd, "Auction ended");
        
        require(token.transferFrom(msg.sender, address(this), _amount), "Token transfer failed");


        if(!isBidder(msg.sender)){
            bidders.push(msg.sender);
        }
        
        bids[msg.sender] += _amount;
        uint256 totalBid = bids[msg.sender];
        
        if (totalBid > highestBid) {
            secondHighestBid = highestBid;
            highestBid = totalBid;
            highestBidder = msg.sender;
        } else if (totalBid > secondHighestBid) {
            secondHighestBid = totalBid;
        }
    }

    function finalizeAuction() external returns(address, uint256) {
        require(block.timestamp >= biddingEnd, "Auction not ended");
        require(!auctionEnded, "Auction already finalized");
        auctionEnded = true;
        
        if (highestBidder != address(0)) {
            require(token.transfer(seller, secondHighestBid), "Token transfer to seller failed");
            require(token.transfer(highestBidder, highestBid - secondHighestBid), "Refund transfer failed");
        }
        
        for (uint256 i = 0; i < bidders.length; i++) {
            address account = bidders[i];
            uint256 refundAmount = bids[account];
            if (account != highestBidder && refundAmount > 0) {
                bids[account] = 0;
                require(token.transfer(account, refundAmount), "Refund transfer failed");
            }
        }
        
        emit AuctionFinalized(highestBidder, secondHighestBid);
        return (highestBidder, secondHighestBid);
    }

    function isBidder(address _account) private view returns (bool) {
        return bids[_account] > 0;
    }

    function getStates() external view override returns (bytes memory)  {
        uint256[] memory bidsArr = new uint256[](bidders.length);
        for (uint256 i = 0; i < bidders.length; i++) {
            bidsArr[i] = bids[bidders[i]];
        }
        return abi.encode(token, seller, biddingEnd, highestBidder, highestBid, secondHighestBid, auctionEnded, bidders, bidsArr);
    }

    function setStates(bytes memory _states) external override {
        uint256[] memory bidsArr;
        (token, seller, biddingEnd, highestBidder, highestBid, secondHighestBid, auctionEnded, bidders, bidsArr) = abi.decode(_states, (IERC20, address, uint256, address, uint256, uint256, bool, address[], uint256[]));

        for (uint256 i = 0; i < bidders.length; i++) {
            bids[bidders[i]] = bidsArr[i];
        }
    }

    function getInteractContracts() external view virtual override returns (address[] memory) {
        address[] memory contracts = new address[](1);
        contracts[0] = address(token);
        return contracts;
    }

}
