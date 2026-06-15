// Package database owns PostgreSQL connectivity: the pgx connection pool, the
// goose migration runner (embedded migrations, executed on boot), and the
// transaction boundary used by application services.
//
// Transaction propagation uses the context: TxManager.WithinTx stores the active
// pgx.Tx in the context, and Querier returns that tx (or the pool) so module
// repositories transparently enlist in the surrounding transaction without the
// application layer ever touching pgx.
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
	"github.com/jackc/pgx/v5/pgconn"
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

// --- Transaction propagation via context -----------------------------------

type txCtxKey struct{}

// PgxDBTX is the structural equivalent of every module's sqlc-generated DBTX
// interface, satisfied by both *pgxpool.Pool and pgx.Tx.
type PgxDBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// Querier returns the transaction stored in ctx, or the pool when there is none.
// Repositories pass the result to their sqlcgen.New(...).
func Querier(ctx context.Context, pool *pgxpool.Pool) PgxDBTX {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// TxManager implements the transaction boundary over a pgx pool.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager builds a TxManager. It satisfies each module's domain TxManager
// port (WithinTx(ctx, func(ctx) error) error).
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx runs fn inside a single transaction, committing on success and rolling
// back on error or panic. A nested call reuses the outer transaction.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return fn(ctx) // already inside a transaction
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	ctxTx := context.WithValue(ctx, txCtxKey{}, tx)
	if err = fn(ctxTx); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// gooseLogger adapts goose's logger interface onto slog.
type gooseLogger struct{ log *slog.Logger }

func (g gooseLogger) Printf(format string, v ...any) {
	if g.log != nil {
		g.log.Info("goose: " + strings.TrimSpace(fmt.Sprintf(format, v...)))
	}
}

func (g gooseLogger) Fatalf(format string, v ...any) {
	if g.log != nil {
		g.log.Error("goose: " + strings.TrimSpace(fmt.Sprintf(format, v...)))
	}
}
