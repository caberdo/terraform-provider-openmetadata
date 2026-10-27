package data_contract

import (
	"context"
	"reflect"
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

	if resp.TypeName != "openmetadata_data_contract" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_data_contract")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"id", "name", "entity", "fully_qualified_name", "entity_status", "version", "updated_at", "updated_by"} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatal("name must be a required string")
	}

	entity, ok := resp.Schema.Attributes["entity"].(rschema.StringAttribute)
	if !ok || !entity.Required {
		t.Fatal("entity must be a required string")
	}

	for _, attribute := range []string{"id", "fully_qualified_name", "entity_status", "version", "updated_at", "updated_by"} {
		switch typed := resp.Schema.Attributes[attribute].(type) {
		case rschema.StringAttribute:
			if !typed.Computed {
				t.Fatalf("%s must be computed", attribute)
			}
		case rschema.Float64Attribute:
			if !typed.Computed {
				t.Fatalf("%s must be computed", attribute)
			}
		default:
			t.Fatalf("%s has unexpected type %T", attribute, typed)
		}
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	fqn := "warehouse.analytics.public.users.monthly"
	entityFQN := "warehouse.analytics.public.users"
	displayName := "Monthly revenue"
	description := "Contract for monthly revenue reporting."
	terms := "Do not share outside finance."
	version := 0.2
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
	effectiveFrom := int64(1_767_225_600_000)
	effectiveUntil := int64(1_798_761_599_000)
	status := ometa.DataContractEntityStatus("Active")

	var model dataContractModel
	apply(&ometa.DataContract{
		Id:                 id,
		Name:               "monthly",
		FullyQualifiedName: &fqn,
		Entity:             ometa.EntityReference{FullyQualifiedName: &entityFQN, Type: "table"},
		DisplayName:        &displayName,
		Description:        &description,
		TermsOfUse:         &ometa.TermsOfUse{Content: &terms},
		EffectiveFrom:      &effectiveFrom,
		EffectiveUntil:     &effectiveUntil,
		EntityStatus:       &status,
		Version:            &version,
		UpdatedAt:          &updatedAt,
		UpdatedBy:          &updatedBy,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "monthly" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Entity.ValueString() != entityFQN {
		t.Fatalf("Entity = %q", model.Entity.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != description {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.TermsOfUse.ValueString() != terms {
		t.Fatalf("TermsOfUse = %q", model.TermsOfUse.ValueString())
	}
	if model.EffectiveFrom.ValueString() != "2026-01-01T00:00:00Z" {
		t.Fatalf("EffectiveFrom = %q", model.EffectiveFrom.ValueString())
	}
	if model.EffectiveUntil.ValueString() != "2026-12-31T23:59:59Z" {
		t.Fatalf("EffectiveUntil = %q", model.EffectiveUntil.ValueString())
	}
	if model.EntityStatus.ValueString() != "Active" {
		t.Fatalf("EntityStatus = %q", model.EntityStatus.ValueString())
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

func TestApplyKeepsOmittedFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	entityFQN := "warehouse.analytics.public.users"

	model := dataContractModel{
		Entity:         types.StringValue(entityFQN),
		TermsOfUse:     types.StringValue("Kept"),
		EffectiveFrom:  types.StringValue("2026-01-01T00:00:00Z"),
		EffectiveUntil: types.StringValue("2027-01-01T00:00:00Z"),
	}
	apply(&ometa.DataContract{Id: id, Name: "bare", Entity: ometa.EntityReference{Type: "table"}}, &model)

	if model.Entity.ValueString() != entityFQN {
		t.Fatalf("Entity = %q, want the prior value", model.Entity.ValueString())
	}
	if model.TermsOfUse.ValueString() != "Kept" {
		t.Fatalf("TermsOfUse = %q, want the prior value", model.TermsOfUse.ValueString())
	}
	if model.EffectiveFrom.ValueString() != "2026-01-01T00:00:00Z" {
		t.Fatalf("EffectiveFrom = %q, want the prior value", model.EffectiveFrom.ValueString())
	}
	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty default", model.Description.ValueString())
	}
	if !model.EntityStatus.IsNull() {
		t.Fatalf("EntityStatus = %v, want null", model.EntityStatus)
	}
}

func TestPatchOps(t *testing.T) {
	base := dataContractModel{
		DisplayName:    types.StringValue("Display"),
		Description:    types.StringValue("before"),
		TermsOfUse:     types.StringValue("old terms"),
		EffectiveFrom:  types.StringValue("2026-01-01T00:00:00Z"),
		EffectiveUntil: types.StringValue("2026-12-31T00:00:00Z"),
	}

	fromMillis := int64(1_767_225_600_000)

	tests := []struct {
		name  string
		plan  dataContractModel
		state dataContractModel
		want  []ometa.JSONPatchOp
	}{
		{
			name:  "no changes",
			plan:  base,
			state: base,
			want:  nil,
		},
		{
			name: "display name replaced",
			plan: dataContractModel{
				DisplayName:    types.StringValue("New display"),
				Description:    base.Description,
				TermsOfUse:     base.TermsOfUse,
				EffectiveFrom:  base.EffectiveFrom,
				EffectiveUntil: base.EffectiveUntil,
			},
			state: base,
			want: []ometa.JSONPatchOp{
				{Op: "add", Path: "/displayName", Value: "New display"},
			},
		},
		{
			name: "display name removed",
			plan: dataContractModel{
				DisplayName:    types.StringNull(),
				Description:    base.Description,
				TermsOfUse:     base.TermsOfUse,
				EffectiveFrom:  base.EffectiveFrom,
				EffectiveUntil: base.EffectiveUntil,
			},
			state: base,
			want: []ometa.JSONPatchOp{
				{Op: "remove", Path: "/displayName"},
			},
		},
		{
			name: "description and terms changed",
			plan: dataContractModel{
				DisplayName:    base.DisplayName,
				Description:    types.StringValue("after"),
				TermsOfUse:     types.StringValue("new terms"),
				EffectiveFrom:  base.EffectiveFrom,
				EffectiveUntil: base.EffectiveUntil,
			},
			state: base,
			want: []ometa.JSONPatchOp{
				{Op: "add", Path: "/description", Value: "after"},
				{Op: "add", Path: "/termsOfUse", Value: map[string]any{"content": "new terms"}},
			},
		},
		{
			name: "terms removed",
			plan: dataContractModel{
				DisplayName:    base.DisplayName,
				Description:    base.Description,
				TermsOfUse:     types.StringNull(),
				EffectiveFrom:  base.EffectiveFrom,
				EffectiveUntil: base.EffectiveUntil,
			},
			state: base,
			want: []ometa.JSONPatchOp{
				{Op: "remove", Path: "/termsOfUse"},
			},
		},
		{
			name: "effective window changed",
			plan: dataContractModel{
				DisplayName:    base.DisplayName,
				Description:    base.Description,
				TermsOfUse:     base.TermsOfUse,
				EffectiveFrom:  types.StringValue("2026-01-01T00:00:00Z"),
				EffectiveUntil: types.StringValue("2028-01-01T00:00:00Z"),
			},
			state: dataContractModel{
				DisplayName:    base.DisplayName,
				Description:    base.Description,
				TermsOfUse:     base.TermsOfUse,
				EffectiveFrom:  types.StringValue("2025-01-01T00:00:00Z"),
				EffectiveUntil: types.StringValue("2026-01-01T00:00:00Z"),
			},
			want: []ometa.JSONPatchOp{
				{Op: "add", Path: "/effectiveFrom", Value: fromMillis},
				{Op: "add", Path: "/effectiveUntil", Value: int64(1_830_297_600_000)},
			},
		},
		{
			name: "effective window cleared",
			plan: dataContractModel{
				DisplayName:    base.DisplayName,
				Description:    base.Description,
				TermsOfUse:     base.TermsOfUse,
				EffectiveFrom:  types.StringNull(),
				EffectiveUntil: types.StringNull(),
			},
			state: base,
			want: []ometa.JSONPatchOp{
				{Op: "remove", Path: "/effectiveFrom"},
				{Op: "remove", Path: "/effectiveUntil"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := patchOps(tt.plan, tt.state)
			if len(ops) != len(tt.want) {
				t.Fatalf("len(ops) = %d, want %d: %#v", len(ops), len(tt.want), ops)
			}
			for i := range tt.want {
				if ops[i].Op != tt.want[i].Op || ops[i].Path != tt.want[i].Path {
					t.Fatalf("op %d = {%q %q}, want {%q %q}", i, ops[i].Op, ops[i].Path, tt.want[i].Op, tt.want[i].Path)
				}
				if tt.want[i].Value != nil && !reflect.DeepEqual(ops[i].Value, tt.want[i].Value) {
					t.Fatalf("op %d value = %#v, want %#v", i, ops[i].Value, tt.want[i].Value)
				}
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	if _, ok := parseTimestamp(types.StringNull()); ok {
		t.Fatal("null timestamp parsed")
	}
	if _, ok := parseTimestamp(types.StringValue("not-a-time")); ok {
		t.Fatal("invalid timestamp parsed")
	}

	millis, ok := parseTimestamp(types.StringValue("2026-01-01T00:00:00Z"))
	if !ok || millis == nil || *millis != 1_767_225_600_000 {
		t.Fatalf("parseTimestamp = %v, %v", millis, ok)
	}
}
