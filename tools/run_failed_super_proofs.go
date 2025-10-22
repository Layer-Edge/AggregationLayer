package main

import (
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/da"
)

var cfg = config.GetConfig()

func main() {
	da.NonBTCTxSuperProofCronJob(&cfg, true)
}
