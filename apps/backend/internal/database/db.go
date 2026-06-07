package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// AppDB wraps sqlx.DB to add custom logging with trace ID
type AppDB struct {
	*sqlx.DB
}

// NewAppDB creates a new AppDB instance
func NewAppDB(db *sqlx.DB) *AppDB {
	return &AppDB{DB: db}
}

// logQuery is a helper to log queries with trace ID and execution time
func logQuery(ctx context.Context, query string, args ...interface{}) {
	// Extract trace ID from context (if exists)
	var traceID string
	if tid, ok := ctx.Value("X-Transaction-ID").(string); ok {
		traceID = tid
	} else {
		traceID = "system" // Fallback for non-HTTP contexts (like cron/bot)
	}

	start := time.Now()

	// We use a closure to capture execution time after the query runs
	defer func() {
		duration := time.Since(start).Milliseconds()
		log.Info().
			Str("trace_id", traceID).
			Str("query", query).
			Interface("args", args).
			Int64("duration_ms", duration).
			Msg("DB Query Executed")
	}()
}

// Override SelectContext to log the query
func (db *AppDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logQuery(ctx, query, args...)
	return db.DB.SelectContext(ctx, dest, query, args...)
}

// Override GetContext to log the query
func (db *AppDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logQuery(ctx, query, args...)
	return db.DB.GetContext(ctx, dest, query, args...)
}

// Override ExecContext to log the query
func (db *AppDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logQuery(ctx, query, args...)
	return db.DB.ExecContext(ctx, query, args...)
}

// Override QueryxContext to log the query
func (db *AppDB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	logQuery(ctx, query, args...)
	return db.DB.QueryxContext(ctx, query, args...)
}
