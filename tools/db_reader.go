package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Layer-Edge/bitcoin-da/config"
	"github.com/Layer-Edge/bitcoin-da/models"
)

var cfg = config.GetConfig()

func main() {
	models.InitDB(cfg.PostgresConnectionURI)

	var proofs []models.AggregatedProof

	ctx := context.Background()

	err := models.
		DB.
		NewSelect().
		Model(&proofs).
		Where("btc_tx_hash IS NOT NULL").
		Order("id DESC").
		Limit(1).
		Offset(0).
		Scan(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response received %s\n", &proofs)
	fmt.Printf("BTC TX: %s\n", string(*proofs[0].BTCTxHash))
	fmt.Printf("Edgen Chain TX: %s\n", string(proofs[0].TransactionHash))
	fmt.Printf("Aggregated Proof: %s\n", string(proofs[0].AggregateProof))
	fmt.Printf("Aggregated Proof Timestamp: %s\n", proofs[0].Timestamp.Format("2006-01-02 15:04:05"))
}
