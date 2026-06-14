package internalsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MarcosHRFerreira/go-gerenciador-orcamento-deck/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	dbPingTimeout     = 5 * time.Second
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 25
	dbConnMaxLifetime = 5 * time.Minute
	dbConnMaxIdleTime = 2 * time.Minute
)

func ConnectPostgres(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(dbMaxOpenConns)
	db.SetMaxIdleConns(dbMaxIdleConns)
	db.SetConnMaxLifetime(dbConnMaxLifetime)
	db.SetConnMaxIdleTime(dbConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), dbPingTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	return db, nil
}
