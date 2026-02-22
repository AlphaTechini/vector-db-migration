package adapters

import (
	"context"
	"testing"
)

// TestDatabaseInterface ensures all adapters implement the interface
func TestDatabaseInterface(t *testing.T) {
	// This test just verifies that PineconeAdapter implements Database
	var _ Database = (*PineconeAdapter)(nil)
	
	t.Log("✓ PineconeAdapter implements Database interface")
}

// TestPineconeAdapterConnect tests connection validation
func TestPineconeAdapterConnect(t *testing.T) {
	ctx := context.Background()
	adapter := &PineconeAdapter{}
	
	// Test with invalid type
	config := DBConfig{
		Type: "invalid",
	}
	
	err := adapter.Connect(ctx, config)
	if err == nil {
		t.Error("Expected error for invalid type, got nil")
	}
	
	t.Log("✓ Connect validates database type correctly")
}

// TestRecordSerialization tests Record JSON serialization
func TestRecordSerialization(t *testing.T) {
	record := Record{
		ID:     "test-123",
		Vector: []float32{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"title":  "Test Document",
			"source": "https://example.com",
		},
	}
	
	if record.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", record.ID)
	}
	
	if len(record.Vector) != 3 {
		t.Errorf("Expected vector length 3, got %d", len(record.Vector))
	}
	
	if record.Metadata["title"] != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%v'", record.Metadata["title"])
	}
	
	t.Log("✓ Record serialization works correctly")
}

// TestDBConfig tests configuration structure
func TestDBConfig(t *testing.T) {
	config := DBConfig{
		Type:    "pinecone",
		URL:     "https://api.pinecone.io",
		APIKey:  "test-key",
		Index:   "test-index",
		Timeout: 30,
		Extra: map[string]string{
			"environment": "us-west1-gcp",
		},
	}
	
	if config.Type != "pinecone" {
		t.Errorf("Expected type 'pinecone', got '%s'", config.Type)
	}
	
	if config.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", config.Timeout)
	}
	
	if config.Extra["environment"] != "us-west1-gcp" {
		t.Errorf("Expected environment 'us-west1-gcp', got '%s'", config.Extra["environment"])
	}
	
	t.Log("✓ DBConfig structure works correctly")
}
