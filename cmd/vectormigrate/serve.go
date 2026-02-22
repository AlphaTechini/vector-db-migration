package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlphaTechini/vector-db-migration/internal/mcp"
	"github.com/AlphaTechini/vector-db-migration/internal/mcp/tools"
	"github.com/spf13/cobra"
)

var (
	mcpAddr string
	apiKey  string

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start MCP server",
		Long:  "Start the Model Context Protocol (MCP) server for AI assistant integration.",
		RunE:  runServe,
	}
)

func init() {
	serveCmd.Flags().StringVar(&mcpAddr, "addr", ":8080", "Address to listen on")
	serveCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for authentication (required)")
	serveCmd.MarkFlagRequired("api-key")
}

func runServe(cmd *cobra.Command, args []string) error {
	log.Printf("🚀 Starting MCP server...")
	log.Printf("   Address: %s", mcpAddr)
	log.Printf("   API Key: %s", maskAPIKey(apiKey))

	// Create context with cancellation
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("\n🛑 Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Create state tracker
	stateTracker, err := createStateTracker("")
	if err != nil {
		return fmt.Errorf("failed to create state tracker: %w", err)
	}
	defer stateTracker.Close()

	// Create tool registry
	registry := mcp.NewToolRegistry()

	// Register tools
	log.Println("   🔧 Registering tools...")

	statusTool := tools.NewMigrationStatusTool(stateTracker)
	if err := statusTool.Register(registry); err != nil {
		return fmt.Errorf("failed to register migration_status tool: %w", err)
	}
	log.Println("   ✅ Registered: migration_status")

	// TODO: Register more tools as they're implemented
	// listTool := tools.NewListMigrationsTool()
	// listTool.Register(registry)

	// Create MCP server with middleware
	server := mcp.NewServer(mcpAddr, registry,
		mcp.WithAPIKey(apiKey),
		mcp.WithRateLimit(100, 20), // 100 req/min, burst of 20
		mcp.WithAuditLog(log.Default()),
	)

	// Start server
	log.Println("   ▶️  Starting HTTP server...")
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	<-ctx.Done()
	log.Println("✅ MCP server stopped")
	return nil
}

// maskAPIKey hides most of the API key for logging
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
