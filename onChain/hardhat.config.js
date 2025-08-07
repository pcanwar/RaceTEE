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
    localhost: {
      url: "http://127.0.0.1:8545"
    },
  },
};
