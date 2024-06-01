package main

import (
	"encoding/hex"
	"fmt"

	"github.com/Layer-Edge/bitcoin-da/relayer"
)

var ExampleConfig = relayer.Config{
	Host:         cfg.WalletRelayer.Host,
	User:         cfg.WalletRelayer.User,
	Pass:         cfg.WalletRelayer.Pass,
	HTTPPostMode: true,
	DisableTLS:   true,
}

var (
	bobPrivateKey      = cfg.PrivateKey.Signer
	internalPrivateKey = cfg.PrivateKey.Internal
)

func ExampleRelayer_Write(data string) {
	// Example usage
	relayer, err := relayer.NewRelayer(ExampleConfig, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Writing...")
	_, err = relayer.Write(bobPrivateKey, internalPrivateKey, PROTOCOL_ID, []byte(data))
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
func ExampleRelayer_Read(data string) {
	// Example usage
	relayer, err := relayer.NewRelayer(ExampleConfig, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = relayer.Write(bobPrivateKey, internalPrivateKey, PROTOCOL_ID, []byte(data))
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: either mock or generate block
	// We're assuming the prev tx was mined at height 146
	blobs, err := relayer.Read(PROTOCOL_ID, 146)
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
