package repository

import (
	"context"
	"database/sql"
)

// DBExecutor is an interface implemented by both *sqlx.DB (via *database.AppDB)
// and *sqlx.Tx. This allows repositories to execute queries in both
// transactional and non-transactional contexts seamlessly.
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}
