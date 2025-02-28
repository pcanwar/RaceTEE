# Project Overview

This is the open-source implementation of the paper "RaceTEE: A Practical Privacy-Preserving Off-Chain Smart Contract Execution Architecture".

This project consists of three primary components: `onChain`, `client`, and `tee`. Each component is designed to address specific responsibilities in a privacy-focused blockchain system.

## 1. `onChain`

The `onChain` directory contains the smart contracts deployed on the blockchain.

### System-Related Contracts
These contracts manage **Trusted Execution Environments (TEEs)** and ensure the secure execution of privacy programs. They handle key functions such as **deployment, execution, and state verification** for privacy-preserving applications.

### User-Defined Contracts
These contracts are written and deployed by users to function as **privacy programs** within our system.


## 2. `client`

The `client` directory provides the interface for users and developers to interact with the blockchain and privacy programs. It includes tools for deploying custom privacy programs, triggering their execution, and retrieving execution results. Additionally, it houses the `userpackage` for defining golang based user-specific privacy program logic.

## 3. `tee`

The `tee` directory hosts the runtime logic that operates within Trusted Execution Environments (TEEs). These secure environments perform computations for privacy programs and communicate with the blockchain to retrieve inputs and send back results.

## System Requirements

Before running the project, ensure the following dependencies are properly installed:

- **Go**: Version 1.21 is required for compiling and running `client` and `tee` programs.
- **EGo Framework**: Version 1.6.0 is necessary only if running in **trusted mode** (SGX). It is not required for untrusted mode.
- **Node.js**: Version 22.13 is used for deploying and interacting with the smart contracts in the `onChain` component.

## How to Use

### Step 1: Deploy Smart Contracts
```bash
cd onChain
npm install
npm run node
npm run ignition
```

### Step 2: Start TEE
#### For SGX (Secure Execution Mode):
```bash
cd tee
ego-go mod tidy
ego-go build
ego sign tee
ego run tee
```
#### For Untrusted Mode (Standard Execution):
```bash
cd tee
go mod tidy
go build
./tee
```

### Step 3: Write Privacy Programs

You can write your privacy programs in one of two ways:

- **Solidity Contracts**: Go to the onChain directory and create your smart contracts inside the contracts folder.
- **Golang Privacy Programs**: Go to the client directory and define your custom privacy programs in the userpackage folder.

### Step 4: Start Client
```bash
cd client
go mod tidy
go build
./client
```
