package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AdrienRCQ/goose-it/internal/contracts"
)

// client pour communiquer avec l'API GOOSE-IT
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New Client API
func New(baseURL string) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")

	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("server URL must use http or https")
	}
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("server URL must contain a host")
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// Get /healthz
// ctx == request body
func (c *Client) Health(ctx context.Context) (contracts.HealthResponse, error) {
	var result contracts.HealthResponse
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/healthz",
		nil,
	)
	if err != nil {
		return result, fmt.Errorf("create health request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return result, fmt.Errorf("contat server: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("server returned HTTP status %s", response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("decode health response: %w", err)
	}

	return result, nil
}
