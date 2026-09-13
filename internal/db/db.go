package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return conn, nil
}

func EnsureSchema(ctx context.Context, conn *sql.DB) error {
	_, err := conn.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS visits (
			id SERIAL PRIMARY KEY,
			count INT NOT NULL DEFAULT 0
		);
		INSERT INTO visits (id, count)
		SELECT 1, 0
		WHERE NOT EXISTS (SELECT 1 FROM visits WHERE id = 1);
	`)
	return err
}
