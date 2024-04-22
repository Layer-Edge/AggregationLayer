package utils

import (
	"encoding/hex"
	"fmt"

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
	type templateMatch struct {
		expectPushData bool
		maxPushDatas   int
		opcode         byte
		extractedData  []byte
	}
	template := [6]templateMatch{
		{opcode: txscript.OP_FALSE},
		{opcode: txscript.OP_IF},
		{expectPushData: true, maxPushDatas: 10},
		{opcode: txscript.OP_ENDIF},
		{expectPushData: true, maxPushDatas: 1},
		{opcode: txscript.OP_CHECKSIG},
	}

	var templateOffset int
	tokenizer := txscript.MakeScriptTokenizer(version, pkScript)
out:
	for tokenizer.Next() {
		// Not a rollkit script if it has more opcodes than expected in the
		// template.
		if templateOffset >= len(template) {
			return nil, nil
		}

		op := tokenizer.Opcode()
		tplEntry := &template[templateOffset]
		if tplEntry.expectPushData {
			for i := 0; i < tplEntry.maxPushDatas; i++ {
				data := tokenizer.Data()
				if data == nil {
					break out
				}
				tplEntry.extractedData = append(tplEntry.extractedData, data...)
				tokenizer.Next()
			}
		} else if op != tplEntry.opcode {
			return nil, nil
		}

		templateOffset++
	}
	// TODO: skipping err checks
	return template[2].extractedData, nil
}
