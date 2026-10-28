package persona

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_persona" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_persona")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok {
			t.Fatalf("%s attribute is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
	}

	for _, attribute := range []string{"name", "display_name", "description", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !computed.Computed || computed.Optional {
			t.Fatalf("%s must be a computed-only string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(dschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatal("version must be a computed-only float")
	}

	defaultAttribute, ok := resp.Schema.Attributes["default"].(dschema.BoolAttribute)
	if !ok || !defaultAttribute.Computed {
		t.Fatal("default must be a computed-only bool")
	}

	users, ok := resp.Schema.Attributes["users"].(dschema.SetAttribute)
	if !ok || !users.Computed {
		t.Fatal("users must be a computed-only set")
	}
	if users.ElementType != types.StringType {
		t.Fatalf("users element type = %v, want string", users.ElementType)
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "DataSteward"
	displayName := "Data Steward"
	description := "Owns data quality."
	userFQN := "alice"
	isDefault := true
	version := 0.4

	var model personaModel
	diags := apply(context.Background(), &ometa.Persona{
		Id:                 id,
		Name:               "DataSteward",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		Default:            &isDefault,
		Users: &[]ometa.EntityReference{
			{Id: uuid.MustParse("2f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c"), Type: "user", FullyQualifiedName: &userFQN},
		},
		Version: &version,
	}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

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
	if !model.Default.ValueBool() {
		t.Fatal("Default = false")
	}
	users := model.Users.Elements()
	if len(users) != 1 || users[0].(types.String).ValueString() != userFQN {
		t.Fatalf("Users = %v", users)
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model personaModel
	diags := apply(context.Background(), &ometa.Persona{Id: id, Name: "Bare"}, &model)
	if diags.HasError() {
		t.Fatalf("apply() diagnostics = %v", diags.Errors())
	}

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want empty", model.Description.ValueString())
	}
	if model.Default.ValueBool() {
		t.Fatal("Default = true, want false")
	}
	if model.Users.IsNull() || len(model.Users.Elements()) != 0 {
		t.Fatalf("Users = %v, want an empty set", model.Users)
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
}
