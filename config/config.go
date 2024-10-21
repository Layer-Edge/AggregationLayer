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

	BtcCliPath string `yaml:"bitcoin-cli-path"`
	BashScriptPath string `yaml:"bash-script-path"`

	EnableWriter bool `yaml:"enable-writer"`

	WriteIntervalBlock int `yaml:"write-interval-blocks"`

	LayerEdgeRPC struct {
		HTTP string `yaml:"http"`
		WSS  string `yaml:"wss"`
	} `yaml:"layer-edge-rpc"`

	Cosmos struct {
		ChainID       string `yaml:"chainId"`
		RpcEndpoint   string `yaml:"rpcEndpoint"`
		AccountPrefix string `yaml:"accountPrefix"`
	} `yaml:"cosmos"`
	// PrivateKey struct {
	// 	// internal key pair is used for tweaking
	// 	Internal string `yaml:"internal"`
	// 	// bob key pair is used for signing reveal tx
	// 	Signer string `yaml:"signer"`
	// } `yaml:"private-key"`

	// Relayer struct {
	// 	Host string `yaml:"host"`
	// 	User string `yaml:"user"`
	// 	Pass string `yaml:"pass"`
	// } `yaml:"relayer"`
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

	// readEnv(&cfg)

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

	if cfg.WriteIntervalBlock == 0 {
		cfg.WriteIntervalBlock = 1 // defaults to 1
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

// func readEnv(cfg *Config) {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// 
// 	cfg.PrivateKey.Internal = os.Getenv("PRIVATE_KEY_INTERNAL")
// 	cfg.PrivateKey.Signer = os.Getenv("PRIVATE_KEY_SIGNER")
// 	log.Println(cfg.PrivateKey.Internal, cfg.PrivateKey.Signer)
// 
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
