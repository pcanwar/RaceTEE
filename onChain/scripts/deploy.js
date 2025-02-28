const fs = require("fs");
const { ethers } = require("hardhat")
const { ensureDirectoryExists } = require("./dirExists");

async function main() {
    const mcFactory = await ethers.getContractFactory("ManagementContract");
    const mc = await mcFactory.deploy();

    // get the address
    address = await mc.getAddress();

    // export accounts to a file
    const content = JSON.stringify({ address }, null, 2);
    ensureDirectoryExists("../client/artifacts");
    ensureDirectoryExists("../tee/artifacts");
    const outputPaths = ["../client/artifacts/managementAddress.json", "../tee/artifacts/managementAddress.json"];
    for (const outputPath of outputPaths) {
        fs.writeFileSync(outputPath, content, "utf-8");
        console.log(`Management contract deployed address has been exported to ${outputPath}`);
    }
}

// run
main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error);
        process.exitCode = 1;
    });
