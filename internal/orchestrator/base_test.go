package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

type failingDatabase struct {
	mockDatabase
	failAtBatch int
	currentBatch int
}

func (f *failingDatabase) GetBatch(ctx context.Context, afterID string, limit int) ([]adapters.Record, error) {
	f.currentBatch++
	if f.currentBatch >= f.failAtBatch {
		return nil, fmt.Errorf("simulated fetch failure")
	}
	return f.mockDatabase.GetBatch(ctx, afterID, limit)
}

func (f *failingDatabase) UpsertBatch(ctx context.Context, records []adapters.Record) error {
	return fmt.Errorf("simulated upsert failure")
}

type trackingStateTracker struct {
	mockStateTracker
	mu sync.Mutex
	states []state.MigrationState
}

func (t *trackingStateTracker) SetState(migrationID string, s state.MigrationState) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.states = append(t.states, s)
	return nil
}

func TestBaseOrchestrator_LifecycleConcurrency(t *testing.T) {
	orchestrator := NewBaseOrchestrator("test-concurrency-lifecycle")

	records := make([]adapters.Record, 100)
	for i := 0; i < 100; i++ {
		records[i] = adapters.Record{ID: fmt.Sprintf("d%d", i)}
	}

	sourceDB := &mockDatabase{records: records}
	targetDB := &mockDatabase{}
	tracker := &mockStateTracker{}

	config := MigrationConfig{
		SourceDB:     sourceDB,
		TargetDB:     targetDB,
		StateTracker: tracker,
		SchemaMapper: &mockMapper{},
		BatchSize:    1,
	}

	ctx := context.Background()
	err := orchestrator.Start(ctx, config)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Rapidly toggle pause/resume
	for i := 0; i < 10; i++ {
		orchestrator.Pause("test-concurrency-lifecycle")
		orchestrator.Resume("test-concurrency-lifecycle")
		time.Sleep(10 * time.Millisecond)
	}

	status, _ := orchestrator.GetStatus("test-concurrency-lifecycle")
	if status.Status != "in_progress" && status.Status != "completed" {
		t.Errorf("Unexpected status during pause/resume: %s", status.Status)
	}

	orchestrator.Stop("test-concurrency-lifecycle")
	status, _ = orchestrator.GetStatus("test-concurrency-lifecycle")
	if status.Status != "stopped" && status.Status != "completed" {
		t.Errorf("Expected stopped or completed, got %s", status.Status)
	}
}

func TestBaseOrchestrator_FetchFailure(t *testing.T) {
	orchestrator := NewBaseOrchestrator("test-fetch-failure")

	records := []adapters.Record{{ID: "d1"}, {ID: "d2"}}
	sourceDB := &failingDatabase{
		mockDatabase: mockDatabase{records: records},
		failAtBatch: 1,
	}
	tracker := &trackingStateTracker{}

	config := MigrationConfig{
		SourceDB:     sourceDB,
		TargetDB:     &mockDatabase{},
		StateTracker: tracker,
		SchemaMapper: &mockMapper{},
		BatchSize:    1,
	}

	_ = orchestrator.Start(context.Background(), config)

	// Wait for failure
	time.Sleep(100 * time.Millisecond)

	status, _ := orchestrator.GetStatus("test-fetch-failure")
	if !stateContains(status.Status, "failed") {
		t.Errorf("Expected status to contain failed, got %s", status.Status)
	}

	tracker.mu.Lock()
	failed := false
	for _, s := range tracker.states {
		if s == state.StateFailed {
			failed = true
			break
		}
	}
	tracker.mu.Unlock()

	if !failed {
		t.Error("Expected StateFailed to be recorded in state tracker")
	}
}

func TestBaseOrchestrator_UpsertFailure(t *testing.T) {
	orchestrator := NewBaseOrchestrator("test-upsert-failure")

	records := []adapters.Record{{ID: "d1"}, {ID: "d2"}}
	sourceDB := &mockDatabase{records: records}
	targetDB := &failingDatabase{failAtBatch: 1}
	tracker := &trackingStateTracker{}

	config := MigrationConfig{
		SourceDB:     sourceDB,
		TargetDB:     targetDB,
		StateTracker: tracker,
		SchemaMapper: &mockMapper{},
		BatchSize:    1,
	}

	_ = orchestrator.Start(context.Background(), config)

	time.Sleep(100 * time.Millisecond)

	status, _ := orchestrator.GetStatus("test-upsert-failure")
	if !stateContains(status.Status, "failed") {
		t.Errorf("Expected status to contain failed, got %s", status.Status)
	}
}

func stateContains(status, substr string) bool {
	return (status != "" && (status == substr || (len(status) > len(substr) && status[:len(substr)] == substr) || (len(status) > 8 && status[:7] == "failed:")))
}
