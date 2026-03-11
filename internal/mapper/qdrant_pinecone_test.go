package mapper

import (
	"reflect"
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

func TestQdrantPineconeMapper_MapRecord(t *testing.T) {
	mapper := NewQdrantPineconeMapper()
	mapping := &SchemaMapping{
		FieldMappings: make(map[string]string),
	}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "Flat primitive fields",
			input: map[string]interface{}{
				"string_field": "hello",
				"int_field":    42,
				"float_field":  3.14,
				"bool_field":   true,
			},
			expected: map[string]interface{}{
				"string_field": "hello",
				"int_field":    42,
				"float_field":  3.14,
				"bool_field":   true,
			},
		},
		{
			name: "Nested map flattening",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "Alice",
					"address": map[string]interface{}{
						"city": "NY",
						"zip":  10001,
					},
				},
			},
			expected: map[string]interface{}{
				"user.name":         "Alice",
				"user.address.city": "NY",
				"user.address.zip":  10001,
			},
		},
		{
			name: "Valid Pinecone string array",
			input: map[string]interface{}{
				"tags": []interface{}{"tech", "go", "vectors"},
				"pure": []string{"a", "b", "c"},
			},
			expected: map[string]interface{}{
				"tags": []string{"tech", "go", "vectors"},
				"pure": []string{"a", "b", "c"},
			},
		},
		{
			name: "Complex arrays undergo JSON stringification",
			input: map[string]interface{}{
				"scores": []interface{}{1.5, 2.5, 3.5},
				"objects": []interface{}{
					map[string]interface{}{"id": 1},
					map[string]interface{}{"id": 2},
				},
				"mixed": []interface{}{"string", 42, true},
			},
			expected: map[string]interface{}{
				"scores":  "[1.5,2.5,3.5]",
				"objects": "[{\"id\":1},{\"id\":2}]",
				"mixed":   "[\"string\",42,true]",
			},
		},
		{
			name: "Nulls are stripped",
			input: map[string]interface{}{
				"valid": "data",
				"empty": nil,
				"nested": map[string]interface{}{
					"deep_null": nil,
					"deep_val":  "ok",
				},
			},
			expected: map[string]interface{}{
				"valid":           "data",
				"nested.deep_val": "ok",
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
				ID:       "1",
				Vector:   []float32{0.1, 0.2},
				Metadata: tt.input,
			}

			result, err := mapper.MapRecord(record, mapping)
			if err != nil {
				t.Fatalf("MapRecord failed: %v", err)
			}

			if !reflect.DeepEqual(result.Metadata, tt.expected) {
				t.Errorf("\nExpected: %#v\nGot:      %#v", tt.expected, result.Metadata)
			}
		})
	}
}
