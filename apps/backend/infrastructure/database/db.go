package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// AppDB wraps sqlx.DB to add custom logging with trace ID and dual pool management
type AppDB struct {
	*sqlx.DB
	userDB  *sqlx.DB
	adminDB *sqlx.DB
}

func NewConnection(dsn string) (*AppDB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool initialized successfully")

	return &AppDB{DB: db, userDB: db, adminDB: db}, nil
}

// NewDualPool initializes both userDB (app role subject to RLS) and adminDB (service role bypassing RLS).
func NewDualPool(appDSN, adminDSN string) (*AppDB, error) {
	userConn, err := sqlx.Connect("postgres", appDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user database pool: %w", err)
	}
	userConn.SetMaxOpenConns(25)
	userConn.SetMaxIdleConns(10)
	userConn.SetConnMaxLifetime(5 * time.Minute)
	userConn.SetConnMaxIdleTime(2 * time.Minute)
	if err := userConn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping user database: %w", err)
	}

	adminConn, err := sqlx.Connect("postgres", adminDSN)
	if err != nil {
		_ = userConn.Close()
		return nil, fmt.Errorf("failed to connect to admin database pool: %w", err)
	}
	adminConn.SetMaxOpenConns(10)
	adminConn.SetMaxIdleConns(5)
	adminConn.SetConnMaxLifetime(5 * time.Minute)
	adminConn.SetConnMaxIdleTime(2 * time.Minute)
	if err := adminConn.Ping(); err != nil {
		_ = userConn.Close()
		return nil, fmt.Errorf("failed to ping admin database: %w", err)
	}

	log.Println("Dual database connection pools (userDB & adminDB) initialized successfully")

	return &AppDB{
		DB:      userConn,
		userDB:  userConn,
		adminDB: adminConn,
	}, nil
}

func (db *AppDB) Close() error {
	var errUser, errAdmin error
	if db.userDB != nil {
		errUser = db.userDB.Close()
	}
	if db.adminDB != nil && db.adminDB != db.userDB {
		errAdmin = db.adminDB.Close()
	}
	if errUser != nil {
		return errUser
	}
	return errAdmin
}

func (db *AppDB) UserDB() *sqlx.DB {
	if db.userDB != nil {
		return db.userDB
	}
	return db.DB
}

func (db *AppDB) AdminDB() *sqlx.DB {
	if db.adminDB != nil {
		return db.adminDB
	}
	return db.DB
}

type userAuthContextKey struct{}

type UserAuthInfo struct {
	UserID string
	Email  string
}

// WithUserAuthContext wraps ctx with UserAuthInfo containing authenticated user credentials.
func WithUserAuthContext(ctx context.Context, userID, email string) context.Context {
	return context.WithValue(ctx, userAuthContextKey{}, UserAuthInfo{
		UserID: userID,
		Email:  email,
	})
}

// UserAuthFromContext retrieves UserAuthInfo from context if present.
func UserAuthFromContext(ctx context.Context) (UserAuthInfo, bool) {
	info, ok := ctx.Value(userAuthContextKey{}).(UserAuthInfo)
	return info, ok
}

// WithUserContext runs fn in a transaction scoped to the authenticated user.
// set_config(..., true) and SET LOCAL ROLE are transaction-scoped: they activate
// Supabase RLS (auth.uid()) for every query inside fn and vanish on commit/rollback.
func (db *AppDB) WithUserContext(ctx context.Context, userID, email string, fn func(tx *sqlx.Tx) error) error {
	if userID == "" {
		if info, ok := UserAuthFromContext(ctx); ok {
			userID = info.UserID
			if email == "" {
				email = info.Email
			}
		}
	}

	pool := db.UserDB()
	tx, err := pool.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after successful commit

	claims, err := json.Marshal(map[string]string{
		"sub":   userID,
		"role":  "authenticated",
		"email": email,
		"aud":   "authenticated",
	})
	if err != nil {
		return err
	}
	// Parameterized: never string-concatenate JWT data into SQL (OWASP A03).
	if _, err := tx.ExecContext(ctx, `SELECT set_config('request.jwt.claims', $1, true)`, string(claims)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE authenticated`); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func NewAppDB(db *sqlx.DB) *AppDB {
	return &AppDB{DB: db, userDB: db, adminDB: db}
}

func logExecution(ctx context.Context, query string, duration time.Duration, err error) {
	var traceID string
	if tid, ok := ctx.Value("X-Transaction-ID").(string); ok {
		traceID = tid
	} else {
		traceID = "system"
	}

	durationMs := duration.Milliseconds()

	if err != nil {
		slog.Error("DB Query Failed",
			slog.String("trace_id", traceID),
			slog.String("query", query),
			slog.Int64("duration_ms", durationMs),
			slog.Any("error", err),
		)
	} else {
		slog.Info("DB Query Executed",
			slog.String("trace_id", traceID),
			slog.String("query", query),
			slog.Int64("duration_ms", durationMs),
		)
	}
}

func (db *AppDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := db.DB.SelectContext(ctx, dest, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return err
}

func (db *AppDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := db.DB.GetContext(ctx, dest, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return err
}

func (db *AppDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := db.DB.ExecContext(ctx, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return res, err
}

func (db *AppDB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryxContext(ctx, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return rows, err
}

func (db *AppDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	logExecution(ctx, query, time.Since(start), nil)
	return row
}
