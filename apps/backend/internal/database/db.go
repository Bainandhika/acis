package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	zerolog "github.com/rs/zerolog/log"
)

// AppDB wraps sqlx.DB to add custom logging with trace ID
type AppDB struct {
	*sqlx.DB
}

func NewConnection(dsn string) (*AppDB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Best practice: Set connection pool limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connected and pool initialized")

	// Langsung return sebagai AppDB wrapper lo
	return &AppDB{DB: db}, nil
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
		zerolog.Info().
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

// QueryRowContext overrides sqlx.DB.QueryRowContext to log the query
// Note: sqlx.Row executes lazily on Scan, so we can't easily measure duration here without wrapping the Row itself.
// For MVP, we just log the query execution attempt.
func (db *AppDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	logQuery(ctx, query, args, 0, nil)
	return db.DB.QueryRowContext(ctx, query, args...)
}
