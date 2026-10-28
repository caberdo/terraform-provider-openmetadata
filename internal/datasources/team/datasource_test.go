package team

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp datasource.MetadataResponse
	NewDataSource().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_team" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_team")
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
			t.Fatalf("%s is %T, want StringAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be optional and computed", attribute)
		}
		if len(selector.Validators) == 0 {
			t.Fatalf("%s must validate that exactly one selector is set", attribute)
		}
	}

	for _, attribute := range []string{"name", "description", "team_type", "email", "default_persona", "version", "updated_at", "updated_by"} {
		computed, ok := resp.Schema.Attributes[attribute]
		if !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
		switch value := computed.(type) {
		case dschema.StringAttribute:
			if !value.Computed {
				t.Fatalf("%s must be computed", attribute)
			}
		case dschema.Float64Attribute:
			if !value.Computed {
				t.Fatalf("%s must be computed", attribute)
			}
		default:
			t.Fatalf("%s is %T, want a computed attribute", attribute, computed)
		}
	}

	for _, attribute := range []string{"parents", "policies", "default_roles", "users"} {
		set, ok := resp.Schema.Attributes[attribute].(dschema.SetAttribute)
		if !ok {
			t.Fatalf("%s is %T, want SetAttribute", attribute, resp.Schema.Attributes[attribute])
		}
		if !set.Computed {
			t.Fatalf("%s must be computed", attribute)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	parentFQN := "department"
	teamType := ometa.TeamTeamTypeDivision
	fqn := "department.division"
	email := "division@example.com"

	var model teamModel
	apply(&ometa.Team{
		Id:                 id,
		Name:               "division",
		FullyQualifiedName: &fqn,
		TeamType:           &teamType,
		Email:              &email,
		Parents:            &[]ometa.EntityReference{{Id: id, Type: "team", FullyQualifiedName: &parentFQN}},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.TeamType.ValueString() != "Division" {
		t.Fatalf("TeamType = %q", model.TeamType.ValueString())
	}
	if model.Email.ValueString() != email {
		t.Fatalf("Email = %q", model.Email.ValueString())
	}
	var parents []string
	if diags := model.Parents.ElementsAs(context.Background(), &parents, false); diags.HasError() {
		t.Fatalf("ElementsAs() diagnostics = %v", diags.Errors())
	}
	sort.Strings(parents)
	if !reflect.DeepEqual(parents, []string{parentFQN}) {
		t.Fatalf("Parents = %v", parents)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model teamModel
	apply(&ometa.Team{Id: id, Name: "bare"}, &model)

	if !model.IsJoinable.ValueBool() {
		t.Fatal("IsJoinable = false, want the server default true")
	}
	if model.Parents.IsNull() || len(model.Parents.Elements()) != 0 {
		t.Fatalf("Parents = %v, want an empty set", model.Parents)
	}
	if !model.DefaultPersona.IsNull() {
		t.Fatalf("DefaultPersona = %v, want null", model.DefaultPersona)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
	if !model.TeamType.IsNull() {
		t.Fatalf("TeamType = %v, want it untouched when the server omits it", model.TeamType)
	}
}
