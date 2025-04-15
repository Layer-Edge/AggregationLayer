package models

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
)

var DB *bun.DB

func InitDB(dsn string) error {
	sqldb, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	DB = bun.NewDB(sqldb, nil)

	return nil
}
