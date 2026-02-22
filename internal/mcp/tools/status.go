package tools

import (
	"context"
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// MigrationStatusTool implements the migration_status MCP tool
type MigrationStatusTool struct {
	stateTracker state.StateTracker
}

// NewMigrationStatusTool creates a new migration_status tool
func NewMigrationStatusTool(stateTracker state.StateTracker) *MigrationStatusTool {
	return &MigrationStatusTool{
		stateTracker: stateTracker,
	}
}

// Register adds the tool to an MCP registry
func (t *MigrationStatusTool) Register(registry *mcp.ToolRegistry) error {
	return registry.Register(&mcp.Tool{
		Name:        "migration_status",
		Description: "Get the current status and progress of a migration",
		Schema:      t.inputSchema(),
		Handler:     t.execute,
	})
}

// inputSchema defines the JSON Schema for tool inputs
func (t *MigrationStatusTool) inputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"migration_id": map[string]interface{}{
				"type":        "string",
				"description": "The unique identifier of the migration",
				"examples":    []string{"mig-123", "migration-abc"},
			},
		},
		"required": []string{"migration_id"},
	}
}

// execute runs the migration_status tool
func (t *MigrationStatusTool) execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Validate inputs
	migrationID, ok := params["migration_id"].(string)
	if !ok || migrationID == "" {
		return nil, fmt.Errorf("migration_id is required and must be a non-empty string")
	}

	// Query state tracker for actual status
	checkpoint, err := t.stateTracker.GetCheckpoint(migrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}

	// Get migration state
	state, err := t.stateTracker.GetState(migrationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Build response
	response := map[string]interface{}{
		"migration_id":      migrationID,
		"status":            string(state),
		"batches_processed": 0,
		"started_at":        nil,
		"ended_at":          nil,
	}

	if checkpoint != nil {
		response["progress"] = map[string]interface{}{
			"total_records":    checkpoint.TotalRecords,
			"migrated_records": checkpoint.ProcessedCount,
			"percentage":       calculatePercentage(checkpoint.ProcessedCount, checkpoint.TotalRecords),
		}
		response["batches_processed"] = checkpoint.ProcessedCount / 100 // Assume 100 records per batch
		if !checkpoint.StartedAt.IsZero() {
			response["started_at"] = checkpoint.StartedAt.Format("2006-01-02T15:04:05Z")
		}
		if !checkpoint.LastCheckpointAt.IsZero() {
			response["ended_at"] = checkpoint.LastCheckpointAt.Format("2006-01-02T15:04:05Z")
		}
	} else {
		response["progress"] = map[string]interface{}{
			"total_records":    0,
			"migrated_records": 0,
			"percentage":       0.0,
		}
	}

	return response, nil
}

// calculatePercentage safely calculates percentage
func calculatePercentage(part, total int64) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(part) / float64(total) * 100.0
}
