package tools

import (
	"testing"
)

type TestStruct struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Score   float64 `json:"score"`
	IsOk    bool    `json:"is_ok"`
	Tags    []string `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

type NestedStruct struct {
	ID   string     `json:"id"`
	Data TestStruct `json:"data"`
}

func TestDecodeParams_Basic(t *testing.T) {
	// Payloads coming from JSON-RPC are typically float64 for all numbers
	// and []interface{} for all arrays.
	input := map[string]interface{}{
		"name":  "jules",
		"age":   float64(30),
		"score": 95.5,
		"is_ok": true,
		"tags":  []interface{}{"tag1", "tag2"},
		"metadata": map[string]interface{}{
			"key": "val",
		},
	}

	var output TestStruct
	err := DecodeParams(input, &output)
	if err != nil {
		t.Fatalf("DecodeParams failed: %v", err)
	}

	if output.Name != "jules" {
		t.Errorf("Expected name jules, got %s", output.Name)
	}
	if output.Age != 30 {
		t.Errorf("Expected age 30, got %d", output.Age)
	}
	if output.Score != 95.5 {
		t.Errorf("Expected score 95.5, got %f", output.Score)
	}
	if !output.IsOk {
		t.Error("Expected is_ok to be true")
	}
	if len(output.Tags) != 2 || output.Tags[0] != "tag1" || output.Tags[1] != "tag2" {
		t.Errorf("Expected tags [tag1 tag2], got %v", output.Tags)
	}
}

func TestDecodeParams_WeaklyTyped(t *testing.T) {
	// JSON often unmarshals numbers as float64
	input := map[string]interface{}{
		"age": float64(25), // float64 instead of int
	}

	var output TestStruct
	err := DecodeParams(input, &output)
	if err != nil {
		t.Fatalf("DecodeParams failed: %v", err)
	}

	if output.Age != 25 {
		t.Errorf("Expected age 25 (coerced from float64), got %d", output.Age)
	}
}

func TestDecodeParams_Nested(t *testing.T) {
	input := map[string]interface{}{
		"id": "nest-1",
		"data": map[string]interface{}{
			"name":  "nested",
			"age":   float64(10),
			"score": float64(88.5),
		},
	}

	var output NestedStruct
	err := DecodeParams(input, &output)
	if err != nil {
		t.Fatalf("DecodeParams failed: %v", err)
	}

	if output.ID != "nest-1" {
		t.Errorf("Expected ID nest-1, got %s", output.ID)
	}
	if output.Data.Name != "nested" {
		t.Errorf("Expected nested name, got %s", output.Data.Name)
	}
	if output.Data.Age != 10 {
		t.Errorf("Expected nested age 10, got %d", output.Data.Age)
	}
	if output.Data.Score != 88.5 {
		t.Errorf("Expected nested score 88.5, got %f", output.Data.Score)
	}
}

func TestDecodeParams_Error(t *testing.T) {
	input := map[string]interface{}{
		"age": "not a number",
	}

	var output TestStruct
	err := DecodeParams(input, &output)
	if err == nil {
		t.Error("Expected error for invalid type conversion, got nil")
	}
}
