package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Server struct {
		Port   int    `mapstructure:"port"`
		Host   string `mapstructure:"host"`
		Env    string `mapstructure:"env"`
		JWTKey string `mapstructure:"jwt_key"`
	} `mapstructure:"server"`
	
	// Database configuration
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"ssl_mode"`
		Schema   string `mapstructure:"schema"`
		PoolMax  int    `mapstructure:"pool_max"`
	} `mapstructure:"database"`
	
	// Firebase configuration
	Firebase struct {
		ProjectID       string `mapstructure:"project_id"`
		CredentialsFile string `mapstructure:"credentials_file"`
	} `mapstructure:"firebase"`
	
	// Hasura configuration
	Hasura struct {
		Endpoint       string `mapstructure:"endpoint"`
		AdminSecret    string `mapstructure:"admin_secret"`
		JWTNamespace   string `mapstructure:"jwt_namespace"`
		WebhookSecret  string `mapstructure:"webhook_secret"`
	} `mapstructure:"hasura"`
	
	// Logging configuration
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
}

// LoadConfig loads configuration from environment variables and config files
func LoadConfig(configName string, paths ...string) (*Config, error) {
	v := viper.New()
	
	// Set default values
	setDefaults(v)
	
	// Handle env variables
	v.SetEnvPrefix("MAMACARE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// Setup config file search paths
	for _, path := range paths {
		v.AddConfigPath(path)
	}
	
	// Add default search paths
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")
	
	// Setup config file name and type
	v.SetConfigName(configName)
	v.SetConfigType("yaml")
	
	// Try to read config file (not required if using env vars)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// It's okay if config file is not found, will use env vars and defaults
	}
	
	// Read and unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.env", "development")
	
	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.schema", "public")
	v.SetDefault("database.pool_max", 10)
	
	// Log defaults
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	
	// Hasura defaults
	v.SetDefault("hasura.jwt_namespace", "https://hasura.io/jwt/claims")
}
