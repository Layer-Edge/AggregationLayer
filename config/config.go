package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ProtocolId string `yaml:"protocol-id"`

	ZmqEndpointDataBlock string `yaml:"zmq-endpoint-data-block"`

	BtcEndpoint string `yaml:"bitcoin-endpoint"`
	User        string `yaml:"bitcoin-user"`
	Auth        string `yaml:"bitcoin-auth"`

	WriteIntervalBlock             int `yaml:"write-interval-blocks"`
	WriteIntervalSeconds           int `yaml:"write-interval-seconds"`
	SuperProofWriteIntervalSeconds int `yaml:"super-proof-write-interval-seconds"`

	LayerEdgeRPC struct {
		ChainID                   int64  `yaml:"chain-id"`
		HTTP                      string `yaml:"http"`
		MerkleTreeStorageContract string `yaml:"merkle-tree-storage-contract"`
		PrivateKey                string `yaml:"private-key"`
	} `yaml:"layer-edge-rpc"`

	MerkleTreeGeneratorServer string `yaml:"merkle-tree-generator-server"`

	PostgresConnectionURI string `yaml:"postgres-connection-uri"`

	CMCAPIKey string `yaml:"cmc-api-key"`
}

var ConfigFilePath = flag.String(
	"c",
	"config.yml",
	"Specify the config path, default: 'config.yml' (root dir)",
)

func GetConfig() Config {
	var cfg Config

	// Check if we're in a test environment
	if isTestEnvironment() {
		// In test environment, use default config file
		// Look for test_config.yml in the project root
		if _, err := os.Stat("test_config.yml"); os.IsNotExist(err) {
			// If not found in current directory, look in parent directories
			*ConfigFilePath = findTestConfigFile()
		} else {
			*ConfigFilePath = "test_config.yml"
		}
	} else {
		// In normal environment, parse flags
		flag.Parse()
	}

	readFile(&cfg)

	return cfg
}

// isTestEnvironment checks if we're running in a test environment
func isTestEnvironment() bool {
	// Check if we're running tests by looking at the command line arguments
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") || strings.Contains(arg, "-test.") {
			return true
		}
	}
	return false
}

// findTestConfigFile searches for test_config.yml in parent directories
func findTestConfigFile() string {
	// Start from current directory and go up
	dir, _ := os.Getwd()

	for {
		configPath := dir + "/test_config.yml"
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		// Go up one directory
		parent := dir + "/.."
		parentAbs, err := filepath.Abs(parent)
		if err != nil || parentAbs == dir {
			// Can't go up further or reached root
			break
		}
		dir = parentAbs
	}

	// Fallback to current directory
	return "test_config.yml"
}

func validateConfig(cfg *Config) {
	if cfg.ProtocolId == "" {
		log.Fatal("Protocol Id is required in config file")
	}

	if cfg.PostgresConnectionURI == "" {
		log.Fatal("Postgres Connection URI is required in config file")
	}

	if cfg.MerkleTreeGeneratorServer == "" {
		log.Fatal("Merkle Tree Generator Server is required")
	}

	if cfg.Auth == "" {
		log.Fatal("BTC Auth is not given")
	}

	if cfg.LayerEdgeRPC.HTTP == "" {
		log.Fatal("LayerEdgeRPC URL is required")
	}

	if cfg.LayerEdgeRPC.PrivateKey == "" {
		log.Fatal("LayerEdgeRPC PrivateKey is required")
	}

	if cfg.LayerEdgeRPC.MerkleTreeStorageContract == "" {
		log.Fatal("LayerEdgeRPC MerkleTreeStorageContract is required")
	}

	if cfg.CMCAPIKey == "" {
		log.Fatal("CMCAPIKey is required")
	}

	if cfg.WriteIntervalBlock == 0 {
		cfg.WriteIntervalBlock = 1 // defaults to 1
	}

	if cfg.WriteIntervalSeconds == 0 {
		cfg.WriteIntervalSeconds = 600 // defaults to 10 min
	}

	if cfg.SuperProofWriteIntervalSeconds == 0 {
		cfg.SuperProofWriteIntervalSeconds = 84600 // defaults to 24 hours
	}
}

func readFile(cfg *Config) {
	var f *os.File
	var err error

	f, err = os.Open(*ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Reading config: %v\n", *ConfigFilePath)
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}

	validateConfig(cfg)
}
