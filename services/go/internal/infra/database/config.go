package database

import (
	"fmt"
)

// Config contains database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// BuildConnectionString creates a PostgreSQL connection string from the configuration
func BuildConnectionString(config Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
	)
}
