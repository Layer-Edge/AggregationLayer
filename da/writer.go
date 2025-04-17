package da

import (
	// "context"

	"github.com/ethereum/go-ethereum/ethclient"

	// "github.com/cosmos/cosmos-sdk/crypto/keyring"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Layer-Edge/bitcoin-da/clients"
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/models"
)

// To be set from Config
var (
	BtcCliPath     = ""
	BashScriptPath = ""
)

func CallScriptWithData(data string) ([]byte, error) {
	cmd := exec.Command(BashScriptPath+"/op_return_transaction.sh", data)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "BTC_CLI_PATH="+BtcCliPath)
	log.Println("Running BTC script", cmd)
	out, err := cmd.Output()
	return out, err
}

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
	btcReader := BlockSubscriber{channeler: nil}
	if btcReader.Subscribe(cfg.ZmqEndpointHashBlock, "hashblock") == false {
		return
	}

	dataReader := BlockSubscriber{channeler: nil}
	if dataReader.Replier(cfg.ZmqEndpointDataBlock) == false {
		return
	}

	err := models.InitDB(cfg.PostgresConnectionURI)

	if err != nil {
		log.Fatal("Error initializing DB Connection: ", err)
		return
	}

	BashScriptPath = cfg.BashScriptPath
	BtcCliPath = cfg.BtcCliPath

	defer btcReader.Reset()
	defer dataReader.Reset()

	layerEdgeClient, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		log.Fatal("Error creating layerEdgeClient: ", err)
	}

	InitOPReturnRPC(cfg.BtcEndpoint, cfg.User, cfg.Auth)

	counter := 0
	aggr := Aggregator{data: ""}
	prf := ZKProof{}
	proof_list := []string{}

	fnAgg := func(msg [][]byte) bool {
		log.Println("Aggregating message: ", string(msg[0]), string(msg[1]))
		aggr.Aggregate(msg[1])
		proof_list = append(proof_list, string(msg[1]))
		return true
	}

	fnBtc := func(msg [][]byte) ([]byte, error) {
		// Process
		hash, err := ProcessMsg(msg[1], cfg.ProtocolId, layerEdgeClient)
		return hash, err
	}

	fnWrite := func() {
		// Generate and process proof
		merkle_root := prf.GenerateAggregatedProof(aggr.data)
		log.Println("Aggregated Data: ", aggr.data)
		log.Println("Aggregated Proof: ", merkle_root)
		aggr.data = ""
		hash, err := btcReader.ProcessOutTuple(fnBtc, [][]byte{nil, []byte(merkle_root)})

		if err != nil {
			log.Println("Error writing -> ", err, "; out:", string(hash))
			return
		}
		log.Println("received btc_tx_hash: ", strings.ReplaceAll(string(hash[:]), "\n", ""))

		out, err := clients.SendCosmosTXWithData(string(merkle_root), "cosmos1c3y4q50cdyaa5mpfaa2k8rx33ydywl35hsvh0d")
		if err != nil {
			log.Fatalf("%v", err)
			return
		}

		btc_tx_hash := strings.ReplaceAll(string(hash[:]), "\n", "")
		cosmos_resp := clients.CosmosTxData{}

		err = json.Unmarshal(out, &cosmos_resp)
		if err != nil {
			log.Fatalf("Failed to parse cosmos response: %v", err)
			return
		}

		aggProof, err := models.CreateAggregatedProof(
			merkle_root,
			proof_list,
			btc_tx_hash,
			cosmos_resp,
		)
		proof_list = make([]string, 0)
		if err != nil {
			log.Fatalf("Failed to store Aggregated Proof in DB: %v", err)
		}

		log.Println("Stored Aggregated Proof: %v", aggProof)
	}

	// Listen for messages
	fmt.Println("Listening for Data Blocks and Hash Blocks (writer)...")
	for {
		select {
		case msg, ok := <-dataReader.channeler.RecvChan:
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
			// Write to Bitcoin
			if (counter % cfg.WriteIntervalBlock) == 0 {
				fnWrite()
			}
		case msg, ok := <-btcReader.channeler.RecvChan:
			log.Println("Received btc block")
			if !btcReader.Validate(ok, msg) {
				continue
			}
			// Write to Bitcoin
			fnWrite()
		}
	}
}
