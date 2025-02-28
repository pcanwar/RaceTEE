const fs = require("fs");
const path = require("path");
const { ensureDirectoryExists } = require("./dirExists");

function copySelectedFiles(srcDir, destDir, files) {
  ensureDirectoryExists(destDir);

  files.forEach((file) => {
    const srcFile = path.join(srcDir, `${file}.sol`, `${file}.json`);
    const destFile = path.join(destDir, `${file}.json`);

    if (fs.existsSync(srcFile)) {
      fs.copyFileSync(srcFile, destFile);
      console.log(`Copied: ${srcFile} -> ${destFile}`);
    } else {
      console.warn(`File not found: ${srcFile}`);
    }
  });
}

const sourceDir = "./artifacts/contracts";
// copy to client
let targetDir = "../client/artifacts";
let filesToCopy = [
  "ProgramContract",
  "UserContract",
  "UserContract2",
  "ManagementContract",
  "ERC20",
  "DEX",
  "QuickSelect",
  "SecondPriceAuction",
  "Calculate",
];
copySelectedFiles(sourceDir, targetDir, filesToCopy);

// copy to client
targetDir = "../tee/artifacts";
filesToCopy = [
  "ManagementContract",
  "StandardProgramContract",
  "SystemContract",
];
copySelectedFiles(sourceDir, targetDir, filesToCopy);

console.log("Selected files have been copied successfully!");
