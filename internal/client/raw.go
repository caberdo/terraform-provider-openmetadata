package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

// Do sends a JSON request and returns the raw response body. It exists for the
// few endpoints whose generated SDK model cannot decode a newer server's
// response shape: OpenMetadata 2.0 returns ingestion-pipeline pipelineStatuses
// as an array, while the pinned 1.13 SDK models it as an object. Callers
// normalize the document before decoding it into the SDK type.
func (c *Client) Do(ctx context.Context, method, path string, body any) ([]byte, error) {
	token, err := c.authToken(ctx)
	if err != nil {
		return nil, err
	}

	var payload []byte
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("openmetadata: failed to encode request body: %w", err)
		}
	}

	backend := ometa.NewHTTPBackend(&ometa.ClientConfig{
		BaseURL:     c.cfg.Host,
		APIVersion:  c.cfg.APIVersion,
		AccessToken: token,
		Retry:       c.cfg.RetryCount,
		RetryWait:   c.cfg.RetryWait,
		RetryCodes:  []int{http.StatusTooManyRequests, http.StatusGatewayTimeout},
		HTTPClient:  c.httpClient,
	})

	return backend.Call(ctx, method, strings.TrimLeft(path, "/"), payload, nil)
}
