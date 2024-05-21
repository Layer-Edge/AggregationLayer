package reader

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layer-Edge/bitcoin-da/config"
)

var cfg = config.GetConfig()

func SubscribeToBlocks() {
	client, err := ethclient.Dial(cfg.LayerEdgeRPC.WSS)
	if err != nil {
		log.Fatal(err)
	}

	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			fmt.Println(header.Hash().Hex())

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Println(err)
				continue
			}

			fmt.Println("Block number:", block.Number().Uint64())                      // 3477413
			fmt.Println("Block time:", block.Time())                                   // 1529525947
			fmt.Println("Block nonce:", block.Nonce())                                 // 130524141876765836
			fmt.Println("Number of transactions in block:", len(block.Transactions())) // 7
		}
	}
}

func ReadBlocks() {
	client, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Chain ID:", chainID.String())

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Current block number:", header.Number.String()) // 5671744

	block, err := client.BlockByNumber(context.Background(), header.Number)
	if err != nil {
		log.Fatal("failed to get block: ", err)
	}

	fmt.Println(block.Number().Uint64())     // 5671744
	fmt.Println(block.Time())                // 1527211625
	fmt.Println(block.Difficulty().Uint64()) // 3217000136609065
	fmt.Println(block.Hash().Hex())          // 0x9e8751ebb5069389b855bba72d94902cc385042661498a415979b7b6ee9ba4b9
	fmt.Println(len(block.Transactions()))   // 144

	for _, tx := range block.Transactions() {
		fmt.Println("Hash:", tx.Hash().Hex())            // 0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2
		fmt.Println("Value:", tx.Value().String())       // 10000000000000000
		fmt.Println("Gas:", tx.Gas())                    // 105000
		fmt.Println("GasPrice:", tx.GasPrice().Uint64()) // 102000000000
		fmt.Println("Nonce:", tx.Nonce())                // 110644
		fmt.Println("Data:", tx.Data())                  // []
		fmt.Println("To:", tx.To().Hex())                // 0x55fE59D8Ad77035154dDd0AD0388D09Dd4047A8e

		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Status:", receipt.Status) // 1
	}
}
