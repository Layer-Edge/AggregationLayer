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
		Order("id ASC").
		Limit(100).
		Offset(0).
		Scan(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response received %s\n", &proofs)
}
