package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Configuration constants
const (
	CosmosBashScriptPath = "/home/ubuntu/repo/bitcoin-da/scripts"
	DefaultCosmosAddress = "cosmos1c3y4q50cdyaa5mpfaa2k8rx33ydywl35hsvh0d"
)

type CosmosTxData struct {
	Success         bool   `json:"success"`
	From            string `json:"from"`
	To              string `json:"to"`
	Amount          string `json:"amount"`
	TransactionHash string `json:"transactionHash"`
	Memo            string `json:"memo"`
	BlockHeight     string `json:"blockHeight"` // can use int64 if you want to parse it directly
	GasUsed         string `json:"gasUsed"`     // same here
}

// Init initializes the Cosmos client with the provided configuration
func SendCosmosTXWithData(data string, addr string) ([]byte, error) {
	// Construct the API endpoint URL
	apiURL := "https://cosmos-api-hcf6.onrender.com/send-tokens"

	// Prepare the request payload
	payload := map[string]string{
		"recipient": addr,
		"memo":      data[1:], // Assuming data starts with a special character, remove it
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client and send request
	client := &http.Client{
		Timeout: 40 * time.Second, // Set a reasonable timeout
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s",
			resp.StatusCode, string(body))
	}

	return body, nil
}
