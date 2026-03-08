package orchestrator

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
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
				BatchesProcessed: o.stats.BatchesProcessed,
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

// Rollback rolls back a migration by deleting migrated records from the target database
func (o *BaseOrchestrator) Rollback(migrationID string) error {
	if migrationID != o.migrationID {
		return fmt.Errorf("migration ID mismatch")
	}

	o.mu.Lock()
	if o.isRunning {
		o.mu.Unlock()
		return fmt.Errorf("cannot rollback a running migration; stop it first")
	}

	// Mark as rolling back
	o.isRunning = true
	o.isPaused = false
	o.stats.Status = "rolling_back"

	// Always create a fresh context for the rollback operation
	o.ctx, o.cancel = context.WithCancel(context.Background())
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		o.isRunning = false
		if o.stats.Status == "rolling_back" {
			// If not updated to rolled_back or failed, assume interrupted
			o.stats.Status = "rollback_interrupted"
		}
		o.mu.Unlock()
	}()

	// 1. Get the checkpoint to know where to stop
	checkpoint, err := o.config.StateTracker.GetCheckpoint(migrationID)
	if err != nil {
		return fmt.Errorf("failed to get checkpoint: %w", err)
	}
	if checkpoint == nil || checkpoint.LastProcessedID == "" {
		return fmt.Errorf("no checkpoint or LastProcessedID found; nothing to rollback")
	}
	targetStopID := checkpoint.LastProcessedID

	// 2. Setup Concurrency (Producer/Consumer)
	numWorkers := 5 // configurable in future
	idChan := make(chan []string, numWorkers*2)
	errChan := make(chan error, numWorkers)
	var wg sync.WaitGroup

	// Start Consumer Workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ids := range idChan {
				// Delete batch from target
				if err := o.config.TargetDB.DeleteBatch(o.ctx, ids); err != nil {
					select {
					case errChan <- fmt.Errorf("failed to delete batch (size %d): %w", len(ids), err):
					default:
					}
					// Signal cancellation so producer stops too
					if o.cancel != nil {
						o.cancel()
					}
					return
				}
			}
		}()
	}

	// 3. Producer: Scan source and feed consumers
	go func() {
		defer close(idChan)
		var afterID string
		batchSize := o.config.BatchSize
		if batchSize == 0 {
			batchSize = 100
		}

		for {
			// Check for cancellation or pause
			o.mu.RLock()
			paused := o.isPaused
			cancelled := o.ctx.Err() != nil
			o.mu.RUnlock()

			if paused || cancelled {
				return
			}

			// Fetch batch from source
			records, err := o.config.SourceDB.GetBatch(o.ctx, afterID, batchSize)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("rollback scanning failed at %s: %w", afterID, err):
				default:
				}
				return
			}

			if len(records) == 0 {
				return // End of source reached before hitting targetStopID
			}

			// Extract IDs and check if we hit the stopping boundary
			ids := make([]string, 0, len(records))
			hitBoundary := false

			for _, r := range records {
				ids = append(ids, r.ID)
				afterID = r.ID
				if r.ID == targetStopID {
					hitBoundary = true
					break
				}
			}

			// Push to workers
			select {
			case <-o.ctx.Done():
				return // Context cancelled while waiting to push
			case idChan <- ids:
				// Pushed successfully
			}

			if hitBoundary {
				return // We reached the last processed ID, stop scanning
			}
		}
	}()

	// Wait for workers to finish
	wg.Wait()

	// Check if any errors occurred
	select {
	case err := <-errChan:
		o.fail(fmt.Sprintf("rollback failed: %v", err))
		return err
	default:
	}

	// Double check context error
	if o.ctx.Err() != nil {
		return fmt.Errorf("rollback interrupted: %w", o.ctx.Err())
	}

	// 4. Update Final State
	o.mu.Lock()
	o.stats.Status = "rolled_back"
	o.mu.Unlock()

	if err := o.config.StateTracker.SetState(migrationID, state.StateRolledBack); err != nil {
		return fmt.Errorf("failed to update state to rolled_back: %w", err)
	}

	// Delete the checkpoint since data is gone
	_ = o.config.StateTracker.DeleteCheckpoint(migrationID)

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

// Validate runs validation on migrated data.
// It uses a sampling strategy by default but can be configured for a full scan.
func (o *BaseOrchestrator) Validate(migrationID string) error {
	if migrationID != o.migrationID {
		return fmt.Errorf("migration ID mismatch")
	}

	o.mu.Lock()
	if o.isRunning {
		o.mu.Unlock()
		return fmt.Errorf("cannot validate a running migration; stop it first")
	}

	// Defensive check
	if o.config.SourceDB == nil || o.config.TargetDB == nil {
		o.mu.Unlock()
		return fmt.Errorf("databases not configured for this migration")
	}

	// Always create a fresh context for the validation operation
	o.ctx, o.cancel = context.WithCancel(context.Background())

	o.isRunning = true
	o.stats.Status = "validating"
	o.mu.Unlock()

	defer func() {
		o.mu.Lock()
		o.isRunning = false
		o.mu.Unlock()
	}()

	// Use configured method, default to sampling
	method := strings.ToLower(o.config.ValidationMethod)
	sampleSize := o.config.SampleSize
	if sampleSize <= 0 {
		sampleSize = 100 // Default sample size
	}

	if method == "full" {
		return o.validateFull()
	}

	return o.validateSampling(sampleSize)
}

