package main

import (
	"github.com/Layer-Edge/bitcoin-da/reader"
	"github.com/Layer-Edge/bitcoin-da/relayer"
)

// PROTOCOL_ID allows data identification by looking at the first few bytes
var PROTOCOL_ID = []byte{0x72, 0x6f, 0x6c, 0x6c}

// Sample data and keys for testing.
// bob key pair is used for signing reveal tx
// internal key pair is used for tweaking
var (
	bobPrivateKey      = "cPbxEJ3UTLAeKzebFy6G38Qr7X5UqjcWv93PkhPJ52hoy9RtNkKD"
	internalPrivateKey = "cNR4CfUPBZNEZE9rShP4ix2NRPUNFfmDjecG7W9ySpupjGTMUKbw"
)

var LayerEdgeRPC = ""

var ExampleConfig = relayer.Config{
	Host:         "localhost:18443",
	User:         "jeet",
	Pass:         "SzKyQMucjU9pd6om64xcuMiEp4FqDtKAn_Q6QA16e6k",
	HTTPPostMode: true,
	DisableTLS:   true,
}

func main() {
	// Call the ExampleRelayer_Write function to write data to the blockchain.
	// ExampleRelayer_Write()
	// Call the ExampleRelayer_Read function to read data from the blockchain.
	// ExampleRelayer_Read()
	reader.SubscribeToBlocks()
}
