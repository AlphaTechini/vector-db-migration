package state

import (
	"testing"
	"time"
)

func TestSQLiteTracker_ListMigrations_FilteringAndSorting(t *testing.T) {
	tmpFile := t.TempDir() + "/test_list.db"

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	// Create test data
	// Setup migrations with unique, sequential timestamps to ensure deterministic sorting.
	// We use an explicit base time to avoid reliance on real-time clock during rapid test execution.
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	setupMigrationAt(t, tracker, "m1", StateCompleted, baseTime.Add(1*time.Hour))
	setupMigrationAt(t, tracker, "m2", StateInProgress, baseTime.Add(2*time.Hour))
	setupMigrationAt(t, tracker, "m3", StateFailed, baseTime.Add(3*time.Hour))

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

func setupMigrationAt(t *testing.T, tracker *SQLiteTracker, id string, state MigrationState, createdAt time.Time) {
	query := `INSERT INTO migrations (migration_id, state, created_at) VALUES (?, ?, ?)`
	_, err := tracker.db.Exec(query, id, state, createdAt)
	if err != nil {
		t.Fatalf("Failed to setup migration %s: %v", id, err)
	}
}

func setupMigration(t *testing.T, tracker *SQLiteTracker, id string, state MigrationState) {
	setupMigrationAt(t, tracker, id, state, time.Now())
}

func TestSQLiteTracker_ComplexCheckpoint(t *testing.T) {
	tmpFile := t.TempDir() + "/test_complex.db"

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

	// Use checked type assertions for nested structure to avoid panics.
	sm, ok := retrieved.SchemaMapping["fields"]
	if !ok {
		t.Fatal("SchemaMapping missing 'fields' key")
	}
	fields, ok := sm.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} for 'fields', got %T", sm)
	}
	if fields["title"] != "description" {
		t.Errorf("Expected title to be 'description', got %v", fields["title"])
	}

	if retrieved.ValidationStats.AvgCosineSimilarity != 0.9999 {
		t.Errorf("Validation stats not preserved correctly, got %f", retrieved.ValidationStats.AvgCosineSimilarity)
	}
}

func TestSQLiteTracker_GetTotalMigrationsCount(t *testing.T) {
	tmpFile := t.TempDir() + "/test_count.db"

	tracker, err := NewSQLiteTracker(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}
	defer tracker.Close()

	setupMigration(t, tracker, "m1", StateCompleted)
	setupMigration(t, tracker, "m2", StateInProgress)
	setupMigration(t, tracker, "m3", StateInProgress)

	count, err := tracker.GetTotalMigrationsCount("")
	if err != nil {
		t.Fatalf("GetTotalMigrationsCount failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected total count 3, got %d", count)
	}

	count, err = tracker.GetTotalMigrationsCount(string(StateInProgress))
	if err != nil {
		t.Fatalf("GetTotalMigrationsCount failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected InProgress count 2, got %d", count)
	}
}

func TestNewSQLiteTracker_Error(t *testing.T) {
	// Try to open a database in a non-existent directory.
	// We use a path inside t.TempDir() but with an extra non-existent sub-directory
	// to ensure deterministic failure across environments.
	badPath := t.TempDir() + "/nonexistent-dir/db.sqlite"
	_, err := NewSQLiteTracker(badPath)
	if err == nil {
		t.Error("Expected error when opening database in non-existent directory, got nil")
	}
}
