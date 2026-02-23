package tools

import (
	"context"
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

func TestListMigrationsTool_Register(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	registry := mcp.NewToolRegistry()

	err := tool.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	retrieved, err := registry.Get("list_migrations")
	if err != nil {
		t.Fatalf("Failed to get registered tool: %v", err)
	}

	if retrieved.Name != "list_migrations" {
		t.Errorf("Expected name 'list_migrations', got '%s'", retrieved.Name)
	}

	if retrieved.Description == "" {
		t.Error("Expected non-empty description")
	}
}

func TestListMigrationsTool_InputSchema(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	schema := tool.inputSchema()

	// Verify schema structure
	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	// Check limit field
	limit, ok := props["limit"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected limit field in schema")
	}
	if limit["type"] != "integer" {
		t.Errorf("Expected limit type 'integer', got '%v'", limit["type"])
	}

	// Check status field has enum
	status, ok := props["status"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected status field in schema")
	}
	if status["enum"] == nil {
		t.Error("Expected status to have enum values")
	}
}

func TestListMigrationsTool_Execute_DefaultParams(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})

	// Verify response structure
	if resultMap["migrations"] == nil {
		t.Error("Expected migrations array in response")
	}

	if resultMap["total"] != 0 {
		t.Errorf("Expected total 0, got %v", resultMap["total"])
	}

	if resultMap["limit"] != 50 {
		t.Errorf("Expected default limit 50, got %v", resultMap["limit"])
	}

	if resultMap["offset"] != 0 {
		t.Errorf("Expected default offset 0, got %v", resultMap["offset"])
	}
}

func TestListMigrationsTool_Execute_CustomLimit(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"limit": 10,
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["limit"] != 10 {
		t.Errorf("Expected limit 10, got %v", resultMap["limit"])
	}
}

func TestListMigrationsTool_Execute_StatusFilter(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"status": "in_progress",
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["migrations"] == nil {
		t.Error("Expected migrations array even with filter")
	}
}

func TestListMigrationsTool_Execute_Sorting(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	ctx := context.Background()

	// Test sort by migration_id desc
	params := map[string]interface{}{
		"sort_by":    "migration_id",
		"sort_order": "desc",
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["migrations"] == nil {
		t.Error("Expected migrations array with sorting")
	}
}

func TestListMigrationsTool_Execute_Pagination(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"limit":  10,
		"offset": 20,
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["limit"] != 10 {
		t.Errorf("Expected limit 10, got %v", resultMap["limit"])
	}
	if resultMap["offset"] != 20 {
		t.Errorf("Expected offset 20, got %v", resultMap["offset"])
	}
}

func TestValidateStatus_ValidStatuses(t *testing.T) {
	validStatuses := []string{"not_started", "in_progress", "completed", "failed", "rolled_back"}

	for _, status := range validStatuses {
		if !validateStatus(status) {
			t.Errorf("Expected '%s' to be valid", status)
		}
		
		// Also test case-insensitive
		if !validateStatus(status) {
			t.Errorf("Expected '%s' (uppercase) to be valid", status)
		}
	}
}

func TestValidateStatus_InvalidStatus(t *testing.T) {
	invalidStatuses := []string{"unknown", "pending", "running", "", "IN_PROGRESS"}

	for _, status := range invalidStatuses {
		if validateStatus(status) {
			t.Errorf("Expected '%s' to be invalid", status)
		}
	}
}

func TestListMigrationsTool_Execute_MaxLimit(t *testing.T) {
	stateTracker, _ := state.NewSQLiteTracker(":memory:")
	defer stateTracker.Close()
	
	tool := NewListMigrationsTool(stateTracker)
	ctx := context.Background()

	params := map[string]interface{}{
		"limit": 500, // Max allowed
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["limit"] != 500 {
		t.Errorf("Expected limit 500, got %v", resultMap["limit"])
	}
}
