package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// BaseOrchestrator provides common orchestration functionality
type BaseOrchestrator struct {
	config      MigrationConfig
	migrationID string
	mu          sync.RWMutex
	isRunning   bool
	isPaused    bool
	ctx         context.Context
	cancel      context.CancelFunc
	stats       *MigrationStats
}

// NewBaseOrchestrator creates a new base orchestrator
func NewBaseOrchestrator(migrationID string) *BaseOrchestrator {
	return &BaseOrchestrator{
		migrationID: migrationID,
		stats: &MigrationStats{
			Status: "not_started",
		},
	}
}

// Start begins the migration process
func (o *BaseOrchestrator) Start(ctx context.Context, config MigrationConfig) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if o.isRunning {
		return fmt.Errorf("migration already running")
	}
	
	o.config = config
	o.ctx, o.cancel = context.WithCancel(ctx)
	o.isRunning = true
	o.isPaused = false
	
	// Initialize stats
	o.stats = &MigrationStats{
		Status:    "in_progress",
		StartTime: time.Now().Format(time.RFC3339),
	}
	
	// Set initial state
	checkpoint := &state.Checkpoint{
		MigrationID:      o.migrationID,
		StartedAt:        time.Now(),
		LastCheckpointAt: time.Now(),
	}
	
	if err := config.StateTracker.SaveCheckpoint(checkpoint); err != nil {
		return fmt.Errorf("failed to save initial checkpoint: %w", err)
	}
	
	// Start migration in background
	go o.runMigration()
	
	return nil
}

// runMigration executes the migration logic
func (o *BaseOrchestrator) runMigration() {
	defer func() {
		o.mu.Lock()
		o.isRunning = false
		o.cancel()
		o.mu.Unlock()
	}()
	
	// Get source stats to know total records
	sourceStats, err := o.config.SourceDB.GetStats(o.ctx)
	if err != nil {
		o.fail(fmt.Sprintf("failed to get source stats: %v", err))
		return
	}
	
	o.mu.Lock()
	o.stats.TotalRecords = sourceStats.TotalRecords
	o.mu.Unlock()
	
	// Process batches
	batchNum := 0
	var afterID string
	
	for {
		// Check if paused or cancelled
		o.mu.RLock()
		if o.isPaused || o.ctx.Err() != nil {
			o.mu.RUnlock()
			return
		}
		o.mu.RUnlock()
		
		// Get next batch
		batchSize := o.config.BatchSize
		if batchSize == 0 {
			batchSize = 100 // Default
		}
		
		records, err := o.config.SourceDB.GetBatch(o.ctx, afterID, batchSize)
		if err != nil {
			o.fail(fmt.Sprintf("failed to get batch %d: %v", batchNum, err))
			return
		}
		
		if len(records) == 0 {
			// No more records, migration complete
			o.complete()
			return
		}
		
		// Map records to target schema
		mappedRecords, err := o.config.SchemaMapper.MapBatch(records, nil)
		if err != nil {
			o.fail(fmt.Sprintf("failed to map batch %d: %v", batchNum, err))
			return
		}
		
		// Upsert to target
		if err := o.config.TargetDB.UpsertBatch(o.ctx, mappedRecords); err != nil {
			o.fail(fmt.Sprintf("failed to upsert batch %d: %v", batchNum, err))
			return
		}
		
		// Update progress
		o.mu.Lock()
		o.stats.BatchesProcessed++
		o.stats.MigratedRecords += int64(len(records))
		if len(records) > 0 {
			afterID = records[len(records)-1].ID
		}
		
		// Save checkpoint every N batches
		validateEvery := o.config.ValidateEvery
		if validateEvery == 0 {
			validateEvery = 10
		}
		
		if batchNum%validateEvery == 0 {
			checkpoint := &state.Checkpoint{
				MigrationID:      o.migrationID,
				LastProcessedID:  afterID,
				TotalRecords:     o.stats.TotalRecords,
				ProcessedCount:   o.stats.MigratedRecords,
				FailedCount:      o.stats.FailedRecords,
				StartedAt:        parseTime(o.stats.StartTime),
				LastCheckpointAt: time.Now(),
			}
			
			if err := o.config.StateTracker.SaveCheckpoint(checkpoint); err != nil {
				o.mu.Unlock()
				o.fail(fmt.Sprintf("failed to save checkpoint: %v", err))
				return
			}
		}
		o.mu.Unlock()
		
		batchNum++
	}
}

