package rdb

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// TODO: expose more APIs if needed
type IDB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	PingContext(ctx context.Context) error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	// Conn(ctx context.Context) (*Conn, error)
	// Driver() driver.Driver
	// SetConnMaxIdleTime(d time.Duration)
	// SetConnMaxLifetime(d time.Duration)
	// SetMaxIdleConns(n int)
	// SetMaxOpenConns(n int)
	// Stats() DBStats
}
