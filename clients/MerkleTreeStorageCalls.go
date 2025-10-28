package clients

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/contracts"
	"github.com/Layer-Edge/bitcoin-da/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TxData struct {
	Success         bool   `json:"success"`
	From            string `json:"from"`
	To              string `json:"to"`
	Amount          string `json:"amount"`
	TransactionHash string `json:"transactionHash"`
	TransactionFee  string `json:"transactionFee"`
	EdgenPrice      string `json:"edgenPrice"`
	Memo            string `json:"memo"`
	BlockHeight     string `json:"blockHeight"` // can use int64 if you want to parse it directly
	GasUsed         string `json:"gasUsed"`     // same here
}

// ContractCircuitBreaker manages the circuit breaker state for contract calls
type ContractCircuitBreaker struct {
	mutex        sync.RWMutex
	failureCount int
	lastFailTime time.Time
	circuitOpen  bool
	timeout      time.Duration
	maxFailures  int
}

// RetryConfig holds configuration for retry mechanisms
type RetryConfig struct {
	MaxRetries    int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

var (
	contractCircuitBreaker = &ContractCircuitBreaker{
		timeout:     60 * time.Second,
		maxFailures: 5,
	}

	retryConfig = &RetryConfig{
		MaxRetries:    3,
		BaseDelay:     2 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 2.0,
	}
)

// CanExecute checks if the circuit breaker allows execution
func (cb *ContractCircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if !cb.circuitOpen {
		return true
	}

	// Check if enough time has passed to try again
	return time.Since(cb.lastFailTime) > cb.timeout
}

// RecordSuccess records a successful operation
func (cb *ContractCircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.circuitOpen = false
}

// RecordFailure records a failed operation
func (cb *ContractCircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.circuitOpen = true
		log.Printf("Contract circuit breaker opened due to %d failures", cb.failureCount)
	}
}

// RetryContractCall executes a contract call with exponential backoff retry
func RetryContractCall(operation func() (*TxData, error)) (*TxData, error) {
	var lastErr error

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if !contractCircuitBreaker.CanExecute() {
			return nil, fmt.Errorf("contract circuit breaker is open, operation rejected")
		}

		if attempt > 0 {
			delay := time.Duration(float64(retryConfig.BaseDelay) *
				utils.PowFloat(retryConfig.BackoffFactor, float64(attempt-1)))
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}

			log.Printf("Retrying contract call after %v delay (attempt %d/%d)",
				delay, attempt+1, retryConfig.MaxRetries+1)

			time.Sleep(delay)
		}

		result, err := operation()
		if err == nil {
			contractCircuitBreaker.RecordSuccess()
			return result, nil
		}

		lastErr = err
		contractCircuitBreaker.RecordFailure()
		log.Printf("Contract call failed (attempt %d/%d): %v",
			attempt+1, retryConfig.MaxRetries+1, err)
	}

	return nil, fmt.Errorf("contract call failed after %d attempts: %w", retryConfig.MaxRetries+1, lastErr)
}

func StoreMerkleTree(cfg *config.Config, contractAddress string, merkle_root string, leaves []string) (*TxData, error) {
	return RetryContractCall(func() (*TxData, error) {
		return storeMerkleTreeWithRetry(cfg, contractAddress, merkle_root, leaves)
	})
}

