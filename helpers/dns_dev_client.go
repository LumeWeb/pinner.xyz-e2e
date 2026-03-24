package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DNSDevRecord represents a DNS record in the dev server
type DNSDevRecord struct {
	Name     string  `json:"name"`  // @ for zone apex, or subdomain like _dnslink
	Type     string  `json:"type"`
	Value    string  `json:"value"`
	TTL      int     `json:"ttl"`
	Priority *int    `json:"priority,omitempty"`
}

// DNSDevZone represents a DNS zone created via the dev server API
type DNSDevZone struct {
	Domain   string        `json:"domain"`
	Records  []DNSDevRecord `json:"records"`
}

// DNSDevZoneResponse is the response from POST /api/zones
type DNSDevZoneResponse struct {
	Message string `json:"message"`
	Domain  string `json:"domain"`
}

// DNSDevZonesResponse is the response from GET /api/zones
type DNSDevZonesResponse struct {
	Zones  []DNSDevZone `json:"zones"`
	Count  int          `json:"count"`
}

// DNSDevZoneDetail is the response from GET /api/zones/{domain}
type DNSDevZoneDetail struct {
	Domain  string          `json:"domain"`
	Records []DNSDevRecord  `json:"records"`
	Count   int             `json:"count"`
}

// DNSDevHealthResponse is the response from GET /health
type DNSDevHealthResponse struct {
	Status     string `json:"status"`
	ZonesCount int    `json:"zones_count"`
}

// DNSDevClient is a mini HTTP client for the DNS development server
type DNSDevClient struct {
	baseURL    string
	httpClient *http.Client
}

// DefaultDNSDevPort is the default port for the DNS dev server API
const DefaultDNSDevPort = 8000
const DefaultDNSDevHost = "127.0.0.1"

// DefaultDNSDevClient creates a DNS dev client with default configuration
func DefaultDNSDevClient() *DNSDevClient {
	return NewDNSDevClient(DefaultDNSDevHost, DefaultDNSDevPort)
}

// NewDNSDevClient creates a DNS dev server API client
func NewDNSDevClient(host string, port int) *DNSDevClient {
	return &DNSDevClient{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Health checks the DNS dev server health
func (c *DNSDevClient) Health(ctx context.Context) (*DNSDevHealthResponse, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get health status: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}
	
	var health DNSDevHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}
	
	return &health, nil
}

// ListZones lists all DNS zones in the dev server
func (c *DNSDevClient) ListZones(ctx context.Context) ([]DNSDevZone, error) {
	url := fmt.Sprintf("%s/api/zones", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list zones failed with status %d", resp.StatusCode)
	}
	
	var response DNSDevZonesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode zones response: %w", err)
	}
	
	return response.Zones, nil
}

// GetZone gets a specific DNS zone by domain
func (c *DNSDevClient) GetZone(ctx context.Context, domain string) (*DNSDevZoneDetail, error) {
	url := fmt.Sprintf("%s/api/zones/%s", c.baseURL, domain)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("zone not found: %s", domain)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get zone failed with status %d", resp.StatusCode)
	}
	
	var zone DNSDevZoneDetail
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return nil, fmt.Errorf("failed to decode zone response: %w", err)
	}
	
	return &zone, nil
}

// AddZone adds a new DNS zone with records to the dev server
func (c *DNSDevClient) AddZone(ctx context.Context, zone DNSDevZone) (*DNSDevZoneResponse, error) {
	url := fmt.Sprintf("%s/api/zones", c.baseURL)
	
	body, err := json.Marshal(zone)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal zone: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to add zone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("add zone failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var response DNSDevZoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode add zone response: %w", err)
	}

	return &response, nil
}

// DeleteZone deletes a DNS zone from the dev server
func (c *DNSDevClient) DeleteZone(ctx context.Context, domain string) error {
	url := fmt.Sprintf("%s/api/zones/%s", c.baseURL, domain)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil // Zone not found is not an error for cleanup
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete zone failed with status %d", resp.StatusCode)
	}
	
	return nil
}

// DNSDevClientKey is the context key for storing DNS dev client
const DNSDevClientKey contextKey = "dns_dev_client"

// GetDNSDevClient retrieves the DNS dev client from context
func GetDNSDevClient(ctx context.Context) (*DNSDevClient, bool) {
	client, ok := ctx.Value(DNSDevClientKey).(*DNSDevClient)
	return client, ok && client != nil
}

// SetDNSDevClient stores the DNS dev client in context
func SetDNSDevClient(ctx context.Context, client *DNSDevClient) context.Context {
	return context.WithValue(ctx, DNSDevClientKey, client)
}

// RequireDNSDevClient retrieves the DNS dev client from context or returns an error
func RequireDNSDevClient(ctx context.Context) (*DNSDevClient, error) {
	client, ok := GetDNSDevClient(ctx)
	if !ok {
		return nil, fmt.Errorf("DNS dev client not found in context")
	}
	return client, nil
}

// EnsureDNSDevClient ensures a DNS dev client exists in context
func EnsureDNSDevClient(ctx context.Context) context.Context {
	if client, ok := GetDNSDevClient(ctx); ok && client != nil {
		return ctx
	}
	
	// Create default client and store in context
	return SetDNSDevClient(ctx, DefaultDNSDevClient())
}
