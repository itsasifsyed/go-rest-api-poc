package main

import (
	"fmt"
	"os"
	"rest_api_poc/internal/infra/config"
	"rest_api_poc/internal/infra/db"
	"rest_api_poc/internal/shared/logger"
)

// CLI tool for managing database migrations
// Usage:
//   go run cmd/migrate/main.go up      # Apply all pending migrations
//   go run cmd/migrate/main.go down    # Rollback last migration
//   go run cmd/migrate/main.go status  # Show current migration status

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load configuration
	cfg := config.LoadConfig()

	command := os.Args[1]

	switch command {
	case "up":
		runMigrations(cfg.DB.ConnectionString)
	case "down":
		rollbackMigration(cfg.DB.ConnectionString)
	case "status":
		showStatus(cfg.DB.ConnectionString)
	default:
		logger.Error("Unknown command: %s", command)
		printUsage()
		os.Exit(1)
	}
}

func runMigrations(connectionString string) {
	logger.Info("Running migrations...")
	if err := db.RunMigrations(connectionString); err != nil {
		logger.Fatal("Migration failed: %v", err)
	}
	logger.Info("Migrations completed successfully!")
}

func rollbackMigration(connectionString string) {
	logger.Info("Rolling back last migration...")
	if err := db.RollbackMigration(connectionString); err != nil {
		logger.Fatal("Rollback failed: %v", err)
	}
	logger.Info("Rollback completed successfully!")
}

func showStatus(connectionString string) {
	logger.Info("Migration Status")
	logger.Info("Connection: %s", maskConnectionString(connectionString))
	logger.Info("\nTo check current version, connect to your database and run:")
	logger.Info("  SELECT * FROM schema_migrations;")
	logger.Info("\nTo view all tables:")
	logger.Info("  \\dt")
}

func printUsage() {
	fmt.Print(`
Database Migration Tool

Usage:
  go run cmd/migrate/main.go <command>

Commands:
  up      Apply all pending migrations
  down    Rollback the last migration (use with caution!)
  status  Show migration status information

Examples:
  go run cmd/migrate/main.go up
  go run cmd/migrate/main.go down
  go run cmd/migrate/main.go status

Note: Migrations also run automatically when starting the main application.
`)
}

func maskConnectionString(connStr string) string {
	if len(connStr) > 50 {
		return connStr[:20] + "..." + connStr[len(connStr)-20:]
	}
	return "***masked***"
}
