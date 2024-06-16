package da

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/wire"
	"gopkg.in/zeromq/goczmq.v4"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/utils"
)

func RawBlockSubscriber(cfg *config.Config) {
	channeler := goczmq.NewSubChanneler(cfg.ZmqEndpoint, "rawblock")

	if channeler == nil {
		log.Fatal("Error creating channeler", channeler)
	}
	defer channeler.Destroy()

	// Listen for messages
	fmt.Println("Listening for Raw Blocks (reader)...")
	for {
		select {
		case msg, ok := <-channeler.RecvChan:
			if !ok {
				log.Println("Failed to receive message")
				continue
			}
			if len(msg) != 3 {
				log.Println("Received message with unexpected number of parts")
				continue
			}

			// Split the message into topic, serialized transaction, and sequence number
			topic := string(msg[0])
			serializedBlock := msg[1]

			// Print out the parts
			fmt.Printf("Topic: %s\n", topic)
			// fmt.Printf("Serialized Transaction: %x\n", serializedBlock) // Print as hex

			parsedBlock, err := parseBlock(serializedBlock)
			if err != nil {
				log.Printf("Failed to parse transaction: %v", err)
				continue
			}
			readPostedData(parsedBlock, []byte(cfg.ProtocolId))
		}
	}
}

func readPostedData(block *wire.MsgBlock, protocolId []byte) {
	var blobs [][]byte
	for _, tx := range block.Transactions {
		if len(tx.TxIn[0].Witness) > 1 {
			witness := tx.TxIn[0].Witness[1]
			pushData, err := utils.ExtractPushData(0, witness)
			if err != nil {
				log.Println("failed to extract push data", err)
			}
			// skip PROTOCOL_ID
			if pushData != nil && bytes.HasPrefix(pushData, protocolId) {
				blobs = append(blobs, pushData[:])
			}
		}
	}
	var data []string
	for _, blob := range blobs {
		got, err := hex.DecodeString(fmt.Sprintf("%x", blob))
		if err != nil {
			log.Fatal("Error decoding blob: ", err)
		}
		data = append(data, string(got))
	}

	log.Println("Relayer Read: ", data)
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
