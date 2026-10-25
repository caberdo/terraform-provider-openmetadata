// Package acctest wires the provider into terraform-plugin-testing and holds
// the acceptance-test preconditions. Keeping it out of the entity packages lets
// every entity test import one small, stable harness.
package acctest

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
	"github.com/caberdo/terraform-provider-openmetadata/internal/provider"
)

// ProviderTypeName is the prefix Terraform uses for every resource and data
// source in this provider.
const ProviderTypeName = "openmetadata"

// NewClient builds a client from the acceptance environment. CheckDestroy
// functions use it to prove an entity is really gone.
func NewClient() (*client.Client, error) {
	return client.New(client.Config{
		Host:       strings.TrimSpace(os.Getenv("OPENMETADATA_HOST")),
		Token:      os.Getenv("OPENMETADATA_TOKEN"),
		Username:   os.Getenv("OPENMETADATA_USERNAME"),
		Password:   os.Getenv("OPENMETADATA_PASSWORD"),
		APIVersion: os.Getenv("OPENMETADATA_API_VERSION"),
	})
}

// ProtoV6ProviderFactories wires the provider into terraform-plugin-testing.
func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		ProviderTypeName: providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

// PreCheck skips the test unless acceptance testing is enabled, verifies that
// credentials are configured, and fails fast when the server is unreachable.
// When OPENMETADATA_EXPECT_VERSION is set it also asserts the server version,
// so a stale image fails here instead of in a confusing capability failure.
func PreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}

	host := strings.TrimSpace(os.Getenv("OPENMETADATA_HOST"))
	if host == "" {
		t.Fatal("OPENMETADATA_HOST must be set for acceptance tests")
	}

	token := os.Getenv("OPENMETADATA_TOKEN")
	username := os.Getenv("OPENMETADATA_USERNAME")
	password := os.Getenv("OPENMETADATA_PASSWORD")
	if token == "" && (username == "" || password == "") {
		t.Fatal("set OPENMETADATA_TOKEN, or OPENMETADATA_USERNAME and OPENMETADATA_PASSWORD, for acceptance tests")
	}

	c, err := NewClient()
	if err != nil {
		t.Fatalf("invalid acceptance configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	version, err := c.Version(ctx)
	if err != nil {
		t.Fatalf("OpenMetadata at %s is not reachable: %v", host, err)
	}

	if expected := strings.TrimSpace(os.Getenv("OPENMETADATA_EXPECT_VERSION")); expected != "" && expected != version {
		t.Fatalf("OPENMETADATA_EXPECT_VERSION=%s but the server reports %q", expected, version)
	}

	t.Logf("acceptance target: OpenMetadata %s at %s", version, host)
}
