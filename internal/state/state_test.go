package state

import (
	"os"
	"testing"
	"time"
)

func TestSQLiteTracker_ListMigrations_FilteringAndSorting(t *testing.T) {
	tmpFile := "/tmp/test_list_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	// Create test data
	// migration-1: completed, 2020-01-01
	// migration-2: in_progress, 2020-01-02
	// migration-3: failed, 2020-01-03

	setupMigration(t, tracker, "m1", StateCompleted)
	setupMigration(t, tracker, "m2", StateInProgress)
	setupMigration(t, tracker, "m3", StateFailed)

	tests := []struct {
		name         string
		statusFilter string
		sortBy       string
		sortOrder    string
		limit        int
		offset       int
		expectedIDs  []string
	}{
		{
			name:        "All migrations, default sort (created_at DESC)",
			limit:       10,
			expectedIDs: []string{"m3", "m2", "m1"},
		},
		{
			name:         "Filter by completed",
			statusFilter: string(StateCompleted),
			limit:        10,
			expectedIDs:  []string{"m1"},
		},
		{
			name:      "Sort by ID ASC",
			sortBy:    "migration_id",
			sortOrder: "asc",
			limit:     10,
			expectedIDs: []string{"m1", "m2", "m3"},
		},
		{
			name:   "Pagination limit 1 offset 1 (default sort)",
			limit:  1,
			offset: 1,
			expectedIDs: []string{"m2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := tracker.ListMigrations(tt.statusFilter, tt.sortBy, tt.sortOrder, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("ListMigrations failed: %v", err)
			}
			if len(ids) != len(tt.expectedIDs) {
				t.Fatalf("Expected %d IDs, got %d", len(tt.expectedIDs), len(ids))
			}
			for i, id := range ids {
				if id != tt.expectedIDs[i] {
					t.Errorf("At index %d: expected %s, got %s", i, tt.expectedIDs[i], id)
				}
			}
		})
	}
}

func setupMigration(t *testing.T, tracker *SQLiteTracker, id string, state MigrationState) {
	// Use explicit INSERT to control created_at if necessary,
	// but here we just want to ensure they are inserted in order.
	query := `INSERT INTO migrations (migration_id, state, created_at) VALUES (?, ?, ?)`
	createdAt := time.Now().Add(time.Duration(len(id)) * time.Second) // Ensure different times
	_, err := tracker.db.Exec(query, id, state, createdAt)
	if err != nil {
		t.Fatalf("Failed to setup migration %s: %v", id, err)
	}
}

func TestSQLiteTracker_ComplexCheckpoint(t *testing.T) {
	tmpFile := "/tmp/test_complex_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	id := "complex-m"
	checkpoint := &Checkpoint{
		MigrationID:     id,
		LastProcessedID: "id-999",
		SchemaMapping: map[string]interface{}{
			"fields": map[string]interface{}{
				"title": "description",
				"tags":  "categories",
			},
			"conversions": []interface{}{"int-to-float", "string-to-bool"},
		},
		ValidationStats: ValidationStats{
			AvgCosineSimilarity: 0.9999,
		},
	}

	err = tracker.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	retrieved, err := tracker.GetCheckpoint(id)
	if err != nil {
		t.Fatalf("GetCheckpoint failed: %v", err)
	}

	if retrieved.SchemaMapping["fields"].(map[string]interface{})["title"] != "description" {
		t.Error("Complex schema mapping not preserved correctly")
	}

	if retrieved.ValidationStats.AvgCosineSimilarity != 0.9999 {
		t.Errorf("Validation stats not preserved correctly, got %f", retrieved.ValidationStats.AvgCosineSimilarity)
	}
}

func TestSQLiteTracker_GetTotalMigrationsCount(t *testing.T) {
	tmpFile := "/tmp/test_count_" + time.Now().Format("20060102_150405") + ".db"
	defer os.Remove(tmpFile)

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	setupMigration(t, tracker, "m1", StateCompleted)
	setupMigration(t, tracker, "m2", StateInProgress)
	setupMigration(t, tracker, "m3", StateInProgress)

	count, _ := tracker.GetTotalMigrationsCount("")
	if count != 3 {
		t.Errorf("Expected total count 3, got %d", count)
	}

	count, _ = tracker.GetTotalMigrationsCount(string(StateInProgress))
	if count != 2 {
		t.Errorf("Expected InProgress count 2, got %d", count)
	}
}

func TestNewSQLiteTracker_Error(t *testing.T) {
	// Try to open a database in a non-existent directory
	_, err := NewSQLiteTracker("/nonexistent/path/db.sqlite")
	if err == nil {
		t.Error("Expected error when opening database in non-existent directory, got nil")
	}
}
