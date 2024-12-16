package da

import (
	// "context"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/ethclient"

	// "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Layer-Edge/bitcoin-da/config"
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
	hash, err := CallScriptWithData(hex.EncodeToString(data))
	return hash, err
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

	mongoc, err := NewMongoSender(
		cfg.Mongo.Endpoint,   // your MongoDB URI
		cfg.Mongo.DB,         // your database name
		cfg.Mongo.Collection, // your collection name
	)
	if err != nil {
		log.Fatal(err)
	}
	// err := mongoc.Init(cfg)
	if err != nil {
		log.Fatal("Error creating Mongo Client: ", err)
		return
	}

	// client := &CosmosClient{}

	BashScriptPath = cfg.BashScriptPath
	BtcCliPath = cfg.BtcCliPath

	defer btcReader.Reset()
	defer dataReader.Reset()

	layerEdgeClient, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
	if err != nil {
		log.Fatal("Error creating layerEdgeClient: ", err)
	}

	counter := 0
	aggr := Aggregator{data: nil}
	prf := ZKProof{}
	lst := make([]map[string]string, 0)

	fnAgg := func(msg [][]byte) bool {
		log.Println("Aggregating message: ", string(msg[0]), string(msg[1]))
		aggr.Aggregate(msg[1])
		m := map[string]string{"length": string(len(msg[1])), "data": string(msg[1])}
		lst = append(lst, m)
		return true
	}

	fnBtc := func(msg [][]byte) ([]byte, error) {
		// Process
		return ProcessMsg(msg[1], cfg.ProtocolId, layerEdgeClient)
	}

	fnWrite := func(msg []byte) {
		// Generate and process proof
		prf := prf.GenerateAggregatedProof(aggr.data)
		log.Println("Aggregated Data: ", aggr.data)
		log.Println("Aggregated Proof: ", prf)
		aggr.data = nil

		payload := map[string]string{
			"recipient": "cosmos1c3y4q50cdyaa5mpfaa2k8rx33ydywl35hsvh0d",
			"memo":      string(prf[:]),
		}

		// Convert payload to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
			return
		}

		// Create HTTP client
		httpClient := &http.Client{
			Timeout: 10 * time.Second,
		}

		// API endpoint
		apiURL := "https://cosmos-api-hcf6.onrender.com/send-tokens"

		// Create HTTP request
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Fatalf("Failed to create request to Cosmos: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Send the request
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatalf("Failed to send data to Cosmos: %v", err)
			return
		}
		defer resp.Body.Close()

		// Read response body
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read Cosmos API response: %v", err)
			return
		}
		fmt.Print(1)
		// Check response status
		if resp.StatusCode != http.StatusOK {
			log.Print("Cosmos API returned non-OK status: %d", resp.StatusCode)
			// return
		}
		// fmt.Print(out)
		var dat map[string]interface{}
		if err := json.Unmarshal(out, &dat); err != nil {
			panic(err)
		}
		dat["proofs"] = lst
		lst = make([]map[string]string, 0)

		hash, err := fnBtc([][]byte{nil, prf[:]})

		if err != nil {
			log.Println("Error writing -> ", err, "; out:", string(hash))
			return
		}
		dat["btc_tx_hash"] = strings.ReplaceAll(string(hash[:]), "\n", "")
		out, err = json.Marshal(dat)
		log.Print("Sending proof info to mongo:", string(out))
		// Send data to Mongo
		err = mongoc.SendData(out)
		if err != nil {
			log.Fatalf("Failed to send data to Mongo: %v", err)
			return
		}
		// if !btcReader.Process(fnBtc, [][]byte{nil, prf[:]}) {
		// 	log.Println("Failed to write proof")
		// }
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
			dataReader.channeler.SendChan <- [][]byte{[]byte("Data Received, will be pushed to next block")}
			counter++
			dataReader.Process(fnAgg, msg)
			// Write to Bitcoin
			if (counter % cfg.WriteIntervalBlock) == 0 {
				fnWrite(aggr.data)
			}
		case msg, ok := <-btcReader.channeler.RecvChan:
			log.Println("Received btc block")
			if !btcReader.Validate(ok, msg) {
				continue
			}
			// Write to Bitcoin
			fnWrite(aggr.data)
		}
	}
}
