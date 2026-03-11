package mapper

import (
	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// WeaviateQdrantMapper converts records from Weaviate to Qdrant format
type WeaviateQdrantMapper struct {
	*BaseMapper
}

// NewWeaviateQdrantMapper creates a new Weaviate to Qdrant mapper
func NewWeaviateQdrantMapper() *WeaviateQdrantMapper {
	return &WeaviateQdrantMapper{
		BaseMapper: NewBaseMapper("weaviate", "qdrant"),
	}
}

// MapRecord transforms a Weaviate record to Qdrant format
// Weaviate: supports flat properties and complex graph arrays (cross-references).
// Qdrant: natively supports arbitrary complex nested JSON payloads.
// Since Qdrant is a superset of Weaviate's structural capabilities, this is a 1:1 passthrough.
func (m *WeaviateQdrantMapper) MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error) {
	// Let BaseMapper apply any user-defined field renaming, default values, or basic type conversions.
	result, err := m.BaseMapper.MapRecord(record, mapping)
	if err != nil {
		return result, err
	}

	// Qdrant natively handles the graph arrays and UUIDs without modification.
	return result, nil
}

// CreateMapping creates an optimized mapping for Weaviate→Qdrant
func (m *WeaviateQdrantMapper) CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error) {
	// Use standard BaseMapper resolution
	return m.BaseMapper.CreateMapping(sourceSchema, targetSchema)
}
