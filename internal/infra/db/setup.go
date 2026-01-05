package db

import (
	"context"
	"errors"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/shared/logger"
	"time"
)

var (
	// ErrNotInitialized is returned when database is not initialized
	ErrNotInitialized = errors.New("database not initialized")
)

// SetupDB initializes the database connection with retry mechanism and returns a dispose function
// for graceful shutdown. This function should be called from main and the returned dispose
// function should be deferred.
func SetupDB(ctx context.Context, cfg *config.DBConfig, environment string) (DB, func(ctx context.Context) error) {
	logger.InfoBlock("Setting up database...")

	// Initialize database with retry mechanism
	pool, err := initDB(ctx, cfg.ConnectionString, cfg.DBRetryCount)
	if err != nil {
		logger.FatalBlock("Failed to initialize database: %v", err)
	}

	// Create DB instance
	dbInstance := &dbImpl{
		pool: pool,
	}

	// Perform initial health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := dbInstance.Health(healthCtx); err != nil {
		logger.FatalBlock("Database health check failed: %v", err)
	}

	// Run database migrations
	if err := RunMigrations(cfg.ConnectionString); err != nil {
		logger.FatalBlock("Failed to run migrations: %v", err)
	}

	// Run database seeds (only in non-production environments)
	if err := RunSeeds(ctx, pool, environment); err != nil {
		logger.FatalBlock("Failed to run seeds: %v", err)
	}

	logger.SuccessBlock("Database setup completed successfully")

	// Return DB instance and dispose function for graceful shutdown
	return dbInstance, func(ctx context.Context) error {
		return disposeDB(ctx, dbInstance)
	}
}

// disposeDB gracefully closes the database connection pool
func disposeDB(ctx context.Context, dbInstance DB) error {
	logger.Warn("Shutting down database connection pool...")

	// Create timeout context for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Perform final health check before shutdown
	if err := dbInstance.Health(shutdownCtx); err != nil {
		logger.Warn("Database health check failed during shutdown: %v", err)
	}

	// Close the connection pool
	dbInstance.Close()

	logger.Success("Database connection pool stopped gracefully")
	return nil
}
