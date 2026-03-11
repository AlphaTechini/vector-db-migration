package mapper

import (
	"encoding/json"
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
)

// QdrantPineconeMapper converts records from Qdrant to Pinecone format
type QdrantPineconeMapper struct {
	*BaseMapper
}

// NewQdrantPineconeMapper creates a new Qdrant to Pinecone mapper
func NewQdrantPineconeMapper() *QdrantPineconeMapper {
	return &QdrantPineconeMapper{
		BaseMapper: NewBaseMapper("qdrant", "pinecone"),
	}
}

// MapRecord transforms a Qdrant record to Pinecone format
// Qdrant: supports nested payloads (objects, mixed arrays, nulls)
// Pinecone: strict flat structure, string/number/bool/[]string only, NO nulls
func (m *QdrantPineconeMapper) MapRecord(record adapters.Record, mapping *SchemaMapping) (adapters.Record, error) {
	// First apply base mapping logic (field renames, default values, simple type conversions)
	result, err := m.BaseMapper.MapRecord(record, mapping)
	if err != nil {
		return result, err
	}

	// Apply Pinecone structural constraints (Hybrid Approach)
	// 1. Flatten nested maps using dot-notation.
	// 2. Stringify complex arrays (arrays of objects/numbers).
	// 3. Keep simple arrays ([]string).
	// 4. Strip out nulls entirely.

	// Pre-allocate to minimize dynamic growth. Multiply by 2 assuming average nesting depth.
	capacity := len(result.Metadata) * 2
	if capacity < 8 {
		capacity = 8
	}
	flattened := make(map[string]interface{}, capacity)

	// Iterative DFS stack to avoid deep recursion overhead
	type stackItem struct {
		prefix string
		value  interface{}
	}

	// Pre-allocate stack
	stack := make([]stackItem, 0, capacity)

	// Initialize stack with root elements
	for k, v := range result.Metadata {
		stack = append(stack, stackItem{prefix: k, value: v})
	}

	for len(stack) > 0 {
		// Pop from stack (LIFO)
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if curr.value == nil {
			continue // Pinecone strictly rejects null values, completely strip them
		}

		switch v := curr.value.(type) {
		case map[string]interface{}:
			// Flatten nested map using dot-notation
			for k, childVal := range v {
				newKey := curr.prefix + "." + k
				stack = append(stack, stackItem{prefix: newKey, value: childVal})
			}

		case []interface{}:
			// Pinecone ONLY supports []string natively.
			// If it contains objects/numbers/mixed, we JSON stringify it.
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
				// Pure string array, passes natively
				flattened[curr.prefix] = stringArr
			} else {
				// Complex array, JSON stringify it to avoid data loss
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					flattened[curr.prefix] = fmt.Sprintf("%v", v) // fallback
				} else {
					flattened[curr.prefix] = string(jsonBytes)
				}
			}

		case []string:
			// Already a valid Pinecone array
			flattened[curr.prefix] = v

		default:
			// Primitive values pass through
			flattened[curr.prefix] = v
		}
	}

	result.Metadata = flattened
	return result, nil
}

// MapBatch transforms multiple records
func (m *QdrantPineconeMapper) MapBatch(records []adapters.Record, mapping *SchemaMapping) ([]adapters.Record, error) {
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

// CreateMapping creates a basic mapping for Qdrant→Pinecone
func (m *QdrantPineconeMapper) CreateMapping(sourceSchema, targetSchema map[string]interface{}) (*SchemaMapping, error) {
	return m.BaseMapper.CreateMapping(sourceSchema, targetSchema)
}
