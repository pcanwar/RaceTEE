const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("Calculate", function () {
  let calculate;

  beforeEach(async function () {
    const Calculate = await ethers.getContractFactory("Calculate");
    calculate = await Calculate.deploy();
  });

  describe("cal function", function () {
    it("Should return correct result for n=1", async function () {
      const result = await calculate.cal(1);
      expect(result).to.equal(1); // 1
    });

    it("Should return correct result for n=2", async function () {
      const result = await calculate.cal(2);
      expect(result).to.equal(-1); // 1 - 2 = -1
    });

    it("Should return correct result for n=3", async function () {
      const result = await calculate.cal(3);
      expect(result).to.equal(2); // 1 - 2 + 3 = 2
    });

    it("Should return correct result for n=4", async function () {
      const result = await calculate.cal(4);
      expect(result).to.equal(-2); // 1 - 2 + 3 - 4 = -2
    });

    it("Should return correct result for n=5", async function () {
      const result = await calculate.cal(5);
      expect(result).to.equal(3); // 1 - 2 + 3 - 4 + 5 = 3
    });

    it("Should handle larger numbers", async function () {
      const result = await calculate.cal(10);
      // 1 - 2 + 3 - 4 + 5 - 6 + 7 - 8 + 9 - 10 = -5
      expect(result).to.equal(-5);
    });

    it("Should return 0 for n=0", async function () {
      const result = await calculate.cal(0);
      expect(result).to.equal(0);
    });
  });

  describe("SystemContract functions", function () {
    it("Should return empty bytes for getStates", async function () {
      const states = await calculate.getStates();
      expect(states).to.equal("0x");
    });

    it("Should not revert when calling setStates", async function () {
      await expect(calculate.setStates("0x1234")).to.not.be.reverted;
    });
  });
});
