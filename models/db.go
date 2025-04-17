package models

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var DB *bun.DB

func InitDB(dsn string) error {
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	// Create a new Bun DB instance with PostgreSQL dialect
	DB = bun.NewDB(sqldb, pgdialect.New())

	return nil
}
