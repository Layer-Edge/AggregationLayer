package models

import (
	"time"

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
