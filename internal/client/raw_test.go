package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoReturnsRawBody(t *testing.T) {
	var method, path, authorization, contentType string
	var payload map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		authorization = r.Header.Get("Authorization")
		contentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"pipelineStatuses":[]}`))
	}))
	defer srv.Close()

	c, err := New(Config{
		Host:       srv.URL,
		Token:      "jwt-token",
		RetryCount: 0,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	raw, err := c.Do(context.Background(), http.MethodPatch, "/services/ingestionPipelines/abc", map[string]string{"name": "demo"})
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}

	if method != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", method)
	}
	if path != "/api/v1/services/ingestionPipelines/abc" {
		t.Errorf("path = %q", path)
	}
	if authorization != "Bearer jwt-token" {
		t.Errorf("Authorization = %q", authorization)
	}
	if contentType != "application/json-patch+json" {
		t.Errorf("Content-Type = %q", contentType)
	}
	if payload["name"] != "demo" {
		t.Errorf("payload = %#v", payload)
	}
	if string(raw) != `{"pipelineStatuses":[]}` {
		t.Errorf("raw = %s", raw)
	}
}

func TestDoReturnsAPIErrorAndRejectsUnencodableBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"missing"}`))
	}))
	defer srv.Close()

	c, err := New(Config{
		Host:       srv.URL,
		Token:      "jwt-token",
		RetryCount: 0,
		HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.Do(context.Background(), http.MethodGet, "/missing", nil)
	if !NotFound(err) {
		t.Fatalf("Do() error = %v, want a 404 API error", err)
	}

	if _, err := c.Do(context.Background(), http.MethodPost, "/widgets", func() {}); err == nil {
		t.Fatal("Do() accepted an unencodable body")
	}
}
