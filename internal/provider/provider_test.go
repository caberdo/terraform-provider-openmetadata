package provider

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/caberdo/terraform-provider-openmetadata/internal/client"
)

func TestNewSetsMetadata(t *testing.T) {
	p := New("1.2.3")()

	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "openmetadata" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata")
	}
	if resp.Version != "1.2.3" {
		t.Fatalf("Version = %q, want %q", resp.Version, "1.2.3")
	}
}

func TestProviderRegistersAllTypes(t *testing.T) {
	ctx := context.Background()
	p := New("dev")()

	want := map[string]bool{
		"openmetadata_ai_application":       false,
		"openmetadata_ai_governance_policy": false,
		"openmetadata_api_collection":       false,
		"openmetadata_api_endpoint":         false,
		"openmetadata_api_service":          false,
		"openmetadata_bot":                  false,
		"openmetadata_chart":                false,
		"openmetadata_classification":       false,
		"openmetadata_container":            false,
		"openmetadata_dashboard":            false,
		"openmetadata_dashboard_data_model": false,
		"openmetadata_dashboard_service":    false,
		"openmetadata_data_product":         false,
		"openmetadata_database":             false,
		"openmetadata_database_schema":      false,
		"openmetadata_database_service":     false,
		"openmetadata_directory":            false,
		"openmetadata_document":             false,
		"openmetadata_domain":               false,
		"openmetadata_drive_service":        false,
		"openmetadata_file":                 false,
		"openmetadata_glossary":             false,
		"openmetadata_glossary_term":        false,
		"openmetadata_learning_resource":    false,
		"openmetadata_llm_model":            false,
		"openmetadata_llm_service":          false,
		"openmetadata_messaging_service":    false,
		"openmetadata_metadata_service":     false,
		"openmetadata_metric":               false,
		"openmetadata_ml_model":             false,
		"openmetadata_ml_model_service":     false,
		"openmetadata_persona":              false,
		"openmetadata_pipeline":             false,
		"openmetadata_pipeline_service":     false,
		"openmetadata_policy":               false,
		"openmetadata_prompt_template":      false,
		"openmetadata_query":                false,
		"openmetadata_role":                 false,
		"openmetadata_search_index":         false,
		"openmetadata_search_service":       false,
		"openmetadata_security_service":     false,
		"openmetadata_spreadsheet":          false,
		"openmetadata_storage_service":      false,
		"openmetadata_stored_procedure":     false,
		"openmetadata_table":                false,
		"openmetadata_tag":                  false,
		"openmetadata_team":                 false,
		"openmetadata_topic":                false,
		"openmetadata_user":                 false,
		"openmetadata_worksheet":            false,
	}

	for _, name := range []string{
		"openmetadata_data_contract",
		"openmetadata_event_subscription",
		"openmetadata_ingestion_pipeline",
		"openmetadata_notification_template",
		"openmetadata_test_case",
		"openmetadata_test_definition",
		"openmetadata_test_suite",
		"openmetadata_workflow_definition",
	} {
		want[name] = false
	}

	resources := p.Resources(ctx)
	if len(resources) != len(want) {
		t.Fatalf("Resources() returned %d factories, want %d", len(resources), len(want))
	}
	for _, factory := range resources {
		var resp resource.MetadataResponse
		factory().Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)
		if _, ok := want[resp.TypeName]; !ok {
			t.Fatalf("unexpected resource %q", resp.TypeName)
		}
		want[resp.TypeName] = true
	}

	dataSources := p.DataSources(ctx)
	if len(dataSources) != len(want) {
		t.Fatalf("DataSources() returned %d factories, want %d", len(dataSources), len(want))
	}
	for _, factory := range dataSources {
		var resp datasource.MetadataResponse
		factory().Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)
		if _, ok := want[resp.TypeName]; !ok {
			t.Fatalf("unexpected data source %q", resp.TypeName)
		}
		if !want[resp.TypeName] {
			t.Fatalf("data source %q was registered without a resource of the same name", resp.TypeName)
		}
	}
}

func TestClientConfig(t *testing.T) {
	tests := []struct {
		name       string
		model      providerModel
		env        map[string]string
		wantErr    string
		wantConfig client.Config
	}{
		{
			name: "token authentication",
			model: providerModel{
				Host:  types.StringValue("https://om.example.com"),
				Token: types.StringValue("jwt"),
			},
			wantConfig: client.Config{
				Host:       "https://om.example.com",
				Token:      "jwt",
				APIVersion: client.DefaultAPIVersion,
				RetryCount: client.DefaultRetryCount,
				RetryWait:  client.DefaultRetryWait,
			},
		},
		{
			name: "login authentication",
			model: providerModel{
				Host:     types.StringValue("https://om.example.com/"),
				Username: types.StringValue("admin@open-metadata.org"),
				Password: types.StringValue("admin"),
			},
			wantConfig: client.Config{
				Host:       "https://om.example.com/",
				Username:   "admin@open-metadata.org",
				Password:   "admin",
				APIVersion: client.DefaultAPIVersion,
				RetryCount: client.DefaultRetryCount,
				RetryWait:  client.DefaultRetryWait,
			},
		},
		{
			name: "environment fallback",
			env: map[string]string{
				"OPENMETADATA_HOST":        "http://localhost:8585",
				"OPENMETADATA_TOKEN":       "env-token",
				"OPENMETADATA_RETRY_COUNT": "1",
				"OPENMETADATA_RETRY_WAIT":  "5s",
				"OPENMETADATA_API_VERSION": "v1",
			},
			wantConfig: client.Config{
				Host:       "http://localhost:8585",
				Token:      "env-token",
				APIVersion: "v1",
				RetryCount: 1,
				RetryWait:  5 * time.Second,
			},
		},
		{
			name:    "missing host",
			model:   providerModel{Token: types.StringValue("jwt")},
			wantErr: "Missing OpenMetadata host",
		},
		{
			name: "missing credentials",
			model: providerModel{
				Host:     types.StringValue("http://localhost:8585"),
				Username: types.StringValue("admin@open-metadata.org"),
			},
			wantErr: "Missing OpenMetadata credentials",
		},
		{
			name: "invalid retry count from environment",
			model: providerModel{
				Host:  types.StringValue("http://localhost:8585"),
				Token: types.StringValue("jwt"),
			},
			env:     map[string]string{"OPENMETADATA_RETRY_COUNT": "many"},
			wantErr: "Invalid retry_count",
		},
		{
			name: "invalid retry wait",
			model: providerModel{
				Host:      types.StringValue("http://localhost:8585"),
				Token:     types.StringValue("jwt"),
				RetryWait: types.StringValue("soon"),
			},
			wantErr: "Invalid retry_wait",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{
				"OPENMETADATA_HOST",
				"OPENMETADATA_TOKEN",
				"OPENMETADATA_USERNAME",
				"OPENMETADATA_PASSWORD",
				"OPENMETADATA_API_VERSION",
				"OPENMETADATA_RETRY_COUNT",
				"OPENMETADATA_RETRY_WAIT",
			} {
				t.Setenv(key, "")
			}
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			cfg, diags := clientConfig(tt.model)

			if tt.wantErr != "" {
				if !diags.HasError() {
					t.Fatalf("expected an error containing %q, got none", tt.wantErr)
				}
				if detail := diags.Errors()[0].Summary(); !strings.Contains(detail, tt.wantErr) {
					t.Fatalf("error summary = %q, want it to contain %q", detail, tt.wantErr)
				}
				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if cfg != tt.wantConfig {
				t.Fatalf("config = %#v, want %#v", cfg, tt.wantConfig)
			}
		})
	}
}
