package da

import (
	// "context"

	"github.com/ethereum/go-ethereum/ethclient"

	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	// "github.com/ethereum/go-ethereum/accounts/abi"

	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/Layer-Edge/bitcoin-da/clients"
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/models"
)

func ProcessMsg(msg []byte, protocolId string, layerEdgeClient *ethclient.Client) ([]byte, error) {
	// layerEdgeHeader, err := layerEdgeClient.HeaderByNumber(context.Background(), nil)
	// if err != nil {
	//     log.Println("Error getting layerEdgeHeader: ", err)
	//     return nil, err
	// }
	// dhash := layerEdgeHeader.Hash()
	// log.Println("Latest LayerEdge Block Hash:", dhash.Hex())

	data := append([]byte(protocolId), msg...)
	hash := CreateOPReturnTransaction(hex.EncodeToString(data))
	return []byte(hash), nil
}

func HashBlockSubscriber(cfg *config.Config) {
	// Init varaibles
	dataReader := BlockSubscriber{channeler: nil}
	if !dataReader.Replier(cfg.ZmqEndpointDataBlock) {
		return
	}

	err := models.InitDB(cfg.PostgresConnectionURI)

	if err != nil {
		log.Fatal("Error initializing DB Connection: ", err)
		return
	}

	defer dataReader.Reset()

	counter := 0
	aggr := Aggregator{data: ""}
	prf := ZKProof{}
	proof_list := []string{}
	last_write := time.Now().Unix()

	fnAgg := func(msg [][]byte) bool {
		log.Println("Aggregating message: ", string(msg[0]), string(msg[1]))
		aggr.Aggregate(msg[1])
		proof_list = append(proof_list, string(msg[1]))
		return true
	}

	fnWrite := func() {
		// Generate and process proof
		merkle_root := prf.GenerateAggregatedProof(aggr.data)
		log.Println("Aggregated Data: ", aggr.data)
		log.Println("Aggregated Proof: ", merkle_root)
		aggr.data = ""

		txData, err := clients.StoreMerkleTree(cfg, merkle_root, proof_list)
		if err != nil {
			log.Println("Error storing merkle  -> ", err)
			return
		}

		aggProof, err := models.CreateAggregatedProof(
			merkle_root,
			proof_list,
			txData.TransactionHash,
			*txData,
		)
		proof_list = make([]string, 0)
		if err != nil {
			log.Fatalf("Failed to store Aggregated Proof in DB: %v", err)
		}

		log.Printf("Stored Aggregated Proof: %v", aggProof)
	}

	// Listen for messages
	fmt.Println("Listening for Data Blocks and Hash Blocks (writer)...")
	for {
		msg, ok := <-dataReader.channeler.RecvChan
		log.Println("Received data for aggregation")
		if !dataReader.Validate(ok, msg) {
			continue
		}
		// Add error handling for SendChan operation
		select {
		case dataReader.channeler.SendChan <- [][]byte{[]byte("Data Received, will be pushed to next block")}:
			// Message sent successfully
		default:
			log.Println("Warning: Could not send response message - channel full or closed")
		}
		counter++
		dataReader.Process(fnAgg, msg)
		// Write to LayerEdge chain
		now := time.Now().Unix()
		if (counter%cfg.WriteIntervalBlock) == 0 || now-last_write > int64(cfg.WriteIntervalSeconds) {
			fnWrite()
			last_write = now
		}
	}
}
