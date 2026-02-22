package main

import (
	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/AlphaTechini/vector-db-migration/internal/mapper"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// MigrationConfig holds migration configuration for CLI
type MigrationConfig struct {
	SourceDB      adapters.Database
	TargetDB      adapters.Database
	SchemaMapper  mapper.SchemaMapper
	StateTracker  state.StateTracker
	BatchSize     int
	MaxRetries    int
	ValidateEvery int
}
