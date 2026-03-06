package orchestrator

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/AlphaTechini/vector-db-migration/internal/mapper"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// TestMigrationOrchestratorInterface ensures orchestrator implements interface
func TestMigrationOrchestratorInterface(t *testing.T) {
	var _ MigrationOrchestrator = (*BaseOrchestrator)(nil)
	t.Log("✓ BaseOrchestrator implements MigrationOrchestrator interface")
}

// TestBaseOrchestrator_New creates orchestrator correctly
func TestBaseOrchestrator_New(t *testing.T) {
	orchestrator := NewBaseOrchestrator("test-123")

	if orchestrator.migrationID != "test-123" {
		t.Errorf("Expected migrationID 'test-123', got '%s'", orchestrator.migrationID)
	}

	if orchestrator.stats.Status != "not_started" {
		t.Errorf("Expected initial status 'not_started', got '%s'", orchestrator.stats.Status)
	}

	t.Log("✓ BaseOrchestrator initializes correctly")
}

// TestBaseOrchestrator_GetStatus tests status retrieval
func TestBaseOrchestrator_GetStatus(t *testing.T) {
	orchestrator := NewBaseOrchestrator("test-status")

	status, err := orchestrator.GetStatus("test-status")
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if status.Status != "not_started" {
		t.Errorf("Expected status 'not_started', got '%s'", status.Status)
	}

	if status.TotalRecords != 0 {
		t.Errorf("Expected TotalRecords 0, got %d", status.TotalRecords)
	}

	t.Log("✓ BaseOrchestrator retrieves status correctly")
}

// TestBaseOrchestrator_ValidateMapping tests validation
func TestBaseOrchestrator_Validate(t *testing.T) {
	orchestrator := NewBaseOrchestrator("test-validate")

	// Validate should not error on non-running migration
	err := orchestrator.Validate("test-validate")
	if err == nil {
		t.Log("✓ Validate method exists (implementation TODO)")
	}
}

// TestMigrationStats tests stats structure
func TestMigrationStats(t *testing.T) {
	stats := &MigrationStats{
		TotalRecords:    1000,
		MigratedRecords: 950,
	}

	if stats.TotalRecords != 1000 {
		t.Errorf("Expected TotalRecords 1000, got %d", stats.TotalRecords)
	}

	if stats.MigratedRecords != 950 {
		t.Errorf("Expected MigratedRecords 950, got %d", stats.MigratedRecords)
	}

	completionRate := float64(stats.MigratedRecords) / float64(stats.TotalRecords) * 100
	if completionRate != 95.0 {
		t.Errorf("Expected 95%% completion, got %.2f%%", completionRate)
	}

	t.Log("✓ MigrationStats structure works correctly")
}

// TestValidationError tests validation error structure
func TestValidationError(t *testing.T) {
	err := ValidationError{
		RecordID: "doc-123",
		Message:  "Cosine similarity below threshold",
	}

	if err.RecordID != "doc-123" {
		t.Errorf("Expected RecordID 'doc-123', got '%s'", err.RecordID)
	}

	if err.Message == "" {
		t.Error("Expected non-empty Message")
	}

	t.Log("✓ ValidationError structure works correctly")
}

// TestMigrationConfig tests config structure
func TestMigrationConfig(t *testing.T) {
	config := MigrationConfig{
		BatchSize:  100,
		MaxRetries: 3,
	}

	if config.BatchSize != 100 {
		t.Errorf("Expected BatchSize 100, got %d", config.BatchSize)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}

	t.Log("✓ MigrationConfig structure works correctly")
}

// Mock implementations for testing
type mockDatabase struct {
	records    []adapters.Record
	deletedIDs []string
}

func (m *mockDatabase) Connect(ctx context.Context, config adapters.DBConfig) error {
	return nil
}

func (m *mockDatabase) Close() error {
	return nil
}

func (m *mockDatabase) GetBatch(ctx context.Context, afterID string, limit int) ([]adapters.Record, error) {
	if len(m.records) == 0 {
		return []adapters.Record{}, nil
	}

	startIndex := 0
	if afterID != "" {
		for i, r := range m.records {
			if r.ID == afterID {
				startIndex = i + 1
				break
			}
		}
	}

	if startIndex >= len(m.records) {
		return []adapters.Record{}, nil
	}

	endIndex := startIndex + limit
	if endIndex > len(m.records) {
		endIndex = len(m.records)
	}

	return m.records[startIndex:endIndex], nil
}

func (m *mockDatabase) UpsertBatch(ctx context.Context, records []adapters.Record) error {
	return nil
}

func (m *mockDatabase) DeleteBatch(ctx context.Context, ids []string) error {
	m.deletedIDs = append(m.deletedIDs, ids...)
	return nil
}

func (m *mockDatabase) ValidateConnection(ctx context.Context) error {
	return nil
}

func (m *mockDatabase) GetStats(ctx context.Context) (*adapters.DBStats, error) {
	return &adapters.DBStats{TotalRecords: 0}, nil
}

