package state

import (
	"os"
	"testing"
	"time"
)

func TestSQLiteTracker_GetSetState(t *testing.T) {
	// Create temp database
	tmpFile := "/tmp/test_tracker_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	migrationID := "test-migration-1"

	// Test initial state (should be NotStarted)
	state, err := tracker.GetState(migrationID)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if state != StateNotStarted {
		t.Errorf("Expected initial state to be NotStarted, got %s", state)
	}

	// Test setting state to InProgress
	err = tracker.SetState(migrationID, StateInProgress)
	if err != nil {
		t.Fatalf("Failed to set state: %v", err)
	}

	state, err = tracker.GetState(migrationID)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if state != StateInProgress {
		t.Errorf("Expected state to be InProgress, got %s", state)
	}

	// Test setting state to Completed
	err = tracker.SetState(migrationID, StateCompleted)
	if err != nil {
		t.Fatalf("Failed to set state: %v", err)
	}

	state, err = tracker.GetState(migrationID)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if state != StateCompleted {
		t.Errorf("Expected state to be Completed, got %s", state)
	}
}

func TestSQLiteTracker_Checkpoint(t *testing.T) {
	// Create temp database
	tmpFile := "/tmp/test_checkpoint_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	migrationID := "test-migration-2"
	startTime := time.Now()

	checkpoint := &Checkpoint{
		MigrationID:      migrationID,
		LastProcessedID:  "doc-100",
		TotalRecords:     1000,
		ProcessedCount:   100,
		FailedCount:      0,
		StartedAt:        startTime,
		LastCheckpointAt: time.Now(),
		SchemaMapping: map[string]interface{}{
			"source_type": "pinecone",
			"target_type": "qdrant",
		},
		ValidationStats: ValidationStats{
			SampledCount:        10,
			AvgCosineSimilarity: 0.987,
			MinCosineSimilarity: 0.982,
			MaxCosineSimilarity: 0.995,
		},
	}

	// Test saving checkpoint
	err = tracker.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Verify state was automatically set to InProgress
	state, err := tracker.GetState(migrationID)
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}
	if state != StateInProgress {
		t.Errorf("Expected state to be InProgress after checkpoint, got %s", state)
	}

	// Test retrieving checkpoint
	retrieved, err := tracker.GetCheckpoint(migrationID)
	if err != nil {
		t.Fatalf("Failed to get checkpoint: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected checkpoint to exist, got nil")
	}

	// Verify checkpoint data
	if retrieved.LastProcessedID != checkpoint.LastProcessedID {
		t.Errorf("Expected LastProcessedID=%s, got %s", checkpoint.LastProcessedID, retrieved.LastProcessedID)
	}
	if retrieved.ProcessedCount != checkpoint.ProcessedCount {
		t.Errorf("Expected ProcessedCount=%d, got %d", checkpoint.ProcessedCount, retrieved.ProcessedCount)
	}
	if retrieved.ValidationStats.AvgCosineSimilarity != checkpoint.ValidationStats.AvgCosineSimilarity {
		t.Errorf("Expected AvgCosineSimilarity=%.3f, got %.3f", checkpoint.ValidationStats.AvgCosineSimilarity, retrieved.ValidationStats.AvgCosineSimilarity)
	}

	// Test deleting checkpoint
	err = tracker.DeleteCheckpoint(migrationID)
	if err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Verify checkpoint is deleted
	retrieved, err = tracker.GetCheckpoint(migrationID)
	if err != nil {
		t.Fatalf("Failed to get checkpoint after deletion: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected checkpoint to be nil after deletion")
	}
}

func TestSQLiteTracker_MultipleMigrations(t *testing.T) {
	// Create temp database
	tmpFile := "/tmp/test_multi_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	// Create multiple migrations
	migrations := []string{"migration-1", "migration-2", "migration-3"}
	
	for i, id := range migrations {
		// Set different states
		states := []MigrationState{StateInProgress, StateCompleted, StateRolledBack}
		err := tracker.SetState(id, states[i])
		if err != nil {
			t.Fatalf("Failed to set state for %s: %v", id, err)
		}

		// Save checkpoints
		checkpoint := &Checkpoint{
			MigrationID:    id,
			ProcessedCount: int64((i + 1) * 100),
			StartedAt:      time.Now(),
		}
		err = tracker.SaveCheckpoint(checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint for %s: %v", id, err)
		}
	}

	// Verify all migrations have correct states
	expectedStates := []MigrationState{StateInProgress, StateCompleted, StateRolledBack}
	for i, id := range migrations {
		state, err := tracker.GetState(id)
		if err != nil {
			t.Fatalf("Failed to get state for %s: %v", id, err)
		}
		if state != expectedStates[i] {
			t.Errorf("Expected %s state to be %s, got %s", id, expectedStates[i], state)
		}

		checkpoint, err := tracker.GetCheckpoint(id)
		if err != nil {
			t.Fatalf("Failed to get checkpoint for %s: %v", id, err)
		}
		if checkpoint.ProcessedCount != int64((i + 1) * 100) {
			t.Errorf("Expected %s ProcessedCount=%d, got %d", id, (i+1)*100, checkpoint.ProcessedCount)
		}
	}
}
