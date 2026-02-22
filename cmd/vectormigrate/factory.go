package main

import (
	"context"
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/AlphaTechini/vector-db-migration/internal/mapper"
	"github.com/AlphaTechini/vector-db-migration/internal/orchestrator"
	"github.com/AlphaTechini/vector-db-migration/internal/state"
)

// createDatabase creates a database adapter based on type
func createDatabase(dbType, url, apiKey, index string, timeout int) (adapters.Database, error) {
	config := adapters.DBConfig{
		Type:    dbType,
		URL:     url,
		APIKey:  apiKey,
		Index:   index,
		Timeout: timeout,
	}

	ctx := context.Background()

	switch dbType {
	case "pinecone":
		adapter := &adapters.PineconeAdapter{}
		if err := adapter.Connect(ctx, config); err != nil {
			return nil, fmt.Errorf("failed to connect to Pinecone: %w", err)
		}
		return adapter, nil

	case "qdrant":
		adapter := &adapters.QdrantAdapter{}
		if err := adapter.Connect(ctx, config); err != nil {
			return nil, fmt.Errorf("failed to connect to Qdrant: %w", err)
		}
		return adapter, nil

	case "weaviate":
		adapter := &adapters.WeaviateAdapter{}
		if err := adapter.Connect(ctx, config); err != nil {
			return nil, fmt.Errorf("failed to connect to Weaviate: %w", err)
		}
		return adapter, nil

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// createMapper creates a schema mapper based on source/target types
func createMapper(sourceType, targetType string) (mapper.SchemaMapper, error) {
	key := sourceType + "_to_" + targetType

	switch key {
	case "pinecone_to_qdrant":
		return mapper.NewPineconeQdrantMapper(), nil

	case "qdrant_to_pinecone":
		// TODO: Implement QdrantToPineconeMapper
		return nil, fmt.Errorf("mapper not implemented: %s", key)

	case "pinecone_to_weaviate":
		// TODO: Implement PineconeToWeaviateMapper
		return nil, fmt.Errorf("mapper not implemented: %s", key)

	case "weaviate_to_pinecone":
		// TODO: Implement WeaviateToPineconeMapper
		return nil, fmt.Errorf("mapper not implemented: %s", key)

	case "qdrant_to_weaviate":
		// TODO: Implement QdrantToWeaviateMapper
		return nil, fmt.Errorf("mapper not implemented: %s", key)

	case "weaviate_to_qdrant":
		// TODO: Implement WeaviateToQdrantMapper
		return nil, fmt.Errorf("mapper not implemented: %s", key)

	default:
		return nil, fmt.Errorf("unsupported migration path: %s → %s", sourceType, targetType)
	}
}

// createStateTracker creates a state tracker
func createStateTracker(dbPath string) (state.StateTracker, error) {
	if dbPath == "" {
		dbPath = "vectormigrate.db" // Default
	}

	tracker, err := state.NewSQLiteTracker(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create state tracker: %w", err)
	}

	return tracker, nil
}

// createOrchestrator creates a migration orchestrator
func createOrchestrator(migrationID string) orchestrator.MigrationOrchestrator {
	return orchestrator.NewBaseOrchestrator(migrationID)
}
