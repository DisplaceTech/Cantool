package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"


	"github.com/displacetech/cantool/internal/output"
)

// HealthResponse describes the node health status.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// Party represents a ledger party.
type Party struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	IsLocal     bool   `json:"is_local"`
}

// Contract represents an active contract.
type Contract struct {
	ContractID string         `json:"contract_id"`
	TemplateID string         `json:"template_id"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  string         `json:"created_at"`
}

// Package represents an uploaded DAR package.
type Package struct {
	PackageID  string `json:"package_id"`
	Size       int64  `json:"size"`
	UploadedAt string `json:"uploaded_at"`
}

// LedgerClient is the interface for Canton Ledger API interactions.
type LedgerClient interface {
	Health(ctx context.Context) (*HealthResponse, error)
	ListParties(ctx context.Context) ([]Party, error)
	AllocateParty(ctx context.Context, name, displayName string) (*Party, error)
	ListContracts(ctx context.Context, party, templateID string) ([]Contract, error)
	UploadDAR(ctx context.Context, darPath string) error
	GetPackages(ctx context.Context) ([]Package, error)
}

// ClientOption configures an HTTPLedgerClient.
type ClientOption func(*HTTPLedgerClient)

// WithToken sets the JWT auth token.
func WithToken(token string) ClientOption {
	return func(c *HTTPLedgerClient) { c.token = token }
}

// WithTimeout sets the HTTP timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *HTTPLedgerClient) { c.httpClient.Timeout = d }
}

// WithHTTPClient injects a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *HTTPLedgerClient) { c.httpClient = hc }
}

// HTTPLedgerClient implements LedgerClient using HTTP.
type HTTPLedgerClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewHTTPLedgerClient creates a new client for the given base URL.
func NewHTTPLedgerClient(baseURL string, opts ...ClientOption) *HTTPLedgerClient {
	c := &HTTPLedgerClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *HTTPLedgerClient) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isConnectionRefused(err) {
			return nil, output.Wrapf(err, "CT6001",
				"Run `cantool dev` to start a local node, or check `cantool env list`",
				"cannot connect to %s", c.baseURL)
		}
		if ctx.Err() != nil {
			return nil, output.Wrapf(err, "CT6003",
				"The node may be slow to respond. Try again or increase the timeout",
				"request to %s timed out", path)
		}
		return nil, output.Wrapf(err, "CT6004",
			"Check that the node is running and accessible",
			"request to %s failed", path)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, output.Wrapf(err, "CT6004",
			"Unexpected response from node",
			"failed to read response body")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, output.Errorf("CT6002",
			"Update your auth token with `cantool env add <name> --token=<jwt>`",
			"authentication failed (HTTP 401)")
	}

	if resp.StatusCode >= 400 {
		return nil, output.Errorf("CT6004",
			"Check the node logs for details",
			"unexpected HTTP %d from %s: %s", resp.StatusCode, path, string(data))
	}

	return data, nil
}

func isConnectionRefused(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr)
}

func decodeJSON[T any](data []byte) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, output.Wrapf(err, "CT6004",
			"The node returned an unexpected response format",
			"failed to decode JSON response")
	}
	return v, nil
}

func encodeJSON(v any) (io.Reader, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}
	return bytes.NewReader(data), nil
}
