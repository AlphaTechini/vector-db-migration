package mcp

import (
	"context"
	"testing"
)

func TestToolRegistry_Register(t *testing.T) {
	registry := NewToolRegistry()

	// Test successful registration
	tool := &Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Test duplicate registration fails
	err = registry.Register(tool)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}

	// Test missing name fails
	badTool := &Tool{
		Handler: tool.Handler,
	}
	err = registry.Register(badTool)
	if err == nil {
		t.Error("Expected error for missing tool name")
	}

	// Test missing handler fails
	noHandler := &Tool{
		Name: "no_handler",
	}
	err = registry.Register(noHandler)
	if err == nil {
		t.Error("Expected error for missing handler")
	}
}

func TestToolRegistry_Get(t *testing.T) {
	registry := NewToolRegistry()

	// Register a tool
	tool := &Tool{
		Name: "get_test",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	}
	registry.Register(tool)

	// Get existing tool
	retrieved, err := registry.Get("get_test")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	if retrieved.Name != "get_test" {
		t.Errorf("Expected name 'get_test', got '%s'", retrieved.Name)
	}

	// Get non-existent tool
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewToolRegistry()

	// List empty registry
	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(tools))
	}

	// Register tools
	registry.Register(&Tool{Name: "tool1", Handler: dummyHandler})
	registry.Register(&Tool{Name: "tool2", Handler: dummyHandler})

	// List should return all tools
	tools = registry.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestToolRegistry_Execute(t *testing.T) {
	registry := NewToolRegistry()

	// Register tool that returns params
	tool := &Tool{
		Name: "echo_tool",
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params, nil
		},
	}
	registry.Register(tool)

	// Execute with params
	ctx := context.Background()
	result, err := registry.Execute(ctx, "echo_tool", map[string]interface{}{
		"key": "value",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be map[string]interface{}")
	}

	if resultMap["key"] != "value" {
		t.Errorf("Expected key='value', got '%v'", resultMap["key"])
	}

	// Execute non-existent tool
	_, err = registry.Execute(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}
}

func TestToolRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewToolRegistry()

	// Register initial tool
	registry.Register(&Tool{Name: "initial", Handler: dummyHandler})

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			registry.List()
			registry.Get("initial")
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			name := "tool" + string(rune('A'+i))
			registry.Register(&Tool{Name: name, Handler: dummyHandler})
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all tools registered
	tools := registry.List()
	if len(tools) != 11 { // initial + 10 concurrent
		t.Errorf("Expected 11 tools, got %d", len(tools))
	}
}

// dummyHandler is a no-op handler for testing
func dummyHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return nil, nil
}