func (o *BaseOrchestrator) validateFull() error {
	// Full scan involves streaming both DBs.
	// This is O(N) but 100% accurate.

	const batchSize = 100
	var afterID string
	var totalSimilarity float64
	var totalCount int64
	var validCount int64
	threshold := 0.99

	for {
		select {
		case <-o.ctx.Done():
			return o.ctx.Err()
		default:
		}

		// 1. Get batch from source
		sourceBatch, err := o.config.SourceDB.GetBatch(o.ctx, afterID, batchSize)
		if err != nil {
			return fmt.Errorf("failed to fetch source batch: %w", err)
		}
		if len(sourceBatch) == 0 {
			break
		}

		// 2. Fetch counterparts from Target using GetByIDs (more efficient than full target scan)
		ids := make([]string, 0, len(sourceBatch))
		sourceMap := make(map[string]adapters.Record)
		for _, r := range sourceBatch {
			ids = append(ids, r.ID)
			sourceMap[r.ID] = r
			afterID = r.ID
		}

		targetBatch, err := o.config.TargetDB.GetByIDs(o.ctx, ids)
		if err != nil {
			return fmt.Errorf("failed to fetch target batch: %w", err)
		}

		// 3. Compare
		for _, tr := range targetBatch {
			sr, ok := sourceMap[tr.ID]
			if !ok {
				continue
			}

			sim := CosineSimilarity(sr.Vector, tr.Vector)
			totalSimilarity += float64(sim)
			totalCount++

			if sim >= float32(threshold) {
				validCount++
			}
		}

		// Update progress
		o.mu.Lock()
		o.stats.Status = fmt.Sprintf("validating: %d records checked...", totalCount)
		o.mu.Unlock()
	}

	o.mu.Lock()
	if totalCount > 0 {
		avgSim := totalSimilarity / float64(totalCount)
		o.stats.Status = fmt.Sprintf("validated_full: %d/%d records passed (avg sim: %.4f)", validCount, totalCount, avgSim)
	} else {
		o.stats.Status = "validated_full: no records found"
	}
	o.mu.Unlock()

	return nil
}

func (o *BaseOrchestrator) validateSampling(sampleSize int) error {
	// 1. Get Source Batch to pick random IDs
	// Actually, for simplicity in this MVP, we just take the FIRST N records.
	// A more robust sampler would use GetStats to find the range and pick truly random.
	records, err := o.config.SourceDB.GetBatch(o.ctx, "", sampleSize)
	if err != nil {
		return fmt.Errorf("failed to fetch sampling source: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no records found in source to validate")
	}

	ids := make([]string, 0, len(records))
	sourceMap := make(map[string]adapters.Record)
	for _, r := range records {
		ids = append(ids, r.ID)
		sourceMap[r.ID] = r
	}

	// 2. Fetch counterparts from Target
	targetRecords, err := o.config.TargetDB.GetByIDs(o.ctx, ids)
	if err != nil {
		return fmt.Errorf("failed to fetch target records for validation: %w", err)
	}

	// 3. Compare
	var totalSimilarity float64
	var validCount int
	threshold := 0.99 // can be configurable

	for _, tr := range targetRecords {
		sr, ok := sourceMap[tr.ID]
		if !ok {
			continue
		}

		sim := CosineSimilarity(sr.Vector, tr.Vector)
		totalSimilarity += float64(sim)

		if sim >= float32(threshold) {
			validCount++
		}
	}

	o.mu.Lock()
	if validCount == len(targetRecords) && len(targetRecords) > 0 {
		o.stats.Status = fmt.Sprintf("validated: %d/%d records passed (avg sim: %.4f)", validCount, len(targetRecords), totalSimilarity/float64(len(targetRecords)))
	} else {
		o.stats.Status = fmt.Sprintf("validation_failed: only %d/%d passed", validCount, len(targetRecords))
	}
	o.mu.Unlock()

	return nil
}

// CosineSimilarity calculates the cosine similarity between two vectors.
// It uses a simple loop that is easy for the compiler to optimize (SIMD-friendly).
func CosineSimilarity(v1, v2 []float32) float32 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0
	}

	var dotProduct, mag1, mag2 float32
	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
		mag1 += v1[i] * v1[i]
		mag2 += v2[i] * v2[i]
	}

	sqrtMag := float32(math.Sqrt(float64(mag1)) * math.Sqrt(float64(mag2)))
	if sqrtMag == 0 {
		return 0
	}
	return dotProduct / sqrtMag
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
		BatchesProcessed: o.stats.BatchesProcessed,
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
