package db

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"rest_api_poc/internal/shared/logger"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seeds/*.sql
var seedsFS embed.FS

// RunSeeds executes all seed files in order
// Seeds are idempotent (use ON CONFLICT DO NOTHING) and safe to run multiple times
// This is separate from migrations - seeds are for development/testing data
func RunSeeds(ctx context.Context, pool *pgxpool.Pool, environment string) error {
	// Only run seeds in development/staging environments
	if environment == "production" {
		logger.Warn("Skipping seeds in production environment")
		return nil
	}

	logger.Info("Running database seeds...")

	// Read all seed files
	entries, err := seedsFS.ReadDir("seeds")
	if err != nil {
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	// Filter and sort SQL files
	var seedFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			seedFiles = append(seedFiles, entry.Name())
		}
	}
	sort.Strings(seedFiles)

	if len(seedFiles) == 0 {
		logger.Info("No seed files found")
		return nil
	}

	logger.Info("Found %d seed file(s)", len(seedFiles))

	// Execute each seed file
	for _, filename := range seedFiles {
		if err := executeSeedFile(ctx, pool, filename); err != nil {
			return fmt.Errorf("failed to execute seed file %s: %w", filename, err)
		}
	}

	logger.Info("All seeds executed successfully!")
	return nil
}

// executeSeedFile executes a single seed file
func executeSeedFile(ctx context.Context, pool *pgxpool.Pool, filename string) error {
	logger.Info("Executing seed: %s", filename)

	// Read seed file content
	content, err := seedsFS.ReadFile(filepath.Join("seeds", filename))
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	// Execute the SQL
	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute seed SQL: %w", err)
	}

	logger.Info("✓ Seed executed: %s", filename)
	return nil
}
