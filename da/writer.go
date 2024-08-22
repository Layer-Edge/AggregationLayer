package da

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/zeromq/goczmq.v4"

	"github.com/Layer-Edge/bitcoin-da/config"
)

// To be set from Config
var (
	BtcCliPath = ""
	BashScriptPath = ""
)

func CallScriptWithData(data string) ([]byte, error) {
	cmd := exec.Command(BashScriptPath + "/op_return_transaction.sh", data)
	cmd.Env = os.Environ()
    cmd.Env = append(cmd.Env, "BTC_CLI_PATH=" + BtcCliPath)
	out,err := cmd.Output()
	return out, err
}

func ProcessMsg(msg [][]byte, protocolId string, layerEdgeClient *ethclient.Client) ([]byte, error) {
	// Split the message into topic, serialized transaction, and sequence number
	topic := string(msg[0])

	// Print out the parts
	fmt.Printf("Topic: %s\n", topic)

	layerEdgeHeader, err := layerEdgeClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Println("Error getting layerEdgeHeader: ", err)
		return nil, err
	}
	dhash := layerEdgeHeader.Hash()
	log.Println("Latest LayerEdge Block Hash:", dhash.Hex())

	data := append([]byte(protocolId), dhash.Bytes()...)
	hash, err := CallScriptWithData(hex.EncodeToString(data))
	return hash, err
}

func HashBlockSubscriber(cfg *config.Config) {
	// Init varaibles
	channeler := goczmq.NewSubChanneler(cfg.ZmqEndpoint, "hashblock")

	BashScriptPath = cfg.BashScriptPath
	BtcCliPath = cfg.BtcCliPath

	if channeler == nil {
		log.Fatal("Error creating channeler", channeler)
	}
	defer channeler.Destroy()

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
			// Process
			hash, err := ProcessMsg(msg, cfg.ProtocolId, layerEdgeClient)
			if err != nil {
				log.Println("Error writing -> ", err)
				continue
			}
			counter++
			log.Println("Relayer Write Done -> ", strings.ReplaceAll(string(hash[:]), "\n",""))
		}
	}
}
