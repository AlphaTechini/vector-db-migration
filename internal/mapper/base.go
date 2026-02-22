package mapper

import (
	"fmt"
	"strings"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// BaseMapper provides common functionality for all schema mappers
type BaseMapper struct {
	sourceDB string
	targetDB string
	matcher  *FieldMatcher
}

// NewBaseMapper creates a new base mapper
func NewBaseMapper(sourceDB, targetDB string) *BaseMapper {
	return &BaseMapper{
		sourceDB: sourceDB,
		targetDB: targetDB,
		matcher:  NewFieldMatcher(),
	}
}

// CreateMapping creates a basic field mapping between schemas
func (m *BaseMapper) CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error) {
	if sourceSchema == nil || targetSchema == nil {
		return nil, fmt.Errorf("source and target schemas must not be nil")
	}
	
	mapping := &SchemaMapping{
		FieldMappings:   make(map[string]string),
		TypeConversions: make(map[string]TypeConversion),
		DefaultValues:   make(map[string]interface{}),
		SourceDB:        m.sourceDB,
		TargetDB:        m.targetDB,
	}
	
	// Auto-match fields with same names
	for sourceField := range sourceSchema {
		// Skip internal fields
		if sourceField == "id" || sourceField == "vector" {
			continue
		}
		
		// Check if target has same field
		if _, exists := targetSchema[sourceField]; exists {
			mapping.FieldMappings[sourceField] = sourceField
		} else {
			// Try fuzzy matching
			matchedField := m.findMatchingField(sourceField, targetSchema)
			if matchedField != "" {
				mapping.FieldMappings[sourceField] = matchedField
			} else {
				// No match found, use default value
				mapping.DefaultValues[sourceField] = nil
			}
		}
	}
	
	return mapping, nil
}

// findMatchingField tries to find a matching field in target schema
func (m *BaseMapper) findMatchingField(sourceField string, targetSchema map[string]interface{}) string {
	// Exact match (case-insensitive)
	if !m.matcher.CaseSensitive {
		sourceLower := strings.ToLower(sourceField)
		for targetField := range targetSchema {
			if strings.ToLower(targetField) == sourceLower {
				return targetField
			}
		}
	}
	
	// TODO: Add fuzzy matching logic if needed
	// For now, just return empty (no match)
	return ""
}

// MapRecord applies mapping to transform a record
func (m *BaseMapper) MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error) {
	result := adapters.Record{
		ID:       record.ID,
		Vector:   record.Vector,
		Metadata: make(map[string]interface{}),
	}
	
	// Apply field mappings
	for sourceField, targetField := range mapping.FieldMappings {
		if value, exists := record.Metadata[sourceField]; exists {
			result.Metadata[targetField] = value
		} else if defaultValue, exists := mapping.DefaultValues[sourceField]; exists {
			result.Metadata[targetField] = defaultValue
		}
	}
	
	// Apply type conversions
	for field, conversion := range mapping.TypeConversions {
		if value, exists := result.Metadata[field]; exists && conversion.Converter != nil {
			converted, err := conversion.Converter(value)
			if err != nil {
				return result, fmt.Errorf("failed to convert field %s: %w", field, err)
			}
			result.Metadata[field] = converted
		}
	}
	
	return result, nil
}

// MapBatch maps multiple records using the same mapping
func (m *BaseMapper) MapBatch(records []adapters.Record, mapping *SchemaMapping) ([]adapters.Record, error) {
	results := make([]adapters.Record, len(records))
	
	for i, record := range records {
		mapped, err := m.MapRecord(record, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to map record %d: %w", i, err)
		}
		results[i] = mapped
	}
	
	return results, nil
}

// ValidateMapping checks if mapping is valid
func (m *BaseMapper) ValidateMapping(mapping *SchemaMapping) error {
	if mapping == nil {
		return fmt.Errorf("mapping cannot be nil")
	}
	
	if mapping.SourceDB == "" {
		return fmt.Errorf("source database type must be specified")
	}
	
	if mapping.TargetDB == "" {
		return fmt.Errorf("target database type must be specified")
	}
	
	if mapping.SourceDB == mapping.TargetDB {
		return fmt.Errorf("source and target databases must be different")
	}
	
	// Check for valid database types
	validDBs := map[string]bool{
		"pinecone": true,
		"qdrant":   true,
		"weaviate": true,
	}
	
	if !validDBs[mapping.SourceDB] {
		return fmt.Errorf("invalid source database type: %s", mapping.SourceDB)
	}
	
	if !validDBs[mapping.TargetDB] {
		return fmt.Errorf("invalid target database type: %s", mapping.TargetDB)
	}
	
	return nil
}

// GetSourceDB returns the source database type
func (m *BaseMapper) GetSourceDB() string {
	return m.sourceDB
}

// GetTargetDB returns the target database type
func (m *BaseMapper) GetTargetDB() string {
	return m.targetDB
}
