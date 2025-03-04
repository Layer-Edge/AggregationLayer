package da

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"os/exec"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Configuration constants
var (
	CosmosContractAddr = ""
	CosmosNodeAddr = ""
	CosmosChainId  = ""
	CosmosKeyring = ""
	CosmosSrc = ""
)

func InitCosmosParams(contractAddr string, node string, id string, keyring string, from string) {
	CosmosContractAddr = contractAddr
	CosmosNodeAddr = node
	CosmosChainId  = id
	CosmosKeyring = keyring
	CosmosSrc = from
}

func CallContractStoreMerkleTree(btc_tx_hash string, root string, leaves string) (bool, string) {
	leaves_arr_str := strings.Join(strings.Split(leaves, ","), "\",\"")
	jsonMsg := fmt.Sprintf(`{"store_merkle_tree":{"id":"%s","root":"%s","leaves":["%s"],"metadata":""}}`, btc_tx_hash, root, leaves_arr_str)

	fmt.Println("%s, %s, %s, %s, %s, %s", CosmosContractAddr, jsonMsg, CosmosSrc, CosmosKeyring, CosmosNodeAddr, CosmosChainId)
	cmd := exec.Command("gaiad", "tx", "wasm", "execute", CosmosContractAddr, jsonMsg,
		"--from", CosmosSrc,
		"--keyring-backend", CosmosKeyring,
		"--gas", "400000",
		"--node", CosmosNodeAddr,
		"--chain-id", CosmosChainId,
		"-y",
	)

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return false, ""
	}
	hash := strings.Split(string(out), "txhash: ")
	fmt.Println(hash[0], "txhash:", hash[1])
	return true, hash[1]
}