func (m *mockDatabase) GetSourceURL() string {
	return "mock://test"
}

type mockMapper struct{}

func (m *mockMapper) CreateMapping(source, target map[string]interface{}) (*mapper.SchemaMapping, error) {
	return &mapper.SchemaMapping{}, nil
}

func (m *mockMapper) MapRecord(record adapters.Record, mapping *mapper.SchemaMapping) (adapters.Record, error) {
	return record, nil
}

func (m *mockMapper) MapBatch(records []adapters.Record, mapping *mapper.SchemaMapping) ([]adapters.Record, error) {
	return records, nil
}

func (m *mockMapper) ValidateMapping(mapping *mapper.SchemaMapping) error {
	return nil
}

func (m *mockMapper) GetSourceDB() string {
	return "mock"
}

func (m *mockMapper) GetTargetDB() string {
	return "mock"
}

type mockStateTracker struct {
	checkpoint *state.Checkpoint
}

func (m *mockStateTracker) GetState(migrationID string) (state.MigrationState, error) {
	return state.StateInProgress, nil
}

func (m *mockStateTracker) SetState(migrationID string, s state.MigrationState) error {
	return nil
}

func (m *mockStateTracker) GetCheckpoint(migrationID string) (*state.Checkpoint, error) {
	if m.checkpoint != nil {
		return m.checkpoint, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockStateTracker) SaveCheckpoint(checkpoint *state.Checkpoint) error {
	return nil
}

func (m *mockStateTracker) DeleteCheckpoint(migrationID string) error {
	return nil
}

func (m *mockStateTracker) Close() error {
	return nil
}

func (m *mockStateTracker) ListMigrations(statusFilter string, limit, offset int) ([]string, error) {
	return []string{}, nil
}

func (m *mockStateTracker) GetMigrationSummary(migrationID string) (*state.Checkpoint, error) {
	return nil, nil
}

// Compile-time assertions to ensure mocks implement the correct interfaces
var _ adapters.Database = (*mockDatabase)(nil)
var _ mapper.SchemaMapper = (*mockMapper)(nil)
var _ state.StateTracker = (*mockStateTracker)(nil)

// TestBaseOrchestrator_Rollback tests rollback logic
func TestBaseOrchestrator_Rollback(t *testing.T) {
orchestrator := NewBaseOrchestrator("test-rollback")

// Setup mocks
sourceDB := &mockDatabase{
records: []adapters.Record{
{ID: "doc1"}, {ID: "doc2"}, {ID: "doc3"}, {ID: "doc4"}, {ID: "doc5"},
},
}
targetDB := &mockDatabase{}
tracker := &mockStateTracker{
checkpoint: &state.Checkpoint{
MigrationID:     "test-rollback",
LastProcessedID: "doc3", // Rollback should stop here
},
}

orchestrator.config = MigrationConfig{
SourceDB:     sourceDB,
TargetDB:     targetDB,
StateTracker: tracker,
BatchSize:    2,
}

// Test rollback execution
err := orchestrator.Rollback("test-rollback")
if err != nil {
t.Fatalf("Failed to rollback: %v", err)
}

// Verify deleted IDs (order isn""t guaranteed due to concurrency, so check length and content)
if len(targetDB.deletedIDs) != 3 {
t.Errorf("Expected 3 deleted IDs, got %d", len(targetDB.deletedIDs))
}

deletedMap := make(map[string]bool)
for _, id := range targetDB.deletedIDs {
deletedMap[id] = true
}

for _, id := range []string{"doc1", "doc2", "doc3"} {
if !deletedMap[id] {
t.Errorf("Expected %s to be deleted", id)
}
}
if deletedMap["doc4"] || deletedMap["doc5"] {
t.Errorf("Did not expect doc4 or doc5 to be deleted")
}

t.Log(" BaseOrchestrator Rolls back correctly up to LastProcessedID")
}


// TestBaseOrchestrator_RollbackConcurrency tests concurrent workers
func TestBaseOrchestrator_RollbackConcurrency(t *testing.T) {
orchestrator := NewBaseOrchestrator("test-concurrency")

records := make([]adapters.Record, 50)
for i := 0; i < 50; i++ {
records[i] = adapters.Record{ID: fmt.Sprintf("doc%d", i)}
}

sourceDB := &mockDatabase{records: records}
targetDB := &mockDatabase{}
tracker := &mockStateTracker{
checkpoint: &state.Checkpoint{
MigrationID:     "test-concurrency",
LastProcessedID: "doc49", 
},
}

orchestrator.config = MigrationConfig{
SourceDB:     sourceDB,
TargetDB:     targetDB,
StateTracker: tracker,
BatchSize:    5, // 10 batches of 5
}

err := orchestrator.Rollback("test-concurrency")
if err != nil {
t.Fatalf("Failed to rollback: %v", err)
}

if len(targetDB.deletedIDs) != 50 {
t.Errorf("Expected 50 deleted IDs, got %d", len(targetDB.deletedIDs))
}
}

