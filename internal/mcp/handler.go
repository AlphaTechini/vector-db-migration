package mcp

import (
	"context"
	"encoding/json"
)

// RequestHandler processes MCP requests with middleware support
type RequestHandler struct {
	registry *ToolRegistry
}

// NewRequestHandler creates a new request handler
func NewRequestHandler(registry *ToolRegistry) *RequestHandler {
	return &RequestHandler{
		registry: registry,
	}
}

// Handle processes a JSON-RPC request and returns a response
func (h *RequestHandler) Handle(ctx context.Context, reqBytes []byte) []byte {
	// Parse request
	var req Request
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		return h.errorResponse(nil, ParseError, "invalid JSON: "+err.Error())
	}

	// Validate version
	if req.JSONRPC != "2.0" {
		return h.errorResponse(req.ID, InvalidRequest, "invalid JSON-RPC version")
	}

	// Execute tool
	result, err := h.executeTool(ctx, req)
	if err != nil {
		return h.errorResponse(req.ID, InternalError, err.Error())
	}

	// Return success response
	return h.successResponse(req.ID, result)
}

// executeTool runs the requested tool
func (h *RequestHandler) executeTool(ctx context.Context, req Request) (interface{}, error) {
	params := h.parseParams(req.Params)
	return h.registry.Execute(ctx, req.Method, params)
}

// parseParams safely converts raw JSON to params map
func (h *RequestHandler) parseParams(raw json.RawMessage) map[string]interface{} {
	if raw == nil {
		return make(map[string]interface{})
	}

	var params map[string]interface{}
	if err := json.Unmarshal(raw, &params); err != nil {
		return make(map[string]interface{})
	}

	return params
}

// successResponse creates a JSON-RPC success response
func (h *RequestHandler) successResponse(id interface{}, result interface{}) []byte {
	response := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	bytes, _ := json.Marshal(response)
	return bytes
}

// errorResponse creates a JSON-RPC error response
func (h *RequestHandler) errorResponse(id interface{}, code int, message string) []byte {
	response := ErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: RPCError{
			Code:    code,
			Message: message,
		},
	}

	bytes, _ := json.Marshal(response)
	return bytes
}
