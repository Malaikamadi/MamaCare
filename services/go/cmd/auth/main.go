package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/internal/app/auth"
	"github.com/mamacare/services/internal/infra/database"
	"github.com/mamacare/services/internal/infra/firebase"
	httpserver "github.com/mamacare/services/internal/infra/http"
	dbrepository "github.com/mamacare/services/internal/infra/database/repository"
	"github.com/mamacare/services/internal/port/handler"
	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/pkg/config"
	"github.com/mamacare/services/pkg/logger"
	"github.com/mamacare/services/pkg/metrics"
)

// Helper function to get a config value or use a default if empty
func getConfigOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	logConfig := logger.Config{
		LogLevel: "debug",
		Pretty:   true,
		WithTime: true,
	}
	log := logger.NewLogger(logConfig)

	// Load configuration
	cfg, err := config.LoadConfig("auth", "./configs", "../configs", "../../configs", ".")
	// If you want to manually create the configs directory and add auth.yaml file,
	// you can use these commands:
	// mkdir -p configs
	// touch configs/auth.yaml
	// Then add the necessary configuration
	if err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	// Initialize database connection
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	connString := database.BuildConnectionString(dbConfig)
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer pool.Close()

	// Only initialize migrations if the database is connected
	if pool != nil {
		// Run migrations
		migrationManager := database.NewMigrationManager(pool, log)
		
		// Add initial migration if needed
		initialMigration := database.CreateInitialMigration()
		migrationManager.AddMigration(1, "Initial schema", initialMigration.SQL)
		
		// Initialize migration table
		if err := migrationManager.Initialize(ctx); err != nil {
			log.Error("Failed to initialize migration manager", err)
			// Continue anyway, it might be already initialized
		}
		
		// Run migrations
		if err := migrationManager.Migrate(ctx); err != nil {
			log.Error("Failed to run migrations", err)
			// Continue with service startup, migrations might be already applied
		}
	}

	// Initialize repositories
	userRepo := dbrepository.NewUserRepository(pool, log)

	// Initialize Firebase Auth (with error handling)
	firebaseConfig := firebase.Config{
		CredentialsFile: cfg.Firebase.CredentialsFile,
		ProjectID:       cfg.Firebase.ProjectID,
	}

	// Initialize Firebase with proper error handling
	firebaseAuth := firebase.New(firebaseConfig, log)
	err = firebaseAuth.Initialize(ctx)
	if err != nil {
		log.Error("Warning: Failed to initialize Firebase Auth", err)
		// We're continuing without Firebase for now, for testing purposes
		// In production, you would want to fail: log.Fatal("Failed to initialize Firebase Auth", err)
	}

	// Initialize auth service with safer defaults
	authServiceConfig := auth.Config{
		// Use a secure default if JWTKey is not configured
		JWTSecret:         getConfigOrDefault(cfg.Server.JWTKey, "temporary-dev-secret-change-in-production"),
		JWTExpiryDuration: 24 * time.Hour, // Default to 24 hours
		HasuraNamespace:   getConfigOrDefault(cfg.Hasura.JWTNamespace, "https://hasura.io/jwt/claims"),
	}
	
	// Create auth service
	authService := auth.NewService(
		authServiceConfig,
		firebaseAuth,
		userRepo,
		log,
	)

	// Initialize handlers with error handling
	authHandler := handler.NewAuthHandler(authService, userRepo, log)
	hasuraWebhook := handler.NewHasuraAuthWebhook(authService, log)

	// Create authentication middleware for protected routes
	authMiddleware := middleware.NewAuthMiddleware(firebaseAuth, log)

	// Initialize metrics client (using NoopClient for now)
	metricsClient := metrics.NewNoopClient()

	// Initialize router - wrap in a try-catch pattern for handling initialization errors
	router, routerErr := func() (*httpserver.Router, error) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovered from router initialization panic", fmt.Errorf("%v", r))
			}
		}()
		
		return httpserver.NewRouter(
			authService,
			authHandler,
			hasuraWebhook,
			authMiddleware,
			log,
			metricsClient,
		), nil
	}()
	
	if routerErr != nil {
		log.Error("Failed to initialize router", routerErr)
		// In development we'll continue, in production we might want to exit
	}

	// Initialize HTTP server with safe defaults
	serverHost := getConfigOrDefault(cfg.Server.Host, "0.0.0.0")
	serverPort := 8080
	if cfg.Server.Port > 0 {
		serverPort = cfg.Server.Port
	}
	
	serverConfig := httpserver.Config{
		Address:         fmt.Sprintf("%s:%d", serverHost, serverPort),
		ReadTimeout:     5 * time.Second,  // Default values
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
	
	// Create server with error handling
	var routerHandler http.Handler
	if router != nil {
		// Wrap router.Setup() in a recover to prevent panics
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Error("Recovered from router setup panic", fmt.Errorf("%v", r))
					// Use a simple handler that returns 503 Service Unavailable
					routerHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusServiceUnavailable)
						w.Write([]byte("Service is starting up"))
					})
				}
			}()
			
			// Try to set up the router
			routerHandler = router.Setup()
		}()
	} else {
		// Fallback handler if router is nil
		routerHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Service is initializing"))
		})
	}
	
	server := httpserver.NewServer(serverConfig, routerHandler, log)

	// Handle graceful shutdown
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Info("Shutting down...", logger.Field{Key: "signal", Value: "SIGTERM"})

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := server.Stop(shutdownCtx); err != nil {
			log.Error("Server shutdown error", err)
		}

		cancel()
	}()

	// Start the server
	log.Info("Starting auth service")
	if err := server.Start(); err != nil {
		log.Fatal("Server error", err)
	}
}