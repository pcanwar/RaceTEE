const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("SystemContract Integration", function () {
  let calculate, quickSelect;
  let owner;

  beforeEach(async function () {
    [owner] = await ethers.getSigners();
    
    const Calculate = await ethers.getContractFactory("Calculate");
    calculate = await Calculate.deploy();
    
    const QuickSelect = await ethers.getContractFactory("QuickSelect");
    quickSelect = await QuickSelect.deploy();
  });

  describe("Abstract SystemContract implementation", function () {
    it("Should have getStates function implemented in Calculate", async function () {
      const states = await calculate.getStates();
      expect(states).to.equal("0x");
    });

    it("Should have getStates function implemented in QuickSelect", async function () {
      const states = await quickSelect.getStates();
      expect(states).to.equal("0x");
    });

    it("Should have setStates function implemented in Calculate", async function () {
      await expect(calculate.setStates("0x1234")).to.not.be.reverted;
    });

    it("Should have setStates function implemented in QuickSelect", async function () {
      await expect(quickSelect.setStates("0x1234")).to.not.be.reverted;
    });

    it("Should have getInteractContracts function (inherited)", async function () {
      const contracts = await calculate.getInteractContracts();
      expect(contracts).to.be.an('array');
      expect(contracts.length).to.equal(0);
    });
  });

  describe("Contract deployment and basic functionality", function () {
    it("Should deploy Calculate contract successfully", async function () {
      expect(calculate.target).to.be.properAddress;
    });

    it("Should deploy QuickSelect contract successfully", async function () {
      expect(quickSelect.target).to.be.properAddress;
    });

    it("Calculate contract should perform basic calculation", async function () {
      const result = await calculate.cal(5);
      expect(result).to.equal(3);
    });

    it("QuickSelect contract should perform basic selection", async function () {
      const arr = [1, 2, 3, 4, 5];
      const result = await quickSelect.quickSelect(arr, 1);
      expect(result).to.equal(5); // largest element
    });
  });
});
