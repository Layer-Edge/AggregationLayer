package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/relayer"
)

// PROTOCOL_ID allows data identification by looking at the first few bytes
var (
	PROTOCOL_ID []byte
	cfg         = config.GetConfig()
)

func main() {
	PROTOCOL_ID = []byte(cfg.ProtocolId)
	fmt.Printf("%+v\n%+v\n", cfg, PROTOCOL_ID)

	client, err := ethclient.Dial(cfg.LayerEdgeRPC.WSS)
	if err != nil {
		log.Fatal(err)
	}

	relayer, err := relayer.NewRelayer(relayer.Config{
		Host:         cfg.Relayer.Host,
		User:         cfg.Relayer.User,
		Pass:         cfg.Relayer.Pass,
		HTTPPostMode: true,
		DisableTLS:   true,
	})
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
