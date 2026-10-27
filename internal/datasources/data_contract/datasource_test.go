package data_contract

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

	if resp.TypeName != "openmetadata_data_contract" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_data_contract")
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
		"entity",
		"display_name",
		"description",
		"terms_of_use",
		"effective_from",
		"effective_until",
		"entity_status",
		"version",
		"updated_at",
		"updated_by",
	} {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}

	for _, attribute := range []string{"id", "fully_qualified_name"} {
		selector, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok || !selector.Optional || !selector.Computed {
			t.Fatalf("%s must be an optional, computed selector", attribute)
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
	effectiveFrom := int64(1_767_225_600_000)
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
		EntityStatus:       &status,
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.Name.ValueString() != "monthly" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.Entity.ValueString() != entityFQN {
		t.Fatalf("Entity = %q", model.Entity.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.TermsOfUse.ValueString() != terms {
		t.Fatalf("TermsOfUse = %q", model.TermsOfUse.ValueString())
	}
	if model.EffectiveFrom.ValueString() != "2026-01-01T00:00:00Z" {
		t.Fatalf("EffectiveFrom = %q", model.EffectiveFrom.ValueString())
	}
	if model.EntityStatus.ValueString() != "Active" {
		t.Fatalf("EntityStatus = %q", model.EntityStatus.ValueString())
	}
}
