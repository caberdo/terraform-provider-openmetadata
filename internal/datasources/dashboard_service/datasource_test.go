package dashboard_service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_dashboard_service" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_dashboard_service")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"name",
		"display_name",
		"description",
		"service_type",
		"connection_config",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	serviceType, ok := resp.Schema.Attributes["service_type"].(dschema.StringAttribute)
	if !ok || !serviceType.Computed {
		t.Fatal("service_type must be a computed string")
	}

	connection, ok := resp.Schema.Attributes["connection_config"].(dschema.StringAttribute)
	if !ok || !connection.Computed {
		t.Fatal("connection must be a computed string")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "custom_dashboard"
	displayName := "Custom Dashboard"
	description := "Dashboards built in-house."
	version := 0.7
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model dashboardServiceModel
	apply(&ometa.DashboardService{
		Id:                 id,
		Name:               "custom_dashboard",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		ServiceType:        ometa.DashboardServiceServiceType("CustomDashboard"),
		Connection: &ometa.DashboardConnection{Config: map[string]any{
			"type": "CustomDashboard",
		}},
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "custom_dashboard" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.ServiceType.ValueString() != "CustomDashboard" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if model.Connection.ValueString() != `{"type":"CustomDashboard"}` {
		t.Fatalf("Connection = %q", model.Connection.ValueString())
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
	if model.UpdatedBy.ValueString() != updatedBy {
		t.Fatalf("UpdatedBy = %q", model.UpdatedBy.ValueString())
	}
}

func TestApplyNullsMissingConnection(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model dashboardServiceModel
	apply(&ometa.DashboardService{
		Id:          id,
		Name:        "bare",
		ServiceType: ometa.DashboardServiceServiceType("CustomDashboard"),
		Connection:  &ometa.DashboardConnection{},
	}, &model)

	if !model.Connection.IsNull() {
		t.Fatalf("Connection = %v, want null", model.Connection)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
}
