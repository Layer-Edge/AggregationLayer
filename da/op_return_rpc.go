package da

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"bytes"
    "log"
    "time"
)

var (
BTCEndpoint = ""
User = "" 
Auth = ""
)

type response struct {
	Result json.RawMessage       `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID     string `json:"id"`
}

type utxo struct {
		Txid string `json:"txid"`
		Vout int `json:"vout"`
		Amount float64 `json:"amount"`	
}

type signedtx struct {
	Hex string `json:"hex"`
}

func Make_RPC_Call(url string, args []byte) string {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", BTCEndpoint, bytes.NewBuffer(args))
	if err != nil {
		log.Fatalf("Failed to create request to BTC: %v", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user", User)
	req.Header.Set("Authorization", "Basic " + Auth)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to send data to BTC: %v", err)
		return ""
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read BTC API response: %v", err)
		return ""
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("BTC API returned non-OK status: %d, %s", resp.StatusCode, out)
		return ""
	}
	log.Print("Successfully sent RPC: ", string(out))
	return ExtractResult(string(out))
}

func ListUnspent() string {
	payload := map[string]interface{} {
		"jsonrpc": "1.0",
		"id": "wallet_txn",
		"method": "listunspent",
		"params": []interface{} {
			 1,
			 9999999,
			 []interface{}{},
			 true,
			 map[string]int{ 
				"maximumCount": 10,
			 },
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal listunspent payload: %v", err)
		return ""
	}
	resp_str := Make_RPC_Call(BTCEndpoint, jsonPayload)
	return resp_str
}

func GetRawAddress() string {
	payload := map[string]interface{} {
		"jsonrpc": "1.0",
		"id": "wallet_address",
		"method": "getrawchangeaddress",
		"params": []interface {}{},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
	}
	return Make_RPC_Call(BTCEndpoint, jsonPayload)
}

func CalculateRequired(numInputs int, dataSize int) float64 {
	return float64(53 + numInputs*68 + dataSize) * float64(0.00000001)
}

func FilterUTXOs(unspent string, length int) ([]map[string]interface{}, float64) {
	inputs := []map[string]interface{}{}
	if unspent == "" {
		return inputs, 0.0
	}
	var t []json.RawMessage
	err := json.Unmarshal([]byte(unspent), &t)
	if err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return inputs, 0.0
	}
	totalAmt := 0.0
	numInputs := 0
	required := 0.0

	log.Printf("Found %d UTXOs to process", len(t))

	for numInputs < len(t) {
		var u utxo
		err := json.Unmarshal(t[numInputs], &u)
		if err != nil {
			log.Printf("Failed to unmarshal response: %v", err)
			return inputs, 0.0
		}else {
			log.Printf("UTXO : %s", u)
		}

		log.Printf("Processing UTXO: txid=%s, vout=%d, amount=%f", u.Txid, u.Vout, u.Amount)

		inputData := map[string]interface{}{
			"txid": u.Txid,
			"vout": u.Vout,
		}
		inputs = append(inputs, inputData)
		totalAmt += float64(u.Amount)
		required = CalculateRequired(numInputs+1, length)

		log.Printf("Current total: %f BTC, required: %f BTC", totalAmt, required)

		if totalAmt >= required {
			break
		}
		numInputs++
		if numInputs >= 10 {
			return []map[string]interface{}{}, 0.0
		}
	}
	return inputs, totalAmt - required
}

func CreateRawTransaction(inputs []map[string]interface{}, address string, change float64, data string) string {
	if len(inputs) == 0 {
		log.Printf("No inputs provided for transaction")
		return ""
	}
	
	if address == "" {
		log.Printf("Empty address provided")
		return ""
	}

	log.Printf("Creating raw transaction with %d inputs, change address %s, change amount %f", len(inputs), address, change)

	payload := map[string]interface{} {
		"jsonrpc": "1.0",
		"id": "op_cat_decode",
		"method": "createrawtransaction",
		"params": []interface{}{
			inputs,
			map[string]interface{}{
				"data": data, 
				address:change,
			},
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal create raw transaction payload: %v", err)
		return ""
	}
	return Make_RPC_Call(BTCEndpoint, jsonPayload)
}

func DecodeRawTransaction(rawtransaction string) {
	if rawtransaction == "" {
		log.Printf("Empty raw transaction provided")
	}

	payload := map[string]interface{} {
		"jsonrpc": "1.0",
		"id": "op_cat_decode",
		"method": "decoderawtransaction",
		"params": []string { 
			rawtransaction,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal decode raw transaction payload: %v", err)
	}
	Make_RPC_Call(BTCEndpoint, jsonPayload)
}

func SignRawTransaction(rawtransaction string) string {
	if rawtransaction == "" {
		log.Printf("Empty raw transaction provided")
		return ""
	}

	log.Printf("Signing raw transaction")

	payload := map[string]interface{} {
		"jsonrpc": "1.0",
		"id": "op_cat_sign_tx",
		"method": "signrawtransactionwithwallet",
		"params": []string { 
			rawtransaction,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal sign raw transaction with wallet payload: %v", err)
		return ""
	}
	return Make_RPC_Call(BTCEndpoint, jsonPayload)
}

func SendSignedTransaction(transaction string) string {
	if transaction == "" {
		log.Printf("Empty transaction provided")
		return ""
	}
	
	log.Printf("Sending signed transaction to network")

	payload := map[string]interface{} {
		"jsonrpc": "1.0",
		"id": "op_cat_send_tx",
		"method": "sendrawtransaction",
		"params": []string { 
			transaction,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal send raw transaction payload: %v", err)
		return ""
	}
	return Make_RPC_Call(BTCEndpoint, jsonPayload)
}

func ExtractResult(responseStr string) string {
	if responseStr == "" {
		log.Print("Empty response string")
		return ""
	}
	
	resp := response{}
	err := json.Unmarshal([]byte(responseStr), &resp)
	if err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return ""
	}
	
	if resp.Error != nil {
		log.Printf("RPC error: code=%d, message=%s", resp.Error.Code, resp.Error.Message)
		return ""
	}
	
	var resultStr string
	err = json.Unmarshal(resp.Result, &resultStr)
	if err == nil {
		return resultStr
	}
	
	return string(resp.Result)
}

func CreateOPReturnTransaction(data string) string {
	log.Printf("Creating OP_RETURN transaction with data of length %d", len(data))

	unspent := ListUnspent()
	inputs, change := FilterUTXOs(unspent, len(data))
	rawaddr := GetRawAddress()
	rawtscn := CreateRawTransaction(inputs, rawaddr, change, data)
	DecodeRawTransaction(rawtscn)
	signtscn := SignRawTransaction(rawtscn)
	var sgn signedtx
	err:= json.Unmarshal([]byte(signtscn), &sgn)
	if err != nil {
		log.Printf("Failed to unmarshal response: %v", err)
		return ""
	}
	sendtscn := SendSignedTransaction(sgn.Hex)

	log.Printf("Successfully created OP_RETURN transaction: %s", sendtscn)
	return sendtscn
}

func InitOPReturnRPC(endpoint string, user string, auth string) {
	BTCEndpoint = endpoint
	User = user
	Auth = auth
}
