package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Layer-Edge/bitcoin-da/clients"
	"github.com/uptrace/bun"
)

type AggregatedProof struct {
	bun.BaseModel `bun:"table:aggregated_proofs,alias:ap"`

	ID              string    `bun:"id,pk,type:char(24),default:generate_mongo_objectid('mongo_objectid_aggregate_proofs_seq')"`
	BlockHeight     int64     `bun:"block_height,unique,notnull"`
	BTCTxHash       *string   `bun:"btc_tx_hash,type:varchar(255)"`
	From            string    `bun:"from,type:varchar(255),notnull"`
	GasUsed         int64     `bun:"gas_used,notnull,default:0"`
	AggregateProof  []byte    `bun:"aggregate_proof,type:bytea,notnull"`
	Proofs          []string  `bun:"proofs,array,type:text[],notnull,default:'{}'"`
	To              string    `bun:"to,type:varchar(255),notnull"`
	TransactionHash string    `bun:"transaction_hash,type:varchar(255),notnull"`
	Amount          string    `bun:"amount,type:text,notnull"`
	Success         bool      `bun:"success,notnull,default:false"`
	Timestamp       time.Time `bun:"timestamp,notnull"`
	CreatedAt       time.Time `bun:"created_at,auto_create"`
	UpdatedAt       time.Time `bun:"updated_at,auto_update"`
}

func CreateAggregatedProof(agg_proof string, proof_list []string, btc_tx_hash string, data clients.TxData) (sql.Result, error) {
	block_height, err := strconv.ParseInt(data.BlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error converting block height ", err)
	}

	gas_used, err := strconv.ParseInt(data.GasUsed, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Error converting gas used ", err)
	}

	ap := &AggregatedProof{
		BTCTxHash:       &btc_tx_hash,
		BlockHeight:     block_height,
		From:            data.From,
		GasUsed:         gas_used,
		AggregateProof:  []byte(agg_proof),
		Proofs:          proof_list,
		To:              data.To,
		TransactionHash: data.TransactionHash,
		Amount:          data.Amount,
		Success:         data.Success,
		Timestamp:       time.Now().UTC(),
	}

	log.Printf("Storing proof info to Postgres DB: %v", *ap)
	newAggProof, err := DB.NewInsert().Model(ap).Exec(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Insert failed: %v", err)
	}

	log.Println("Inserted AggregatedProof successfully")

	return newAggProof, nil
}