// Pause pauses an in-progress migration
func (o *BaseOrchestrator) Pause(migrationID string) error {
	if migrationID != o.migrationID {
		return fmt.Errorf("migration ID mismatch")
	}
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if !o.isRunning {
		return fmt.Errorf("migration not running")
	}
	
	o.isPaused = true
	o.stats.Status = "paused"
	
	return nil
}

// Resume resumes a paused migration
func (o *BaseOrchestrator) Resume(migrationID string) error {
	if migrationID != o.migrationID {
		return fmt.Errorf("migration ID mismatch")
	}
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if !o.isPaused {
		return fmt.Errorf("migration not paused")
	}
	
	o.isPaused = false
	o.stats.Status = "in_progress"
	
	return nil
}

// Stop stops a migration gracefully
func (o *BaseOrchestrator) Stop(migrationID string) error {
	if migrationID != o.migrationID {
		return fmt.Errorf("migration ID mismatch")
	}
	
	o.mu.Lock()
	defer o.mu.Unlock()
	
	if !o.isRunning {
		return fmt.Errorf("migration not running")
	}
	
	o.cancel()
	o.stats.Status = "stopped"
	o.isRunning = false
	
	return nil
}

// Rollback rolls back a migration
func (o *BaseOrchestrator) Rollback(migrationID string) error {
	// TODO: Implement rollback logic
	// For now, just mark as rolled back
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.stats.Status = "rolled_back"
	
	if err := o.config.StateTracker.SetState(migrationID, state.StateRolledBack); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}
	
	return nil
}

// GetStatus returns current migration status
func (o *BaseOrchestrator) GetStatus(migrationID string) (*MigrationStats, error) {
	if migrationID != o.migrationID {
		return nil, fmt.Errorf("migration ID mismatch")
	}
	
	o.mu.RLock()
	defer o.mu.RUnlock()
	
	// Return a copy
	statsCopy := *o.stats
	return &statsCopy, nil
}

// Validate runs validation on migrated data
func (o *BaseOrchestrator) Validate(migrationID string) error {
	if migrationID != o.migrationID {
		return fmt.Errorf("migration ID mismatch")
	}
	
	// TODO: Implement validation logic
	// Sample records from source and target
	// Compare vectors (cosine similarity)
	// Compare metadata
	// Report discrepancies
	
	return nil
}

// complete marks migration as complete
func (o *BaseOrchestrator) complete() {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.stats.Status = "completed"
	o.stats.EndTime = time.Now().Format(time.RFC3339)
	o.isRunning = false
	
	// Save final checkpoint
	checkpoint := &state.Checkpoint{
		MigrationID:      o.migrationID,
		TotalRecords:     o.stats.TotalRecords,
		ProcessedCount:   o.stats.MigratedRecords,
		FailedCount:      o.stats.FailedRecords,
		StartedAt:        parseTime(o.stats.StartTime),
		LastCheckpointAt: time.Now(),
	}
	
	_ = o.config.StateTracker.SaveCheckpoint(checkpoint)
	_ = o.config.StateTracker.SetState(o.migrationID, state.StateCompleted)
}

// fail marks migration as failed
func (o *BaseOrchestrator) fail(reason string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	
	o.stats.Status = fmt.Sprintf("failed: %s", reason)
	o.stats.EndTime = time.Now().Format(time.RFC3339)
	o.isRunning = false
	
	_ = o.config.StateTracker.SetState(o.migrationID, state.StateFailed)
}

// parseTime parses RFC3339 time string
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// Ensure BaseOrchestrator implements MigrationOrchestrator
var _ MigrationOrchestrator = (*BaseOrchestrator)(nil)
