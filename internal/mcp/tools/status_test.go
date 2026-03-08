package tools

import (
	"context"
	"testing"
	"time"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

func TestMigrationStatusTool_Register(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	registry := mcp.NewToolRegistry()

	err := tool.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Verify tool is registered
	retrieved, err := registry.Get("migration_status")
	if err != nil {
		t.Fatalf("Failed to get registered tool: %v", err)
	}

	if retrieved.Name != "migration_status" {
		t.Errorf("Expected name 'migration_status', got '%s'", retrieved.Name)
	}

	if retrieved.Description == "" {
		t.Error("Expected non-empty description")
	}

	if retrieved.Schema == nil {
		t.Error("Expected non-nil schema")
	}
}

func TestMigrationStatusTool_InputSchema(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	schema := tool.inputSchema()

	// Verify schema structure
	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	// Verify migration_id field
	migrationID, ok := props["migration_id"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected migration_id field in schema")
	}

	if migrationID["type"] != "string" {
		t.Errorf("Expected migration_id type 'string', got '%v'", migrationID["type"])
	}

	if migrationID["description"] == "" {
		t.Error("Expected migration_id description")
	}

	// Verify required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}

	if len(required) != 1 || required[0] != "migration_id" {
		t.Errorf("Expected ['migration_id'] as required, got %v", required)
	}
}

func TestMigrationStatusTool_Execute_Success(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"migration_id": "mig-123",
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	// Verify response structure
	if resultMap["migration_id"] != "mig-123" {
		t.Errorf("Expected migration_id 'mig-123', got '%v'", resultMap["migration_id"])
	}

	if resultMap["status"] != "not_started" {
		t.Errorf("Expected status 'not_started', got '%v'", resultMap["status"])
	}

	progress, ok := resultMap["progress"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected progress to be map[string]interface{}")
	}

	if progress["total_records"] != 0 {
		t.Errorf("Expected total_records 0, got %v", progress["total_records"])
	}

	if progress["percentage"] != 0.0 {
		t.Errorf("Expected percentage 0.0, got %v", progress["percentage"])
	}
}

func TestMigrationStatusTool_Execute_MissingParam(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	// Missing migration_id
	params := map[string]interface{}{}

	_, err := tool.execute(ctx, params)
	if err == nil {
		t.Error("Expected error for missing migration_id")
	}

	expectedMsg := "migration_id is required"
	if err.Error() != expectedMsg && err.Error() != "migration_id is required and must be a non-empty string" {
		t.Errorf("Expected error message about migration_id, got '%s'", err.Error())
	}
}

func TestMigrationStatusTool_Execute_EmptyParam(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	// Empty migration_id
	params := map[string]interface{}{
		"migration_id": "",
	}

	_, err := tool.execute(ctx, params)
	if err == nil {
		t.Error("Expected error for empty migration_id")
	}
}

func TestMigrationStatusTool_Execute_WeaklyTyped(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	// Weakly typed input (number instead of string) is supported
	params := map[string]interface{}{
		"migration_id": 123,
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Expected weakly typed numeric input to be cast to string successfully, got error: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["migration_id"] != "123" {
		t.Errorf("Expected migration_id '123', got '%v'", resultMap["migration_id"])
	}
}

func TestMigrationStatusTool_Execute_WithAllFields(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"migration_id": "mig-456",
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})

	// Verify all expected fields present
	expectedFields := []string{
		"migration_id",
		"status",
		"progress",
		"batches_processed",
		"started_at",
		"ended_at",
	}

	for _, field := range expectedFields {
		if _, exists := resultMap[field]; !exists {
			t.Errorf("Expected field '%s' in response", field)
		}
	}
}
func TestMigrationStatusTool_Execute_WithBatches(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()

	// 1. Save a checkpoint with specific BatchesProcessed
	checkpoint := &state.Checkpoint{
		MigrationID:      "mig-batches",
		TotalRecords:     1000,
		ProcessedCount:   500,
		BatchesProcessed: 5, // 100 per batch
		StartedAt:        time.Now(),
	}
	stateTracker.SaveCheckpoint(checkpoint)

	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"migration_id": "mig-batches",
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})

	// 2. Verify batches_processed matches exactly what we saved (not calculated)
	if resultMap["batches_processed"] != int64(5) {
		t.Errorf("Expected batches_processed 5, got %v", resultMap["batches_processed"])
	}

	// 3. Test with a different batch ratio to ensure no hardcoding
	checkpoint.MigrationID = "mig-custom"
	checkpoint.ProcessedCount = 500
	checkpoint.BatchesProcessed = 2 // maybe batch size 250
	stateTracker.SaveCheckpoint(checkpoint)

	params["migration_id"] = "mig-custom"
	result, _ = tool.execute(ctx, params)
	resultMap = result.(map[string]interface{})

	if resultMap["batches_processed"] != int64(2) {
		t.Errorf("Expected batches_processed 2, got %v", resultMap["batches_processed"])
	}
}
