package domain

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

	if resp.TypeName != "openmetadata_domain" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_domain")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	for _, attribute := range []string{"name", "domain_type"} {
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
	fqn := "Engineering.Platform"
	displayName := "Platform"
	parent := "Engineering"
	version := 0.2
	updatedAt := int64(1_700_000_000_000)

	var model domainModel
	apply(&ometa.Domain{
		Id:                 id,
		Name:               "Platform",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        "Platform engineering assets.",
		DomainType:         ometa.DomainDomainTypeSourceAligned,
		Parent: &ometa.EntityReference{
			FullyQualifiedName: &parent,
			Type:               "domain",
		},
		Version:   &version,
		UpdatedAt: &updatedAt,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "Platform" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DomainType.ValueString() != "Source-aligned" {
		t.Fatalf("DomainType = %q", model.DomainType.ValueString())
	}
	if model.Parent.ValueString() != parent {
		t.Fatalf("Parent = %q", model.Parent.ValueString())
	}
	if model.Version.ValueFloat64() != version {
		t.Fatalf("Version = %v", model.Version.ValueFloat64())
	}
	if model.UpdatedAt.ValueString() != "2023-11-14T22:13:20Z" {
		t.Fatalf("UpdatedAt = %q", model.UpdatedAt.ValueString())
	}
}

func TestApplyKeepsDomainTypeWhenMissing(t *testing.T) {
	var model domainModel
	model.DomainType = types.StringValue("Aggregate")

	apply(&ometa.Domain{Id: uuid.New(), Name: "bare"}, &model)

	if model.DomainType.ValueString() != "Aggregate" {
		t.Fatalf("DomainType = %q, want the prior state", model.DomainType.ValueString())
	}
	if !model.Parent.IsNull() {
		t.Fatalf("Parent = %v, want null", model.Parent)
	}
}

func TestPatchOps(t *testing.T) {
	base := domainModel{
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
		DomainType:  types.StringValue("Aggregate"),
	}

	tests := []struct {
		name      string
		plan      domainModel
		state     domainModel
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
			plan: domainModel{
				DisplayName: types.StringValue("New display"),
				Description: base.Description,
				DomainType:  base.DomainType,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/displayName",
			wantOp:    "add",
			wantValue: "New display",
		},
		{
			name: "description changed",
			plan: domainModel{
				DisplayName: base.DisplayName,
				Description: types.StringValue("after"),
				DomainType:  base.DomainType,
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/description",
			wantOp:    "add",
			wantValue: "after",
		},
		{
			name: "domain type changed",
			plan: domainModel{
				DisplayName: base.DisplayName,
				Description: base.Description,
				DomainType:  types.StringValue("Consumer-aligned"),
			},
			state:     base,
			wantOps:   1,
			wantPath:  "/domainType",
			wantOp:    "add",
			wantValue: "Consumer-aligned",
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
