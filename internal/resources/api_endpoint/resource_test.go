package api_endpoint

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_api_endpoint" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_api_endpoint")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "api_collection"} {
		attr, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !attr.Required {
			t.Fatalf("%s must be required", attribute)
		}
		if len(attr.PlanModifiers) == 0 {
			t.Fatalf("%s must carry a plan modifier", attribute)
		}
	}

	for _, attribute := range []string{"service", "service_type"} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed || computed.Optional || computed.Required {
			t.Fatalf("%s must be computed only", attribute)
		}
	}

	endpointURL, ok := resp.Schema.Attributes["endpoint_url"].(rschema.StringAttribute)
	if !ok || !endpointURL.Optional || !endpointURL.Computed || endpointURL.Default == nil {
		t.Fatal("endpoint_url must be optional and computed with an empty default")
	}

	requestMethod, ok := resp.Schema.Attributes["request_method"].(rschema.StringAttribute)
	if !ok || !requestMethod.Optional || !requestMethod.Computed || requestMethod.Default == nil {
		t.Fatal("request_method must be optional and computed with a default")
	}
	if len(requestMethod.Validators) == 0 {
		t.Fatal("request_method must validate the HTTP method")
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed || description.Default == nil {
		t.Fatal("description must be optional and computed with an empty default")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Required || displayName.Computed {
		t.Fatal("display_name must be optional only")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	collectionName := "pets"
	collectionFQN := "payments-api.pets"
	serviceName := "payments-api"
	serviceFQN := "payments-api"
	serviceType := ometa.APIEndpointServiceTypeRest
	endpointURL := "https://api.example.com/v1/pets"
	requestMethod := ometa.APIEndpointRequestMethodPOST
	fqn := "payments-api.pets.list"
	displayName := "List pets"
	description := "Lists every pet."
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model apiEndpointModel
	apply(&ometa.APIEndpoint{
		Id:                 id,
		Name:               "list",
		ApiCollection:      &ometa.EntityReference{Name: &collectionName, FullyQualifiedName: &collectionFQN},
		Service:            ometa.EntityReference{Name: &serviceName, FullyQualifiedName: &serviceFQN},
		ServiceType:        &serviceType,
		EndpointURL:        &endpointURL,
		RequestMethod:      &requestMethod,
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
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

func TestApplyKeepsPriorReferencesWhenServerOmitsThem(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "payments-api"

	model := apiEndpointModel{
		APICollection: types.StringValue("payments-api.pets"),
		Service:       types.StringValue("payments-api"),
		ServiceType:   types.StringValue("Rest"),
		RequestMethod: types.StringValue("POST"),
	}
	apply(&ometa.APIEndpoint{
		Id:      id,
		Name:    "bare",
		Service: ometa.EntityReference{Name: &serviceName},
	}, &model)

	if model.APICollection.ValueString() != "payments-api.pets" {
		t.Fatalf("APICollection = %q, want the prior value preserved", model.APICollection.ValueString())
	}
	if model.Service.ValueString() != "payments-api" {
		t.Fatalf("Service = %q, want the name fallback", model.Service.ValueString())
	}
	if model.ServiceType.ValueString() != "Rest" {
		t.Fatalf("ServiceType = %q, want the prior value preserved", model.ServiceType.ValueString())
	}
	if model.RequestMethod.ValueString() != "POST" {
		t.Fatalf("RequestMethod = %q, want the prior value preserved", model.RequestMethod.ValueString())
	}
	if model.EndpointURL.ValueString() != "" {
		t.Fatalf("EndpointURL = %q, want the empty default", model.EndpointURL.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := apiEndpointModel{
		Name:          types.StringValue("list"),
		APICollection: types.StringValue("payments-api.pets"),
		Service:       types.StringValue("payments-api"),
		ServiceType:   types.StringValue("Rest"),
		EndpointURL:   types.StringValue("https://api.example.com/v1/pets"),
		RequestMethod: types.StringValue("GET"),
		DisplayName:   types.StringValue("Display"),
		Description:   types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      apiEndpointModel
		state     apiEndpointModel
		wantOps   int
		wantPath  string
		wantOp    string
		wantValue any
	}{
		{
			name:    "no changes",
			plan:    base,
			state:   base,
			wantOps: 0,
		},
		{
			name: "display name replaced",
			plan: apiEndpointModel{
				Name:          base.Name,
				APICollection: base.APICollection,
				Service:       base.Service,
				ServiceType:   base.ServiceType,
				EndpointURL:   base.EndpointURL,
				RequestMethod: base.RequestMethod,
				DisplayName:   types.StringValue("New display"),
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: apiEndpointModel{
				Name:          base.Name,
				APICollection: base.APICollection,
				Service:       base.Service,
				ServiceType:   base.ServiceType,
				EndpointURL:   base.EndpointURL,
				RequestMethod: base.RequestMethod,
				DisplayName:   types.StringNull(),
				Description:   base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: apiEndpointModel{
				Name:          base.Name,
				APICollection: base.APICollection,
				Service:       base.Service,
				ServiceType:   base.ServiceType,
				EndpointURL:   base.EndpointURL,
				RequestMethod: base.RequestMethod,
				DisplayName:   base.DisplayName,
				Description:   types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "endpoint url changed",
			plan: apiEndpointModel{
				Name:          base.Name,
				APICollection: base.APICollection,
				Service:       base.Service,
				ServiceType:   base.ServiceType,
				EndpointURL:   types.StringValue("https://api.example.com/v2/pets"),
				RequestMethod: base.RequestMethod,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/endpointURL",
			wantOp:    "add",
			wantValue: "https://api.example.com/v2/pets",
		},
		{
			name: "endpoint url removed",
			plan: apiEndpointModel{
				Name:          base.Name,
				APICollection: base.APICollection,
				Service:       base.Service,
				ServiceType:   base.ServiceType,
				EndpointURL:   types.StringValue(""),
				RequestMethod: base.RequestMethod,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/endpointURL",
			wantOp:   "remove",
		},
		{
			name: "request method changed",
			plan: apiEndpointModel{
				Name:          base.Name,
				APICollection: base.APICollection,
				Service:       base.Service,
				ServiceType:   base.ServiceType,
				EndpointURL:   base.EndpointURL,
				RequestMethod: types.StringValue("POST"),
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/requestMethod",
			wantOp:    "add",
			wantValue: "POST",
		},
		{
			name: "name and collection cannot be patched",
			plan: apiEndpointModel{
				Name:          types.StringValue("renamed"),
				APICollection: types.StringValue("other-api.other"),
				Service:       types.StringValue("other-api"),
				ServiceType:   types.StringValue("WEBHOOK"),
				EndpointURL:   base.EndpointURL,
				RequestMethod: base.RequestMethod,
				DisplayName:   base.DisplayName,
				Description:   base.Description,
			},
			state:   base,
			wantOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
			if len(ops) != tt.wantOps {
				t.Fatalf("len(ops) = %d, want %d: %#v", len(ops), tt.wantOps, ops)
			}
			if tt.wantOps == 0 {
				return
			}
			if ops[0].Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", ops[0].Path, tt.wantPath)
			}
			if ops[0].Op != tt.wantOp {
				t.Fatalf("op = %q, want %q", ops[0].Op, tt.wantOp)
			}
			if tt.wantValue != nil && ops[0].Value != tt.wantValue {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
