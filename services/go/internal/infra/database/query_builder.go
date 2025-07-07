package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryBuilder provides a fluent interface for building SQL queries
type QueryBuilder struct {
	columns     []string
	table       string
	joins       []string
	conditions  []string
	args        []interface{}
	orderBy     string
	limit       int
	offset      int
	currentArg  int
	returningSql string
}

// Querier defines the interface for database querying
type Querier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		columns:     []string{},
		joins:       []string{},
		conditions:  []string{},
		args:        []interface{}{},
		currentArg:  1,
	}
}

// Select specifies columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = columns
	return qb
}

// From specifies the table to select from
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(join string) *QueryBuilder {
	qb.joins = append(qb.joins, join)
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	// Replace ? placeholders with $n
	for strings.Contains(condition, "?") && len(args) > 0 {
		condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", qb.currentArg), 1)
		qb.currentArg++
		qb.args = append(qb.args, args[0])
		args = args[1:]
	}
	
	qb.conditions = append(qb.conditions, condition)
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// Limit adds a LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset adds an OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Returning adds a RETURNING clause for INSERT and UPDATE
func (qb *QueryBuilder) Returning(columns ...string) *QueryBuilder {
	qb.returningSql = "RETURNING " + strings.Join(columns, ", ")
	return qb
}

// ToSQL builds the SQL query string and returns it with arguments
func (qb *QueryBuilder) ToSQL() (string, []interface{}) {
	var builder strings.Builder
	
	// SELECT clause
	builder.WriteString("SELECT ")
	if len(qb.columns) == 0 {
		builder.WriteString("*")
	} else {
		builder.WriteString(strings.Join(qb.columns, ", "))
	}
	
	// FROM clause
	builder.WriteString(" FROM ")
	builder.WriteString(qb.table)
	
	// JOIN clauses
	for _, join := range qb.joins {
		builder.WriteString(" ")
		builder.WriteString(join)
	}
	
	// WHERE clauses
	if len(qb.conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(qb.conditions, " AND "))
	}
	
	// ORDER BY clause
	if qb.orderBy != "" {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(qb.orderBy)
	}
	
	// LIMIT clause
	if qb.limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}
	
	// OFFSET clause
	if qb.offset > 0 {
		builder.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}
	
	return builder.String(), qb.args
}

// Insert builds an INSERT query
func (qb *QueryBuilder) Insert(columns []string, values []interface{}) (string, []interface{}) {
	var builder strings.Builder
	var placeholders []string
	
	// Start INSERT statement
	builder.WriteString(fmt.Sprintf("INSERT INTO %s (", qb.table))
	
	// Add columns
	builder.WriteString(strings.Join(columns, ", "))
	
	// Add VALUES
	builder.WriteString(") VALUES (")
	
	// Create placeholders
	for i := range values {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
	}
	
	builder.WriteString(strings.Join(placeholders, ", "))
	builder.WriteString(")")
	
	// Add RETURNING if specified
	if qb.returningSql != "" {
		builder.WriteString(" ")
		builder.WriteString(qb.returningSql)
	}
	
	return builder.String(), values
}

// Update builds an UPDATE query
func (qb *QueryBuilder) Update(columns []string, values []interface{}) (string, []interface{}) {
	var builder strings.Builder
	var sets []string
	
	// Start UPDATE statement
	builder.WriteString(fmt.Sprintf("UPDATE %s SET ", qb.table))
	
	// Add SET statements
	for i, col := range columns {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i+1))
	}
	
	builder.WriteString(strings.Join(sets, ", "))
	
	// Add WHERE if conditions exist
	if len(qb.conditions) > 0 {
		builder.WriteString(" WHERE ")
		
		// Adjust placeholders in conditions to continue from the last value
		condStr := strings.Join(qb.conditions, " AND ")
		for i := 1; i <= qb.currentArg-1; i++ {
			oldPlaceholder := fmt.Sprintf("$%d", i)
			newPlaceholder := fmt.Sprintf("$%d", i+len(values))
			condStr = strings.Replace(condStr, oldPlaceholder, newPlaceholder, -1)
		}
		
		builder.WriteString(condStr)
	}
	
	// Add RETURNING if specified
	if qb.returningSql != "" {
		builder.WriteString(" ")
		builder.WriteString(qb.returningSql)
	}
	
	// Combine values with condition args
	allArgs := append(values, qb.args...)
	
	return builder.String(), allArgs
}

// Delete builds a DELETE query
func (qb *QueryBuilder) Delete() (string, []interface{}) {
	var builder strings.Builder
	
	// Start DELETE statement
	builder.WriteString(fmt.Sprintf("DELETE FROM %s", qb.table))
	
	// Add WHERE if conditions exist
	if len(qb.conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(qb.conditions, " AND "))
	}
	
	return builder.String(), qb.args
}

// Execute executes the query with the given context and querier
func (qb *QueryBuilder) Execute(ctx context.Context, querier Querier) (pgx.Rows, error) {
	sql, args := qb.ToSQL()
	return querier.Query(ctx, sql, args...)
}

// ExecuteRow executes the query and returns a single row
func (qb *QueryBuilder) ExecuteRow(ctx context.Context, querier Querier) pgx.Row {
	sql, args := qb.ToSQL()
	return querier.QueryRow(ctx, sql, args...)
}

// ExecuteInsert executes an insert with the given columns and values
func (qb *QueryBuilder) ExecuteInsert(ctx context.Context, querier Querier, columns []string, values []interface{}) (pgconn.CommandTag, error) {
	sql, args := qb.Insert(columns, values)
	return querier.Exec(ctx, sql, args...)
}

// ExecuteUpdate executes an update with the given columns and values
func (qb *QueryBuilder) ExecuteUpdate(ctx context.Context, querier Querier, columns []string, values []interface{}) (pgconn.CommandTag, error) {
	sql, args := qb.Update(columns, values)
	return querier.Exec(ctx, sql, args...)
}

// ExecuteDelete executes a delete query
func (qb *QueryBuilder) ExecuteDelete(ctx context.Context, querier Querier) (pgconn.CommandTag, error) {
	sql, args := qb.Delete()
	return querier.Exec(ctx, sql, args...)
}

// GetQuerier returns a querier from the context or the pool
func GetQuerier(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}
