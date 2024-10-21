package da

import (
    "bytes"
    // "encoding/hex"
    "fmt"
    "log"

    "github.com/btcsuite/btcd/wire"

    "github.com/Layer-Edge/bitcoin-da/config"
    "github.com/Layer-Edge/bitcoin-da/utils"
)

func RawBlockSubscriber(cfg *config.Config) {
    // channelReader := ZmqChannelReader{channeler : nil}
    channelReader := BlockSubscriber{channeler : nil}
    processor := BitcoinBlockProcessor{}

    if channelReader.Subscribe(cfg.ZmqEndpointRawBlock, "rawblock") == false {
        return
    }

    defer channelReader.Reset()

    fn := func(msg [][]byte) bool {
        log.Println("Processing message")
        return processor.process(msg, cfg.ProtocolId)
    }
    // Listen for messages
    fmt.Println("Listening for Raw Blocks (reader) from ZMQ channel...", cfg.ZmqEndpointRawBlock)
    for {
            msg, ok := <-channelReader.channeler.RecvChan
            if !channelReader.Validate(ok, msg) {
                continue
            }
            channelReader.Process(fn, msg)
    }
}

type RawBlockProcessor interface
{
    // validate() bool
    process(data [][]byte, protocolId string) bool
    // printBlock()
    // printTransaction()
}

type BitcoinBlockProcessor struct
{
}


func (btcProc BitcoinBlockProcessor) process(msg [][]byte, protocolId string) bool{
    // Split the message into topic, serialized transaction, and sequence number
    topic := string(msg[0])
    serializedBlock := msg[1]

    // Print out the parts
    fmt.Printf("Topic: %s\n", topic)
    fmt.Printf("Serialized block: %x\n", serializedBlock)
    // fmt.Printf("Serialized Transaction: %x\n", serializedBlock) // Print as hex
    parsedBlock, err := parseBlock(serializedBlock)
    if err != nil {
        log.Printf("Failed to parse transaction: %v", err)
        return false
    }
    printBlock(parsedBlock)
    readPostedData(parsedBlock, []byte(protocolId))
    return true
}

// parseBlock parses a serialized Bitcoin block
func parseBlock(data []byte) (*wire.MsgBlock, error) {
    var block wire.MsgBlock
    err := block.Deserialize(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    return &block, nil
}

func readPostedData(block *wire.MsgBlock, protocolId []byte) {
    var blobs [][]byte
    for _, tx := range block.Transactions {
        for _, txout := range tx.TxOut {
            pushData, err := utils.ExtractPushData(1, txout.PkScript)
            if err != nil {
                log.Println("failed to extract push data", err)
            }
            log.Println("Data: ", pushData)
            if pushData != nil && bytes.HasPrefix(pushData, protocolId) {
                blobs = append(blobs, pushData[:])
            }
        }
    }
    var data []string
    for _, blob := range blobs {
        data = append(data, fmt.Sprintf("%s:%x", blob[:len(protocolId)], blob[len(protocolId):]))
    }

    log.Println("Relayer Read: ", data)
}

// printBlock prints the details of a Bitcoin block
func printBlock(block *wire.MsgBlock) {
    fmt.Println("Block Details:")
    fmt.Printf("  Block Header:\n")
    fmt.Printf("    Version: %d\n", block.Header.Version)
    fmt.Printf("    Previous Block: %s\n", block.Header.PrevBlock)
    fmt.Printf("    Merkle Root: %s\n", block.Header.MerkleRoot)
    fmt.Printf("    Timestamp: %s\n", block.Header.Timestamp)
    fmt.Printf("    Bits: %d\n", block.Header.Bits)
    fmt.Printf("    Nonce: %d\n", block.Header.Nonce)

    fmt.Println("  Transactions:")
    for i, tx := range block.Transactions {
        fmt.Printf("    Transaction #%d:\n", i+1)
        printTransaction(tx)
    }
    fmt.Println("  Block Height: [unknown]")
}

// printTransaction prints the details of a Bitcoin transaction
func printTransaction(tx *wire.MsgTx) {
    fmt.Println("Transaction Details:")
    fmt.Printf("  Version: %d\n", tx.Version)
    fmt.Printf("  LockTime: %d\n", tx.LockTime)

    fmt.Println("  Inputs:")
    for i, txIn := range tx.TxIn {
        fmt.Printf("    Input #%d:\n", i+1)
        fmt.Printf("      Previous Outpoint: %s\n", txIn.PreviousOutPoint)
        fmt.Printf("      Signature Script: %x\n", txIn.SignatureScript)
        fmt.Printf("      Sequence: %d\n", txIn.Sequence)
    }

    fmt.Println("  Outputs:")
    for i, txOut := range tx.TxOut {
        fmt.Printf("    Output #%d:\n", i+1)
        fmt.Printf("      Value: %d\n", txOut.Value)
        fmt.Printf("      PkScript: %x\n", txOut.PkScript)
    }
}
