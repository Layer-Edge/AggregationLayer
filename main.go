package main

import (
	"context"
	"log"
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/relayer"
)

// PROTOCOL_ID allows data identification by looking at the first few bytes
var (
	PROTOCOL_ID = []byte(config.GetConfig().ProtocolId)
	cfg         = config.GetConfig()
)

func main() {
	certs, err := os.ReadFile("./rpc.cert")
	if err != nil {
		log.Fatal(err)
	}

	layerEdgeClient, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		log.Fatal(err)
	}

	relayer, err := relayer.NewRelayer(relayer.Config{
		Host:         cfg.WalletRelayer.Host,
		User:         cfg.WalletRelayer.User,
		Pass:         cfg.WalletRelayer.Pass,
		HTTPPostMode: true,
		Certificates: certs,
	}, nil)
	if err != nil {
		log.Fatal("Error creating http relayer: ", err)
	}

	// Only override the handlers for notifications you care about.
	// Also note most of these handlers will only be called if you register
	// for notifications.  See the documentation of the rpcclient
	// NotificationHandlers type for more details about each handler.
	ntfnHandlers := rpcclient.NotificationHandlers{
		OnFilteredBlockConnected: func(height int32, header *wire.BlockHeader, txns []*btcutil.Tx) {
			log.Printf("\n\nBlock connected: %v (%d) %v\n", header.BlockHash(), height, header.Timestamp)

			layerEdgeHeader, err := layerEdgeClient.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Latest LayerEdge Block Hash:", layerEdgeHeader.Hash().Hex())
			hash, err := relayer.Write(
				cfg.PrivateKey.Signer,
				cfg.PrivateKey.Internal,
				PROTOCOL_ID,
				[]byte(layerEdgeHeader.Hash().Hex()),
			)
			if err != nil {
				log.Println(err)
				return
			}
			log.Println("Relayer Write Done: ", hash)
		},
		OnFilteredBlockDisconnected: func(height int32, header *wire.BlockHeader) {
			log.Printf("Block disconnected: %v (%d) %v",
				header.BlockHash(), height, header.Timestamp)
		},
	}

	btcdListener, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         cfg.WsRelayer.Host,
		User:         cfg.WsRelayer.User,
		Pass:         cfg.WsRelayer.Pass,
		Endpoint:     "ws",
		Certificates: certs,
	}, &ntfnHandlers)
	if err != nil {
		log.Fatal("Error creating btcd wss listener: \n", err)
	}

	// Register for block connect and disconnect notifications.
	if err := btcdListener.NotifyBlocks(); err != nil {
		log.Fatal(err, "\nHOST:", cfg.WsRelayer.Host)
	}
	log.Println("NotifyBlocks: Registration Complete")

	// Get the current block count.
	blockCount, err := btcdListener.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	hash, err := btcdListener.GetBlockHash(blockCount)
	if err != nil {
		log.Fatal(err)
	}
	block, err := btcdListener.GetBlock(hash)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block: %v (%d)", block.Header.Timestamp, blockCount)

	// // For this example gracefully shutdown the client after 10 seconds.
	// log.Println("Client shutdown in 30 minutes...")
	// time.AfterFunc(time.Minute*30, func() {
	// 	log.Println("Client shutting down...")
	// 	btcdListener.Shutdown()
	// 	log.Println("Client shutdown complete.")
	// })

	// Wait until the client either shuts down gracefully (or the user
	// terminates the process with Ctrl+C).
	btcdListener.WaitForShutdown()
}
