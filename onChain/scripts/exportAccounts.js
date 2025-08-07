const fs = require("fs");
const { ethers, network } = require("hardhat");
const { ensureDirectoryExists } = require("./dirExists");

async function main() {
  // load mnemonic from hardhat configuration
  let hardhatConfig = require("hardhat").config.networks.hardhat.accounts;
  let m = hardhatConfig.mnemonic || "test test test test test test test test test test test junk"; // default hardhat mnemonic
  let count = hardhatConfig.count || 20; // default to 20 accounts
  
  const mnemonic = ethers.Mnemonic.fromPhrase(m)


  const accounts = [];
  for (let i = 0; i < count; i++) {
    const wallet = ethers.HDNodeWallet.fromMnemonic(mnemonic, `m/44'/60'/0'/0`).deriveChild(i);
    accounts.push({
      address: wallet.address.substring(2),
      privateKey: wallet.privateKey.substring(2),
    });
  }

  // export accounts to a file
  const content = JSON.stringify(accounts, null, 2);
  ensureDirectoryExists("../client/artifacts");
  ensureDirectoryExists("../tee/artifacts");
  const outputPaths = ["../client/artifacts/accounts.json", "../tee/artifacts/accounts.json"];
  for (const outputPath of outputPaths) {
      fs.writeFileSync(outputPath, content, "utf-8");
      console.log(`Account details have been exported to  ${outputPath}`);
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
