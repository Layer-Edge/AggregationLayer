package relayer_test

import (
	"encoding/hex"
	"fmt"

	bitcoinda "github.com/Layer-Edge/bitcoin-da/relayer"
)

// PROTOCOL_ID allows data identification by looking at the first few bytes
var PROTOCOL_ID = []byte{0x6C, 0x61, 0x79, 0x65, 0x72, 0x65, 0x64, 0x67, 0x65}

// Sample data and keys for testing.
// bob key pair is used for signing reveal tx
// internal key pair is used for tweaking
var (
	bobPrivateKey      = "cPbxEJ3UTLAeKzebFy6G38Qr7X5UqjcWv93PkhPJ52hoy9RtNkKD"
	internalPrivateKey = "cNR4CfUPBZNEZE9rShP4ix2NRPUNFfmDjecG7W9ySpupjGTMUKbw"
)

var ExampleConfig = bitcoinda.Config{
	Host:         "localhost:18443",
	User:         "jeet",
	Pass:         "SzKyQMucjU9pd6om64xcuMiEp4FqDtKAn_Q6QA16e6k",
	HTTPPostMode: true,
	DisableTLS:   true,
}

// ExampleRelayer_Write tests that writing data to the blockchain works as
// expected.
func ExampleRelayer_Write() {
	// Example usage
	relayer, err := bitcoinda.NewRelayer(ExampleConfig, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Writing...")
	_, err = relayer.Write(bobPrivateKey, internalPrivateKey, PROTOCOL_ID, []byte("rollkit-btc: gm"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("done")
	// Output: Writing...
	// done
}

// ExampleRelayer_Read tests that reading data from the blockchain works as
// expected.
func ExampleRelayer_Read() {
	// Example usage
	relayer, err := bitcoinda.NewRelayer(ExampleConfig, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = relayer.Write(bobPrivateKey, internalPrivateKey, PROTOCOL_ID, []byte("rollkit-btc: gm"))
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: either mock or generate block
	// We're assuming the prev tx was mined at height 146
	blobs, err := relayer.Read(PROTOCOL_ID, 146)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, blob := range blobs {
		got, err := hex.DecodeString(fmt.Sprintf("%x", blob))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(got))
	}
	// Output: rollkit-btc: gm
}
