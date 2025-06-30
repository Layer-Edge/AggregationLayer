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

	ZmqEndpointRawBlock  string `yaml:"zmq-endpoint-raw-block"`
	ZmqEndpointHashBlock string `yaml:"zmq-endpoint-hash-block"`
	ZmqEndpointDataBlock string `yaml:"zmq-endpoint-data-block"`

	BtcCliPath  string `yaml:"bitcoin-cli-path"`
	BtcEndpoint string `yaml:"bitcoin-endpoint"`
	User        string `yaml:"bitcoin-user"`
	Auth        string `yaml:"bitcoin-auth"`

	BashScriptPath string `yaml:"bash-script-path"`

	EnableWriter bool `yaml:"enable-writer"`

	WriteIntervalBlock int `yaml:"write-interval-blocks"`
	WriteIntervalSeconds int `yaml:"write-interval-seconds"`

	LayerEdgeRPC struct {
		ChainID                   int64  `yaml:"chain-id"`
		HTTP                      string `yaml:"http"`
		WSS                       string `yaml:"wss"`
		MerkleTreeStorageContract string `yaml:"merkle-tree-storage-contract"`
		PrivateKey                string `yaml:"private-key"`
	} `yaml:"layer-edge-rpc"`

	Cosmos struct {
		ChainID                   string `yaml:"chainId"`
		RpcEndpoint               string `yaml:"rpcEndpoint"`
		AccountPrefix             string `yaml:"accountPrefix"`
		NodeAddr                  string `yaml:"nodeAddr"`
		ContractAddr              string `yaml:"contractAddr"`
		Keyring                   string `yaml:"keyring"`
		From                      string `yaml:"from"`
		KeyringBackend            string `yaml:"keyringBackend"`
		KeyName                   string `yaml:"keyName"`
		MerkleTreeStorageContract string `yaml:"merkleTreeStorageContract"`
	} `yaml:"cosmos"`

	MerkleTreeGeneratorServer string `yaml:"merkle-tree-generator-server"`

	PostgresConnectionURI string `yaml:"postgres-connection-uri"`
}

// Define a command-line flag
var IsWriter = flag.Bool(
	"w",
	false,
	"Run DA Writer",
)

var ConfigFilePath = flag.String(
	"c",
	"config.yml",
	"Specify the config path, default: 'config.yml' (root dir)",
)

func GetConfig() Config {
	var cfg Config
	flag.Parse()

	readFile(&cfg)

	if *IsWriter {
		cfg.EnableWriter = true
	}

	return cfg
}

func validateConfig(cfg *Config) {
	if cfg.ProtocolId == "" {
		log.Fatal("Protocol Id is required in config file")
	}

	if cfg.BtcCliPath == "" {
		log.Fatal("Bitcoin CLI Path not given")
	}

	if cfg.BashScriptPath == "" {
		log.Fatal("Bash Script Path not given")
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

	if cfg.Cosmos.ChainID == "" {
		log.Fatal("Cosmos ChainID is required")
	}

	if cfg.Cosmos.RpcEndpoint == "" {
		log.Fatal("Cosmos RpcEndpoint is required")
	}

	if cfg.Cosmos.KeyringBackend == "" {
		log.Fatal("Cosmos KeyringBackend is required")
	}

	if cfg.Cosmos.KeyName == "" {
		log.Fatal("Cosmos KeyName is required")
	}

	if cfg.Cosmos.MerkleTreeStorageContract == "" {
		log.Fatal("Cosmos MerkleTreeStorageContract is required")
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
