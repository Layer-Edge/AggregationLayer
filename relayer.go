package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
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

// chunkSlice splits input slice into max chunkSize length slices
func chunkSlice(slice []byte, chunkSize int) [][]byte {
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
func createTaprootAddress(embeddedData []byte) (string, error) {
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
	chunks := chunkSlice(embeddedData, 520)
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
func payToTaprootScript(taprootKey *btcec.PublicKey) ([]byte, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_1).
		AddData(schnorr.SerializePubKey(taprootKey)).
		Script()
}

// Relayer is a bitcoin client wrapper which provides reader and writer methods
// to write binary blobs to the blockchain.
type Relayer struct {
	client *rpcclient.Client
}

// close shuts down the client.
func (r Relayer) close() {
	r.client.Shutdown()
}

// commitTx commits an output to the given taproot address, such that the
// output is only spendable by posting the embedded data on chain, as part of
// the script satisfying the tapscript spend path that commits to the data. It
// returns the hash of the commit transaction and error, if any.
func (r Relayer) commitTx(addr string) (*chainhash.Hash, error) {
	// Create a transaction that sends 0.001 BTC to the given address.
	address, err := btcutil.DecodeAddress(addr, &chaincfg.RegressionNetParams)
	if err != nil {
		return nil, fmt.Errorf("error decoding recipient address: %v", err)
	}

	amount, err := btcutil.NewAmount(0.001)
	if err != nil {
		return nil, fmt.Errorf("error creating new amount: %v", err)
	}

	hash, err := r.client.SendToAddress(address, amount)
	// Print address to send
	fmt.Println("Address to send:", address.EncodeAddress())
	if err != nil {
		return nil, fmt.Errorf("error sending to address: %v", err)
	}

	return hash, nil
}

// revealTx spends the output from the commit transaction and as part of the
// script satisfying the tapscript spend path, posts the embedded data on
// chain. It returns the hash of the reveal transaction and error, if any.
func (r Relayer) revealTx(embeddedData []byte, commitHash *chainhash.Hash) (*chainhash.Hash, error) {
	rawCommitTx, err := r.client.GetRawTransaction(commitHash)
	if err != nil {
		return nil, fmt.Errorf("error getting raw commit tx: %v", err)
	}

	// TODO: use a better way to find our output
	var commitIndex int
	var commitOutput *wire.TxOut
	for i, out := range rawCommitTx.MsgTx().TxOut {
		if out.Value == 100000 {
			commitIndex = i
			commitOutput = out
			break
		}
	}

	privKey, err := btcutil.DecodeWIF(bobPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding bob private key: %v", err)
	}

	pubKey := privKey.PrivKey.PubKey()

	internalPrivKey, err := btcutil.DecodeWIF(internalPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding internal private key: %v", err)
	}

	internalPubKey := internalPrivKey.PrivKey.PubKey()

	// Our script will be a simple <embedded-data> OP_DROP OP_CHECKSIG as the
	// sole leaf of a tapscript tree.
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0)
	builder.AddOp(txscript.OP_IF)
	chunks := chunkSlice(embeddedData, 520)
	for _, chunk := range chunks {
		builder.AddData(chunk)
	}
	builder.AddOp(txscript.OP_ENDIF)
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	pkScript, err := builder.Script()
	if err != nil {
		return nil, fmt.Errorf("error building script: %v", err)
	}

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(
		internalPubKey,
	)

	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		internalPubKey, tapScriptRootHash[:],
	)
	p2trScript, err := payToTaprootScript(outputKey)
	if err != nil {
		return nil, fmt.Errorf("error building p2tr script: %v", err)
	}

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  *rawCommitTx.Hash(),
			Index: uint32(commitIndex),
		},
	})
	txOut := &wire.TxOut{
		Value: 1e3, PkScript: p2trScript,
	}
	tx.AddTxOut(txOut)

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		commitOutput.PkScript,
		commitOutput.Value,
	)
	sigHashes := txscript.NewTxSigHashes(tx, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(
		tx, sigHashes, 0, txOut.Value,
		txOut.PkScript, tapLeaf, txscript.SigHashDefault,
		privKey.PrivKey,
	)

	if err != nil {
		return nil, fmt.Errorf("error signing tapscript: %v", err)
	}

	// Now that we have the sig, we'll make a valid witness
	// including the control block.
	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("error including control block: %v", err)
	}
	tx.TxIn[0].Witness = wire.TxWitness{
		sig, pkScript, ctrlBlockBytes,
	}

	hash, err := r.client.SendRawTransaction(tx, true)
	if err != nil {
		return nil, fmt.Errorf("error sending reveal transaction: %v", err)
	}
	return hash, nil
}

