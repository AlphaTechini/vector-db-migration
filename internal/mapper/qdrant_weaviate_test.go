package mapper

import (
	"testing"

	"github.com/AlphaTechini/vector-db-migration/internal/adapters"
	"github.com/google/uuid"
)

func TestQdrantWeaviateMapper_MapRecord(t *testing.T) {
	m := NewQdrantWeaviateMapper()

	t.Run("Deterministic_UUIDv5_ID_Mapping", func(t *testing.T) {
		rec := adapters.Record{
			ID: "product_123",
			Metadata: map[string]interface{}{
				"name": "Widget",
			},
		}

		mapped, err := m.MapRecord(rec, nil)
		if err != nil {
			t.Fatalf("Failed to map record: %v", err)
		}

		// Verify UUIDv5 format
		_, err = uuid.Parse(mapped.ID)
		if err != nil {
			t.Errorf("Mapped ID is not a valid UUID: %v", err)
		}

		// Verify determinism
		expectedUUID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("product_123")).String()
		if mapped.ID != expectedUUID {
			t.Errorf("Expected UUID %s, got %s", expectedUUID, mapped.ID)
		}

		// Verify traceability
		if mapped.Metadata["_original_id"] != "product_123" {
			t.Errorf("Original ID not preserved in metadata, got %v", mapped.Metadata["_original_id"])
		}
	})

	t.Run("Lossless_Nested_Payload_Mapping", func(t *testing.T) {
		nestedPayload := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "Alice",
				"address": map[string]interface{}{
					"city": "London",
				},
			},
			"tags": []interface{}{"ai", "search"},
		}

		rec := adapters.Record{
			ID:       "user_456",
			Metadata: nestedPayload,
		}

		mapped, err := m.MapRecord(rec, nil)
		if err != nil {
			t.Fatalf("Failed to map record: %v", err)
		}

		// Verify deep structure is preserved
		userObj, ok := mapped.Metadata["user"].(map[string]interface{})
		if !ok {
			t.Fatal("Nested 'user' object not preserved")
		}

		addressObj, ok := userObj["address"].(map[string]interface{})
		if !ok {
			t.Fatal("Nested 'address' object not preserved")
		}

		if addressObj["city"] != "London" {
			t.Errorf("Expected city 'London', got %v", addressObj["city"])
		}
	})
}
