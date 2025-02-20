package da

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

const (
    serverEndpoint = "http://localhost:3000"
)

type ProcessRequest struct {
    Operation    string   `json:"operation"`
    Data        []string `json:"data"`
    ProofRequest interface{} `json:"proof_request"`
    Proof       interface{} `json:"proof"`
}

// Updated to handle both possible response formats
type ProcessResponse struct {
    Root         string `json:"root"`
    Visualization struct {
        // This can handle either array or map format
        DataToHashMapping json.RawMessage `json:"data_to_hash_mapping"`
    } `json:"visualization"`
}

func GetMerkleRoot(input string) (string, error) {
    dataArray := strings.Split(input, ",")

    // Create request payload
    reqData := ProcessRequest{
        Operation: "insert",
        Data:      dataArray,
    }

    jsonData, err := json.Marshal(reqData)
    if err != nil {
        return "", fmt.Errorf("error creating JSON request: %v", err)
    }

    // Make POST request
    resp, err := http.Post(
        serverEndpoint+"/process",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return "", fmt.Errorf("error making HTTP request: %v", err)
    }
    defer resp.Body.Close()

    // Parse response
    var processResp ProcessResponse
    if err := json.NewDecoder(resp.Body).Decode(&processResp); err != nil {
        return "", fmt.Errorf("error decoding response: %v", err)
    }

    if processResp.Root == "" {
        return "", fmt.Errorf("failed to insert data into Merkle Tree")
    }

    return processResp.Root, nil
}