package main

import (
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/da"
)

var cfg = config.GetConfig()

func main() {
	if cfg.EnableWriter {
		da.HashBlockSubscriber(&cfg)
	} else {
		da.RawBlockSubscriber(&cfg)
	}
}
