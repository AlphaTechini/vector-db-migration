package mapper

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

func TestBaseMapper_FuzzyMatching(t *testing.T) {
	m := NewBaseMapper("pinecone", "qdrant")

	sourceSchema := map[string]interface{}{
		"UserEmail": "string",
		"post_id":   "int",
	}

	targetSchema := map[string]interface{}{
		"useremail": "string",
		"POST_ID":   "int",
	}

	mapping, err := m.CreateMapping(sourceSchema, targetSchema)
	if err != nil {
		t.Fatalf("CreateMapping failed: %v", err)
	}

	if mapping.FieldMappings["UserEmail"] != "useremail" {
		t.Errorf("Expected UserEmail to match useremail, got %s", mapping.FieldMappings["UserEmail"])
	}
	if mapping.FieldMappings["post_id"] != "POST_ID" {
		t.Errorf("Expected post_id to match POST_ID, got %s", mapping.FieldMappings["post_id"])
	}
}

func TestBaseMapper_TypeConversion(t *testing.T) {
	m := NewBaseMapper("pinecone", "qdrant")

	record := adapters.Record{
		ID: "r1",
		Metadata: map[string]interface{}{
			"age": "25",
		},
	}

	mapping := &SchemaMapping{
		FieldMappings: map[string]string{
			"age": "age",
		},
		TypeConversions: map[string]TypeConversion{
			"age": {
				FromType: "string",
				ToType:   "int",
				Converter: func(v interface{}) (interface{}, error) {
					s, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("value is not a string (got %T)", v)
					}
					return strconv.Atoi(s)
				},
			},
		},
	}

	result, err := m.MapRecord(record, mapping)
	if err != nil {
		t.Fatalf("MapRecord failed: %v", err)
	}

	age, ok := result.Metadata["age"].(int)
	if !ok || age != 25 {
		t.Errorf("Expected age 25 (int), got %v (%T)", result.Metadata["age"], result.Metadata["age"])
	}
}

func TestBaseMapper_TypeConversionError(t *testing.T) {
	m := NewBaseMapper("pinecone", "qdrant")

	record := adapters.Record{
		ID: "r1",
		Metadata: map[string]interface{}{
			"age": "not-a-number",
		},
	}

	mapping := &SchemaMapping{
		FieldMappings: map[string]string{
			"age": "age",
		},
		TypeConversions: map[string]TypeConversion{
			"age": {
				Converter: func(v interface{}) (interface{}, error) {
					s, ok := v.(string)
					if !ok {
						return nil, fmt.Errorf("value is not a string (got %T)", v)
					}
					return strconv.Atoi(s)
				},
			},
		},
	}

	_, err := m.MapRecord(record, mapping)
	if err == nil {
		t.Error("Expected error during MapRecord due to failed conversion, got nil")
	}
	if !strings.Contains(err.Error(), "failed to convert field age") {
		t.Errorf("Expected error message to mention field conversion failure, got: %v", err)
	}
}

func TestBaseMapper_DefaultValues(t *testing.T) {
	m := NewBaseMapper("pinecone", "qdrant")

	record := adapters.Record{
		ID:       "r1",
		Metadata: map[string]interface{}{
			// "title" is missing
		},
	}

	mapping := &SchemaMapping{
		FieldMappings: map[string]string{
			"title": "title",
		},
		DefaultValues: map[string]interface{}{
			"title": "Untitled",
		},
	}

	result, err := m.MapRecord(record, mapping)
	if err != nil {
		t.Fatalf("MapRecord failed: %v", err)
	}

	if result.Metadata["title"] != "Untitled" {
		t.Errorf("Expected default title 'Untitled', got %v", result.Metadata["title"])
	}
}
