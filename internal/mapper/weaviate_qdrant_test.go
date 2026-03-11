package mapper

import (
	"reflect"
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

func TestWeaviateQdrantMapper_MapRecord(t *testing.T) {
	mapper := NewWeaviateQdrantMapper()
	mapping := &SchemaMapping{
		FieldMappings: make(map[string]string),
	}

	tests := []struct {
		name     string
		recordID string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "Weaviate UUID and Flat Primitive Passthrough",
			recordID: "36b48590-db1a-4286-9ac6-0b196144dc61",
			input: map[string]interface{}{
				"title":       "An Article",
				"wordCount":   1250,
				"isPublished": true,
			},
			expected: map[string]interface{}{
				"title":       "An Article",
				"wordCount":   1250,
				"isPublished": true,
			},
		},
		{
			name:     "Weaviate Cross-Reference Graph Array Passthrough",
			recordID: "55b48590-db1a-4286-9ac6-0b196144dc61",
			input: map[string]interface{}{
				"title": "Related Documents",
				"hasAuthors": []interface{}{
					map[string]interface{}{
						"beacon": "weaviate://localhost/Author/123",
						"href":   "/v1/objects/Author/123",
					},
				},
			},
			expected: map[string]interface{}{
				"title": "Related Documents",
				"hasAuthors": []interface{}{
					map[string]interface{}{
						"beacon": "weaviate://localhost/Author/123",
						"href":   "/v1/objects/Author/123",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Auto mappings for root fields
			mapping.FieldMappings = make(map[string]string)
			for k := range tt.input {
				mapping.FieldMappings[k] = k
			}

			record := adapters.Record{
				ID:       tt.recordID,
				Vector:   []float32{1.0, 0.0},
				Metadata: tt.input,
			}

			result, err := mapper.MapRecord(record, mapping)
			if err != nil {
				t.Fatalf("MapRecord failed: %v", err)
			}

			// Ensure metadata is identical
			if !reflect.DeepEqual(result.Metadata, tt.expected) {
				t.Errorf("\nExpected: %#v\nGot:      %#v", tt.expected, result.Metadata)
			}

			// Ensure UUID wasn't touched
			if result.ID != tt.recordID {
				t.Errorf("Expected ID %s, got %s", tt.recordID, result.ID)
			}
		})
	}
}
