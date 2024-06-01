package relayer

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	"github.com/Layer-Edge/bitcoin-da/utils"
)

type Config = rpcclient.ConnConfig

// Relayer is a bitcoin client wrapper which provides reader and writer methods
// to write binary blobs to the blockchain.
type Relayer struct {
	Client *rpcclient.Client
}

// NewRelayer returns a new relayer. It can error if there's an RPC connection
// error with the connection config.
func NewRelayer(connCfg Config, ntfnHandlers *rpcclient.NotificationHandlers) (*Relayer, error) {
	client, err := rpcclient.New(&connCfg, ntfnHandlers)
	if err != nil {
		return nil, fmt.Errorf("error creating btcd RPC client: %v", err)
	}
	return &Relayer{
		Client: client,
	}, nil
}

// close shuts down the client.
func (r Relayer) close() {
	r.Client.Shutdown()
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

	hash, err := r.Client.SendToAddress(address, amount)
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
func (r Relayer) revealTx(bobPrivateKey, internalPrivateKey string, embeddedData []byte, commitHash *chainhash.Hash) (*chainhash.Hash, error) {
	rawCommitTx, err := r.Client.GetRawTransaction(commitHash)
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
	chunks := utils.ChunkSlice(embeddedData, 520)
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
	p2trScript, err := utils.PayToTaprootScript(outputKey)
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

	hash, err := r.Client.SendRawTransaction(tx, false)
	if err != nil {
		return nil, fmt.Errorf("error sending reveal transaction: %v", err)
	}
	return hash, nil
}

func (r Relayer) ReadTransaction(PROTOCOL_ID []byte, hash *chainhash.Hash) ([]byte, error) {
	tx, err := r.Client.GetRawTransaction(hash)
	if err != nil {
		return nil, err
	}
	if len(tx.MsgTx().TxIn[0].Witness) > 1 {
		witness := tx.MsgTx().TxIn[0].Witness[1]
		pushData, err := utils.ExtractPushData(0, witness)
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

func (r Relayer) ReadFromTxns(PROTOCOL_ID []byte, txns []*btcutil.Tx) ([][]byte, error) {
	var data [][]byte
	for _, tx := range txns {
		if len(tx.MsgTx().TxIn[0].Witness) > 1 {
			witness := tx.MsgTx().TxIn[0].Witness[1]
			pushData, err := utils.ExtractPushData(0, witness)
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

func (r Relayer) Read(PROTOCOL_ID []byte, height int64) ([][]byte, error) {
	hash, err := r.Client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	block, err := r.Client.GetBlock(hash)
	if err != nil {
		return nil, err
	}

	var data [][]byte
	for _, tx := range block.Transactions {
		if len(tx.TxIn[0].Witness) > 1 {
			witness := tx.TxIn[0].Witness[1]
			pushData, err := utils.ExtractPushData(0, witness)
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

func (r Relayer) Write(bobPrivateKey, internalPrivateKey string, PROTOCOL_ID []byte, data []byte) (*chainhash.Hash, error) {
	data = append(PROTOCOL_ID, data...)
	address, err := utils.CreateTaprootAddress(bobPrivateKey, internalPrivateKey, data)
	if err != nil {
		return nil, err
	}
	hash, err := r.commitTx(address)
	if err != nil {
		return nil, err
	}
	hash, err = r.revealTx(bobPrivateKey, internalPrivateKey, data, hash)
	if err != nil {
		return nil, err
	}
	return hash, nil
}
