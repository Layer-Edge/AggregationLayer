package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ProtocolId   string `yaml:"protocol-id"`
	LayerEdgeRPC struct {
		HTTP string `yaml:"http"`
		WSS  string `yaml:"wss"`
	} `yaml:"layer-edge-rpc"`

	ZmqEndpoint string `yaml:"zmq-endpoint"`

	PrivateKey struct {
		// internal key pair is used for tweaking
		Internal string `yaml:"internal"`
		// bob key pair is used for signing reveal tx
		Signer string `yaml:"signer"`
	} `yaml:"private-key"`

	Relayer struct {
		Host string `yaml:"host"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"relayer"`
}

func readFile(cfg *Config) {
	var f *os.File
	var err error

	if len(os.Args) > 1 {
		arg := os.Args[1]
		f, err = os.Open(arg)
	} else {
		f, err = os.Open("config.yml")
	}
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func readEnv(cfg *Config) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	privateKeyInternal := os.Getenv("PRIVATE_KEY_INTERNAL")
	privateKeySigner := os.Getenv("PRIVATE_KEY_SIGNER")

	if privateKeyInternal != "" {
		cfg.PrivateKey.Internal = privateKeyInternal
	}
	if privateKeySigner != "" {
		cfg.PrivateKey.Signer = privateKeySigner
	}

	if err != nil {
		log.Fatal(err)
	}
}

func GetConfig() Config {
	var cfg Config
	readFile(&cfg)
	readEnv(&cfg)
	return cfg
}
