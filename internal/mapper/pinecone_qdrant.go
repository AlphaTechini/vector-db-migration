package mapper

import (
	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// PineconeQdrantMapper converts records from Pinecone to Qdrant format
type PineconeQdrantMapper struct {
	*BaseMapper
}

// NewPineconeQdrantMapper creates a new Pinecone to Qdrant mapper
func NewPineconeQdrantMapper() *PineconeQdrantMapper {
	return &PineconeQdrantMapper{
		BaseMapper: NewBaseMapper("pinecone", "qdrant"),
	}
}

// MapRecord transforms a Pinecone record to Qdrant format
// Pinecone: flat metadata only
// Qdrant: supports nested payloads
func (m *PineconeQdrantMapper) MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error) {
	// Start with base mapping
	result, err := m.BaseMapper.MapRecord(record, mapping)
	if err != nil {
		return result, err
	}
	
	// Pinecone to Qdrant: flatten any nested structures
	// (Pinecone doesn't support nested metadata, but be safe)
	result.Metadata = m.flattenMetadata(result.Metadata)
	
	return result, nil
}

// flattenMetadata ensures all values are Qdrant-compatible
func (m *PineconeQdrantMapper) flattenMetadata(metadata map[string]interface{}) map[string]interface{} {
	flat := make(map[string]interface{})
	
	for key, value := range metadata {
		switch v := value.(type) {
		case map[string]interface{}:
			// Flatten nested maps with dot notation
			for subKey, subValue := range v {
				flat[key+"."+subKey] = subValue
			}
		case []interface{}:
			// Keep arrays as-is (Qdrant supports them)
			flat[key] = v
		default:
			// Primitive types pass through
			flat[key] = value
		}
	}
	
	return flat
}

// CreateMapping creates optimized mapping for Pinecone→Qdrant
func (m *PineconeQdrantMapper) CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error) {
	mapping, err := m.BaseMapper.CreateMapping(sourceSchema, targetSchema)
	if err != nil {
		return nil, err
	}
	
	// Add Pinecone→Qdrant specific type conversions
	// Pinecone stores all numbers as float64, Qdrant can handle int/float
	for field := range sourceSchema {
		mapping.TypeConversions[field] = TypeConversion{
			FromType:  "float64",
			ToType:    "auto",
			Converter: autoConvertNumber,
		}
	}
	
	return mapping, nil
}

// autoConvertNumber attempts to convert float64 to int if possible
func autoConvertNumber(value interface{}) (interface{}, error) {
	if f, ok := value.(float64); ok {
		// If it's a whole number, convert to int
		if f == float64(int64(f)) {
			return int64(f), nil
		}
	}
	return value, nil
}
