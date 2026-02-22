package mapper

import (
	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// SchemaMapping defines how to transform metadata between databases
type SchemaMapping struct {
	// FieldMappings maps source field names to target field names
	FieldMappings map[string]string `json:"field_mappings"`
	
	// TypeConversions defines how to convert field types
	TypeConversions map[string]TypeConversion `json:"type_conversions"`
	
	// DefaultValues for missing fields
	DefaultValues map[string]interface{} `json:"default_values"`
	
	// SourceDB type (pinecone, qdrant, weaviate)
	SourceDB string `json:"source_db"`
	
	// TargetDB type (pinecone, qdrant, weaviate)
	TargetDB string `json:"target_db"`
}

// TypeConversion defines how to convert a field type
type TypeConversion struct {
	FromType string `json:"from_type"`
	ToType   string `json:"to_type"`
	Converter func(interface{}) (interface{}, error) `json:"-"`
}

// SchemaMapper interface for converting records between database schemas
type SchemaMapper interface {
	// CreateMapping analyzes source and target schemas and creates a mapping
	CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error)
	
	// MapRecord transforms a record from source format to target format
	MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error)
	
	// MapBatch transforms multiple records using the same mapping
	MapBatch(records []adapters.Record, mapping *SchemaMapping) ([]adapters.Record, error)
	
	// ValidateMapping checks if a mapping is valid and complete
	ValidateMapping(mapping *SchemaMapping) error
	
	// GetSourceDB returns the source database type
	GetSourceDB() string
	
	// GetTargetDB returns the target database type
	GetTargetDB() string
}

// FieldMatcher helps match fields between different schemas
type FieldMatcher struct {
	// CaseSensitive matching (default: false)
	CaseSensitive bool
	
	// FuzzyMatch enables fuzzy matching (default: true)
	FuzzyMatch bool
	
	// IgnoreFields lists fields to ignore during matching
	IgnoreFields []string
}

// NewFieldMatcher creates a new field matcher with default settings
func NewFieldMatcher() *FieldMatcher {
	return &FieldMatcher{
		CaseSensitive: false,
		FuzzyMatch:    true,
		IgnoreFields:  []string{"id", "vector"}, // Always preserve these
	}
}
