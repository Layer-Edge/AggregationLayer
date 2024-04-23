package main

import (
	"encoding/hex"
	"fmt"

	"github.com/Layer-Edge/bitcoin-da/relayer"
)

func ExampleRelayer_Write() {
	// Example usage
	relayer, err := relayer.NewRelayer(ExampleConfig)
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
	relayer, err := relayer.NewRelayer(ExampleConfig)
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

	height := uint64(146)
	blobs, err := relayer.Read(PROTOCOL_ID, height)
	// Print the blobs
	if err != nil {
		fmt.Println(err)
		return
	}
	// Print the length of blobs
	fmt.Println(len(blobs))

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
