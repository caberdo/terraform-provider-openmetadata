package role

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

	if resp.TypeName != "openmetadata_role" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_role")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, selector := range []string{"id", "fully_qualified_name"} {
		attribute, ok := resp.Schema.Attributes[selector].(dschema.StringAttribute)
		if !ok || !attribute.Optional || !attribute.Computed {
			t.Fatalf("%s must be optional and computed", selector)
		}
		if len(attribute.Validators) == 0 {
			t.Fatalf("%s must validate that exactly one selector is set", selector)
		}
	}

	policies, ok := resp.Schema.Attributes["policies"].(dschema.SetAttribute)
	if !ok || !policies.Computed {
		t.Fatal("policies must be a computed set")
	}

	for _, attribute := range []string{"name", "display_name", "description", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(dschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed float")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "DataSteward"
	displayName := "Data Steward"
	policyName := "DataStewardPolicy"
	policyFQN := "DataStewardPolicy"
	version := 0.1
	description := "Can curate data assets."

	var model roleModel
	apply(&ometa.Role{
		Id:                 id,
		Name:               "DataSteward",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Policies: &[]ometa.EntityReference{{
			Id:                 uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"),
			Type:               "policy",
			Name:               &policyName,
			FullyQualifiedName: &policyFQN,
		}},
		Version: &version,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "DataSteward" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Can curate data assets." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}

	var policies []string
	if diags := model.Policies.ElementsAs(context.Background(), &policies, false); diags.HasError() {
		t.Fatalf("Policies elements: %v", diags.Errors())
	}
	if len(policies) != 1 || policies[0] != policyFQN {
		t.Fatalf("Policies = %#v, want [%q]", policies, policyFQN)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model roleModel
	apply(&ometa.Role{Id: id, Name: "bare"}, &model)

	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.Policies.IsNull() {
		t.Fatalf("Policies = %v, want null", model.Policies)
	}
	if !model.Version.IsNull() {
		t.Fatalf("Version = %v, want null", model.Version)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPolicySetValueFallsBackToName(t *testing.T) {
	name := "FallbackPolicy"

	value := policySetValue([]ometa.EntityReference{{Name: &name}})

	var policies []string
	if diags := value.ElementsAs(context.Background(), &policies, false); diags.HasError() {
		t.Fatalf("Policies elements: %v", diags.Errors())
	}
	if len(policies) != 1 || policies[0] != name {
		t.Fatalf("Policies = %#v, want [%q]", policies, name)
	}
}
