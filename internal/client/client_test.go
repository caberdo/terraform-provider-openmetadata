package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoginResponseShapes(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "json string", body: `"jwt-token"`},
		{name: "plain text", body: "jwt-token"},
		{name: "accessToken object", body: `{"accessToken":"jwt-token"}`},
		{name: "token object", body: `{"token":"jwt-token"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/users/login" {
					t.Errorf("unexpected path %q", r.URL.Path)
					http.NotFound(w, r)
					return
				}
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}

				var payload struct {
					Email    string `json:"email"`
					Password string `json:"password"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Errorf("decode login body: %v", err)
				}
				if payload.Email != "admin@open-metadata.org" || payload.Password != "YWRtaW4=" {
					t.Errorf("login payload = %#v", payload)
				}

				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c, err := New(Config{
				Host:       srv.URL,
				Username:   "admin@open-metadata.org",
				Password:   "admin",
				RetryCount: 0,
				HTTPClient: srv.Client(),
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			token, err := c.login(context.Background())
			if err != nil {
				t.Fatalf("login() error = %v", err)
			}
			if token != "jwt-token" {
				t.Fatalf("token = %q, want %q", token, "jwt-token")
			}
		})
	}
}

func TestSDKUsesLoginTokenAndBearerHeader(t *testing.T) {
	var loginCalls atomic.Int32
	var bearerHeader string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/users/login":
			loginCalls.Add(1)
			_, _ = w.Write([]byte(`"jwt-token"`))
		case "/api/v1/glossaries/abc":
			bearerHeader = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"id":"1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c","name":"demo","description":""}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := New(Config{
		Host:       srv.URL,
		Username:   "admin@open-metadata.org",
		Password:   "admin",
		RetryCount: 0,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	sdk1, err := c.SDK(context.Background())
	if err != nil {
		t.Fatalf("SDK() error = %v", err)
	}
	sdk2, err := c.SDK(context.Background())
	if err != nil {
		t.Fatalf("second SDK() error = %v", err)
	}
	if sdk1 != sdk2 {
		t.Fatal("SDK() returned a different client on the second call")
	}

	if _, err := sdk1.Glossaries.GetByID(context.Background(), "abc", nil); err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got := loginCalls.Load(); got != 1 {
		t.Fatalf("login calls = %d, want 1", got)
	}
	if bearerHeader != "Bearer jwt-token" {
		t.Fatalf("Authorization = %q, want %q", bearerHeader, "Bearer jwt-token")
	}
}

func TestTokenSkipsLogin(t *testing.T) {
	var loginCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/users/login" {
			loginCalls.Add(1)
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c, err := New(Config{
		Host:       srv.URL,
		Token:      "static-token",
		Username:   "admin@open-metadata.org",
		Password:   "admin",
		RetryCount: 0,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := c.SDK(context.Background()); err != nil {
		t.Fatalf("SDK() error = %v", err)
	}
	if got := loginCalls.Load(); got != 0 {
		t.Fatalf("login calls = %d, want 0 when a token is configured", got)
	}
}

func TestLoginFailureIsClassified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"message":"Incorrect username or password"}`))
	}))
	defer srv.Close()

	c, err := New(Config{
		Host:       srv.URL,
		Username:   "admin@open-metadata.org",
		Password:   "wrong",
		RetryCount: 0,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.SDK(context.Background())
	if err == nil {
		t.Fatal("SDK() error = nil, want an authentication error")
	}
	if !Unauthorized(err) {
		t.Fatalf("Unauthorized() = false for %v", err)
	}
	detail := ErrorDetail(err)
	if !strings.Contains(detail, "401") || !strings.Contains(detail, "Incorrect username or password") {
		t.Fatalf("ErrorDetail() = %q", detail)
	}
}

func TestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/system/version" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"version":"1.13.0","revision":"abc"}`))
	}))
	defer srv.Close()

	c, err := New(Config{Host: srv.URL, Token: "t", HTTPClient: srv.Client(), RetryCount: 0})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	version, err := c.Version(context.Background())
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}
	if version != "1.13.0" {
		t.Fatalf("version = %q, want %q", version, "1.13.0")
	}
}

func TestNewValidation(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{name: "empty host", cfg: Config{}, wantErr: "host must be set"},
		{name: "missing scheme", cfg: Config{Host: "openmetadata.example.com"}, wantErr: "must start with http"},
		{name: "negative retries", cfg: Config{Host: "http://localhost:8585", RetryCount: -1}, wantErr: "retry count"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.cfg)
			if err == nil {
				t.Fatal("New() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("New() error = %q, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

func TestNewDefaultsAndNormalization(t *testing.T) {
	c, err := New(Config{Host: "https://om.example.com/", RetryCount: 0})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if c.Host() != "https://om.example.com" {
		t.Fatalf("Host() = %q, want the trailing slash removed", c.Host())
	}
	if c.APIVersion() != DefaultAPIVersion {
		t.Fatalf("APIVersion() = %q, want %q", c.APIVersion(), DefaultAPIVersion)
	}
	if c.cfg.RetryWait != DefaultRetryWait {
		t.Fatalf("RetryWait = %s, want %s", c.cfg.RetryWait, DefaultRetryWait)
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Fatalf("HTTP client timeout = %s, want %s", c.httpClient.Timeout, DefaultTimeout)
	}
}

func TestErrorClassification(t *testing.T) {
	if !NotFound(newAPIError(http.StatusNotFound, []byte(`{"message":"gone"}`))) {
		t.Fatal("NotFound() = false for a 404")
	}
	if !Conflict(newAPIError(http.StatusConflict, []byte(`{"message":"exists"}`))) {
		t.Fatal("Conflict() = false for a 409")
	}
	if !BadRequest(newAPIError(http.StatusBadRequest, []byte(`{"message":"bad"}`))) {
		t.Fatal("BadRequest() = false for a 400")
	}
	if NotFound(newAPIError(http.StatusInternalServerError, nil)) {
		t.Fatal("NotFound() = true for a 500")
	}
	if _, ok := StatusCode(context.DeadlineExceeded); ok {
		t.Fatal("StatusCode() reported a status for a non-API error")
	}
}

func TestLoginRequiresCredentials(t *testing.T) {
	c, err := New(Config{Host: "http://localhost:8585", RetryCount: 0, RetryWait: time.Millisecond})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.SDK(context.Background())
	if err == nil || !strings.Contains(err.Error(), "requires either token or username and password") {
		t.Fatalf("SDK() error = %v, want a credential error", err)
	}
}
