package utils

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// chunkSlice splits input slice into max chunkSize length slices
func ChunkSlice(slice []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// createTaprootAddress returns an address committing to a Taproot script with
// a single leaf containing the spend path with the script:
// <embedded data> OP_DROP <pubkey> OP_CHECKSIG
func CreateTaprootAddress(bobPrivateKey, internalPrivateKey string, embeddedData []byte) (string, error) {
	privKey, err := btcutil.DecodeWIF(bobPrivateKey)
	if err != nil {
		return "", fmt.Errorf("error decoding bob private key: %v", err)
	}

	pubKey := privKey.PrivKey.PubKey()
	// Print the pubkey
	fmt.Println("Public key:", hex.EncodeToString(pubKey.SerializeCompressed()))

	// Step 1: Construct the Taproot script with one leaf.
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0)
	builder.AddOp(txscript.OP_IF)
	chunks := ChunkSlice(embeddedData, 520)
	for _, chunk := range chunks {
		builder.AddData(chunk)
	}
	builder.AddOp(txscript.OP_ENDIF)
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	pkScript, err := builder.Script()
	if err != nil {
		return "", fmt.Errorf("error building script: %v", err)
	}

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

	internalPrivKey, err := btcutil.DecodeWIF(internalPrivateKey)
	if err != nil {
		return "", fmt.Errorf("error decoding internal private key: %v", err)
	}

	internalPubKey := internalPrivKey.PrivKey.PubKey()

	// Step 2: Generate the Taproot tree.
	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		internalPubKey, tapScriptRootHash[:],
	)

	// Step 3: Generate the Bech32m address.
	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(outputKey), &chaincfg.RegressionNetParams)
	if err != nil {
		return "", fmt.Errorf("error encoding Taproot address: %v", err)
	}

	return address.String(), nil
}

// payToTaprootScript creates a pk script for a pay-to-taproot output key.
func PayToTaprootScript(taprootKey *btcec.PublicKey) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_1).
		AddData(schnorr.SerializePubKey(taprootKey)).
		Script()
}

func ExtractPushData(version uint16, pkScript []byte) ([]byte, error) {
	type OpData struct {
		opcode     byte
		datalength byte
		data       []byte
	}
	result := OpData{pkScript[0], pkScript[1], pkScript[2:]}
	if result.opcode != txscript.OP_RETURN {
		return nil, nil
	}
	return result.data, nil
}

func FormatAmount(value *big.Int, decimals, places int) float64 {
	if value == nil {
		return 0
	}
	if decimals < 0 {
		decimals = 0
	}

	// Do the division at high precision first, then convert to float64.
	f := new(big.Float).SetPrec(256).SetInt(value)
	denInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	den := new(big.Float).SetPrec(256).SetInt(denInt)
	f.Quo(f, den)

	x, _ := f.Float64() // convert to double precision

	if places < 0 {
		return x
	}
	pow := math.Pow(10, float64(places))
	return math.Round(x*pow) / pow
}
