package goe

import (
	"context"
	"database/sql"
	"time"
)

type Select interface {
	WhereSelect
	Join
	Result
}

type WhereSelect interface {
	Where(...operator) SelectWhere
}

type SelectWhere interface {
	Result
}

type Join interface {
	Join(...any) Select
}

type Result interface {
	Result(any)
}

type Insert interface {
	Value
}

type InsertBetwent interface {
	Values
}

type Update interface {
	WhereUpdate
	Value
}

type WhereUpdate interface {
	Where(...operator) UpdateWhere
}

type UpdateWhere interface {
	Value
}

type Value interface {
	Value(any)
}

type Values interface {
	Values(any, any)
}

type Delete interface {
	Where(...operator)
}

type operator interface {
	operation() string
}

type DeleteIn interface {
	Where(...any)
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
