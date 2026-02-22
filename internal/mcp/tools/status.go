package tools

import (
	"context"
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
)

// MigrationStatusTool implements the migration_status MCP tool
type MigrationStatusTool struct {
	// TODO: Inject orchestrator or state tracker
	// For now, this is a stub that will be wired up later
}

// NewMigrationStatusTool creates a new migration_status tool
func NewMigrationStatusTool() *MigrationStatusTool {
	return &MigrationStatusTool{}
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

	// TODO: Query orchestrator/state tracker for actual status
	// For now, return a stub response
	return map[string]interface{}{
		"migration_id": migrationID,
		"status":       "not_started",
		"progress": map[string]interface{}{
			"total_records":    0,
			"migrated_records": 0,
			"percentage":       0.0,
		},
		"batches_processed": 0,
		"started_at":        nil,
		"ended_at":          nil,
	}, nil
}
