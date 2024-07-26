package relayer

import (
	"bytes"
	"fmt"
	"crypto/md5"
	"log"
	"encoding/hex"
    "os/exec"
	"strings"
	"strconv"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcec/v2"
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


var (
		unspenttxid = "69840245a0044884ff3c67a069b5e6e80fc2964a8e53e62c96afc3f05862f481"
		unspentamt = 50.00
		unspentvout = 0
		customchangeaddress = "bcrt1qtz9kga4v6vkr0gpxfmnjjwps8s0ms9f2mjp850"
		unspentscriptpubkey = "00145ead7adba9d3f0fae63bf16574fc55d639a153df"
		cli_path = "/home/rishabh/bitcoin-cli/bitcoin-27.0/bin/bitcoin-cli"
)

func getFirstUnspentTransaction() (string, string, uint32, float64) {
	cmd_btc := exec.Command(cli_path, "listunspent")
	out, err := cmd_btc.Output()
	// log.Println(string(out))
	outstr := string(out)
	if err != nil {
		log.Fatal("Cannot get unspent", err)
	}
	cmd_btc2 := exec.Command(cli_path, "getrawchangeaddress")
	address, err := cmd_btc2.Output()
	if err != nil {
		log.Fatal("Cannot get address", err)
	}
	outstr = strings.ReplaceAll(outstr, "\n", "")
	cmd0 := "echo '" + outstr + "' | jq -r '.[0] | .vout'"
	// fmt.Println(cmd0)
	cmd_jq1 := exec.Command("bash", "-c", cmd0)
	vout, err := cmd_jq1.Output()
	if err != nil {
		log.Fatal("Cannot get vout ", vout, err)
	}
	cmd0 = "echo '" + outstr + "' | jq -r '.[0] | .txid'"
	cmd_jq2 := exec.Command("bash", "-c", cmd0)
	txid, err := cmd_jq2.Output()
	if err != nil {
		log.Fatal("Cannot get txid", err)
	}
	cmd0 = "echo '" + outstr + "' | jq -r '.[0] | .amount'"
	cmd_jq3 := exec.Command("bash", "-c", cmd0)
	amount, err := cmd_jq3.Output()
	log.Println("------------------ ", string(amount))
	if err != nil {
		log.Fatal("Cannot get amount", err)
	}
	vout_int, err := strconv.ParseUint(strings.ReplaceAll(string(vout), "\n", ""), 10, 32)
	amount_flt, err := strconv.ParseFloat(strings.ReplaceAll(string(amount), "\n", ""), 64)
	return string(address), string(txid), uint32(vout_int), amount_flt
}

func signTransaction(tx string) string {
	cmd_btc := exec.Command(cli_path, "signrawtransactionwithwallet", tx)
	out, err := cmd_btc.Output()
	log.Println(string(out))
	outstr := string(out)
	if err != nil {
		log.Fatal("Cannot get signature ", err)
	}
	return outstr
}

func sendTransaction(signedtx string) string {
	cmd_btc := exec.Command(cli_path, "sendrawtransaction", signedtx)
	out, err := cmd_btc.Output()
	log.Println(string(out))
	outstr := string(out)
	if err != nil {
		log.Fatal("Cannot send transaction ", err)
	}
	return outstr
}

func getFieldFromJson(json string, field string) string {
	cmd0 := "echo '" + json + "' | jq -r '." + field + "'"
	fmt.Println(cmd0)
	cmd_jq := exec.Command("bash", "-c", cmd0)
	out, err := cmd_jq.Output()
	if err != nil {
		log.Fatal("Cannot get " + field + " ", err)
	}
	return strings.ReplaceAll(string(out), "\n", "")
}

func buildTxOPRETURN(key *btcec.PrivateKey, changeAddress btcutil.Address, hash *chainhash.Hash, script []byte, data string) *wire.MsgTx {
	log.Println("Getting info")
	addr,txid,vout,amt := getFirstUnspentTransaction()
	addr = strings.ReplaceAll(addr, "\n", "")
	txid = strings.ReplaceAll(txid, "\n", "")
	log.Println(addr,txid,vout,amt)

	txhash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		log.Fatal("Failed to create hash key: ", err, "\n")
	}
	changeAddress, err = btcutil.DecodeAddress(addr, &chaincfg.RegressionNetParams)
	if err != nil {
		log.Fatal("Failed to decode address: ", err, "\n")
	}

	tx := wire.NewMsgTx(1 /*Version */)

	txin := wire.NewTxIn(wire.NewOutPoint(txhash, vout /*prevVOutIndex*/), []byte{}, nil)
	tx.AddTxIn(txin)

	b := txscript.NewScriptBuilder()
	b.AddOp(txscript.OP_RETURN)
	b.AddData([]byte(data))

	pkScript, err := txscript.PayToAddrScript(changeAddress)
	if err != nil {
		log.Fatal(err)
	}

	tx.AddTxOut(wire.NewTxOut(int64(100000000 * (float32(amt) - 0.0001)/*change*/), pkScript))
	scriptByte, err := b.Script()
	tx.AddTxOut(wire.NewTxOut(0, scriptByte))

	var txHex bytes.Buffer
	if err := tx.Serialize(&txHex); err != nil {
		log.Fatal(err)
	}
	hexString := hex.EncodeToString(txHex.Bytes())
	fmt.Printf("Raw TX: %s\n", hexString)
	signedTxJson := signTransaction(hexString)
	signedTxJson = strings.ReplaceAll(signedTxJson, "\n", "")
	fmt.Printf("Signed Tx: %s\n", signedTxJson)
	signedTxHex := getFieldFromJson(signedTxJson, "hex")
	sentTxHex := sendTransaction(signedTxHex)
	fmt.Printf("Sent Tx: %s\n", strings.ReplaceAll(sentTxHex, "\n", ""))
	return tx // signedTxHex.Bytes()
}

