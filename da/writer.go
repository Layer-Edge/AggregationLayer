package da

import (
    "context"
    "encoding/hex"
    "fmt"
    "github.com/ethereum/go-ethereum/ethclient"
    // "github.com/cosmos/cosmos-sdk/crypto/keyring"
    "log"
    "os"
    "os/exec"
    "strings"

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
    out, err := cmd.Output()
    return out, err
}

func ProcessMsg(msg []byte, protocolId string, layerEdgeClient *ethclient.Client) ([]byte, error) {
    // // Split the message into topic, serialized transaction, and sequence number
    // topic := string(msg[0])

    // // Print out the parts
    // fmt.Printf("Topic: %s\n", topic)

    layerEdgeHeader, err := layerEdgeClient.HeaderByNumber(context.Background(), nil)
    // if err != nil {
    //     log.Println("Error getting layerEdgeHeader: ", err)
    //     return nil, err
    // }
    dhash := layerEdgeHeader.Hash()
    log.Println("Latest LayerEdge Block Hash:", dhash.Hex())

    // data := append([]byte(protocolId), dhash.Bytes()...)
    data := append([]byte(protocolId), msg...)
    hash, err := CallScriptWithData(hex.EncodeToString(data))
    return hash, err
}

func HashBlockSubscriber(cfg *config.Config) {
    // Init varaibles
    // processor := BitcoinBlockProcessor{}
    btcReader := BlockSubscriber{channeler : nil}
    if btcReader.Subscribe(cfg.ZmqEndpointHashBlock, "hashblock") == false {
        return
    }

    dataReader := BlockSubscriber{channeler : nil}
    if dataReader.Subscribe(cfg.ZmqEndpointDataBlock, "datablock") == false {
        return
    }

    //coscfg := CosmosClientConfig{
    //    ChainID:        config.GetConfig().Cosmos.ChainID,
    //    RPCEndpoint:    config.GetConfig().Cosmos.RpcEndpoint,
    //    AccountPrefix:  config.GetConfig().Cosmos.AccountPrefix,
    //    KeyringBackend: keyring.BackendTest,
    //    KeyName:        "layeredge", // Replace with actual key name
    //    HomeDir:        "~/repo/bitcoin-da",      // Replace with actual home directory
    //}

    client := &CosmosClient{}
    // err := client.Init(coscfg)
    // if err != nil {
    //     log.Fatalf("Failed to initialize Cosmos client: %v", err)
    // }

    BashScriptPath = cfg.BashScriptPath
    BtcCliPath = cfg.BtcCliPath

    defer btcReader.Reset()
    defer dataReader.Reset()

    layerEdgeClient, err := ethclient.Dial(cfg.LayerEdgeRPC.HTTP)
    if err != nil {
        log.Fatal("Error creating layerEdgeClient: ", err)
    }

    counter := 0
    aggr := Aggregator{data : nil}
    prf := ZKProof{}

    fnAgg := func(msg [][]byte) bool {
        log.Println("Aggregating message: ", string(msg[0]), string(msg[1]))
        aggr.Aggregate(msg[1])
        return true
    }

    fnBtc := func(msg [][]byte) bool {
        // if len(msg) != 3 {
        //     log.Println("Received message with unexpected number of parts")
        //     return false
        // }
        // Process
        hash, err := ProcessMsg(msg[1], cfg.ProtocolId, layerEdgeClient)
        if err != nil {
            log.Println("Error writing -> ", err)
            return false
        }
        log.Println("Relayer Write Done -> ", strings.ReplaceAll(string(hash[:]), "\n", ""))
        return true
    }

    fnWrite := func(msg []byte) {
        err = client.Send(string(aggr.data[:]), config.GetConfig().Cosmos.RpcEndpoint)
        if err != nil {
            log.Fatalf("Failed to send data: %v", err)
            return
        }

        prf := prf.GenerateAggregatedProof(aggr.data)
        aggr.data = nil
        log.Println("Aggregated Proof: ", prf)
        if !btcReader.Process(fnBtc, [][]byte{nil,prf}) {
            log.Println("Failed to write proof")
        }
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
