package mapper

import (
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// TestSchemaMapperInterface ensures all mappers implement the interface
func TestSchemaMapperInterface(t *testing.T) {
	var _ SchemaMapper = (*PineconeQdrantMapper)(nil)
	t.Log("✓ PineconeQdrantMapper implements SchemaMapper interface")
}

// TestBaseMapper_CreateMapping tests basic mapping creation
func TestBaseMapper_CreateMapping(t *testing.T) {
	mapper := NewBaseMapper("pinecone", "qdrant")
	
	sourceSchema := map[string]interface{}{
		"title":  "string",
		"url":    "string",
		"score":  "float",
	}
	
	targetSchema := map[string]interface{}{
		"title":  "text",
		"url":    "text",
		"score":  "float",
	}
	
	mapping, err := mapper.CreateMapping(sourceSchema, targetSchema)
	if err != nil {
		t.Fatalf("Failed to create mapping: %v", err)
	}
	
	if mapping.SourceDB != "pinecone" {
		t.Errorf("Expected SourceDB 'pinecone', got '%s'", mapping.SourceDB)
	}
	
	if mapping.TargetDB != "qdrant" {
		t.Errorf("Expected TargetDB 'qdrant', got '%s'", mapping.TargetDB)
	}
	
	// Check field mappings
	if len(mapping.FieldMappings) != 3 {
		t.Errorf("Expected 3 field mappings, got %d", len(mapping.FieldMappings))
	}
	
	t.Log("✓ BaseMapper creates mappings correctly")
}

// TestBaseMapper_MapRecord tests record transformation
func TestBaseMapper_MapRecord(t *testing.T) {
	mapper := NewBaseMapper("pinecone", "qdrant")
	
	record := adapters.Record{
		ID:     "test-123",
		Vector: []float32{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"title": "Test Document",
			"url":   "https://example.com",
		},
	}
	
	mapping := &SchemaMapping{
		FieldMappings: map[string]string{
			"title": "title",
			"url":   "url",
		},
		SourceDB: "pinecone",
		TargetDB: "qdrant",
	}
	
	result, err := mapper.MapRecord(record, mapping)
	if err != nil {
		t.Fatalf("Failed to map record: %v", err)
	}
	
	if result.ID != record.ID {
		t.Errorf("Expected ID '%s', got '%s'", record.ID, result.ID)
	}
	
	if len(result.Vector) != 3 {
		t.Errorf("Expected vector length 3, got %d", len(result.Vector))
	}
	
	if result.Metadata["title"] != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%v'", result.Metadata["title"])
	}
	
	t.Log("✓ BaseMapper maps records correctly")
}

// TestBaseMapper_ValidateMapping tests validation logic
func TestBaseMapper_ValidateMapping(t *testing.T) {
	mapper := NewBaseMapper("pinecone", "qdrant")
	
	// Test valid mapping
	validMapping := &SchemaMapping{
		SourceDB: "pinecone",
		TargetDB: "qdrant",
	}
	
	err := mapper.ValidateMapping(validMapping)
	if err != nil {
		t.Errorf("Expected valid mapping to pass validation: %v", err)
	}
	
	// Test nil mapping
	err = mapper.ValidateMapping(nil)
	if err == nil {
		t.Error("Expected error for nil mapping")
	}
	
	// Test missing source DB
	invalidMapping := &SchemaMapping{
		TargetDB: "qdrant",
	}
	err = mapper.ValidateMapping(invalidMapping)
	if err == nil {
		t.Error("Expected error for missing source DB")
	}
	
	// Test same source and target
	sameDBMapping := &SchemaMapping{
		SourceDB: "pinecone",
		TargetDB: "pinecone",
	}
	err = mapper.ValidateMapping(sameDBMapping)
	if err == nil {
		t.Error("Expected error for same source and target DB")
	}
	
	t.Log("✓ BaseMapper validates mappings correctly")
}

// TestPineconeQdrantMapper_FlattenMetadata tests metadata flattening
func TestPineconeQdrantMapper_FlattenMetadata(t *testing.T) {
	mapper := NewPineconeQdrantMapper()
	
	// Test with flat metadata (should pass through)
	flat := map[string]interface{}{
		"title": "Test",
		"score": 0.95,
	}
	
	result := mapper.flattenMetadata(flat)
	if len(result) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(result))
	}
	
	// Test with nested metadata (should flatten)
	nested := map[string]interface{}{
		"author": map[string]interface{}{
			"name": "John Doe",
			"email": "john@example.com",
		},
	}
	
	result = mapper.flattenMetadata(nested)
	if _, exists := result["author.name"]; !exists {
		t.Error("Expected flattened 'author.name' field")
	}
	if _, exists := result["author.email"]; !exists {
		t.Error("Expected flattened 'author.email' field")
	}
	
	t.Log("✓ PineconeQdrantMapper flattens metadata correctly")
}

// TestBaseMapper_MapBatch tests batch mapping
func TestBaseMapper_MapBatch(t *testing.T) {
	mapper := NewBaseMapper("pinecone", "qdrant")
	
	records := []adapters.Record{
		{
			ID:       "doc-1",
			Vector:   []float32{0.1, 0.2},
			Metadata: map[string]interface{}{"title": "Doc 1"},
		},
		{
			ID:       "doc-2",
			Vector:   []float32{0.3, 0.4},
			Metadata: map[string]interface{}{"title": "Doc 2"},
		},
	}
	
	mapping := &SchemaMapping{
		FieldMappings: map[string]string{
			"title": "title",
		},
		SourceDB: "pinecone",
		TargetDB: "qdrant",
	}
	
	results, err := mapper.MapBatch(records, mapping)
	if err != nil {
		t.Fatalf("Failed to map batch: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	if results[0].ID != "doc-1" {
		t.Errorf("Expected first result ID 'doc-1', got '%s'", results[0].ID)
	}
	
	if results[1].ID != "doc-2" {
		t.Errorf("Expected second result ID 'doc-2', got '%s'", results[1].ID)
	}
	
	t.Log("✓ BaseMapper maps batches correctly")
}
