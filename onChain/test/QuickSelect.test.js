const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("QuickSelect", function () {
  let quickSelect;

  beforeEach(async function () {
    const QuickSelect = await ethers.getContractFactory("QuickSelect");
    quickSelect = await QuickSelect.deploy();
  });

  describe("quickSelect function", function () {
    it("Should return the k-th largest element", async function () {
      const arr = [3, 2, 1, 5, 6, 4];
      const k = 2;
      const result = await quickSelect.quickSelect(arr, k);
      expect(result).to.equal(5); // 2nd largest element
    });

    it("Should return the largest element when k=1", async function () {
      const arr = [3, 2, 1, 5, 6, 4];
      const k = 1;
      const result = await quickSelect.quickSelect(arr, k);
      expect(result).to.equal(6); // largest element
    });

    it("Should return the smallest element when k=length", async function () {
      const arr = [3, 2, 1, 5, 6, 4];
      const k = arr.length;
      const result = await quickSelect.quickSelect(arr, k);
      expect(result).to.equal(1); // smallest element
    });

    it("Should work with negative numbers", async function () {
      const arr = [-1, -3, 2, 0, -2];
      const k = 3;
      const result = await quickSelect.quickSelect(arr, k);
      expect(result).to.equal(-1); // 3rd largest: 2, 0, -1
    });

    it("Should work with single element array", async function () {
      const arr = [42];
      const k = 1;
      const result = await quickSelect.quickSelect(arr, k);
      expect(result).to.equal(42);
    });

    it("Should revert when k is 0", async function () {
      const arr = [3, 2, 1];
      const k = 0;
      await expect(quickSelect.quickSelect(arr, k))
        .to.be.revertedWith("Invalid k");
    });

    it("Should revert when k is greater than array length", async function () {
      const arr = [3, 2, 1];
      const k = 4;
      await expect(quickSelect.quickSelect(arr, k))
        .to.be.revertedWith("Invalid k");
    });
  });

  describe("SystemContract functions", function () {
    it("Should return empty bytes for getStates", async function () {
      const states = await quickSelect.getStates();
      expect(states).to.equal("0x");
    });

    it("Should not revert when calling setStates", async function () {
      await expect(quickSelect.setStates("0x1234")).to.not.be.reverted;
    });
  });
});
