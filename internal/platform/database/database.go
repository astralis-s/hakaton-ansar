// Package database owns PostgreSQL connectivity: the pgx connection pool, the
// goose migration runner (embedded migrations, executed on boot), and the
// WithinTx transaction helper used by application services.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver (used by goose)
	"github.com/pressly/goose/v3"

	"github.com/astralis-s/hakaton-ansar/internal/platform/config"
)

// NewPool creates and verifies a pgx connection pool.
func NewPool(ctx context.Context, cfg config.DB) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Migrate applies all pending goose migrations from the embedded filesystem.
// It uses a short-lived database/sql connection (goose works over database/sql)
// while the application itself runs on the pgx pool.
func Migrate(ctx context.Context, dbURL string, fsys fs.FS, log *slog.Logger) error {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return fmt.Errorf("open sql db for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(fsys)
	goose.SetLogger(gooseLogger{log: log})
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// WithinTx runs fn inside a single transaction, committing on success and
// rolling back on error or panic. Application services use this for atomic
// use-cases (e.g. create+activate a contract).
func WithinTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-raise after cleanup
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// gooseLogger adapts goose's logger interface onto slog so migration output is
// structured like the rest of the application.
type gooseLogger struct{ log *slog.Logger }

func (g gooseLogger) Printf(format string, v ...any) {
	if g.log == nil {
		return
	}
	g.log.Info("goose: " + strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func (g gooseLogger) Fatalf(format string, v ...any) {
	if g.log != nil {
		g.log.Error("goose: " + strings.TrimSpace(fmt.Sprintf(format, v...)))
	}
}
