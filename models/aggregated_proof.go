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
	BTCBlockNumber  *int64    `bun:"btc_block_number,type:bigint"`
	BTCTxHash       *string   `bun:"btc_tx_hash,type:varchar(255)"`
	From            string    `bun:"from,type:varchar(255),notnull"`
	GasUsed         int64     `bun:"gas_used,notnull,default:0"`
	AggregateProof  []byte    `bun:"aggregate_proof,type:bytea,notnull"`
	Proofs          []string  `bun:"proofs,array,type:text[],notnull,default:'{}'"`
	To              string    `bun:"to,type:varchar(255),notnull"`
	TransactionHash string    `bun:"transaction_hash,type:varchar(255),notnull"`
	TransactionFee  string    `bun:"transaction_fee,type:double precision,default:0"`
	EdgenPrice      string    `bun:"edgen_price,type:double precision,default:0"`
	Amount          string    `bun:"amount,type:double precision,notnull"`
	Success         bool      `bun:"success,notnull,default:false"`
	Timestamp       time.Time `bun:"timestamp,notnull"`
	CreatedAt       time.Time `bun:"created_at,auto_create"`
	UpdatedAt       time.Time `bun:"updated_at,auto_update"`
}

func CreateAggregatedProof(agg_proof string, proof_list []string, data clients.TxData) (sql.Result, error) {
	return CreateAggregatedProofWithBTC(agg_proof, proof_list, nil, nil, data)
}

func CreateAggregatedProofWithBTC(agg_proof string, proof_list []string, btc_tx_hash *string, btc_block_number *int64, data clients.TxData) (sql.Result, error) {
	block_height, err := strconv.ParseInt(data.BlockHeight, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting block height: %w", err)
	}

	gas_used, err := strconv.ParseInt(data.GasUsed, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting gas used: %w", err)
	}

	var transaction_fee *string = nil
	if btc_tx_hash == nil {
		transaction_fee = &data.TransactionFee
	}

	ap := &AggregatedProof{
		BTCTxHash:       btc_tx_hash,
		BTCBlockNumber:  btc_block_number,
		BlockHeight:     block_height,
		From:            data.From,
		GasUsed:         gas_used,
		AggregateProof:  []byte(agg_proof),
		Proofs:          proof_list,
		To:              data.To,
		TransactionHash: data.TransactionHash,
		TransactionFee:  *transaction_fee,
		EdgenPrice:      data.EdgenPrice,
		Amount:          data.Amount,
		Success:         data.Success,
		Timestamp:       time.Now().UTC(),
	}

	log.Printf("Storing proof info to Postgres DB: %v", *ap)

	// Use retry mechanism for database operation
	var newAggProof sql.Result
	err = RetryDBOperation(func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := db.NewInsert().Model(ap).Exec(ctx)
		if err != nil {
			return fmt.Errorf("insert operation failed: %w", err)
		}

		newAggProof = result
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create aggregated proof after retries: %w", err)
	}

	log.Println("Inserted AggregatedProof successfully")
	return newAggProof, nil
}

// GetUnprocessedMerkleRoots fetches all merkle roots from aggregated_proofs that haven't been processed in a super proof yet
func GetUnprocessedMerkleRoots(lastProcessedTimestamp time.Time) ([]string, error) {
	var merkleRoots []string

	err := RetryDBOperation(func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Query for merkle roots (aggregate_proof field) from records created after lastProcessedTimestamp
		err = db.NewSelect().
			Model(&AggregatedProof{}).
			Column("aggregate_proof").
			Where("timestamp > ?", lastProcessedTimestamp).
			Order("timestamp ASC").
			Scan(ctx, &merkleRoots)

		if err != nil {
			return fmt.Errorf("failed to fetch merkle roots: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get unprocessed merkle roots after retries: %w", err)
	}

	log.Printf("Fetched %d unprocessed merkle roots since %v", len(merkleRoots), lastProcessedTimestamp)
	return merkleRoots, nil
}

// GetLastSuperProofTimestamp returns the timestamp of the last super proof creation
// Super proofs are identified by having a non-null BTCTxHash
func GetLastSuperProofTimestamp() (time.Time, error) {
	var lastTimestamp time.Time

	err := RetryDBOperation(func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Query for the most recent super proof (where btc_tx_hash is not null)
		err = db.NewSelect().
			Model(&AggregatedProof{}).
			Column("timestamp").
			Where("btc_tx_hash IS NOT NULL").
			Order("timestamp DESC").
			Limit(1).
			Scan(ctx, &lastTimestamp)

		if err != nil {
			// If no super proof exists yet, return a default timestamp (24 hours ago)
			if err == sql.ErrNoRows {
				return nil // This will be handled in the calling code
			}
			return fmt.Errorf("failed to fetch last super proof timestamp: %w", err)
		}

		return nil
	})

	if err != nil {
		return time.Now().UTC().Add(-24 * time.Hour), nil // Default to 24 hours ago
	}

	log.Printf("Last super proof timestamp: %v", lastTimestamp)
	return lastTimestamp, nil
}

func GetSuperProofsWithoutBTCTxHash() ([]AggregatedProof, error) {
	var proofs []AggregatedProof

	err := RetryDBOperation(func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Query for the most recent super proof (where btc_tx_hash is not null)
		err = db.NewSelect().
			Model(&proofs).
			Where("btc_tx_hash = ''").
			Order("id ASC").
			Limit(1).
			Scan(ctx)

		if err != nil {
			// If no super proof exists yet, return a default timestamp (24 hours ago)
			if err == sql.ErrNoRows {
				return nil // This will be handled in the calling code
			}
			return fmt.Errorf("failed to fetch last super proof timestamp: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err // Default to 24 hours ago
	}

	log.Printf("Found %d Super Proofs Without BTC TX Hash", len(proofs))
	return proofs, nil
}

func UpdateSuperProofWithBTCTxHash(id string, btc_tx_hash *string, btc_block_number *int64) error {
	err := RetryDBOperation(func() error {
		db, err := GetDB()
		if err != nil {
			return fmt.Errorf("failed to get database connection: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, err = db.NewUpdate().
			Model(&AggregatedProof{}).
			Where("id = ?", id).
			Set("btc_tx_hash = ?", btc_tx_hash).
			Set("btc_block_number = ?", btc_block_number).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to update super proof with BTC TX hash: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update super proof with BTC TX hash after retries: %w", err)
	}

	log.Printf("Updated super proof with BTC TX hash: %s", id)
	return nil
}
