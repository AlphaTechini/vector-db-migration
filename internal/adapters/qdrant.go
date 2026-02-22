package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// QdrantAdapter implements Database interface for Qdrant
type QdrantAdapter struct {
	config     DBConfig
	httpClient *http.Client
	baseURL    string
	sourceURL  string
}

// qdrantPoint represents Qdrant's point format
type qdrantPoint struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// qdrantUpsertRequest represents Qdrant upsert request
type qdrantUpsertRequest struct {
	Collection string         `json:"collection"`
	Points     []qdrantPoint  `json:"points"`
}

// qdrantSearchRequest represents Qdrant search request
type qdrantSearchRequest struct {
	Vector []float32 `json:"vector"`
	Limit  int       `json:"limit"`
	Offset string    `json:"offset,omitempty"`
}

// qdrantSearchResponse represents Qdrant search response
type qdrantSearchResponse struct {
	Result []qdrantPoint `json:"result"`
}

// Connect establishes connection to Qdrant
func (a *QdrantAdapter) Connect(ctx context.Context, config DBConfig) error {
	if config.Type != "qdrant" {
		return fmt.Errorf("expected type 'qdrant', got '%s'", config.Type)
	}
	
	a.config = config
	a.sourceURL = config.URL
	a.baseURL = config.URL
	
	// Create HTTP client with timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	a.httpClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	
	// Validate connection
	return a.ValidateConnection(ctx)
}

// Close closes the HTTP client
func (a *QdrantAdapter) Close() error {
	if a.httpClient != nil {
		a.httpClient.CloseIdleConnections()
	}
	return nil
}

// GetBatch retrieves a batch of records from Qdrant
func (a *QdrantAdapter) GetBatch(ctx context.Context, afterID string, limit int) ([]Record, error) {
	url := fmt.Sprintf("%s/collections/%s/points/scroll", a.baseURL, a.config.Index)
	
	request := struct {
		Limit  int    `json:"limit"`
		Offset string `json:"offset,omitempty"`
		WithPayload bool `json:"with_payload"`
		WithVector bool `json:"with_vector"`
	}{
		Limit:       limit,
		Offset:      afterID,
		WithPayload: true,
		WithVector:  true,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to scroll Qdrant: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Qdrant API error (%d): %s", resp.StatusCode, string(body))
	}
	
	var scrollResp struct {
		Result struct {
			Points []qdrantPoint `json:"points"`
			NextPageOffset string `json:"next_page_offset"`
		} `json:"result"`
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&scrollResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to our Record format
	records := make([]Record, len(scrollResp.Result.Points))
	for i, p := range scrollResp.Result.Points {
		records[i] = Record{
			ID:       p.ID,
			Vector:   p.Vector,
			Metadata: p.Payload,
		}
	}
	
	return records, nil
}

// UpsertBatch inserts or updates records in Qdrant
func (a *QdrantAdapter) UpsertBatch(ctx context.Context, records []Record) error {
	url := fmt.Sprintf("%s/collections/%s/points", a.baseURL, a.config.Index)
	
	// Convert to Qdrant format
	points := make([]qdrantPoint, len(records))
	for i, r := range records {
		points[i] = qdrantPoint{
			ID:      r.ID,
			Vector:  r.Vector,
			Payload: r.Metadata,
		}
	}
	
	request := qdrantUpsertRequest{
		Collection: a.config.Index,
		Points:     points,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upsert to Qdrant: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Qdrant API error (%d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// DeleteBatch deletes records from Qdrant by IDs
func (a *QdrantAdapter) DeleteBatch(ctx context.Context, ids []string) error {
	url := fmt.Sprintf("%s/collections/%s/points/delete", a.baseURL, a.config.Index)
	
	request := struct {
		Points []string `json:"points"`
	}{
		Points: ids,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete from Qdrant: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Qdrant API error (%d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// ValidateConnection checks if Qdrant is accessible
func (a *QdrantAdapter) ValidateConnection(ctx context.Context) error {
	// Check cluster status
	url := fmt.Sprintf("%s/cluster", a.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Qdrant: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Qdrant connection failed (status %d)", resp.StatusCode)
	}
	
	return nil
}

// GetStats returns Qdrant statistics
func (a *QdrantAdapter) GetStats(ctx context.Context) (*DBStats, error) {
	// Get collection info
	url := fmt.Sprintf("%s/collections/%s", a.baseURL, a.config.Index)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats request: %w", err)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats from Qdrant: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Qdrant API error (%d)", resp.StatusCode)
	}
	
	var collectionInfo struct {
		Result struct {
			Status      string `json:"status"`
			VectorsCount int64  `json:"vectors_count"`
			PointsCount int64  `json:"points_count"`
			Config      struct {
				Params struct {
					Vectors struct {
						Size     int    `json:"size"`
						Distance string `json:"distance"`
					} `json:"vectors"`
				} `json:"params"`
			} `json:"config"`
		} `json:"result"`
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&collectionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}
	
	return &DBStats{
		TotalRecords: collectionInfo.Result.VectorsCount,
		Dimensions:   collectionInfo.Result.Config.Params.Vectors.Size,
		IndexType:    "qdrant-hnsw",
		MemoryUsage:  0, // Not available via API
	}, nil
}

// GetSourceURL returns the Qdrant source URL
func (a *QdrantAdapter) GetSourceURL() string {
	return a.sourceURL
}

// Ensure QdrantAdapter implements Database interface
var _ Database = (*QdrantAdapter)(nil)
