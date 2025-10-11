package da

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Layer-Edge/bitcoin-da/utils"
)

var (
	BTCEndpoint      = ""
	Auth             = ""
	WalletPassphrase = ""

	// RPC configuration
	maxRetries     = 3
	baseDelay      = 1 * time.Second
	maxDelay       = 30 * time.Second
	backoffFactor  = 2.0
	requestTimeout = 30 * time.Second

	// Circuit breaker for RPC calls
	rpcMutex       sync.RWMutex
	failureCount   int
	lastFailTime   time.Time
	circuitOpen    bool
	circuitTimeout = 60 * time.Second
	maxFailures    = 5
)

type response struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	ID string `json:"id"`
}

type utxo struct {
	Txid   string  `json:"txid"`
	Vout   int     `json:"vout"`
	Amount float64 `json:"amount"`
}

type signedtx struct {
	Hex string `json:"hex"`
}

// RPCCircuitBreaker manages the circuit breaker state for RPC calls
type RPCCircuitBreaker struct {
	mutex        sync.RWMutex
	failureCount int
	lastFailTime time.Time
	circuitOpen  bool
}

// CanExecute checks if the circuit breaker allows execution
func (cb *RPCCircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if !cb.circuitOpen {
		return true
	}

	// Check if enough time has passed to try again
	return time.Since(cb.lastFailTime) > circuitTimeout
}

// RecordSuccess records a successful operation
func (cb *RPCCircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.circuitOpen = false
}

// RecordFailure records a failed operation
func (cb *RPCCircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.failureCount >= maxFailures {
		cb.circuitOpen = true
		log.Printf("RPC circuit breaker opened due to %d failures", cb.failureCount)
	}
}

var rpcCircuitBreaker = &RPCCircuitBreaker{}

// RetryRPCCall executes an RPC call with exponential backoff retry
func RetryRPCCall(operation func() (string, error)) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if !rpcCircuitBreaker.CanExecute() {
			return "", fmt.Errorf("RPC circuit breaker is open, operation rejected")
		}

		if attempt > 0 {
			delay := time.Duration(float64(baseDelay) *
				utils.PowFloat(backoffFactor, float64(attempt-1)))
			if delay > maxDelay {
				delay = maxDelay
			}

			log.Printf("Retrying RPC call after %v delay (attempt %d/%d)",
				delay, attempt+1, maxRetries+1)

			time.Sleep(delay)
		}

		result, err := operation()
		if err == nil {
			rpcCircuitBreaker.RecordSuccess()
			return result, nil
		}

		lastErr = err
		rpcCircuitBreaker.RecordFailure()
		log.Printf("RPC call failed (attempt %d/%d): %v",
			attempt+1, maxRetries+1, err)
	}

	return "", fmt.Errorf("RPC call failed after %d attempts: %w", maxRetries+1, lastErr)
}

func Make_RPC_Call(url string, args []byte) string {
	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(url, args)
	})

	if err != nil {
		log.Printf("RPC call failed after retries: %v", err)
		return ""
	}

	return result
}

// makeRPCCallWithTimeout makes a single RPC call with timeout and proper error handling
func makeRPCCallWithTimeout(url string, args []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", BTCEndpoint, bytes.NewBuffer(args))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+Auth)

	httpClient := &http.Client{
		Timeout: requestTimeout,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("BTC API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully sent RPC: %s", string(body))
	result := ExtractResult(string(body))
	return result, nil
}

func UnlockWallet() string {
	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "unlock",
		"method":  "walletpassphrase",
		"params": []interface{}{
			WalletPassphrase,
			180,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal listunspent payload: %v", err)
		return ""
	}

	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("ListUnspent RPC call failed: %v", err)
		return ""
	}

	return result
}

func ListUnspent() string {
	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "wallet_txn",
		"method":  "listunspent",
		"params": []interface{}{
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

	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("ListUnspent RPC call failed: %v", err)
		return ""
	}

	return result
}

func GetRawAddress() string {
	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "wallet_address",
		"method":  "getrawchangeaddress",
		"params":  []interface{}{},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal getrawchangeaddress payload: %v", err)
		return ""
	}

	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("GetRawAddress RPC call failed: %v", err)
		return ""
	}

	return result
}

func CalculateRequired(numInputs int, dataSize int) float64 {
	return float64(53+numInputs*68+dataSize) * float64(0.00000001)
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
		} else {
			log.Printf("UTXO : %+v", u)
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
	change := ((totalAmt - required) * 100000000) / float64(100000000)
	log.Printf("Inputs: %v, Change: %f", inputs, change)
	return inputs, float64(change)
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

	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "op_cat_decode",
		"method":  "createrawtransaction",
		"params": []interface{}{
			inputs,
			map[string]interface{}{
				"data":  data,
				address: change,
			},
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal create raw transaction payload: %v", err)
		return ""
	}

	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("CreateRawTransaction RPC call failed: %v", err)
		return ""
	}

	return result
}

