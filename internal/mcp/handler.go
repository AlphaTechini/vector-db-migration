// Package mcp implements the MCP (Model Context Protocol) server.
//
// NOTE: RequestHandler was removed. All JSON-RPC request handling is done
// directly by Server.handleRequest in server.go, which calls the ToolRegistry.
// Use Server and ToolRegistry instead.
package mcp
