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

// WeaviateAdapter implements Database interface for Weaviate
type WeaviateAdapter struct {
	config     DBConfig
	httpClient *http.Client
	baseURL    string
	sourceURL  string
	className  string
}

// weaviateObject represents Weaviate's object format
type weaviateObject struct {
	Class      string                 `json:"class"`
	ID         string                 `json:"id"`
	Vector     []float32              `json:"vector"`
	Properties map[string]interface{} `json:"properties"`
}

// weaviateGetResponse represents Weaviate get response
type weaviateGetResponse struct {
	Result []struct {
		ID         string                 `json:"id"`
		Vector     []float32              `json:"vector"`
		Properties map[string]interface{} `json:"properties"`
	} `json:"data"`
}

// Connect establishes connection to Weaviate
func (a *WeaviateAdapter) Connect(ctx context.Context, config DBConfig) error {
	if config.Type != "weaviate" {
		return fmt.Errorf("expected type 'weaviate', got '%s'", config.Type)
	}
	
	a.config = config
	a.sourceURL = config.URL
	a.baseURL = config.URL
	a.className = config.Index // Weaviate uses "class" instead of "index"
	
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
func (a *WeaviateAdapter) Close() error {
	if a.httpClient != nil {
		a.httpClient.CloseIdleConnections()
	}
	return nil
}

// GetBatch retrieves a batch of objects from Weaviate
func (a *WeaviateAdapter) GetBatch(ctx context.Context, afterID string, limit int) ([]Record, error) {
	// Use GraphQL-style query via REST
	query := fmt.Sprintf(`
		{
			Get {
				%s(limit: %d, after: "%s") {
					_additional {
						id
						vector
					}
				}
			}
		}
	`, a.className, limit, afterID)
	
	request := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	url := fmt.Sprintf("%s/v1/graphql", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if a.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Weaviate: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Weaviate API error (%d): %s", resp.StatusCode, string(body))
	}
	
	var graphqlResp struct {
		Data struct {
			Get []map[string]interface{} `json:"Get"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors,omitempty"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("Weaviate GraphQL error: %s", graphqlResp.Errors[0].Message)
	}
	
	// Extract objects from response
	objects := graphqlResp.Data.Get
	if len(objects) == 0 {
		return []Record{}, nil
	}
	
	// Get the class data
	classData, ok := objects[0][a.className]
	if !ok {
		return []Record{}, nil
	}
	
	items, ok := classData.([]interface{})
	if !ok {
		return []Record{}, nil
	}
	
	// Convert to our Record format
	records := make([]Record, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		record := Record{
			Metadata: make(map[string]interface{}),
		}
		
		// Extract ID and vector from _additional
		if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
			if id, ok := additional["id"].(string); ok {
				record.ID = id
			}
			if vector, ok := additional["vector"].([]interface{}); ok {
				record.Vector = make([]float32, len(vector))
				for i, v := range vector {
					if vf, ok := v.(float64); ok {
						record.Vector[i] = float32(vf)
					}
				}
			}
		}
		
		// Copy properties to metadata
		for key, value := range itemMap {
			if key != "_additional" {
				record.Metadata[key] = value
			}
		}
		
		if record.ID != "" {
			records = append(records, record)
		}
	}
	
	return records, nil
}

// UpsertBatch inserts or updates objects in Weaviate
func (a *WeaviateAdapter) UpsertBatch(ctx context.Context, records []Record) error {
	// Batch upsert using REST API
	url := fmt.Sprintf("%s/v1/batch/objects", a.baseURL)
	
	// Convert to Weaviate format
	objects := make([]weaviateObject, len(records))
	for i, r := range records {
		objects[i] = weaviateObject{
			Class:      a.className,
			ID:         r.ID,
			Vector:     r.Vector,
			Properties: r.Metadata,
		}
	}
	
	request := struct {
		Fields []string        `json:"fields"`
		Objects []weaviateObject `json:"objects"`
	}{
		Fields:  []string{"ALL"},
		Objects: objects,
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
	if a.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to batch upsert to Weaviate: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Weaviate API error (%d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// DeleteBatch deletes objects from Weaviate by IDs
func (a *WeaviateAdapter) DeleteBatch(ctx context.Context, ids []string) error {
	// Delete each object individually (Weaviate doesn't support batch delete by ID list)
	for _, id := range ids {
		url := fmt.Sprintf("%s/v1/objects/%s/%s", a.baseURL, a.className, id)
		
		req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create delete request: %w", err)
		}
		
		if a.config.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
		}
		
		resp, err := a.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to delete from Weaviate: %w", err)
		}
		resp.Body.Close()
	}
	
	return nil
}

// ValidateConnection checks if Weaviate is accessible
func (a *WeaviateAdapter) ValidateConnection(ctx context.Context) error {
	// Check readiness endpoint
	url := fmt.Sprintf("%s/v1/.well-known/ready", a.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}
	
	if a.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Weaviate: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Weaviate connection failed (status %d)", resp.StatusCode)
	}
	
	return nil
}

// GetStats returns Weaviate statistics
func (a *WeaviateAdapter) GetStats(ctx context.Context) (*DBStats, error) {
	// Get class schema
	url := fmt.Sprintf("%s/v1/schema/%s", a.baseURL, a.className)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create stats request: %w", err)
	}
	
	if a.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats from Weaviate: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Weaviate API error (%d)", resp.StatusCode)
	}
	
	var classSchema struct {
		Class             string `json:"class"`
		VectorIndexType   string `json:"vectorIndexType"`
		VectorIndexConfig struct {
			Distance string `json:"distance"`
		} `json:"vectorIndexConfig"`
		Properties []struct {
			Name     string `json:"name"`
			DataType []string `json:"dataType"`
		} `json:"properties"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&classSchema); err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}
	
	// Get object count via aggregate query
	aggQuery := fmt.Sprintf(`
		{
			Aggregate {
				%s {
					meta {
						count
					}
				}
			}
		}
	`, a.className)
	
	aggRequest := struct {
		Query string `json:"query"`
	}{
		Query: aggQuery,
	}
	
	jsonData, err := json.Marshal(aggRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal aggregate request: %w", err)
	}
	
	aggURL := fmt.Sprintf("%s/v1/graphql", a.baseURL)
	aggReq, err := http.NewRequestWithContext(ctx, "POST", aggURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create aggregate request: %w", err)
	}
	
	aggReq.Header.Set("Content-Type", "application/json")
	if a.config.APIKey != "" {
		aggReq.Header.Set("Authorization", "Bearer "+a.config.APIKey)
	}
	
	aggResp, err := a.httpClient.Do(aggReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregate: %w", err)
	}
	defer aggResp.Body.Close()
	
	var aggGraphQLResp struct {
		Data struct {
			Aggregate struct {
				Class []struct {
					Meta struct {
						Count int64 `json:"count"`
					} `json:"meta"`
				} `json:""`
			} `json:"Aggregate"`
		} `json:"data"`
	}
	
	// Default stats if we can't get count
	stats := &DBStats{
		TotalRecords: 0,
		Dimensions:   0, // Not available in schema
		IndexType:    classSchema.VectorIndexType,
		MemoryUsage:  0,
	}
	
	// Try to extract count
	if len(aggGraphQLResp.Data.Aggregate.Class) > 0 {
		stats.TotalRecords = aggGraphQLResp.Data.Aggregate.Class[0].Meta.Count
	}
	
	return stats, nil
}

// GetSourceURL returns the Weaviate source URL
func (a *WeaviateAdapter) GetSourceURL() string {
	return a.sourceURL
}

// Ensure WeaviateAdapter implements Database interface
var _ Database = (*WeaviateAdapter)(nil)
