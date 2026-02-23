package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// ListMigrationsTool implements the list_migrations MCP tool
type ListMigrationsTool struct {
	stateTracker state.StateTracker
}

// NewListMigrationsTool creates a new list_migrations tool
func NewListMigrationsTool(stateTracker state.StateTracker) *ListMigrationsTool {
	return &ListMigrationsTool{
		stateTracker: stateTracker,
	}
}

// Register adds the tool to an MCP registry
func (t *ListMigrationsTool) Register(registry *mcp.ToolRegistry) error {
	return registry.Register(&mcp.Tool{
		Name:        "list_migrations",
		Description: "List all migrations with optional filtering by status and date range",
		Schema:      t.inputSchema(),
		Handler:     t.execute,
	})
}

// inputSchema defines the JSON Schema for tool inputs
func (t *ListMigrationsTool) inputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type": "string",
				"description": "Filter by migration status (not_started, in_progress, completed, failed, rolled_back)",
				"enum": []string{"not_started", "in_progress", "completed", "failed", "rolled_back"},
			},
			"limit": map[string]interface{}{
				"type": "integer",
				"description": "Maximum number of migrations to return",
				"default": 50,
				"minimum": 1,
				"maximum": 500,
			},
			"offset": map[string]interface{}{
				"type": "integer",
				"description": "Number of migrations to skip (for pagination)",
				"default": 0,
				"minimum": 0,
			},
			"sort_by": map[string]interface{}{
				"type": "string",
				"description": "Field to sort by",
				"enum": []string{"created_at", "status", "migration_id"},
				"default": "created_at",
			},
			"sort_order": map[string]interface{}{
				"type": "string",
				"description": "Sort order",
				"enum": []string{"asc", "desc"},
				"default": "desc",
			},
		},
	}
}

// MigrationSummary is a simplified migration info for listing
type MigrationSummary struct {
	MigrationID string `json:"migration_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	Progress    *struct {
		Total    int64   `json:"total"`
		Current  int64   `json:"current"`
		Percent  float64 `json:"percent"`
	} `json:"progress,omitempty"`
}

// execute runs the list_migrations tool
func (t *ListMigrationsTool) execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Parse parameters
	statusFilter, _ := params["status"].(string)
	limit, ok := params["limit"].(float64)
	if !ok {
		limit = 50
	}
	offset, ok := params["offset"].(float64)
	if !ok {
		offset = 0
	}
	sortBy, _ := params["sort_by"].(string)
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder, _ := params["sort_order"].(string)
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Query state tracker for migrations
	statusStr := ""
	if statusFilter != "" && validateStatus(statusFilter) {
		statusStr = statusFilter
	}
	
	migrationIDs, err := t.stateTracker.ListMigrations(statusStr, int(limit)+int(offset), 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %w", err)
	}
	
	// Build migration summaries
	migrations := make([]MigrationSummary, 0, len(migrationIDs))
	for _, id := range migrationIDs {
		checkpoint, err := t.stateTracker.GetCheckpoint(id)
		if err != nil {
			continue // Skip if checkpoint not found
		}
		
		summary := MigrationSummary{
			MigrationID: id,
		}
		
		if checkpoint != nil {
			state, _ := t.stateTracker.GetState(id)
			summary.Status = string(state)
			
			if !checkpoint.StartedAt.IsZero() {
				summary.CreatedAt = checkpoint.StartedAt.Format(time.RFC3339)
			}
			
			if checkpoint.TotalRecords > 0 {
				percent := float64(checkpoint.ProcessedCount) / float64(checkpoint.TotalRecords) * 100.0
				summary.Progress = &struct {
					Total   int64   `json:"total"`
					Current int64   `json:"current"`
					Percent float64 `json:"percent"`
				}{
					Total:   checkpoint.TotalRecords,
					Current: checkpoint.ProcessedCount,
					Percent: percent,
				}
			}
		}
		
		migrations = append(migrations, summary)
	}
	
	// Apply sorting
	sort.Slice(migrations, func(i, j int) bool {
		switch sortBy {
		case "migration_id":
			if sortOrder == "desc" {
				return migrations[i].MigrationID > migrations[j].MigrationID
			}
			return migrations[i].MigrationID < migrations[j].MigrationID
		case "status":
			if sortOrder == "desc" {
				return migrations[i].Status > migrations[j].Status
			}
			return migrations[i].Status < migrations[j].Status
		default: // created_at
			if sortOrder == "desc" {
				return migrations[i].CreatedAt > migrations[j].CreatedAt
			}
			return migrations[i].CreatedAt < migrations[j].CreatedAt
		}
	})

	// Apply pagination
	start := int(offset)
	end := start + int(limit)
	if start > len(migrations) {
		migrations = []MigrationSummary{}
	} else if end > len(migrations) {
		migrations = migrations[start:]
	} else {
		migrations = migrations[start:end]
	}

	return map[string]interface{}{
		"migrations": migrations,
		"total":      len(migrationIDs),
		"limit":      limit,
		"offset":     offset,
	}, nil
}

// validateStatus checks if a status string is valid
func validateStatus(status string) bool {
	validStatuses := []string{"not_started", "in_progress", "completed", "failed", "rolled_back"}
	for _, s := range validStatuses {
		if strings.EqualFold(status, s) {
			return true
		}
	}
	return false
}
