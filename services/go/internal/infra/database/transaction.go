package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// TxKey is the key used to store transaction in context
type TxKey struct{}

// Transaction represents a database transaction
type Transaction interface {
	// Exec executes SQL query that doesn't return rows
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	
	// Query executes SQL query that returns rows
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	
	// QueryRow executes SQL query that returns at most one row
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	
	// Begin starts a new transaction
	Begin(ctx context.Context) (pgx.Tx, error)
	
	// Commit commits the transaction
	Commit(ctx context.Context) error
	
	// Rollback aborts the transaction
	Rollback(ctx context.Context) error
}

// TxManager manages database transactions
type TxManager struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewTxManager creates a new transaction manager
func NewTxManager(pool *pgxpool.Pool, logger logger.Logger) *TxManager {
	return &TxManager{
		pool:   pool,
		logger: logger,
	}
}

// WithinTransaction executes a function within a transaction
func (tm *TxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Check if we're already in a transaction
	tx, ok := ctx.Value(TxKey{}).(pgx.Tx)
	if ok {
		// Already in a transaction, reuse it
		return fn(ctx)
	}

	// Start a new transaction
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to begin transaction")
	}

	// Add transaction to context
	txCtx := context.WithValue(ctx, TxKey{}, tx)

	// Execute the function
	err = fn(txCtx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			tm.logger.Error("Failed to rollback transaction", rbErr, 
				logger.Field{Key: "rollback_error", Value: rbErr.Error()},
				logger.Field{Key: "original_error", Value: err.Error()})
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to commit transaction")
	}

	return nil
}

// ExecuteInTransaction executes a query within a transaction
func (tm *TxManager) ExecuteInTransaction(ctx context.Context, queryFn func(tx pgx.Tx) error) error {
	// Check if we're already in a transaction
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		// Already in a transaction, execute within it
		return queryFn(tx)
	}

	// Start a new transaction
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to begin transaction")
	}

	// Execute the query function
	if err := queryFn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			tm.logger.Error("Failed to rollback transaction", rbErr,
				logger.Field{Key: "rollback_error", Value: rbErr.Error()},
				logger.Field{Key: "original_error", Value: err.Error()})
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to commit transaction")
	}

	return nil
}

// GetTx gets the transaction from context or starts a new one
func (tm *TxManager) GetTx(ctx context.Context) (pgx.Tx, error) {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx, nil
	}

	// Start a new transaction
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to begin transaction")
	}

	return tx, nil
}

// GetQuerier returns either a transaction from context or the connection pool
func (tm *TxManager) GetQuerier(ctx context.Context) interface{} {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx
	}
	return tm.pool
}