func DecodeRawTransaction(rawtransaction string) {
	if rawtransaction == "" {
		log.Printf("Empty raw transaction provided")
	}

	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "op_cat_decode",
		"method":  "decoderawtransaction",
		"params": []string{
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

	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "op_cat_sign_tx",
		"method":  "signrawtransactionwithwallet",
		"params": []string{
			rawtransaction,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal sign raw transaction with wallet payload: %v", err)
		return ""
	}

	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("SignRawTransaction RPC call failed: %v", err)
		return ""
	}

	return result
}

func SendSignedTransaction(transaction string) string {
	if transaction == "" {
		log.Printf("Empty transaction provided")
		return ""
	}

	log.Printf("Sending signed transaction to network")

	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "op_cat_send_tx",
		"method":  "sendrawtransaction",
		"params": []string{
			transaction,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal send raw transaction payload: %v", err)
		return ""
	}

	result, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("SendSignedTransaction RPC call failed: %v", err)
		return ""
	}

	return result
}

// GetTransactionInfo retrieves detailed transaction information including block details
func GetTransactionInfo(txHash string) (string, *int64) {
	if txHash == "" {
		log.Printf("Empty transaction hash provided")
		return "", nil
	}

	log.Printf("Getting transaction info for: %s", txHash)

	payload := map[string]interface{}{
		"jsonrpc": "1.0",
		"id":      "get_tx_info",
		"method":  "gettransaction",
		"params": []interface{}{
			txHash,
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal get transaction payload: %v", err)
		return "", nil
	}

	response, err := RetryRPCCall(func() (string, error) {
		return makeRPCCallWithTimeout(BTCEndpoint, jsonPayload)
	})

	if err != nil {
		log.Printf("GetTransactionInfo RPC call failed: %v", err)
		return "", nil
	}

	// Parse the response to extract block height
	var txResponse struct {
		Result struct {
			Blockhash     string `json:"blockhash,omitempty"`
			Blockheight   *int64 `json:"blockheight,omitempty"`
			Blockindex    *int   `json:"blockindex,omitempty"`
			Blocktime     *int64 `json:"blocktime,omitempty"`
			Confirmations *int   `json:"confirmations,omitempty"`
			Txid          string `json:"txid"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	err = json.Unmarshal([]byte(response), &txResponse)
	if err != nil {
		log.Printf("Failed to unmarshal transaction response: %v", err)
		return response, nil
	}

	if txResponse.Error != nil {
		log.Printf("RPC error in gettransaction: code=%d, message=%s", txResponse.Error.Code, txResponse.Error.Message)
		return response, nil
	}

	log.Printf("Transaction info retrieved - Block height: %v, Confirmations: %v", txResponse.Result.Blockheight, txResponse.Result.Confirmations)
	return response, txResponse.Result.Blockheight
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

	unlocked := UnlockWallet()
	if unlocked == "" {
		log.Printf("Failed to unlock wallet")
		return ""
	}

	log.Printf("Wallet unlocked: %s", unlocked)

	// Step 1: Get unspent outputs
	unspent := ListUnspent()
	if unspent == "" {
		log.Printf("Failed to get unspent outputs")
		return ""
	}

	// Step 2: Filter UTXOs
	inputs, change := FilterUTXOs(unspent, len(data))
	if len(inputs) == 0 {
		log.Printf("No suitable UTXOs found for transaction")
		return ""
	}

	// Step 3: Get raw address
	rawaddr := GetRawAddress()
	if rawaddr == "" {
		log.Printf("Failed to get raw address")
		return ""
	}

	// Step 4: Create raw transaction
	rawtscn := CreateRawTransaction(inputs, rawaddr, change, data)
	if rawtscn == "" {
		log.Printf("Failed to create raw transaction")
		return ""
	}

	// Step 5: Decode for verification (optional)
	DecodeRawTransaction(rawtscn)

	// Step 6: Sign transaction
	signtscn := SignRawTransaction(rawtscn)
	if signtscn == "" {
		log.Printf("Failed to sign transaction")
		return ""
	}

	// Step 7: Parse signed transaction
	var sgn signedtx
	err := json.Unmarshal([]byte(signtscn), &sgn)
	if err != nil {
		log.Printf("Failed to unmarshal signed transaction response: %v", err)
		return ""
	}

	// Step 8: Send signed transaction
	sendtscn := SendSignedTransaction(sgn.Hex)
	if sendtscn == "" {
		log.Printf("Failed to send signed transaction")
		return ""
	}

	log.Printf("Successfully created OP_RETURN transaction: %s", sendtscn)
	return sendtscn
}

func InitOPReturnRPC(endpoint string, auth string, passphrase string) {
	BTCEndpoint = endpoint
	Auth = auth
	WalletPassphrase = passphrase
}
