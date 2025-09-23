package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ProtocolId string `yaml:"protocol-id"`

	ZmqEndpointDataBlock string `yaml:"zmq-endpoint-data-block"`

	BtcEndpoint string `yaml:"bitcoin-endpoint"`
	User        string `yaml:"bitcoin-user"`
	Auth        string `yaml:"bitcoin-auth"`

	WriteIntervalBlock   int `yaml:"write-interval-blocks"`
	WriteIntervalSeconds int `yaml:"write-interval-seconds"`

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
	flag.Parse()

	readFile(&cfg)

	return cfg
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
