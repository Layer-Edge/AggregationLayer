#!/bin/bash

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
LOG_DIR="${SCRIPT_DIR}/logs"
LOG_FILE="${LOG_DIR}/transactions.log"

# Default values
RECIPIENT=""
MEMO=""

# Create logs directory
mkdir -p "$LOG_DIR"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to print usage
print_usage() {
    echo "Usage: $0 -r <recipient_address> -m <memo>"
    echo "  -r : Recipient address (required)"
    echo "  -m : Memo (required)"
    echo "Example: $0 -r cosmos1abc... -m \"test transfer\""
    exit 1
}

# Parse command line arguments
while getopts "r:m:" opt; do
    case $opt in
        r)
            RECIPIENT="$OPTARG"
            ;;
        m)
            MEMO="$OPTARG"
            ;;
        ?)
            print_usage
            ;;
    esac
done

# Check if required parameters are provided
if [ -z "$RECIPIENT" ] || [ -z "$MEMO" ]; then
    log "Error: Both recipient (-r) and memo (-m) are required"
    print_usage
fi

# Check for Node.js
if ! command -v node &> /dev/null; then
    log "Installing Node.js..."
    curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
    sudo apt-get install -y nodejs
fi

# Check if package.json exists and create only if it doesn't
if [ ! -f "package.json" ]; then
    log "Creating package.json..."
    cat > "package.json" << EOL
{
  "name": "cosmos-tx-script",
  "version": "1.0.0",
  "type": "module",
  "dependencies": {
    "@cosmjs/stargate": "0.29.5",
    "@cosmjs/proto-signing": "0.29.5",
    "@cosmjs/tendermint-rpc": "0.29.5",
    "@cosmjs/encoding": "0.29.5",
    "dotenv": "16.0.3"
  }
}
EOL

    # Install dependencies only if package.json was just created
    log "Installing dependencies for the first time..."
    npm install
fi

# Create or update index.js with the recipient and memo as command line arguments
log "Updating index.js..."
cat > "index.js" << 'EOL'
import pkg from '@cosmjs/stargate';
import protoPkg from '@cosmjs/proto-signing';
import { Tendermint34Client } from '@cosmjs/tendermint-rpc';
import dotenv from 'dotenv';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const { SigningStargateClient, StargateClient } = pkg;
const { DirectSecp256k1HdWallet } = protoPkg;

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
dotenv.config();

// Get recipient and memo from command line arguments
const recipient = process.argv[2];
const memo = process.argv[3];

const main = async () => {
    try {
        const mnemonic = process.env.MNEMONIC;
        const rpcEndpoint = "http://34.31.74.109:26657";

        if (!mnemonic) {
            throw new Error("MNEMONIC not found in environment variables");
        }

        // Initialize wallet
        const wallet = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);
        const [firstAccount] = await wallet.getAccounts();
        console.log("Account Address:", firstAccount.address);

        // Initialize client
        const client = await StargateClient.connect(rpcEndpoint);
        const balance = await client.getAllBalances(firstAccount.address);
        console.log("Account balance:", balance);

        if (balance.length === 0) {
            throw new Error("Account has no tokens. Please fund the account first.");
        }

        // Initialize signing client
        const signingClient = await SigningStargateClient.connectWithSigner(
            rpcEndpoint,
            wallet
        );

        const amount = {
            denom: "token",
            amount: "1",
        };

        console.log("\nPreparing transaction...");
        console.log("From:", firstAccount.address);
        console.log("To:", recipient);
        console.log("Amount:", amount);
        console.log("Memo:", memo);

        // Send transaction
        try {
            const fee = {
                amount: [{
                    denom: "token",
                    amount: "5000",
                }],
                gas: "200000",
            };

            console.log("\nSending transaction...");
            const result = await signingClient.sendTokens(
                firstAccount.address,
                recipient,
                [amount],
                fee,
                memo
            );

            console.log("\nTransaction sent successfully!");
            console.log("Transaction hash:", result.transactionHash);

            // Basic transaction confirmation
            const txResult = await client.getTx(result.transactionHash);
            if (txResult) {
                console.log("\nTransaction confirmed!");
                console.log("Gas used:", txResult.gasUsed);
                console.log("Height:", txResult.height);
            }

        } catch (sendError) {
            console.error("\nTransaction failed!");
            console.error("Error details:", sendError.message);
            
            if (sendError.message.includes("insufficient funds")) {
                console.log("\nPossible solution: Make sure you have enough tokens to cover both the transfer amount and gas fees.");
            }
            if (sendError.message.includes("invalid address")) {
                console.log("\nPossible solution: Check if the recipient address is correct.");
            }
        }

    } catch (error) {
        console.error("\nSetup error:", error.message);
        if (error.message.includes("does not exist on chain")) {
            console.log("\nSolution: Please fund your account first with some tokens before trying to send transactions.");
        }
        if (error.message.includes("invalid mnemonic")) {
            console.log("\nSolution: Check if your mnemonic phrase in .env file is correct.");
        }
    }
};

main().catch(error => {
    console.error("Fatal error:", error);
    process.exit(1);
});
EOL

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    log "Creating .env file..."
    echo "MNEMONIC=your_mnemonic_here" > .env
    chmod 600 .env
    log "Please update the .env file with your mnemonic"
fi

# Run the script with command line arguments
log "Executing transaction script..."
node index.js "$RECIPIENT" "$MEMO"
