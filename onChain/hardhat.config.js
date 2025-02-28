require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    version: "0.8.28",
    settings: {
      optimizer: {
        enabled: true,
        runs: 200,
      },
      viaIR: true,
    }
  },

  networks: {
    hardhat: {
      mining: {
        auto: true, 
        // interval: 12000
      },
    },
    ganache: {
      url: "http://localhost:8545",
      accounts: {
        count: 200,
        mnemonic: "lift focus style steel census extra glory visual mercy mind differ pelican",
      },
    },
  },
};
