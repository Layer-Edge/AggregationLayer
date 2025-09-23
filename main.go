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
	da.HashBlockSubscriber(&cfg)
}
