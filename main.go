package main

import (
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/da"
)

var cfg = config.GetConfig()
// var processor = store.GetProcessor(cfg)
// var channelReader = store.GetChannelReader(cfg)
// var relayer = store.GetRelayer(cfg)

func main() {
	if cfg.EnableWriter {
		da.HashBlockSubscriber(&cfg)
	} else {
		da.RawBlockSubscriber(&cfg)
	}
}