// storeMerkleTreeWithRetry performs the actual contract interaction with enhanced error handling
func storeMerkleTreeWithRetry(cfg *config.Config, contractAddress string, merkle_root string, leaves []string) (*TxData, error) {
	// Create client with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	layerEdgeClient, err := ethclient.DialContext(ctx, cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		return nil, fmt.Errorf("error creating layerEdgeClient: %w", err)
	}
	defer layerEdgeClient.Close()

	// Your private key
	privateKeyStr := cfg.LayerEdgeRPC.PrivateKey
	// Remove 0x prefix if present, as crypto.HexToECDSA expects hex without prefix
	privateKeyStr = strings.TrimPrefix(privateKeyStr, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	// Get public address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get nonce with timeout
	nonceCtx, nonceCancel := context.WithTimeout(ctx, 10*time.Second)
	nonce, err := layerEdgeClient.PendingNonceAt(nonceCtx, fromAddress)
	nonceCancel()
	if err != nil {
		return nil, fmt.Errorf("error getting nonce: %w", err)
	}

	// Set gas price with timeout
	gasPriceCtx, gasPriceCancel := context.WithTimeout(ctx, 10*time.Second)
	gasPrice, err := layerEdgeClient.SuggestGasPrice(gasPriceCtx)
	gasPriceCancel()
	if err != nil {
		return nil, fmt.Errorf("error getting gas price: %w", err)
	}

	// Create transactor
	chainID := big.NewInt(cfg.LayerEdgeRPC.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("error creating transactor: %w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(10000000) // gas limit
	auth.GasPrice = gasPrice

	contractAddr := common.HexToAddress(contractAddress)
	merkleTreeStorageContract, err := contracts.NewMerkleTreeStorage(contractAddr, layerEdgeClient)
	if err != nil {
		return nil, fmt.Errorf("error creating merkleTreeStorageContract: %w", err)
	}

	// Parse merkle root string into [32]byte
	// Expected format: "0xhash" or "hash" (will add 0x prefix if not present)
	merkleRootStr := strings.TrimSpace(merkle_root)
	if !strings.HasPrefix(merkleRootStr, "0x") {
		merkleRootStr = "0x" + merkleRootStr
	}
	merkleRootHash := common.HexToHash(merkleRootStr)

	log.Printf("\nMerkleroothash: %v", merkleRootHash)

	// Parse leaves string into array of bytes
	// Expected format: plain strings that will be converted to bytes
	var leafHashes [][32]byte

	for _, leafStr := range leaves {
		leafStr = strings.TrimSpace(leafStr)
		if leafStr == "" {
			continue
		}

		if !strings.HasPrefix(leafStr, "0x") {
			leafStr = "0x" + leafStr
		}

		// Convert plain string to bytes
		leafHashes = append(leafHashes, common.HexToHash(leafStr))
	}

	// Prepare call data for gas estimation
	// Use the ABI from the contracts package
	storeTreeABI, err := abi.JSON(strings.NewReader(contracts.MerkleTreeStorageABI))
	if err != nil {
		return nil, fmt.Errorf("error parsing ABI: %w", err)
	}
	storeTreeData, err := storeTreeABI.Pack("storeTree", merkleRootHash, leafHashes)
	if err != nil {
		return nil, fmt.Errorf("error packing storeTree data for gas estimation: %w", err)
	}

	callMsg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &contractAddr,
		Gas:      0,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     storeTreeData,
	}

	// Estimate gas with timeout
	gasEstimateCtx, gasEstimateCancel := context.WithTimeout(ctx, 15*time.Second)
	estimatedGas, err := layerEdgeClient.EstimateGas(gasEstimateCtx, callMsg)
	gasEstimateCancel()
	if err != nil {
		return nil, fmt.Errorf("error estimating gas: %w", err)
	}

	auth.GasLimit = estimatedGas + 10000 // gas limit with buffer

	// Call a write function (e.g., addLeaf)
	tx, err := merkleTreeStorageContract.StoreTree(auth, merkleRootHash, leafHashes)
	if err != nil {
		return nil, fmt.Errorf("error in store merkle tree contract call: %w", err)
	}

	log.Println("Transaction sent:", tx.Hash().Hex())
	log.Println("Waiting for transaction to be mined...")

	// Wait for transaction to be mined with timeout
	waitCtx, waitCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer waitCancel()

	receipt, err := bind.WaitMined(waitCtx, layerEdgeClient, tx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for transaction to be mined: %w", err)
	}

	TransactionFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), receipt.EffectiveGasPrice)
	TransactionFee18Decimals := utils.FormatAmount(TransactionFee, 18, 18)

	EdgenPrice := GetPrice(cfg, "EDGEN")

	return &TxData{
		Success:         receipt.Status == 1,
		From:            fromAddress.Hex(),
		To:              contractAddress,
		Amount:          fmt.Sprintf("%.18f", EdgenPrice*TransactionFee18Decimals),
		TransactionHash: tx.Hash().Hex(),
		TransactionFee:  fmt.Sprintf("%.18f", TransactionFee18Decimals),
		EdgenPrice:      fmt.Sprintf("%.18f", EdgenPrice),
		Memo:            "",
		BlockHeight:     receipt.BlockNumber.String(),
		GasUsed:         strconv.FormatUint(receipt.GasUsed, 10),
	}, nil
}