func verifyTransaction(tx *wire.MsgTx) bool {
	hashstr := tx.TxHash().String()
	cmd_btc := exec.Command(cli_path, "getrawtransaction", hashstr)
	out, err := cmd_btc.Output()
	outstr := strings.ReplaceAll(string(out), "\n", "")
	log.Println(string(outstr))
	if err != nil {
		log.Fatal("Cannot find transaction ", err)
	}
	cmd_btc = exec.Command(cli_path, "decoderawtransaction", outstr)
	out, err = cmd_btc.Output()
	log.Println(string(out))
	if err != nil {
		log.Fatal("Cannot find transaction ", err)
	}
	return true
}

func (r Relayer) Write(bobPrivateKey, internalPrivateKey string, PROTOCOL_ID []byte, data []byte) (*chainhash.Hash, error) {

	fmt.Printf("Keys: ", bobPrivateKey, internalPrivateKey, "\n")

	data = append(PROTOCOL_ID, data...)

	md5hash := md5.Sum(data)
	fmt.Printf("MD5 Hash: %x\n", md5hash)

	privKey, err := btcutil.DecodeWIF(bobPrivateKey)
	fmt.Printf("Key", privKey.PrivKey,"\n")
	if err != nil {
		log.Fatal("Failed to decode private key: ", err, "\n")
	}
	txhash, err := chainhash.NewHashFromStr(unspenttxid)
	if err != nil {
		log.Fatal("Failed to create hash key: ", err, "\n")
	}
	changeaddress, err := btcutil.DecodeAddress(customchangeaddress, &chaincfg.RegressionNetParams)
	if err != nil {
		log.Fatal("Failed to decode address: ", err, "\n")
	}
	sourcePKScript, err := hex.DecodeString(unspentscriptpubkey)
    if err != nil {
		log.Fatal("Failed to parse sourcePKS\n")
    }

	// Build Transaction and send it on the wire
	tx := buildTxOPRETURN(privKey.PrivKey, changeaddress, txhash, sourcePKScript, string(md5hash[:]))
	fmt.Printf("OP RETURN hex: %x" , tx, tx, "\n")

	// Verify transaction existence
	verifyTransaction(tx)
	ret := tx.TxHash()
	return &ret, nil
}
