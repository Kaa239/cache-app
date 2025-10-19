package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresDB(ctx context.Context, dataBaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dataBaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse postgres url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool (to postgres ) : %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database (to postgres ) : %w", err)
	}
	log.Printf("successfully connected to database (to postgres )", dataBaseURL)

	// Автоматическое создание таблиц если они не существуют
	if err := createTables(ctx, pool); err != nil {
		return nil, fmt.Errorf("unable to create tables: %w", err)
	}

	return pool, nil
}

func createTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			order_uid VARCHAR(255) PRIMARY KEY,
			track_number VARCHAR(255),
			entry VARCHAR(50),
			delivery JSONB,
			payment JSONB,
			items JSONB,
			locale VARCHAR(10),
			internal_signature VARCHAR(255),
			customer_id VARCHAR(255),
			delivery_service VARCHAR(100),
			shardkey VARCHAR(50),
			sm_id INTEGER,
			date_created TIMESTAMP WITH TIME ZONE,
			oof_shard VARCHAR(50)
		);

		CREATE INDEX IF NOT EXISTS idx_orders_order_uid ON orders(order_uid);
		CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created);
	`

	_, err := pool.Exec(ctx, query)
	return err
}
