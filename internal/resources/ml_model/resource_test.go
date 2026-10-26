package ml_model

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

	if resp.TypeName != "openmetadata_ml_model" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_ml_model")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "service"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !required.Required {
			t.Fatalf("%s must be required", attribute)
		}
		if len(required.PlanModifiers) == 0 {
			t.Fatalf("%s must carry a plan modifier", attribute)
		}
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"version",
		"updated_at",
		"updated_by",
	} {
		if attribute == "version" {
			floatAttribute, ok := resp.Schema.Attributes[attribute].(rschema.Float64Attribute)
			if !ok || !floatAttribute.Computed {
				t.Fatalf("%s must be a computed float", attribute)
			}
			continue
		}
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	algorithm, ok := resp.Schema.Attributes["algorithm"].(rschema.StringAttribute)
	if !ok || !algorithm.Optional || !algorithm.Computed || algorithm.Required {
		t.Fatal("algorithm must be optional and computed")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed {
		t.Fatal("description must be optional and computed")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional || displayName.Required || displayName.Computed {
		t.Fatal("display_name must be optional only")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "mlflow"
	serviceFQN := "mlflow"
	fqn := "mlflow.ChurnModel"
	displayName := "Churn model"
	description := "Predicts whether a customer will churn."
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"

	var model mlModelModel
	apply(&ometa.MlModel{
		Id:                 id,
		Name:               "ChurnModel",
		Algorithm:          "linear_regression",
		Service:            ometa.EntityReference{Name: &serviceName, FullyQualifiedName: &serviceFQN},
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
	if model.Name.ValueString() != "ChurnModel" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Algorithm.ValueString() != "linear_regression" {
		t.Fatalf("Algorithm = %q", model.Algorithm.ValueString())
	}
	if model.Service.ValueString() != serviceFQN {
		t.Fatalf("Service = %q", model.Service.ValueString())
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

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := mlModelModel{
		Algorithm: types.StringValue("linear_regression"),
		Service:   types.StringValue("mlflow.ExistingModel"),
	}
	apply(&ometa.MlModel{Id: id, Name: "bare"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if model.Algorithm.ValueString() != "linear_regression" {
		t.Fatalf("Algorithm = %q, want the prior value preserved", model.Algorithm.ValueString())
	}
	if model.Service.ValueString() != "mlflow.ExistingModel" {
		t.Fatalf("Service = %q, want the prior value preserved", model.Service.ValueString())
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestApplyFallsBackToServiceName(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	serviceName := "mlflow"

	var model mlModelModel
	apply(&ometa.MlModel{
		Id:      id,
		Name:    "ChurnModel",
		Service: ometa.EntityReference{Name: &serviceName},
	}, &model)

	if model.Service.ValueString() != serviceName {
		t.Fatalf("Service = %q, want %q", model.Service.ValueString(), serviceName)
	}
}

func TestPatchOps(t *testing.T) {
	base := mlModelModel{
		Name:        types.StringValue("ChurnModel"),
		Service:     types.StringValue("mlflow"),
		Algorithm:   types.StringValue("linear_regression"),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      mlModelModel
		state     mlModelModel
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
			plan: mlModelModel{
				Name:        base.Name,
				Service:     base.Service,
				Algorithm:   base.Algorithm,
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "display name removed",
			plan: mlModelModel{
				Name:        base.Name,
				Service:     base.Service,
				Algorithm:   base.Algorithm,
				DisplayName: types.StringNull(),
				Description: base.Description,
			},
			state:    base,
			wantOps:  1,
			wantPath: "/displayName",
			wantOp:   "remove",
		},
		{
			name: "description changed",
			plan: mlModelModel{
				Name:        base.Name,
				Service:     base.Service,
				Algorithm:   base.Algorithm,
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "algorithm changed",
			plan: mlModelModel{
				Name:        base.Name,
				Service:     base.Service,
				Algorithm:   types.StringValue("logistic_regression"),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/algorithm",
			wantOp:    "add",
			wantValue: "logistic_regression",
		},
		{
			name: "algorithm null is not patched",
			plan: mlModelModel{
				Name:        base.Name,
				Service:     base.Service,
				Algorithm:   types.StringNull(),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "algorithm unknown is not patched",
			plan: mlModelModel{
				Name:        base.Name,
				Service:     base.Service,
				Algorithm:   types.StringUnknown(),
				DisplayName: base.DisplayName,
				Description: base.Description,
			},
			state:   base,
			wantOps: 0,
		},
		{
			name: "name and service cannot be patched",
			plan: mlModelModel{
				Name:        types.StringValue("RenamedModel"),
				Service:     types.StringValue("sklearn"),
				Algorithm:   base.Algorithm,
				DisplayName: base.DisplayName,
				Description: base.Description,
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
