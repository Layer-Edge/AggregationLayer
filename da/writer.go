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
	"github.com/Layer-Edge/bitcoin-da/utils"
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

	// Initialize replier with retry
	if !dataReader.Replier(cfg.ZmqEndpointDataBlock) {
		log.Fatal("Failed to initialize replier after retries")
		return
	}

	counter := 0
	aggr := Aggregator{data: ""}
	prf := ZKProof{}
	proof_list := []string{}    // For database storage (hex-encoded proofs)
	merkle_leaves := []string{} // For merkle tree storage (proof hashes)
	// Initialize last_write to the current aligned boundary so the first write occurs at the next boundary
	writePeriod := time.Duration(cfg.WriteIntervalSeconds) * time.Second
	last_write := time.Now().Truncate(writePeriod).Unix()

	fnAgg := func(msg [][]byte) bool {
		log.Println("Aggregating message: ", string(msg[0]), "proof length:", len(msg[1]))

		// Store hex-encoded ABI proof for database
		hexProof := "0x" + hex.EncodeToString(msg[1])
		proof_list = append(proof_list, hexProof)

		// Use keccak256 hash of the ABI proof as leaf for merkle tree
		proofHash := utils.Keccak256Hash(msg[1])
		aggr.Aggregate(proofHash)

		// Store proof hash for merkle tree storage (without 0x prefix for contract)
		merkle_leaves = append(merkle_leaves, proofHash)

		log.Printf("Stored proof: %s, hash for merkle: %s", hexProof, proofHash)
		return true
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

		// Store merkle tree with retry mechanism
		txData, err := clients.StoreMerkleTree(cfg, cfg.LayerEdgeRPC.MerkleTreeStorageContract, merkle_root, merkle_leaves)
		if err != nil {
			log.Printf("Error storing merkle tree: %v", err)
			// Don't return, continue with database storage attempt
		}

		// Store in database with retry mechanism
		if txData != nil {
			aggProof, err := models.CreateAggregatedProof(
				merkle_root,
				proof_list,
				*txData,
			)
			proof_list = make([]string, 0)
			merkle_leaves = make([]string, 0)
			if err != nil {
				log.Printf("Failed to store Aggregated Proof in DB: %v", err)
				// Continue execution, don't crash
			} else {
				log.Printf("Stored Aggregated Proof: %v", aggProof)
			}
		} else {
			log.Println("No transaction data available, skipping database storage")
			proof_list = make([]string, 0)
			merkle_leaves = make([]string, 0)
		}
	}

	// Listen for messages with enhanced error handling
	fmt.Println("Listening for Data Blocks and Hash Blocks (writer)...")

	for {
		// Get message with timeout protection
		ok, msg := dataReader.GetMessage()
		if !ok {
			log.Println("Failed to receive message or channel closed")
			time.Sleep(1 * time.Second) // Brief pause before retry
			continue
		}

		log.Println("Received data for aggregation")
		if !dataReader.Validate(ok, msg) {
			log.Println("Message validation failed, skipping")
			continue
		}

		// Send acknowledgment with error handling
		select {
		case dataReader.channeler.SendChan <- [][]byte{[]byte("Data Received, will be pushed to next block")}:
			// Message sent successfully
		case <-time.After(5 * time.Second):
			log.Println("Warning: Could not send response message - timeout")
		default:
			log.Println("Warning: Could not send response message - channel full or closed")
		}

		// Process message with error handling
		counter++
		if !dataReader.Process(fnAgg, msg) {
			log.Println("Failed to process message, skipping")
			continue
		}

		// Write to LayerEdge chain with error handling
		nowTime := time.Now()
		alignedNow := nowTime.Truncate(writePeriod).Unix()
		if (counter%cfg.WriteIntervalBlock) == 0 || alignedNow > last_write {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic in fnWrite: %v", r)
					}
				}()
				fnWrite()
			}()
			// Record the boundary we just wrote for, so we only write once per aligned interval
			last_write = alignedNow
		}
	}
}