type Config struct {
	Host         string
	User         string
	Pass         string
	HTTPPostMode bool
	DisableTLS   bool
}

// NewRelayer returns a new relayer. It can error if there's an RPC connection
// error with the connection config.
func NewRelayer(config Config) (*Relayer, error) {
	// Set up the connection to the btcd RPC server.
	// NOTE: for testing bitcoind can be used in regtest with the following params -
	// bitcoind -chain=regtest -rpcport=18332 -rpcuser=rpcuser -rpcpassword=rpcpass -fallbackfee=0.000001 -txindex=1
	connCfg := &rpcclient.ConnConfig{
		Host:         config.Host,
		User:         config.User,
		Pass:         config.Pass,
		HTTPPostMode: config.HTTPPostMode,
		DisableTLS:   config.DisableTLS,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating btcd RPC client: %v", err)
	}
	return &Relayer{
		client: client,
	}, nil
}

func (r Relayer) ReadTransaction(hash *chainhash.Hash) ([]byte, error) {
	tx, err := r.client.GetRawTransaction(hash)
	if err != nil {
		return nil, err
	}
	if len(tx.MsgTx().TxIn[0].Witness) > 1 {
		witness := tx.MsgTx().TxIn[0].Witness[1]
		pushData, err := ExtractPushData(0, witness)
		if err != nil {
			return nil, err
		}
		// skip PROTOCOL_ID
		if pushData != nil && bytes.HasPrefix(pushData, PROTOCOL_ID) {
			return pushData[4:], nil
		}
	}
	return nil, nil
}

func (r Relayer) Read(height uint64) ([][]byte, error) {
	hash, err := r.client.GetBlockHash(int64(height))
	if err != nil {
		return nil, err
	}
	block, err := r.client.GetBlock(hash)
	if err != nil {
		return nil, err
	}

	var data [][]byte
	for _, tx := range block.Transactions {
		if len(tx.TxIn[0].Witness) > 1 {
			witness := tx.TxIn[0].Witness[1]
			pushData, err := ExtractPushData(0, witness)
			if err != nil {
				return nil, err
			}
			// skip PROTOCOL_ID
			if pushData != nil && bytes.HasPrefix(pushData, PROTOCOL_ID) {
				data = append(data, pushData[4:])
			}
		}
	}
	return data, nil
}

func (r Relayer) Write(data []byte) (*chainhash.Hash, error) {
	data = append(PROTOCOL_ID, data...)
	address, err := createTaprootAddress(data)
	if err != nil {
		return nil, err
	}
	hash, err := r.commitTx(address)
	if err != nil {
		return nil, err
	}
	hash, err = r.revealTx(data, hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func ExtractPushData(version uint16, pkScript []byte) ([]byte, error) {
	type templateMatch struct {
		expectPushData bool
		maxPushDatas   int
		opcode         byte
		extractedData  []byte
	}
	var template = [6]templateMatch{
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

func ExampleRelayer_Write() {
	// Example usage
	relayer, err := NewRelayer(Config{
		Host:         "localhost:18443",
		User:         "rpcuser",
		Pass:         "rpcpass",
		HTTPPostMode: true,
		DisableTLS:   true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Writing...")
	_, err = relayer.Write([]byte("rollkit-btc: gm"))
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
	relayer, err := NewRelayer(Config{
		Host:         "localhost:18443",
		User:         "rpcuser",
		Pass:         "rpcpass",
		HTTPPostMode: true,
		DisableTLS:   true,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = relayer.Write([]byte("rollkit-btc: gm"))
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: either mock or generate block
	// We're assuming the prev tx was mined at height 146

	height := uint64(146)
	blobs, err := relayer.Read(height)
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

func main() {
	// Call the ExampleRelayer_Write function to write data to the blockchain.
	ExampleRelayer_Write()
	// Call the ExampleRelayer_Read function to read data from the blockchain.
	// ExampleRelayer_Read()
}
