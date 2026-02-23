package tools

import (
	"context"
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
)

func TestSchemaRecommendationTool_Register(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	registry := mcp.NewToolRegistry()

	err := tool.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	retrieved, err := registry.Get("schema_recommendation")
	if err != nil {
		t.Fatalf("Failed to get registered tool: %v", err)
	}

	if retrieved.Name != "schema_recommendation" {
		t.Errorf("Expected name 'schema_recommendation', got '%s'", retrieved.Name)
	}
}

func TestSchemaRecommendationTool_InputSchema(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	schema := tool.inputSchema()

	if schema["type"] != "object" {
		t.Errorf("Expected type 'object', got '%v'", schema["type"])
	}

	_, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok || len(required) != 2 {
		t.Fatal("Expected 2 required fields")
	}

	hasSourceType := false
	hasTargetType := false
	for _, field := range required {
		if field == "source_type" {
			hasSourceType = true
		}
		if field == "target_type" {
			hasTargetType = true
		}
	}

	if !hasSourceType || !hasTargetType {
		t.Error("Expected source_type and target_type to be required")
	}
}

func TestSchemaRecommendationTool_Execute_Success(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"source_type": "pinecone",
		"target_type": "qdrant",
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	rec, ok := result.(*SchemaRecommendation)
	if !ok {
		t.Fatal("Expected result to be *SchemaRecommendation")
	}

	if rec.SourceType != "pinecone" {
		t.Errorf("Expected source_type 'pinecone', got '%s'", rec.SourceType)
	}

	if rec.TargetType != "qdrant" {
		t.Errorf("Expected target_type 'qdrant', got '%s'", rec.TargetType)
	}

	if len(rec.FieldMappings) == 0 {
		t.Error("Expected field mappings")
	}

	if rec.OverallConfidence <= 0 || rec.OverallConfidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", rec.OverallConfidence)
	}
}

func TestSchemaRecommendationTool_Execute_MissingSourceType(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"target_type": "qdrant",
	}

	_, err := tool.execute(ctx, params)
	if err == nil {
		t.Error("Expected error for missing source_type")
	}
}

func TestSchemaRecommendationTool_Execute_MissingTargetType(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"source_type": "pinecone",
	}

	_, err := tool.execute(ctx, params)
	if err == nil {
		t.Error("Expected error for missing target_type")
	}
}

func TestSchemaRecommendationTool_Execute_SameSourceTarget(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"source_type": "pinecone",
		"target_type": "pinecone",
	}

	_, err := tool.execute(ctx, params)
	if err == nil {
		t.Error("Expected error when source and target are the same")
	}
}

func TestSchemaRecommendationTool_Execute_WithSourceSchema(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	params := map[string]interface{}{
		"source_type": "pinecone",
		"target_type": "qdrant",
		"source_schema": map[string]interface{}{
			"id":      "string",
			"title":   "string",
			"content": "text",
			"custom_field": "string",
		},
	}

	result, err := tool.execute(ctx, params)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	rec := result.(*SchemaRecommendation)

	// Should have mappings for common fields + custom field
	if len(rec.FieldMappings) < 4 {
		t.Errorf("Expected at least 4 field mappings, got %d", len(rec.FieldMappings))
	}

	// Check if custom field was mapped
	foundCustom := false
	for _, mapping := range rec.FieldMappings {
		if mapping.SourceField == "custom_field" {
			foundCustom = true
			break
		}
	}

	if !foundCustom {
		t.Error("Expected custom_field to be mapped")
	}
}

func TestSchemaRecommendationTool_DatabaseSpecificWarnings(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	testCases := []struct {
		source string
		target string
		expectWarning string
	}{
		{"pinecone", "qdrant", "flat metadata"},
		{"pinecone", "weaviate", "schema definition"},
		{"qdrant", "pinecone", "flattened"},
		{"weaviate", "pinecone", "untyped"},
	}

	for _, tc := range testCases {
		params := map[string]interface{}{
			"source_type": tc.source,
			"target_type": tc.target,
		}

		result, _ := tool.execute(ctx, params)
		rec := result.(*SchemaRecommendation)

		foundWarning := false
		for _, warning := range rec.Warnings {
			if containsIgnoreCase(warning, tc.expectWarning) {
				foundWarning = true
				break
			}
		}

		if !foundWarning {
			t.Errorf("Expected warning about '%s' for %s→%s migration", 
				tc.expectWarning, tc.source, tc.target)
		}
	}
}

func TestSchemaRecommendationTool_ConfidenceScores(t *testing.T) {
	tool := NewSchemaRecommendationTool()
	ctx := context.Background()

	testCases := []struct {
		source string
		target string
		minConfidence float64
	}{
		{"pinecone", "qdrant", 0.85},
		{"pinecone", "weaviate", 0.8},
		{"qdrant", "pinecone", 0.8},
		{"weaviate", "pinecone", 0.75},
		{"milvus", "qdrant", 0.7}, // Generic path
	}

	for _, tc := range testCases {
		params := map[string]interface{}{
			"source_type": tc.source,
			"target_type": tc.target,
		}

		result, _ := tool.execute(ctx, params)
		rec := result.(*SchemaRecommendation)

		if rec.OverallConfidence < tc.minConfidence {
			t.Errorf("Expected confidence >= %.2f for %s→%s, got %.2f",
				tc.minConfidence, tc.source, tc.target, rec.OverallConfidence)
		}
	}
}

func TestFieldRecommendation_Structure(t *testing.T) {
	rec := FieldRecommendation{
		SourceField: "test",
		TargetField: "test_mapped",
		Confidence:  0.9,
		ConversionNeeded: true,
		Notes:       "Test notes",
	}

	if rec.SourceField != "test" {
		t.Errorf("Expected SourceField 'test', got '%s'", rec.SourceField)
	}

	if rec.Confidence != 0.9 {
		t.Errorf("Expected Confidence 0.9, got %f", rec.Confidence)
	}

	if !rec.ConversionNeeded {
		t.Error("Expected ConversionNeeded to be true")
	}
}

// Helper function
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 len(s) > len(substr) && 
		 (containsLower(s, substr)))
}

func containsLower(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
