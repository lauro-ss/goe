package goe

import (
	"context"
	"database/sql"
	"time"
)

type operator interface {
	operation() string
}

type Driver interface {
	Migrate(*Migrator, Connection)
	Init(*DB)
}

type Connection interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Close() error
}

type ConnectionPool interface {
	Connection
	Conn(ctx context.Context) (*sql.Conn, error)
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}
