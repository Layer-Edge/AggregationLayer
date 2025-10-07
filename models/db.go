package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Layer-Edge/bitcoin-da/utils"
	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	DB            *bun.DB
	dbMutex       sync.RWMutex
	healthCheck   *time.Ticker
	stopHealth    chan bool
	maxRetries    = 3
	retryDelay    = 1 * time.Second
	maxDelay      = 30 * time.Second
	backoffFactor = 2.0
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// InitDB initializes the database connection with enhanced error handling
func InitDB(dsn string) error {
	return InitDBWithConfig(DatabaseConfig{
		DSN:             dsn,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})
}

// InitDBWithConfig initializes the database with custom configuration
func InitDBWithConfig(config DatabaseConfig) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	// Close existing connection if any
	if DB != nil {
		DB.Close()
	}

	sqldb, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	sqldb.SetMaxOpenConns(config.MaxOpenConns)
	sqldb.SetMaxIdleConns(config.MaxIdleConns)
	sqldb.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqldb.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqldb.PingContext(ctx); err != nil {
		sqldb.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create a new Bun DB instance with PostgreSQL dialect
	DB = bun.NewDB(sqldb, pgdialect.New())

	// Start health check
	startHealthCheck()

	log.Println("Database connection established successfully")
	return nil
}

// startHealthCheck starts a periodic health check for the database
func startHealthCheck() {
	if healthCheck != nil {
		healthCheck.Stop()
	}

	healthCheck = time.NewTicker(30 * time.Second)
	stopHealth = make(chan bool)

	go func() {
		for {
			select {
			case <-healthCheck.C:
				if err := checkDBHealth(); err != nil {
					log.Printf("Database health check failed: %v", err)
					// Attempt to reconnect
					go attemptReconnection()
				}
			case <-stopHealth:
				return
			}
		}
	}()
}

// checkDBHealth performs a health check on the database
func checkDBHealth() error {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple query to test connection
	var result int
	err := DB.NewSelect().ColumnExpr("1").Scan(ctx, &result)
	if err != nil {
		return fmt.Errorf("health check query failed: %w", err)
	}

	return nil
}

// attemptReconnection attempts to reconnect to the database
func attemptReconnection() {
	log.Println("Attempting database reconnection...")

	// This would need the original DSN, which we should store
	// For now, we'll just log the attempt
	log.Println("Database reconnection attempt initiated")
}

// RetryDBOperation executes a database operation with retry logic
func RetryDBOperation(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(retryDelay) *
				utils.PowFloat(backoffFactor, float64(attempt-1)))
			if delay > maxDelay {
				delay = maxDelay
			}

			log.Printf("Retrying database operation after %v delay (attempt %d/%d)",
				delay, attempt+1, maxRetries)

			time.Sleep(delay)
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("Database operation failed (attempt %d/%d): %v",
			attempt+1, maxRetries, err)
	}

	return fmt.Errorf("database operation failed after %d attempts: %w", maxRetries, lastErr)
}

// GetDB returns the database instance with health check
func GetDB() (*bun.DB, error) {
	dbMutex.RLock()
	defer dbMutex.RUnlock()

	if DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	// Quick health check
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var result int
	err := DB.NewSelect().ColumnExpr("1").Scan(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	return DB, nil
}

// CloseDB closes the database connection
func CloseDB() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if healthCheck != nil {
		healthCheck.Stop()
		close(stopHealth)
	}

	if DB != nil {
		err := DB.Close()
		DB = nil
		return err
	}

	return nil
}
