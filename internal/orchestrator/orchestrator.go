package orchestrator

import (
	"context"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/AlphaTechini/vector-db-migration/internal/mapper"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	SourceDB      adapters.Database
	TargetDB      adapters.Database
	SchemaMapper  mapper.SchemaMapper
	StateTracker  state.StateTracker
	BatchSize     int
	MaxRetries    int
	ValidateEvery int // Validate every N batches
}

// MigrationStats tracks migration progress
type MigrationStats struct {
	TotalRecords     int64 `json:"total_records"`
	MigratedRecords  int64 `json:"migrated_records"`
	FailedRecords    int64 `json:"failed_records"`
	BatchesProcessed int64 `json:"batches_processed"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time,omitempty"`
	Status           string `json:"status"`
}

// MigrationOrchestrator interface for coordinating migrations
type MigrationOrchestrator interface {
	// Start begins the migration process
	Start(ctx context.Context, config MigrationConfig) error
	
	// Pause pauses an in-progress migration
	Pause(migrationID string) error
	
	// Resume resumes a paused migration
	Resume(migrationID string) error
	
	// Stop stops a migration gracefully
	Stop(migrationID string) error
	
	// Rollback rolls back a completed or failed migration
	Rollback(migrationID string) error
	
	// GetStatus returns current migration status
	GetStatus(migrationID string) (*MigrationStats, error)
	
	// Validate runs validation on migrated data
	Validate(migrationID string) error
}

// BatchProcessor handles batch operations
type BatchProcessor interface {
	// ProcessBatch processes a single batch of records
	ProcessBatch(ctx context.Context, batch []adapters.Record) error
	
	// GetProgress returns current batch processing progress
	GetProgress() (processed, total int64)
}

// ValidationResult holds validation results
type ValidationResult struct {
	// TotalRecords validated
	TotalRecords int64 `json:"total_records"`
	
	// ValidRecords passed validation
	ValidRecords int64 `json:"valid_records"`
	
	// InvalidRecords failed validation
	InvalidRecords int64 `json:"invalid_records"`
	
	// AvgCosineSimilarity average similarity score
	AvgCosineSimilarity float64 `json:"avg_cosine_similarity"`
	
	// MinCosineSimilarity minimum similarity score
	MinCosineSimilarity float64 `json:"min_cosine_similarity"`
	
	// Errors encountered during validation
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationError represents a validation failure
type ValidationError struct {
	RecordID string `json:"record_id"`
	Message  string `json:"message"`
	Field    string `json:"field,omitempty"`
}
