package main

import (
	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/notification_service"
)

// PROTOCOL_ID allows data identification by looking at the first few bytes
var (
	PROTOCOL_ID = []byte(config.GetConfig().ProtocolId)
	cfg         = config.GetConfig()
)

func main() {
	notification_service.ZeromqBlockSubscriber()
}
