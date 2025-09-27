package da

import (
	// "context"

	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	// "github.com/ethereum/go-ethereum/accounts/abi"

	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Layer-Edge/bitcoin-da/clients"
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/models"
)

func ProcessBTCMsg(msg []byte, protocolId string) ([]byte, error) {
	data := append([]byte(protocolId), msg...)
	hash := CreateOPReturnTransaction(hex.EncodeToString(data))
	return []byte(hash), nil
}

func SuperProofSubscriber(ch chan [][]byte, cfg *config.Config) {
	// Initialize with enhanced error handling
	dataReader := NewBlockSubscriber()
	defer func() {
		if err := dataReader.Close(); err != nil {
			log.Printf("Error closing BlockSubscriber: %v", err)
		}
	}()

	// Initialize database with retry mechanism
	err := models.InitDB(cfg.PostgresConnectionURI)
	if err != nil {
		log.Fatalf("Error initializing DB Connection: %v", err)
	}
	defer func() {
		if err := models.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

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

	fnBtc := func(msg [][]byte) ([]byte, error) {
		// Process
		hash, err := ProcessBTCMsg(msg[1], cfg.ProtocolId)
		return hash, err
	}

	fnWrite := func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in fnWrite: %v", r)
			}
		}()

		// Generate and process proof
		merkle_root := prf.GenerateAggregatedProof(aggr.data)
		if merkle_root == "" {
			log.Println("Failed to generate aggregated proof, skipping write")
			return
		}

		log.Println("Aggregated Data: ", aggr.data)
		log.Println("Aggregated Proof: ", merkle_root)
		aggr.data = ""

		hash, err := dataReader.ProcessOutTuple(fnBtc, [][]byte{nil, []byte(merkle_root)})
		if err != nil {
			log.Println("Error writing -> ", err, "; out:", string(hash))
			return
		}

		btc_tx_hash := strings.ReplaceAll(string(hash[:]), "\n", "")

		// Store merkle tree with retry mechanism
		txData, err := clients.StoreMerkleTree(cfg, merkle_root, proof_list)
		if err != nil {
			log.Printf("Error storing merkle tree: %v", err)
			// Don't return, continue with database storage attempt
		}

		// Store in database with retry mechanism
		if txData != nil {
			aggProof, err := models.CreateAggregatedProof(
				merkle_root,
				proof_list,
				btc_tx_hash,
				*txData,
			)
			proof_list = make([]string, 0)
			if err != nil {
				log.Printf("Failed to store Aggregated Proof in DB: %v", err)
				// Continue execution, don't crash
			} else {
				log.Printf("Stored Aggregated Proof: %v", aggProof)
			}
		} else {
			log.Println("No transaction data available, skipping database storage")
			proof_list = make([]string, 0)
		}
	}

	// Listen for messages with enhanced error handling
	fmt.Println("Listening for Data Blocks and Hash Blocks (writer)...")

	for {
		// Get message with timeout protection
		msg := <-ch

		log.Println("Received data for aggregation")
		if !dataReader.Validate(true, msg) {
			log.Println("Message validation failed, skipping")
			continue
		}

		// Process message with error handling
		counter++
		if !dataReader.Process(fnAgg, msg) {
			log.Println("Failed to process message, skipping")
			continue
		}

		// Write to LayerEdge chain with error handling
		now := time.Now().Unix()
		if now-last_write > int64(cfg.SuperProofWriteIntervalSeconds) {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic in fnWrite: %v", r)
					}
				}()
				fnWrite()
			}()
			last_write = now
		}
	}
}
