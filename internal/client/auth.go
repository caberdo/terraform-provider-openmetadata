package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

// login exchanges username and password for a JWT.
//
// The SDK deliberately exposes bearer-token auth only, so this is the one place
// the provider speaks to the API directly. OpenMetadata returns the token as a
// JSON string, as a plain-text body, or occasionally as an object; all three
// shapes are accepted.
func (c *Client) login(ctx context.Context) (string, error) {
	if c.cfg.Username == "" || c.cfg.Password == "" {
		return "", fmt.Errorf("openmetadata: authentication requires either token or username and password")
	}

	payload, err := json.Marshal(ometa.LoginRequest{
		Email: c.cfg.Username,
		// The login endpoint rejects a plaintext password with HTTP 400
		// ("Password needs to be encoded in Base-64.") and decodes it before
		// authentication.
		Password: base64.StdEncoding.EncodeToString([]byte(c.cfg.Password)),
	})
	if err != nil {
		return "", fmt.Errorf("openmetadata: failed to encode login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/users/login", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("openmetadata: failed to build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openmetadata: login request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := readLimited(res)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= http.StatusBadRequest {
		return "", newAPIError(res.StatusCode, raw)
	}

	if token := tokenFromBody(raw); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("openmetadata: login succeeded but returned an empty token")
}

// tokenFromBody extracts a JWT from the known response shapes.
func tokenFromBody(raw []byte) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return ""
	}

	var asString string
	if err := json.Unmarshal([]byte(trimmed), &asString); err == nil && asString != "" {
		return asString
	}

	var asObject map[string]any
	if err := json.Unmarshal([]byte(trimmed), &asObject); err == nil {
		for _, key := range []string{"accessToken", "token", "jwt", "id_token"} {
			if value, ok := asObject[key].(string); ok && value != "" {
				return value
			}
		}
	}

	// A plain-text body is a valid token. Reject anything with whitespace
	// because that is an HTML error page rather than a JWT.
	if !strings.ContainsAny(trimmed, " \t\r\n<") {
		return trimmed
	}

	return ""
}
