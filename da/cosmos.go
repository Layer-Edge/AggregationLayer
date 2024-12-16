package da

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Configuration constants
const (
	CosmosBashScriptPath = "/home/ubuntu/repo/bitcoin-da/scripts"
	DefaultCosmosAddress = "cosmos1c3y4q50cdyaa5mpfaa2k8rx33ydywl35hsvh0d"
)

// CosmosClientConfig holds the configuration for the Cosmos client
type CosmosClientConfig struct {
	ChainID        string
	RPCEndpoint    string
	AccountPrefix  string
	KeyringBackend string
	KeyName        string
	HomeDir        string
}

// CosmosClient represents a client for interacting with the Cosmos blockchain
type CosmosClient struct {
	config     CosmosClientConfig
	clientCtx  client.Context
	kr         keyring.Keyring
	senderAddr sdk.AccAddress
}

// Init initializes the Cosmos client with the provided configuration
func (c *CosmosClient) Init(cfg CosmosClientConfig) error {
	c.config = cfg

	clientCtx := client.Context{}
	clientCtx = clientCtx.
		WithChainID(cfg.ChainID).
		WithNodeURI(cfg.RPCEndpoint).
		WithBroadcastMode("block")

	// Configure SDK
	sdkConfig := sdk.GetConfig()
	sdkConfig.SetBech32PrefixForAccount(cfg.AccountPrefix, cfg.AccountPrefix+"pub")

	// Set up codec
	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	// Initialize keyring
	kr, err := keyring.New(
		"layeredge.info",
		"test",
		cfg.HomeDir,
		nil,
		marshaler,
	)
	if err != nil {
		return fmt.Errorf("failed to create keyring: %v", err)
	}

	// Get sender information
	senderInfo, err := kr.Key(cfg.KeyName)
	if err != nil {
		return fmt.Errorf("failed to get sender info: %v", err)
	}

	senderAddr, err := senderInfo.GetAddress()
	if err != nil {
		return fmt.Errorf("failed to get sender address: %v", err)
	}

	c.clientCtx = clientCtx.WithFromAddress(senderAddr)
	c.kr = kr
	c.senderAddr = senderAddr

	return nil
}

// SendData sends data to the Cosmos blockchain using a transaction
func (c *CosmosClient) SendData(data string) error {
	// Create a token transfer message
	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1)))
	msg := banktypes.NewMsgSend(c.senderAddr, c.senderAddr, amount)

	// Get account details
	accountNumber, sequence, err := c.getAccountNumberAndSequence()
	if err != nil {
		return fmt.Errorf("failed to get account number and sequence: %v", err)
	}

	// Set up transaction parameters
	gasPrices := sdk.NewDecCoins(sdk.NewDecCoin("stake", math.NewInt(1)))
	txf := tx.Factory{}.
		WithTxConfig(c.clientCtx.TxConfig).
		WithAccountNumber(accountNumber).
		WithSequence(sequence).
		WithGas(200000).
		WithGasPrices(gasPrices.String()).
		WithChainID(c.config.ChainID).
		WithMemo(data).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	// Build transaction
	txBuilder, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return fmt.Errorf("failed to build unsigned transaction: %v", err)
	}

	// Sign transaction
	err = tx.Sign(context.Background(), txf, c.config.KeyName, txBuilder, true)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Encode transaction
	txBytes, err := c.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return fmt.Errorf("failed to encode transaction: %v", err)
	}

	// Broadcast transaction
	res, err := c.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %v", err)
	}

	fmt.Printf("Transaction broadcasted successfully. Hash: %s\n", res.TxHash)
	fmt.Printf("Data sent in memo: %s\n", data)

	return nil
}

func (c *CosmosClient) getAccountNumberAndSequence() (uint64, uint64, error) {
	accNum, seq, err := c.clientCtx.AccountRetriever.GetAccountNumberSequence(c.clientCtx, c.senderAddr)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to retrieve account info: %w", err)
	}
	return accNum, seq, nil
}

func (c *CosmosClient) Send(data string, addr string) ([]byte, error) {
	// Construct the API endpoint URL
	apiURL := "https://cosmos-api-hcf6.onrender.com/send-tokens"

	// Prepare the request payload
	payload := map[string]string{
		"recipient": addr,
		"memo":      data[1:], // Assuming data starts with a special character, remove it
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client and send request
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a reasonable timeout
	}

	// Debug logging
	fmt.Printf("Debug - API URL: %s\n", apiURL)
	fmt.Printf("Debug - Request Payload: %s\n", string(jsonPayload))

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s",
			resp.StatusCode, string(body))
	}

	// Debug logging for response
	fmt.Printf("Debug - Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Debug - Response Body: %s\n", string(body))

	return body, nil
}
func (c *CosmosClient) ValidateScript() error {
	scriptPath := fmt.Sprintf("%s/run-cosmos-tx.sh", CosmosBashScriptPath)

	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("script not found at path: %s", scriptPath)
	}
	if err != nil {
		return fmt.Errorf("error checking script: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("script is not executable: %s", scriptPath)
	}

	return nil
}
