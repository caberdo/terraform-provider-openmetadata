package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

// newAPIError normalizes a non-2xx response into the SDK's APIError so callers
// only have one error shape to reason about.
func newAPIError(status int, raw []byte) error {
	apiErr := &ometa.APIError{StatusCode: status}

	var decoded ometa.APIError
	if err := json.Unmarshal(raw, &decoded); err == nil && strings.TrimSpace(decoded.Message) != "" {
		apiErr.Code = decoded.Code
		apiErr.Message = decoded.Message
	} else {
		apiErr.Message = strings.TrimSpace(string(raw))
	}

	return apiErr
}

// ErrorDetail renders an error for a Terraform diagnostic detail field.
func ErrorDetail(err error) string {
	if err == nil {
		return ""
	}

	var apiErr *ometa.APIError
	if errors.As(err, &apiErr) {
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			return fmt.Sprintf("OpenMetadata returned HTTP %d.", apiErr.StatusCode)
		}
		return fmt.Sprintf("OpenMetadata returned HTTP %d: %s", apiErr.StatusCode, message)
	}

	return err.Error()
}

// ErrorDiagnostic builds a framework diagnostic carrying the API detail.
func ErrorDiagnostic(summary string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(summary, ErrorDetail(err))
}

// StatusCode reports the HTTP status carried by err, if any.
func StatusCode(err error) (int, bool) {
	if err == nil {
		return 0, false
	}
	var apiErr *ometa.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode, true
	}
	return 0, false
}

// NotFound reports whether the server answered 404.
func NotFound(err error) bool {
	status, ok := StatusCode(err)
	return ok && status == http.StatusNotFound
}

// Conflict reports whether the server answered 409.
func Conflict(err error) bool {
	status, ok := StatusCode(err)
	return ok && status == http.StatusConflict
}

// BadRequest reports whether the server answered 400. Identifier lookups use it
// to fall back from UUID to fully qualified name.
func BadRequest(err error) bool {
	status, ok := StatusCode(err)
	return ok && status == http.StatusBadRequest
}

// Unauthorized reports whether the server answered 401 or 403. The provider
// turns this into a hint about token expiry or missing permissions.
func Unauthorized(err error) bool {
	status, ok := StatusCode(err)
	return ok && (status == http.StatusUnauthorized || status == http.StatusForbidden)
}
