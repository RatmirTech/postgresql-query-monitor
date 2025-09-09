package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ratmirtech/postgresql-query-monitor/internal/collectors"
	"github.com/ratmirtech/postgresql-query-monitor/internal/models"
)

// Client represents the PostgreSQL config analyzer client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new analyzer client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AnalyzeConfig sends server configuration for analysis
func (c *Client) AnalyzeConfig(ctx context.Context, serverData models.ServerData, isSchedulerTask bool) (*models.Recommendation, error) {
	if isSchedulerTask {
		c.baseURL = c.BaseURL() + "/scheduler"
	}
	url := fmt.Sprintf("%s/config/analyze", c.baseURL)

	fmt.Printf("Using Review API URL: %s\n", url)

	// Create request body
	requestBody := serverData

	// Marshal to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("%s, %s", req.Body, req.URL.String())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Println(string(body))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var report models.Recommendation
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &report, nil
}

// ReviewSingleQuery sends a single SQL query for analysis
func (c *Client) ReviewSingleQuery(ctx context.Context, query models.QueryReviewRequest) (*models.QueryReviewResponse, error) {
	url := fmt.Sprintf("%s/review/", c.baseURL)

	// Marshal to JSON
	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response models.QueryReviewResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ReviewBatchQueries sends multiple SQL queries for batch analysis
func (c *Client) ReviewBatchQueries(ctx context.Context, batch models.BatchReviewRequest) (*models.BatchReviewResponse, error) {
	url := fmt.Sprintf("%s/review/batch", c.baseURL)

	// Marshal to JSON
	jsonData, err := json.Marshal(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response models.BatchReviewResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ReviewMigration sends a migration SQL script for analysis
func (c *Client) ReviewMigration(ctx context.Context, migration models.MigrationReviewRequest) (*models.MigrationReviewResponse, error) {
	url := fmt.Sprintf("%s/review/", c.baseURL)

	// Marshal to JSON
	jsonData, err := json.Marshal(migration)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response models.MigrationReviewResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// WithHTTPClient allows to set custom HTTP client
func (c *Client) WithHTTPClient(client *http.Client) *Client {
	c.httpClient = client
	return c
}

// WithTimeout sets custom timeout for HTTP requests
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// BaseURL returns the base URL of the client
func (c *Client) BaseURL() string {
	return c.baseURL
}

// SetBaseURL sets the base URL of the client
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// AnalyzeSystemMetrics sends system metrics for analysis
func (c *Client) AnalyzeSystemMetrics(ctx context.Context, metrics collectors.SystemMetrics, serverInfo models.ServerInfo, environment string) (*models.Recommendation, error) {
	url := fmt.Sprintf("%s/config/analyze", c.baseURL)

	// Create request body
	requestBody := struct {
		Config      collectors.SystemMetrics `json:"config"`
		Environment string                   `json:"environment"`
		ServerInfo  models.ServerInfo        `json:"server_info"`
	}{
		Config:      metrics,
		Environment: environment,
		ServerInfo:  serverInfo,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var report models.Recommendation
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &report, nil
}
