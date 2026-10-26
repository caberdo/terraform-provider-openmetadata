package tag

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

	if resp.TypeName != "openmetadata_tag" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_tag")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"classification", "name"} {
		required, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !required.Required {
			t.Fatalf("%s must be a required string", attribute)
		}
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
	classificationFQN := "Sensitivity"
	fqn := "Sensitivity.Confidential"
	displayName := "Confidential"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	mutuallyExclusive := true

	var model tagModel
	apply(&ometa.Tag{
		Id:                 id,
		Name:               "Confidential",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        "Confidential data.",
		Classification: &ometa.EntityReference{
			FullyQualifiedName: &classificationFQN,
			Type:               "classification",
		},
		MutuallyExclusive: &mutuallyExclusive,
		Version:           &version,
		UpdatedAt:         &updatedAt,
		UpdatedBy:         stringPointer("admin"),
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Classification.ValueString() != classificationFQN {
		t.Fatalf("Classification = %q", model.Classification.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if !model.MutuallyExclusive.ValueBool() {
		t.Fatal("MutuallyExclusive = false")
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyKeepsClassificationWhenReferenceIsMissing(t *testing.T) {
	var model tagModel
	model.Classification = types.StringValue("Sensitivity")

	apply(&ometa.Tag{Id: uuid.New(), Name: "Confidential", Description: ""}, &model)

	if model.Classification.ValueString() != "Sensitivity" {
		t.Fatalf("Classification = %q, want the prior state", model.Classification.ValueString())
	}
}

func TestPatchOps(t *testing.T) {
	base := tagModel{
		Classification: types.StringValue("Sensitivity"),
		DisplayName:    types.StringValue("Display"),
		Description:    types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      tagModel
		state     tagModel
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
			plan: tagModel{
				Classification: base.Classification,
				DisplayName:    types.StringValue("New display"),
				Description:    base.Description,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "description changed",
			plan: tagModel{
				Classification: base.Classification,
				DisplayName:    base.DisplayName,
				Description:    types.StringValue("after"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "mutually exclusive changed",
			plan: tagModel{
				Classification:    base.Classification,
				DisplayName:       base.DisplayName,
				Description:       base.Description,
				MutuallyExclusive: types.BoolValue(true),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/mutuallyExclusive",
			wantOp:    "add",
			wantValue: true,
		},
		{
			name: "classification cannot be patched",
			plan: tagModel{
				Classification: types.StringValue("Other"),
				DisplayName:    base.DisplayName,
				Description:    base.Description,
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

func stringPointer(value string) *string { return &value }
