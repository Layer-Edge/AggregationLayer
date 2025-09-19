package clients

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

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

func StoreMerkleTree(cfg *config.Config, merkle_root string, leaves []string) (*TxData, error) {
	layerEdgeClient, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		return nil, fmt.Errorf("error creating layerEdgeClient: %v", err)
	}

	// Your private key
	privateKeyStr := cfg.LayerEdgeRPC.PrivateKey
	// Remove 0x prefix if present, as crypto.HexToECDSA expects hex without prefix
	if strings.HasPrefix(privateKeyStr, "0x") {
		privateKeyStr = privateKeyStr[2:]
	}
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %v", err)
	}

	// Get public address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get nonce
	nonce, err := layerEdgeClient.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("error getting nonce: %v", err)
	}

	// Set gas price
	gasPrice, err := layerEdgeClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting gas price: %v", err)
	}

	// Create transactor
	chainID := big.NewInt(cfg.LayerEdgeRPC.ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("error creating transactor: %v", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)       // in wei
	auth.GasLimit = uint64(10000000) // gas limit
	auth.GasPrice = gasPrice

	contractAddress := common.HexToAddress(cfg.LayerEdgeRPC.MerkleTreeStorageContract)
	merkleTreeStorageContract, err := contracts.NewMerkleTreeStorage(contractAddress, layerEdgeClient)
	if err != nil {
		return nil, fmt.Errorf("error creating merkleTreeStorageContract: %v", err)
	}

	// Parse merkle root string into [32]byte
	// Expected format: "0xhash" or "hash" (will add 0x prefix if not present)
	merkleRootStr := strings.TrimSpace(merkle_root)
	if !strings.HasPrefix(merkleRootStr, "0x") {
		merkleRootStr = "0x" + merkleRootStr
	}
	merkleRootHash := common.HexToHash(merkleRootStr)

	// Parse leaves string into array of bytes
	// Expected format: plain strings that will be converted to bytes
	var leafHashes [][]byte

	for _, leafStr := range leaves {
		leafStr = strings.TrimSpace(leafStr)
		if leafStr == "" {
			continue
		}

		// Convert plain string to bytes
		leafHashes = append(leafHashes, []byte(leafStr))
	}

	// Prepare call data for gas estimation
	// Use the ABI from the contracts package
	storeTreeABI, err := abi.JSON(strings.NewReader(contracts.MerkleTreeStorageABI))
	if err != nil {
		return nil, fmt.Errorf("error parsing ABI: %v", err)
	}
	storeTreeData, err := storeTreeABI.Pack("storeTree", merkleRootHash, leafHashes)
	if err != nil {
		return nil, fmt.Errorf("error packing storeTree data for gas estimation: %v", err)
	}

	callMsg := ethereum.CallMsg{
		From:     fromAddress,
		To:       &contractAddress,
		Gas:      0,
		GasPrice: gasPrice,
		Value:    big.NewInt(0),
		Data:     storeTreeData,
	}

	estimatedGas, err := layerEdgeClient.EstimateGas(context.Background(), callMsg)
	if err != nil {
		return nil, fmt.Errorf("error estimating gas: %v", err)
	}

	auth.GasLimit = estimatedGas + 10000 // gas limit with buffer

	// Call a write function (e.g., addLeaf)
	tx, err := merkleTreeStorageContract.StoreTree(auth, merkleRootHash, leafHashes)
	if err != nil {
		return nil, fmt.Errorf("error in store merkle tree contract call: %v", err)
	}

	log.Println("Transaction sent:", tx.Hash().Hex())
	log.Println("Waiting for transaction to be mined...")

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), layerEdgeClient, tx)
	if err != nil {
		return nil, fmt.Errorf("error waiting for transaction to be mined: %v", err)
	}

	TransactionFee := new(big.Int).Mul(big.NewInt(int64(receipt.GasUsed)), receipt.EffectiveGasPrice)
	TransactionFee18Decimals := utils.FormatAmount(TransactionFee, 18, 18)

	EdgenPrice := GetPrice(cfg, "EDGEN")

	return &TxData{
		Success:         receipt.Status == 1,
		From:            fromAddress.Hex(),
		To:              cfg.LayerEdgeRPC.MerkleTreeStorageContract,
		Amount:          fmt.Sprintf("%f", EdgenPrice*TransactionFee18Decimals),
		TransactionHash: tx.Hash().Hex(),
		TransactionFee:  fmt.Sprintf("%f", TransactionFee18Decimals),
		EdgenPrice:      fmt.Sprintf("%f", EdgenPrice),
		Memo:            "",
		BlockHeight:     receipt.BlockNumber.String(),
		GasUsed:         strconv.FormatUint(receipt.GasUsed, 10),
	}, nil
}
