package mapper

import (
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/google/uuid"
)

// QdrantWeaviateMapper maps records from Qdrant to Weaviate.
// It leverages Weaviate's native object support (v1.22+) for lossless 1:1 payload mapping.
// It uses UUIDv5 for deterministic translation of Qdrant IDs to Weaviate-compatible UUIDs.
type QdrantWeaviateMapper struct {
	*BaseMapper
}

// NewQdrantWeaviateMapper creates a new QdrantWeaviateMapper.
func NewQdrantWeaviateMapper() *QdrantWeaviateMapper {
	return &QdrantWeaviateMapper{
		BaseMapper: NewBaseMapper("qdrant", "weaviate"),
	}
}

// MapRecord transforms a single Qdrant record into a Weaviate-compatible format.
func (m *QdrantWeaviateMapper) MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error) {
	var result adapters.Record
	var err error

	// If a mapping is provided, we allow BaseMapper to handle filtering/renaming.
	// Otherwise, we default to 1:1 passthrough to maintain lossless properties.
	if mapping != nil {
		result, err = m.BaseMapper.MapRecord(record, mapping)
		if err != nil {
			return result, err
		}
	} else {
		result = adapters.Record{
			ID:       record.ID,
			Vector:   record.Vector,
			Metadata: make(map[string]interface{}, len(record.Metadata)),
		}
		for k, v := range record.Metadata {
			result.Metadata[k] = v
		}
	}

	// 1. ID Mapping: Deterministic UUIDv5
	// We use uuid.NameSpaceURL as a base namespace for SHA1-based UUID generation.
	// This ensures that non-UUID Qdrant IDs become valid Weaviate UUIDs deterministically.
	originalID := record.ID // Use original record ID for hashing
	newUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte(originalID)).String()

	// 2. Metadata Updates: Add traceability
	// We ensure result.Metadata is initialized
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}

	// Add traceability field
	result.Metadata["_original_id"] = originalID
	result.ID = newUUID

	return result, nil
}

// MapBatch transforms a batch of Qdrant records into Weaviate records.
func (m *QdrantWeaviateMapper) MapBatch(records []adapters.Record, mapping *SchemaMapping) ([]adapters.Record, error) {
	if len(records) == 0 {
		return nil, nil
	}

	results := make([]adapters.Record, len(records))

	for i, rec := range records {
		mapped, err := m.MapRecord(rec, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to map record %s: %w", rec.ID, err)
		}
		results[i] = mapped
	}

	return results, nil
}

// CreateMapping creates a basic mapping for Qdrant→Weaviate
func (m *QdrantWeaviateMapper) CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error) {
	return m.BaseMapper.CreateMapping(sourceSchema, targetSchema)
}
