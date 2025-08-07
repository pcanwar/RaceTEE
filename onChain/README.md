# OnChain Component - Smart Contracts

This directory contains the smart contracts for the RaceTEE project, built with **Hardhat v6.1.0** and modern development practices.

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Start local blockchain
npm run start

# Deploy contracts and export artifacts
npm run ignition

# Run tests
npm run test
```

## 📁 Project Structure

```
onChain/
├── contracts/           # Solidity smart contracts
│   ├── SystemContract.sol          # Abstract base contract
│   ├── ManagementContract.sol      # TEE management
│   ├── ProgramContract.sol         # Privacy program execution
│   ├── Calculate.sol               # Example calculation contract
│   ├── QuickSelect.sol            # Example sorting algorithm
│   └── ...
├── scripts/            # Deployment and utility scripts
│   ├── deploy.js                  # Main deployment script
│   ├── exportAccounts.js         # Export test accounts
│   └── exportArtifacts.js        # Export contract ABIs
├── test/              # Comprehensive test suite
│   ├── Calculate.test.js          # Calculate contract tests
│   ├── QuickSelect.test.js        # QuickSelect contract tests
│   └── SystemContract.test.js     # Integration tests
└── artifacts/         # Compiled contracts (auto-generated)
```

## 🛠 Scripts

| Command | Description |
|---------|-------------|
| `npm run start` | Start Hardhat local blockchain node |
| `npm run ignition` | Complete deployment pipeline (compile + deploy + export) |
| `npm run compile` | Compile smart contracts |
| `npm run deploy` | Deploy contracts to Hardhat network |
| `npm run test` | Run the complete test suite |
| `npm run exportArtifacts` | Export contract ABIs to client/tee |
| `npm run exportAccounts` | Export test accounts to client/tee |

## 🔧 Technology Stack

- **Hardhat**: v2.26.2 - Ethereum development environment
- **Hardhat Toolbox**: v6.1.0 - Essential plugins and tools
- **Solidity**: Latest stable version
- **Chai/Mocha**: Testing framework

## 🔒 Security Features

This project includes security overrides for known vulnerabilities:
- Cookie >= 0.7.0
- Elliptic >= 6.6.1
- Undici >= 5.28.6
- And other security patches

## 📊 Smart Contracts Overview

### System Contracts
- **SystemContract.sol**: Abstract base contract defining core interfaces
- **ManagementContract.sol**: Manages TEE registration and verification
- **ProgramContract.sol**: Handles privacy program deployment and execution

### Example Contracts
- **Calculate.sol**: Demonstrates alternating sum calculation
- **QuickSelect.sol**: Implements k-th largest element selection algorithm
- **DEX.sol**: Decentralized exchange functionality
- **ERC20.sol**: Standard token implementation

## 🧪 Testing

The project includes comprehensive test coverage:

```bash
# Run all tests
npm run test

# Tests cover:
# - Contract deployment
# - Function correctness
# - Edge cases and error handling
# - Integration between contracts
```

## 🚀 Deployment

### Local Development
```bash
# Terminal 1: Start local blockchain
npm run start

# Terminal 2: Deploy contracts
npm run ignition
```

### Network Configuration
The project is configured for Hardhat's built-in network with:
- 20 test accounts with 10,000 ETH each
- Deterministic account generation
- Fast mining for development

## 📤 Integration with Other Components

After deployment, contracts and accounts are automatically exported to:
- `../client/artifacts/` - For client component integration
- `../tee/artifacts/` - For TEE component integration

## 🔄 Recent Updates

- ✅ Upgraded to Hardhat v6.1.0
- ✅ Removed Ganache dependencies
- ✅ Added comprehensive test suite
- ✅ Simplified deployment workflow
- ✅ Enhanced security with dependency overrides
- ✅ Streamlined account management

## 🤝 Contributing

When adding new contracts:
1. Place Solidity files in `contracts/`
2. Add corresponding tests in `test/`
3. Update deployment script if needed
4. Run tests to ensure compatibility

## 📝 License

This project is part of the RaceTEE system. Please refer to the main repository license.
