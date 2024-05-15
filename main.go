package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layer-Edge/bitcoin-da/relayer"
)

// PROTOCOL_ID allows data identification by looking at the first few bytes
var PROTOCOL_ID = []byte{0x72, 0x6f, 0x6c, 0x6c}

// Sample data and keys for testing.
// bob key pair is used for signing reveal tx
// internal key pair is used for tweaking
var (
	bobPrivateKey      = "cPbxEJ3UTLAeKzebFy6G38Qr7X5UqjcWv93PkhPJ52hoy9RtNkKD"
	internalPrivateKey = "cNR4CfUPBZNEZE9rShP4ix2NRPUNFfmDjecG7W9ySpupjGTMUKbw"
)

var LayerEdgeRPC = struct {
	WSS  string
	HTTP string
}{
	WSS:  "wss://testnet-rpc.layeredge.io/ws",
	HTTP: "https://testnet-rpc.layeredge.io/http",
}

var ExampleConfig = relayer.Config{
	Host:         "localhost:18443",
	User:         "jeet",
	Pass:         "SzKyQMucjU9pd6om64xcuMiEp4FqDtKAn_Q6QA16e6k",
	HTTPPostMode: true,
	DisableTLS:   true,
}

func main() {
	client, err := ethclient.Dial(LayerEdgeRPC.WSS)
	if err != nil {
		log.Fatal(err)
	}

	relayer, err := relayer.NewRelayer(ExampleConfig)
	if err != nil {
		fmt.Println(err)
		return
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
			fmt.Println("Hash: ", header.Hash().Hex())
			data, err := relayer.Read(PROTOCOL_ID)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Data: ", data)
		}
	}
}
