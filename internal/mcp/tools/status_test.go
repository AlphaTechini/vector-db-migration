package tools

import (
	"context"
	"testing"

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

func TestMigrationStatusTool_Execute_WrongType(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewMigrationStatusTool(stateTracker)
	ctx := context.Background()

	// Wrong type (number instead of string)
	params := map[string]interface{}{
		"migration_id": 123,
	}

	_, err := tool.execute(ctx, params)
	if err == nil {
		t.Error("Expected error for wrong type")
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
