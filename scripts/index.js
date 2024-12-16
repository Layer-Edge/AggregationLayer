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

// Get command line arguments
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

        // Initialize client and check balance
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

            // Confirm transaction
            const txResult = await client.getTx(result.transactionHash);
            if (txResult) {
                console.log("\nTransaction confirmed!");
                console.log("Gas used:", txResult.gasUsed);
                console.log("Height:", txResult.height);
            }

        } catch (sendError) {
            console.error("\nTransaction failed!");
            console.error("Error details:", sendError.message);
            
            // Provide helpful error messages
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
