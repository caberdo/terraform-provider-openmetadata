package glossary_term

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"
)

func TestMetadata(t *testing.T) {
	var resp resource.MetadataResponse
	NewResource().Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "openmetadata"}, &resp)

	if resp.TypeName != "openmetadata_glossary_term" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_glossary_term")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"glossary", "name"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
	}

	synonyms, ok := resp.Schema.Attributes["synonyms"].(rschema.ListAttribute)
	if !ok || !synonyms.Optional || !synonyms.Computed {
		t.Fatal("synonyms must be an optional computed list")
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "version", "updated_at", "updated_by"} {
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
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	glossary := "finance"
	parent := "finance.Cash"
	fqn := "finance.Cash.PettyCash"
	displayName := "Petty cash"
	synonyms := []string{"cash on hand", "float"}
	version := 0.4
	updatedAt := int64(1_700_000_000_000)

	var model glossaryTermModel
	apply(&ometa.GlossaryTerm{
		Id:                 id,
		Name:               "PettyCash",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        "Small cash expenses.",
		Glossary: ometa.EntityReference{
			FullyQualifiedName: &glossary,
			Type:               "glossary",
		},
		Parent: &ometa.EntityReference{
			FullyQualifiedName: &parent,
			Type:               "glossaryTerm",
		},
		Synonyms:  &synonyms,
		Version:   &version,
		UpdatedAt: &updatedAt,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Glossary.ValueString() != glossary {
		t.Fatalf("Glossary = %q", model.Glossary.ValueString())
	}
	if model.Parent.ValueString() != parent {
		t.Fatalf("Parent = %q", model.Parent.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	var gotSynonyms []string
	if diags := model.Synonyms.ElementsAs(context.Background(), &gotSynonyms, false); diags.HasError() {
		t.Fatalf("Synonyms conversion: %v", diags.Errors())
	}
	if !reflect.DeepEqual(gotSynonyms, synonyms) {
		t.Fatalf("Synonyms = %#v, want %#v", gotSynonyms, synonyms)
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyDefaultsMissingCollections(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model glossaryTermModel
	apply(&ometa.GlossaryTerm{Id: id, Name: "bare", Description: ""}, &model)

	if !model.Synonyms.IsNull() {
		t.Fatalf("Synonyms = %v, want null", model.Synonyms)
	}
	if !model.Parent.IsNull() {
		t.Fatalf("Parent = %v, want null", model.Parent)
	}
	if model.MutuallyExclusive.ValueBool() {
		t.Fatal("MutuallyExclusive = true, want the false default")
	}
}

func TestPatchOps(t *testing.T) {
	base := glossaryTermModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		Synonyms:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("old")}),
	}
	otherSynonyms := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("new")})

	tests := []struct {
		name      string
		plan      glossaryTermModel
		state     glossaryTermModel
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
			plan: glossaryTermModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				Synonyms:    base.Synonyms,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "description changed",
			plan: glossaryTermModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				Synonyms:    base.Synonyms,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "synonyms replaced",
			plan: glossaryTermModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Synonyms:    otherSynonyms,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/synonyms",
			wantOp:    "add",
			wantValue: []string{"new"},
		},
		{
			name: "synonyms removed",
			plan: glossaryTermModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				Synonyms:    types.ListNull(types.StringType),
			},
			state:    base,
			wantOps:  1,
			wantPath: "/synonyms",
			wantOp:   "remove",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, diags := patchOps(context.Background(), tt.plan, tt.state)
			if diags.HasError() {
				t.Fatalf("patchOps() diagnostics = %v", diags.Errors())
			}
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
			if tt.wantValue != nil && !reflect.DeepEqual(ops[0].Value, tt.wantValue) {
				t.Fatalf("value = %#v, want %#v", ops[0].Value, tt.wantValue)
			}
		})
	}
}
