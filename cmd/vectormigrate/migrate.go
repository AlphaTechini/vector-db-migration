package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AlphaTechini/vector-db-migration/internal/orchestrator"
	"github.com/spf13/cobra"
)

var (
	sourceType     string
	sourceURL      string
	sourceAPIKey   string
	sourceIndex    string
	targetType     string
	targetURL      string
	targetAPIKey   string
	targetIndex    string
	batchSize      int
	maxRetries     int
	validateEvery  int
	dryRun         bool

	migrateCmd = &cobra.Command{
		Use:   "migrate [migration-id]",
		Short: "Start a migration",
		Long:  "Migrate data from source vector database to target with zero downtime.",
		Args:  cobra.ExactArgs(1),
		RunE:  runMigrate,
	}
)

func init() {
	// Source flags
	migrateCmd.Flags().StringVar(&sourceType, "source-type", "", "Source database type (pinecone, qdrant, weaviate)")
	migrateCmd.Flags().StringVar(&sourceURL, "source-url", "", "Source database URL")
	migrateCmd.Flags().StringVar(&sourceAPIKey, "source-api-key", "", "Source database API key")
	migrateCmd.Flags().StringVar(&sourceIndex, "source-index", "", "Source index/collection name")
	migrateCmd.MarkFlagRequired("source-type")
	migrateCmd.MarkFlagRequired("source-url")
	migrateCmd.MarkFlagRequired("source-index")

	// Target flags
	migrateCmd.Flags().StringVar(&targetType, "target-type", "", "Target database type (pinecone, qdrant, weaviate)")
	migrateCmd.Flags().StringVar(&targetURL, "target-url", "", "Target database URL")
	migrateCmd.Flags().StringVar(&targetAPIKey, "target-api-key", "", "Target database API key")
	migrateCmd.Flags().StringVar(&targetIndex, "target-index", "", "Target index/collection name")
	migrateCmd.MarkFlagRequired("target-type")
	migrateCmd.MarkFlagRequired("target-url")
	migrateCmd.MarkFlagRequired("target-index")

	// Migration options
	migrateCmd.Flags().IntVar(&batchSize, "batch-size", 100, "Number of records per batch")
	migrateCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum retry attempts per batch")
	migrateCmd.Flags().IntVar(&validateEvery, "validate-every", 10, "Validate every N batches")
	migrateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate migration without writing")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	migrationID := args[0]

	// Validate database types
	if err := validateDatabaseType(sourceType); err != nil {
		return fmt.Errorf("invalid source type: %w", err)
	}
	if err := validateDatabaseType(targetType); err != nil {
		return fmt.Errorf("invalid target type: %w", err)
	}

	log.Printf("🚀 Starting migration: %s", migrationID)
	log.Printf("   Source: %s (%s)", sourceType, sourceIndex)
	log.Printf("   Target: %s (%s)", targetType, targetIndex)
	log.Printf("   Batch size: %d", batchSize)
	log.Printf("   Validate every: %d batches", validateEvery)

	if dryRun {
		log.Println("   📝 DRY RUN - no data will be written")
		return nil
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Initialize components
	log.Println("   🔧 Initializing components...")

	sourceDB, err := createDatabase(sourceType, sourceURL, sourceAPIKey, sourceIndex, 30)
	if err != nil {
		return err
	}
	defer sourceDB.Close()

	targetDB, err := createDatabase(targetType, targetURL, targetAPIKey, targetIndex, 30)
	if err != nil {
		return err
	}
	defer targetDB.Close()

	schemaMapper, err := createMapper(sourceType, targetType)
	if err != nil {
		return err
	}

	stateTracker, err := createStateTracker("")
	if err != nil {
		return err
	}
	defer stateTracker.Close()

	// Create orchestrator
	migrator := createOrchestrator(migrationID)

	// Configure migration
	orchConfig := orchestrator.MigrationConfig{
		SourceDB:      sourceDB,
		TargetDB:      targetDB,
		SchemaMapper:  schemaMapper,
		StateTracker:  stateTracker,
		BatchSize:     batchSize,
		MaxRetries:    maxRetries,
		ValidateEvery: validateEvery,
	}

	// Start migration
	log.Println("   ▶️  Starting migration...")
	if err := migrator.Start(ctx, orchConfig); err != nil {
		return fmt.Errorf("failed to start migration: %w", err)
	}

	// Monitor progress
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⚠️  Migration cancelled")
			return ctx.Err()

		case t := <-ticker.C:
			_ = t // Use ticker time
			status, err := migrator.GetStatus(migrationID)
			if err != nil {
				return fmt.Errorf("failed to get status: %w", err)
			}

			var progress float64
			if status.TotalRecords > 0 {
				progress = float64(status.MigratedRecords) / float64(status.TotalRecords) * 100
			}

			log.Printf("   📊 Progress: %d/%d records (%.1f%%) - Status: %s",
				status.MigratedRecords, status.TotalRecords, progress, status.Status)

			if status.Status == "completed" {
				log.Printf("✅ Migration completed successfully!")
				log.Printf("   Total: %d records, %d batches", status.MigratedRecords, status.BatchesProcessed)
				return nil
			}

			if status.Status == "failed" || status.Status == "stopped" {
				return fmt.Errorf("migration %s", status.Status)
			}
		}
	}
}

// validateDatabaseType checks if the database type is supported
func validateDatabaseType(dbType string) error {
	supportedTypes := map[string]bool{
		"pinecone": true,
		"qdrant":   true,
		"weaviate": true,
		"milvus":   true,
	}

	if !supportedTypes[dbType] {
		return fmt.Errorf("unsupported database type: %s (supported: pinecone, qdrant, weaviate, milvus)", dbType)
	}
	return nil
}
