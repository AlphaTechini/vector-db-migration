package tools

import (
	"context"
	"fmt"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
)

// SchemaRecommendationTool implements the schema_recommendation MCP tool
type SchemaRecommendationTool struct{}

// FieldRecommendation represents a recommended field mapping
type FieldRecommendation struct {
	SourceField      string  `json:"source_field"`
	TargetField      string  `json:"target_field"`
	Confidence       float64 `json:"confidence"`
	ConversionNeeded bool    `json:"conversion_needed"`
	Notes            string  `json:"notes,omitempty"`
}

// SchemaRecommendation is the full recommendation response
type SchemaRecommendation struct {
	SourceType      string              `json:"source_type"`
	TargetType      string              `json:"target_type"`
	FieldMappings   []FieldRecommendation `json:"field_mappings"`
	OverallConfidence float64           `json:"overall_confidence"`
	Warnings        []string            `json:"warnings,omitempty"`
}

// NewSchemaRecommendationTool creates a new schema_recommendation tool
func NewSchemaRecommendationTool() *SchemaRecommendationTool {
	return &SchemaRecommendationTool{}
}

// Register adds the tool to an MCP registry
func (t *SchemaRecommendationTool) Register(registry *mcp.ToolRegistry) error {
	return registry.Register(&mcp.Tool{
		Name:        "schema_recommendation",
		Description: "Get schema mapping recommendations for migrating between vector databases",
		Schema:      t.inputSchema(),
		Handler:     t.execute,
	})
}

// inputSchema defines the JSON Schema for tool inputs
func (t *SchemaRecommendationTool) inputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_type": map[string]interface{}{
				"type": "string",
				"description": "Source database type",
				"enum": []string{"pinecone", "qdrant", "weaviate", "milvus"},
			},
			"target_type": map[string]interface{}{
				"type": "string",
				"description": "Target database type",
				"enum": []string{"pinecone", "qdrant", "weaviate", "milvus"},
			},
			"source_schema": map[string]interface{}{
				"type": "object",
				"description": "Source database schema (field names and types)",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"source_type", "target_type"},
	}
}

// execute runs the schema_recommendation tool
func (t *SchemaRecommendationTool) execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Validate inputs
	sourceType, ok := params["source_type"].(string)
	if !ok || sourceType == "" {
		return nil, fmt.Errorf("source_type is required")
	}

	targetType, ok := params["target_type"].(string)
	if !ok || targetType == "" {
		return nil, fmt.Errorf("target_type is required")
	}

	if sourceType == targetType {
		return nil, fmt.Errorf("source_type and target_type must be different")
	}

	// Get source schema if provided
	sourceSchema, _ := params["source_schema"].(map[string]interface{})

	// Generate recommendations based on migration path
	recommendation := t.generateRecommendations(sourceType, targetType, sourceSchema)

	return recommendation, nil
}

// generateRecommendations creates schema mapping recommendations
func (t *SchemaRecommendationTool) generateRecommendations(sourceType, targetType string, sourceSchema map[string]interface{}) *SchemaRecommendation {
	rec := &SchemaRecommendation{
		SourceType: sourceType,
		TargetType: targetType,
		FieldMappings: []FieldRecommendation{},
		Warnings: []string{},
	}

	// Common field mappings across all migrations
	commonFields := map[string]FieldRecommendation{
		"id": {
			SourceField: "id",
			TargetField: "id",
			Confidence:  1.0,
			Notes:       "Primary identifier, direct mapping",
		},
		"title": {
			SourceField: "title",
			TargetField: "title",
			Confidence:  0.95,
			Notes:       "Common metadata field",
		},
		"url": {
			SourceField: "url",
			TargetField: "url",
			Confidence:  0.95,
			Notes:       "URL reference field",
		},
		"content": {
			SourceField: "content",
			TargetField: "content",
			Confidence:  0.9,
			Notes:       "Main content field",
		},
	}

	// Add common fields
	for _, fieldRec := range commonFields {
		rec.FieldMappings = append(rec.FieldMappings, fieldRec)
	}

	// Database-specific recommendations
	switch sourceType + "_to_" + targetType {
	case "pinecone_to_qdrant":
		rec.Warnings = append(rec.Warnings, "Pinecone flat metadata will be flattened in Qdrant with dot notation")
		rec.OverallConfidence = 0.9

	case "pinecone_to_weaviate":
		rec.Warnings = append(rec.Warnings, "Weaviate requires schema definition before upsert")
		rec.Warnings = append(rec.Warnings, "Nested metadata not supported in Pinecone, but supported in Weaviate")
		rec.OverallConfidence = 0.85

	case "qdrant_to_pinecone":
		rec.Warnings = append(rec.Warnings, "Qdrant nested payloads must be flattened for Pinecone")
		rec.Warnings = append(rec.Warnings, "Use dot notation: author.name → author_name")
		rec.OverallConfidence = 0.85

	case "weaviate_to_pinecone":
		rec.Warnings = append(rec.Warnings, "Weaviate typed properties will become untyped in Pinecone")
		rec.Warnings = append(rec.Warnings, "Type information will be lost")
		rec.OverallConfidence = 0.8

	default:
		rec.OverallConfidence = 0.75
		rec.Warnings = append(rec.Warnings, "Generic migration path - review mappings carefully")
	}

	// If source schema provided, add specific recommendations
	if len(sourceSchema) > 0 {
		for fieldName := range sourceSchema {
			// Check if we have a recommendation for this field
			found := false
			for _, mapping := range rec.FieldMappings {
				if mapping.SourceField == fieldName {
					found = true
					break
				}
			}

			if !found {
				// Add generic recommendation for unknown fields
				rec.FieldMappings = append(rec.FieldMappings, FieldRecommendation{
					SourceField: fieldName,
					TargetField: fieldName,
					Confidence:  0.7,
					Notes:       "Auto-mapped by name - verify type compatibility",
				})
			}
		}
	}

	return rec
}
