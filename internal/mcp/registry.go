package mcp

import (
	"context"
	"fmt"
	"sync"
)

// ToolHandler executes a tool with given parameters
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	Schema      map[string]interface{}
	Handler     ToolHandler
}

// ToolRegistry manages registered tools
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]*Tool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*Tool),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(tool *Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool.Name == "" {
		return fmt.Errorf("tool name is required")
	}

	if tool.Handler == nil {
		return fmt.Errorf("tool handler is required for %s", tool.Name)
	}

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}

	r.tools[tool.Name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (*Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

// List returns all registered tools
func (r *ToolRegistry) List() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}

	return tools
}

// Execute runs a tool with the given parameters
func (r *ToolRegistry) Execute(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	return tool.Handler(ctx, params)
}
