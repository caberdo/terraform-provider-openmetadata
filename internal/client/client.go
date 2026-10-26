// Package client wraps the official OpenMetadata Go SDK with the pieces the
// Terraform provider needs: environment-driven configuration, username/password
// login, a stable HTTP client, and error classification.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

// maxResponseBytes bounds how much of a response body we buffer. OpenMetadata
// error payloads are small; anything larger is not worth logging verbatim.
const maxResponseBytes = 1 << 20

const (
	// DefaultAPIVersion is the OpenMetadata REST API version the SDK targets.
	DefaultAPIVersion = "v1"

	// DefaultRetryCount matches the SDK default.
	DefaultRetryCount = 3

	// DefaultRetryWait matches the SDK default.
	DefaultRetryWait = 30 * time.Second

	// DefaultTimeout bounds a single API call, including login.
	DefaultTimeout = 2 * time.Minute
)

// Config is the fully resolved provider configuration. Values are already
// merged with their environment-variable fallbacks by the provider.
type Config struct {
	// Host is the OpenMetadata server root, for example
	// https://openmetadata.example.com. It must include the scheme.
	Host string

	// Token is a JWT or personal access token. When set it takes precedence
	// over Username and Password.
	Token string

	// Username and Password are used to obtain a JWT through
	// POST /api/v1/users/login when Token is empty.
	Username string
	Password string

	// APIVersion defaults to DefaultAPIVersion.
	APIVersion string

	// RetryCount is the number of retries after the initial attempt for
	// retryable status codes.
	RetryCount int

	// RetryWait is the delay between retries.
	RetryWait time.Duration

	// HTTPClient overrides the default client. Primarily for tests.
	HTTPClient *http.Client
}

// Client lazily authenticates and memoizes the SDK client, so provider
// configuration and `terraform validate` do not require a reachable server.
type Client struct {
	cfg        Config
	httpClient *http.Client
	baseURL    string

	mu    sync.Mutex
	sdk   *ometa.Client
	token string
}

// New validates cfg and returns a client that authenticates on first use.
func New(cfg Config) (*Client, error) {
	cfg.Host = strings.TrimRight(strings.TrimSpace(cfg.Host), "/")
	if cfg.Host == "" {
		return nil, fmt.Errorf("host must be set")
	}
	if !strings.HasPrefix(cfg.Host, "http://") && !strings.HasPrefix(cfg.Host, "https://") {
		return nil, fmt.Errorf("host %q must start with http:// or https://", cfg.Host)
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = DefaultAPIVersion
	}
	if cfg.RetryCount < 0 {
		return nil, fmt.Errorf("retry count must be zero or greater, got %d", cfg.RetryCount)
	}
	if cfg.RetryWait <= 0 {
		cfg.RetryWait = DefaultRetryWait
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: DefaultTimeout}
	}

	return &Client{
		cfg:        cfg,
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("%s/api/%s", cfg.Host, cfg.APIVersion),
	}, nil
}

// Host returns the normalized server root.
func (c *Client) Host() string { return c.cfg.Host }

// APIVersion returns the configured REST API version.
func (c *Client) APIVersion() string { return c.cfg.APIVersion }

// SDK returns an authenticated SDK client. The first call performs the login
// exchange when no token was configured; subsequent calls are free.
func (c *Client) SDK(ctx context.Context) (*ometa.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sdk != nil {
		return c.sdk, nil
	}

	token, err := c.tokenLocked(ctx)
	if err != nil {
		return nil, err
	}

	c.sdk = ometa.NewClient(
		c.cfg.Host,
		ometa.WithToken(token),
		ometa.WithAPIVersion(c.cfg.APIVersion),
		ometa.WithRetry(c.cfg.RetryCount, c.cfg.RetryWait),
		ometa.WithHTTPClient(c.httpClient),
	)

	return c.sdk, nil
}

// Delete sends DELETE /api/{version}/{path} with optional query parameters.
// The SDK omits Delete for a few services (domains and data products), so
// those resources use this narrow helper instead of a no-op destroy.
func (c *Client) Delete(ctx context.Context, path string, params url.Values) error {
	token, err := c.authToken(ctx)
	if err != nil {
		return err
	}

	endpoint := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("openmetadata: failed to build delete request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openmetadata: delete request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := readLimited(res)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return newAPIError(res.StatusCode, raw)
	}

	return nil
}

// authToken returns the configured token or the cached login token.
func (c *Client) authToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tokenLocked(ctx)
}

// tokenLocked implements authToken for callers already holding the mutex.
func (c *Client) tokenLocked(ctx context.Context) (string, error) {
	if c.token != "" {
		return c.token, nil
	}

	token := c.cfg.Token
	if token == "" {
		var err error
		token, err = c.login(ctx)
		if err != nil {
			return "", err
		}
	}

	c.token = token
	return token, nil
}

// Version probes GET /api/v{version}/system/version. Acceptance prechecks use
// it to fail fast on an unreachable or mis-versioned server.
func (c *Client) Version(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/system/version", nil)
	if err != nil {
		return "", fmt.Errorf("openmetadata: failed to build version request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openmetadata: version request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := readLimited(res)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return "", newAPIError(res.StatusCode, raw)
	}

	var payload struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", fmt.Errorf("openmetadata: failed to decode version response: %w", err)
	}

	return payload.Version, nil
}

// readLimited reads a bounded response body.
func readLimited(res *http.Response) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("openmetadata: failed to read response body: %w", err)
	}
	return raw, nil
}
