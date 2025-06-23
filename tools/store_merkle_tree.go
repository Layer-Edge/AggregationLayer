package main

import (
	"fmt"
	"log"

	"github.com/Layer-Edge/bitcoin-da/clients"
	"github.com/Layer-Edge/bitcoin-da/config"
)

var cfg = config.GetConfig()

func main() {
	// Test merkle root (32-byte hash)
	merkleRoot := "0x1234567890123456789012345678901234567890123456789012345678901234"

	// Test leaves - multiple 32-byte hashes separated by commas
	leaves := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890,0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321,0x1111111111111111111111111111111111111111111111111111111111111111"

	fmt.Printf("Testing StoreMerkleTree function...\n")
	fmt.Printf("Merkle Root: %s\n", merkleRoot)
	fmt.Printf("Leaves: %s\n", leaves)

	err := clients.StoreMerkleTree(&cfg, merkleRoot, leaves)
	if err != nil {
		log.Fatalf("Error storing merkle tree: %v", err)
	}

	fmt.Println("Merkle tree stored successfully!")
}
