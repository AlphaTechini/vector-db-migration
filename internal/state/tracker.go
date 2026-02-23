package state

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// MigrationState represents the current state of a migration
type MigrationState string

const (
	StateNotStarted   MigrationState = "not_started"
	StateInProgress   MigrationState = "in_progress"
	StateCompleted    MigrationState = "completed"
	StateRolledBack   MigrationState = "rolled_back"
	StateFailed       MigrationState = "failed"
)

// Checkpoint represents a migration checkpoint for resume-on-failure
type Checkpoint struct {
	MigrationID        string                 `json:"migration_id"`
	LastProcessedID    string                 `json:"last_processed_id"`
	TotalRecords       int64                  `json:"total_records"`
	ProcessedCount     int64                  `json:"processed_count"`
	FailedCount        int64                  `json:"failed_count"`
	StartedAt          time.Time              `json:"started_at"`
	LastCheckpointAt   time.Time              `json:"last_checkpoint_at"`
	SchemaMapping      map[string]interface{} `json:"schema_mapping,omitempty"`
	ValidationStats    ValidationStats        `json:"validation_stats,omitempty"`
}

// ValidationStats tracks validation metrics
type ValidationStats struct {
	SampledCount      int64   `json:"sampled_count"`
	AvgCosineSimilarity float64 `json:"avg_cosine_similarity"`
	MinCosineSimilarity float64 `json:"min_cosine_similarity"`
	MaxCosineSimilarity float64 `json:"max_cosine_similarity"`
}

// StateTracker interface for persisting and retrieving migration state
type StateTracker interface {
	// GetState returns the current state of a migration
	GetState(migrationID string) (MigrationState, error)
	
	// SetState updates the state of a migration
	SetState(migrationID string, state MigrationState) error
	
	// GetCheckpoint returns the last checkpoint for a migration
	GetCheckpoint(migrationID string) (*Checkpoint, error)
	
	// SaveCheckpoint saves a checkpoint for resume-on-failure
	SaveCheckpoint(checkpoint *Checkpoint) error
	
	// DeleteCheckpoint removes a checkpoint (cleanup after completion)
	DeleteCheckpoint(migrationID string) error
	
	// Close closes the underlying storage connection
	Close() error
	
	// ListMigrations returns migration IDs with optional filtering
	ListMigrations(statusFilter string, limit, offset int) ([]string, error)
	
	// GetMigrationSummary returns a migration summary by ID
	GetMigrationSummary(migrationID string) (*Checkpoint, error)
}

// SQLiteTracker implements StateTracker using SQLite
type SQLiteTracker struct {
	db *sql.DB
}

// NewSQLiteTracker creates a new SQLite-based state tracker
func NewSQLiteTracker(dbPath string) (*SQLiteTracker, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &SQLiteTracker{db: db}, nil
}

// createTables creates the necessary database tables
func createTables(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS migrations (
		migration_id TEXT PRIMARY KEY,
		state TEXT NOT NULL DEFAULT 'not_started',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS checkpoints (
		migration_id TEXT PRIMARY KEY,
		checkpoint_data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (migration_id) REFERENCES migrations(migration_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_migrations_state ON migrations(state);
	`

	_, err := db.Exec(schema)
	return err
}

// GetState returns the current state of a migration
func (t *SQLiteTracker) GetState(migrationID string) (MigrationState, error) {
	query := `SELECT state FROM migrations WHERE migration_id = ?`
	
	var state string
	err := t.db.QueryRow(query, migrationID).Scan(&state)
	if err == sql.ErrNoRows {
		return StateNotStarted, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}

	return MigrationState(state), nil
}

// SetState updates the state of a migration
func (t *SQLiteTracker) SetState(migrationID string, state MigrationState) error {
	query := `
	INSERT INTO migrations (migration_id, state, updated_at) 
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(migration_id) DO UPDATE SET 
		state = excluded.state,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := t.db.Exec(query, migrationID, state)
	if err != nil {
		return fmt.Errorf("failed to set state: %w", err)
	}

	return nil
}

// GetCheckpoint returns the last checkpoint for a migration
func (t *SQLiteTracker) GetCheckpoint(migrationID string) (*Checkpoint, error) {
	query := `SELECT checkpoint_data FROM checkpoints WHERE migration_id = ?`
	
	var jsonData string
	err := t.db.QueryRow(query, migrationID).Scan(&jsonData)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal([]byte(jsonData), &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// SaveCheckpoint saves a checkpoint for resume-on-failure
func (t *SQLiteTracker) SaveCheckpoint(checkpoint *Checkpoint) error {
	jsonData, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	query := `
	INSERT INTO checkpoints (migration_id, checkpoint_data, updated_at)
	VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(migration_id) DO UPDATE SET
		checkpoint_data = excluded.checkpoint_data,
		updated_at = CURRENT_TIMESTAMP
	`

	_, err = t.db.Exec(query, checkpoint.MigrationID, jsonData)
	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// Also update migration state to in_progress if not already set
	state, err := t.GetState(checkpoint.MigrationID)
	if err != nil {
		return err
	}
	if state == StateNotStarted {
		if err := t.SetState(checkpoint.MigrationID, StateInProgress); err != nil {
			return err
		}
	}

	return nil
}

// DeleteCheckpoint removes a checkpoint (cleanup after completion)
func (t *SQLiteTracker) DeleteCheckpoint(migrationID string) error {
	query := `DELETE FROM checkpoints WHERE migration_id = ?`
	_, err := t.db.Exec(query, migrationID)
	if err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}
	return nil
}

// Close closes the underlying database connection
func (t *SQLiteTracker) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}

// ListMigrations returns a list of all migration IDs with optional filtering
func (t *SQLiteTracker) ListMigrations(statusFilter string, limit, offset int) ([]string, error) {
	query := `SELECT migration_id FROM migrations`
	args := []interface{}{}
	
	if statusFilter != "" {
		query += ` WHERE state = ?`
		args = append(args, statusFilter)
	}
	
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	
	rows, err := t.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list migrations: %w", err)
	}
	defer rows.Close()
	
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan migration ID: %w", err)
		}
		ids = append(ids, id)
	}
	
	return ids, nil
}

// GetMigrationSummary returns a summary of a migration by ID
func (t *SQLiteTracker) GetMigrationSummary(migrationID string) (*Checkpoint, error) {
	return t.GetCheckpoint(migrationID)
}

// Ensure SQLiteTracker implements StateTracker interface
var _ StateTracker = (*SQLiteTracker)(nil)
