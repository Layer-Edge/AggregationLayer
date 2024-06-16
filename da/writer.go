package da

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/zeromq/goczmq.v4"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/relayer"
)

func HashBlockSubscriber(cfg *config.Config) {
	channeler := goczmq.NewSubChanneler(cfg.ZmqEndpoint, "hashblock")

	if channeler == nil {
		log.Fatal("Error creating channeler", channeler)
	}
	defer channeler.Destroy()

	relayer, err := relayer.NewRelayer(relayer.Config{
		Host:         cfg.Relayer.Host,
		User:         cfg.Relayer.User,
		Pass:         cfg.Relayer.Pass,
		DisableTLS:   true,
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		log.Fatal("Error creating http relayer: ", err)
	}

	layerEdgeClient, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		log.Fatal("Error creating layerEdgeClient: ", err)
	}

	counter := 0

	// Listen for messages
	fmt.Println("Listening for Hash Blocks (writer)...")
	for {
		select {
		case msg, ok := <-channeler.RecvChan:
			if !ok {
				log.Println("Failed to receive message")
				continue
			}
			if (counter % cfg.WriteIntervalBlock) != 0 {
				continue
			}
			if len(msg) != 3 {
				log.Println("Received message with unexpected number of parts")
				continue
			}

			// Split the message into topic, serialized transaction, and sequence number
			topic := string(msg[0])
			// serializedTx := msg[1]

			// Print out the parts
			fmt.Printf("Topic: %s\n", topic)
			// fmt.Printf("Serialized Transaction: %x\n", serializedTx) // Print as hex

			layerEdgeHeader, err := layerEdgeClient.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Println("Error getting layerEdgeHeader: ", err)
				continue
			}
			log.Println("Latest LayerEdge Block Hash:", layerEdgeHeader.Hash().Hex())

			hash, err := relayer.Write(
				cfg.PrivateKey.Signer,
				cfg.PrivateKey.Internal,
				[]byte(cfg.ProtocolId),
				[]byte(layerEdgeHeader.Hash().Hex()),
			)
			if err != nil {
				log.Println("Error writing -> ", err)
				continue
			}
			counter++
			log.Println("Relayer Write Done -> ", hash)
		}
	}
}
