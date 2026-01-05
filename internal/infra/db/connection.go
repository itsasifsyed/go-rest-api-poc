package db

import (
	"context"
	"rest_api_poc/internal/utils/logger"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB interface provides access to the database connection pool
type DB interface {
	// Pool returns the underlying pgxpool.Pool for direct database operations
	Pool() *pgxpool.Pool
	// Close gracefully closes the database connection pool
	Close()
	// Health checks if the database connection is healthy
	Health(ctx context.Context) error
}

// dbImpl implements the DB interface
type dbImpl struct {
	pool *pgxpool.Pool
}

// Pool returns the underlying connection pool
func (d *dbImpl) Pool() *pgxpool.Pool {
	return d.pool
}

// Close gracefully closes all connections in the pool
func (d *dbImpl) Close() {
	if d.pool != nil {
		logger.Info("Closing database connection pool...")
		d.pool.Close()
		logger.Success("Database connection pool closed successfully")
	}
}

// Health performs a health check on the database connection
func (d *dbImpl) Health(ctx context.Context) error {
	if d.pool == nil {
		return ErrNotInitialized
	}

	// Use a timeout context for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return d.pool.Ping(healthCtx)
}

// initDB initializes the database connection with retry mechanism
// It accepts a context to allow cancellation during initialization
func initDB(ctx context.Context, connectionString string, retryCount int) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error

	logger.InfoBlock("Initializing database connection...")
	logger.Info("Connection string: %s", maskConnectionString(connectionString))
	logger.Info("Retry count: %d", retryCount)

	// Configure connection pool with production-ready settings
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	// Production-ready pool configuration
	config.MaxConns = 25                      // Maximum number of connections
	config.MinConns = 5                       // Minimum number of connections
	config.MaxConnLifetime = time.Hour        // Maximum connection lifetime
	config.MaxConnIdleTime = 30 * time.Minute // Maximum idle time
	config.HealthCheckPeriod = time.Minute    // Health check interval
	// Connection timeout is handled via context in NewWithConfig

	// Retry mechanism with exponential backoff
	for attempt := 1; attempt <= retryCount; attempt++ {
		// Check if context is cancelled before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		logger.Info("Attempting to connect to database (attempt %d/%d)...", attempt, retryCount)

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			logger.Warn("Failed to create connection pool: %v", err)
			if attempt < retryCount {
				backoff := time.Duration(attempt) * time.Second
				logger.Info("Retrying in %v...", backoff)
				time.Sleep(backoff)
			}
			continue
		}

		// Test the connection with timeout derived from parent context
		pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err = pool.Ping(pingCtx)
		cancel()

		if err != nil {
			logger.Warn("Failed to ping database: %v", err)
			pool.Close()
			if attempt < retryCount {
				backoff := time.Duration(attempt) * time.Second
				logger.Info("Retrying in %v...", backoff)
				time.Sleep(backoff)
			}
			continue
		}

		// Success
		logger.Success("Database connection established successfully")
		logger.Info("Connection pool stats: MaxConns=%d, MinConns=%d", config.MaxConns, config.MinConns)
		return pool, nil
	}

	// All retries failed
	logger.ErrorBlock("Failed to connect to database after %d attempts", retryCount)
	return nil, err
}

// maskConnectionString masks sensitive information in connection string for logging
func maskConnectionString(connStr string) string {
	// Simple masking - in production, you might want more sophisticated masking
	// This masks passwords in connection strings
	if len(connStr) > 50 {
		return connStr[:20] + "..." + connStr[len(connStr)-20:]
	}
	return "***masked***"
}
