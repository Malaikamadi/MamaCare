package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/pkg/logger"
)

// PostgresConfig contains the PostgreSQL specific database configuration
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	Schema   string
	PoolMax  int
}

// PostgresDB manages the connection to PostgreSQL
type PostgresDB struct {
	pool   *pgxpool.Pool
	config PostgresConfig
	logger logger.Logger
}

// New creates a new PostgresDB instance
func New(config PostgresConfig, logger logger.Logger) *PostgresDB {
	return &PostgresDB{
		config: config,
		logger: logger,
	}
}

// Connect establishes a connection to the database
func (p *PostgresDB) Connect(ctx context.Context) error {
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s pool_max_conns=%d",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.DBName,
		p.config.SSLMode,
		p.config.Schema,
		p.config.PoolMax,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("failed to parse postgres configuration: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(p.config.PoolMax)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	p.pool = pool
	p.logger.Info("Connected to PostgreSQL database", logger.Field{Key: "host", Value: p.config.Host})

	return nil
}

// Close closes the database connection
func (p *PostgresDB) Close() {
	if p.pool != nil {
		p.pool.Close()
		p.logger.Info("Closed PostgreSQL connection")
	}
}

// GetPool returns the connection pool
func (p *PostgresDB) GetPool() *pgxpool.Pool {
	return p.pool
}

// Health checks the health of the database
func (p *PostgresDB) Health(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("database connection not established")
	}

	return p.pool.Ping(ctx)
}
