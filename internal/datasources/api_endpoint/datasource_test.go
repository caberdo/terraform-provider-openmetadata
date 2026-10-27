package api_endpoint

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_api_endpoint" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_api_endpoint")
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
		"api_collection",
		"service",
		"service_type",
		"endpoint_url",
		"request_method",
		"display_name",
		"description",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	collectionFQN := "payments-api.pets"
	serviceFQN := "payments-api"
	serviceType := ometa.APIEndpointServiceTypeRest
	endpointURL := "https://api.example.com/v1/pets"
	requestMethod := ometa.APIEndpointRequestMethodPOST
	fqn := "payments-api.pets.list"
	displayName := "List pets"
	description := "Lists every pet."

	var model apiEndpointModel
	apply(&ometa.APIEndpoint{
		Id:                 id,
		Name:               "list",
		ApiCollection:      &ometa.EntityReference{FullyQualifiedName: &collectionFQN},
		Service:            ometa.EntityReference{FullyQualifiedName: &serviceFQN},
		ServiceType:        &serviceType,
		EndpointURL:        &endpointURL,
		RequestMethod:      &requestMethod,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "list" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.APICollection.ValueString() != collectionFQN {
		t.Fatalf("APICollection = %q", model.APICollection.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
	}
	if model.ServiceType.ValueString() != "Rest" {
		t.Fatalf("ServiceType = %q", model.ServiceType.ValueString())
	}
	if model.EndpointURL.ValueString() != endpointURL {
		t.Fatalf("EndpointURL = %q", model.EndpointURL.ValueString())
	}
	if model.RequestMethod.ValueString() != "POST" {
		t.Fatalf("RequestMethod = %q", model.RequestMethod.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
}

func TestApplyMissingReferencesKeepPriorValues(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := apiEndpointModel{
		APICollection: types.StringValue("payments-api.pets"),
		Service:       types.StringValue("payments-api"),
		ServiceType:   types.StringValue("Rest"),
		RequestMethod: types.StringValue("POST"),
	}
	apply(&ometa.APIEndpoint{Id: id, Name: "bare"}, &model)

	if model.APICollection.ValueString() != "payments-api.pets" {
		t.Fatalf("APICollection = %q, want the prior value preserved", model.APICollection.ValueString())
	}
	if model.Service.ValueString() != "payments-api" {
		t.Fatalf("Service = %q, want the prior value preserved", model.Service.ValueString())
	}
	if model.ServiceType.ValueString() != "Rest" {
		t.Fatalf("ServiceType = %q, want the prior value preserved", model.ServiceType.ValueString())
	}
	if model.RequestMethod.ValueString() != "POST" {
		t.Fatalf("RequestMethod = %q, want the prior value preserved", model.RequestMethod.ValueString())
	}
}
