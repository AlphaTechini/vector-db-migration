package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// WeaviatePineconeMapper converts records from Weaviate to Pinecone format
type WeaviatePineconeMapper struct {
	*BaseMapper
}

// NewWeaviatePineconeMapper creates a new Weaviate to Pinecone mapper
func NewWeaviatePineconeMapper() *WeaviatePineconeMapper {
	return &WeaviatePineconeMapper{
		BaseMapper: NewBaseMapper("weaviate", "pinecone"),
	}
}

// MapRecord transforms a Weaviate record to Pinecone format
// Weaviate: supports flat properties and complex graph arrays (cross-references).
// Pinecone: strict flat structure, NO complex objects/arrays, NO nulls.
func (m *WeaviatePineconeMapper) MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error) {
	result, err := m.BaseMapper.MapRecord(record, mapping)
	if err != nil {
		return result, err
	}

	capacity := len(result.Metadata) * 2
	if capacity < 8 {
		capacity = 8
	}
	flattened := make(map[string]interface{}, capacity)

	// We use the exact same Iterative DFS stack approach here to unpack
	// Weaviate's Cross-Reference arrays safely into Pinecone strings.
	type stackItem struct {
		prefix string
		value  interface{}
	}

	stack := make([]stackItem, 0, capacity)

	for k, v := range result.Metadata {
		stack = append(stack, stackItem{prefix: k, value: v})
	}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if curr.value == nil {
			continue // Pinecone strictly rejects null values
		}

		switch v := curr.value.(type) {
		case map[string]interface{}:
			// Flatten nested map using dot-notation
			for k, childVal := range v {
				newKey := curr.prefix + "." + k
				stack = append(stack, stackItem{prefix: newKey, value: childVal})
			}

		case []interface{}:
			if len(v) == 0 {
				flattened[curr.prefix] = []string{}
				continue
			}

			isAllStrings := true
			stringArr := make([]string, len(v))

			for i, elem := range v {
				if str, ok := elem.(string); ok {
					stringArr[i] = str
				} else {
					isAllStrings = false
					break
				}
			}

			if isAllStrings {
				// Pure string array natively supported by Pinecone
				flattened[curr.prefix] = stringArr
			} else {
				// Weaviate Cross-Reference array containing maps -> JSON stringify
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					flattened[curr.prefix] = fmt.Sprintf("%v", v)
				} else {
					flattened[curr.prefix] = string(jsonBytes)
				}
			}

		case []string:
			flattened[curr.prefix] = v

		default:
			flattened[curr.prefix] = v
		}
	}

	result.Metadata = flattened
	return result, nil
}

// MapBatch transforms multiple records
func (m *WeaviatePineconeMapper) MapBatch(records []adapters.Record, mapping *SchemaMapping) ([]adapters.Record, error) {
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

// CreateMapping creates a basic mapping for Weaviate→Pinecone
func (m *WeaviatePineconeMapper) CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error) {
	return m.BaseMapper.CreateMapping(sourceSchema, targetSchema)
}
