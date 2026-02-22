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

// PineconeAdapter implements Database interface for Pinecone
type PineconeAdapter struct {
	config   DBConfig
	httpClient *http.Client
	baseURL    string
	sourceURL  string
}

// pineconeRecord represents Pinecone's record format
type pineconeRecord struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// pineconeFetchResponse represents Pinecone fetch response
type pineconeFetchResponse struct {
	Vectors []pineconeRecord `json:"vectors"`
}

// Connect establishes connection to Pinecone
func (a *PineconeAdapter) Connect(ctx context.Context, config DBConfig) error {
	if config.Type != "pinecone" {
		return fmt.Errorf("expected type 'pinecone', got '%s'", config.Type)
	}
	
	a.config = config
	a.sourceURL = config.URL
	
	// Pinecone API base URL
	a.baseURL = "https://api.pinecone.io"
	
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
func (a *PineconeAdapter) Close() error {
	if a.httpClient != nil {
		a.httpClient.CloseIdleConnections()
	}
	return nil
}

// GetBatch retrieves a batch of records from Pinecone
func (a *PineconeAdapter) GetBatch(ctx context.Context, afterID string, limit int) ([]Record, error) {
	// Pinecone doesn't have native pagination, so we'll use list + fetch
	// In production, this would use Pinecone's list endpoint with pagination
	
	url := fmt.Sprintf("%s/vectors/list?index=%s&limit=%d", a.baseURL, a.config.Index, limit)
	if afterID != "" {
		url += fmt.Sprintf("&pagination_token=%s", afterID)
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Api-Key", a.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Pinecone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Pinecone API error (%d): %s", resp.StatusCode, string(body))
	}
	
	var listResp struct {
		Vectors      []pineconeRecord `json:"vectors"`
		Pagination   struct {
			NextToken string `json:"next"`
		} `json:"pagination"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Convert to our Record format
	records := make([]Record, len(listResp.Vectors))
	for i, v := range listResp.Vectors {
		records[i] = Record{
			ID:       v.ID,
			Vector:   v.Values,
			Metadata: v.Metadata,
		}
	}
	
	return records, nil
}

// UpsertBatch inserts or updates records in Pinecone
func (a *PineconeAdapter) UpsertBatch(ctx context.Context, records []Record) error {
	url := fmt.Sprintf("%s/vectors/upsert", a.baseURL)
	
	// Convert to Pinecone format
	pineconeRecords := make([]pineconeRecord, len(records))
	for i, r := range records {
		pineconeRecords[i] = pineconeRecord{
			ID:       r.ID,
			Values:   r.Vector,
			Metadata: r.Metadata,
		}
	}
	
	payload := struct {
		Vectors   []pineconeRecord `json:"vectors"`
		Namespace string           `json:"namespace,omitempty"`
	}{
		Vectors: pineconeRecords,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Api-Key", a.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upsert to Pinecone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Pinecone API error (%d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// DeleteBatch deletes records from Pinecone by IDs
func (a *PineconeAdapter) DeleteBatch(ctx context.Context, ids []string) error {
	url := fmt.Sprintf("%s/vectors/delete", a.baseURL)
	
	payload := struct {
		IDs []string `json:"ids"`
	}{
		IDs: ids,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Api-Key", a.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete from Pinecone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Pinecone API error (%d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// ValidateConnection checks if Pinecone is accessible
func (a *PineconeAdapter) ValidateConnection(ctx context.Context) error {
	// Simple health check - try to describe index
	url := fmt.Sprintf("%s/indexes/%s", a.baseURL, a.config.Index)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}
	
	req.Header.Set("Api-Key", a.config.APIKey)
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Pinecone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Pinecone connection failed (status %d)", resp.StatusCode)
	}
	
	return nil
}

// GetStats returns Pinecone statistics
func (a *PineconeAdapter) GetStats(ctx context.Context) (*DBStats, error) {
	// Describe index to get stats
	url := fmt.Sprintf("%s/indexes/%s", a.baseURL, a.config.Index)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats request: %w", err)
	}
	
	req.Header.Set("Api-Key", a.config.APIKey)
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats from Pinecone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pinecone API error (%d)", resp.StatusCode)
	}
	
	var indexInfo struct {
		Database struct {
			TotalVectorCount int64 `json:"totalVectorCount"`
		} `json:"database"`
		Dimension int `json:"dimension"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&indexInfo); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}
	
	return &DBStats{
		TotalRecords: indexInfo.Database.TotalVectorCount,
		Dimensions:   indexInfo.Dimension,
		IndexType:    "pinecone-serverless",
		MemoryUsage:  0, // Not available via API
	}, nil
}

// GetSourceURL returns the Pinecone source URL
func (a *PineconeAdapter) GetSourceURL() string {
	return a.sourceURL
}

// Ensure PineconeAdapter implements Database interface
var _ Database = (*PineconeAdapter)(nil)
