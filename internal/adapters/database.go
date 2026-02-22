package adapters

import (
	"context"
)

// Record represents a vector record with metadata
type Record struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DBStats holds database statistics
type DBStats struct {
	TotalRecords int64   `json:"total_records"`
	Dimensions   int     `json:"dimensions"`
	IndexType    string  `json:"index_type"`
	MemoryUsage  float64 `json:"memory_usage_mb"`
}

// Database interface for vector database operations
type Database interface {
	// Connect establishes connection to the database
	Connect(ctx context.Context, config DBConfig) error
	
	// Close closes the database connection
	Close() error
	
	// GetBatch retrieves a batch of records after the given ID
	GetBatch(ctx context.Context, afterID string, limit int) ([]Record, error)
	
	// UpsertBatch inserts or updates a batch of records
	UpsertBatch(ctx context.Context, records []Record) error
	
	// DeleteBatch deletes records by IDs
	DeleteBatch(ctx context.Context, ids []string) error
	
	// ValidateConnection checks if the database is accessible
	ValidateConnection(ctx context.Context) error
	
	// GetStats returns database statistics
	GetStats(ctx context.Context) (*DBStats, error)
	
	// GetSourceURL returns the database source URL (for logging)
	GetSourceURL() string
}

// DBConfig holds database connection configuration
type DBConfig struct {
	Type     string            `json:"type"` // pinecone, qdrant, weaviate
	URL      string            `json:"url"`
	APIKey   string            `json:"api_key"`
	Index    string            `json:"index"` // Pinecone index name / Qdrant collection
	Timeout  int               `json:"timeout_seconds"`
	Extra    map[string]string `json:"extra,omitempty"` // Provider-specific settings
}
