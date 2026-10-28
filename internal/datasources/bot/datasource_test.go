package bot

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

	if resp.TypeName != "openmetadata_bot" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_bot")
	}
}

func TestSchema(t *testing.T) {
	var resp datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	selectors := map[string]struct{ optional, computed bool }{
		"id":                   {true, true},
		"fully_qualified_name": {true, true},
	}
	computed := []string{"name", "display_name", "description", "bot_user", "updated_at", "updated_by"}

	for attribute, want := range selectors {
		got, ok := resp.Schema.Attributes[attribute].(dschema.StringAttribute)
		if !ok {
			t.Fatalf("%s is not a string attribute", attribute)
		}
		if got.Optional != want.optional || got.Computed != want.computed {
			t.Fatalf("%s optional=%v computed=%v, want optional=%v computed=%v",
				attribute, got.Optional, got.Computed, want.optional, want.computed)
		}
	}

	for _, attribute := range computed {
		if _, ok := resp.Schema.Attributes[attribute]; !ok {
			t.Fatalf("schema is missing %q", attribute)
		}
	}
	if _, ok := resp.Schema.Attributes["version"]; !ok {
		t.Fatal("schema is missing \"version\"")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	botUserFQN := "automation-bot"
	displayName := "Automation bot"
	fqn := "reporting-bot"
	description := "Runs scheduled reports."

	var model botModel
	apply(&ometa.Bot{
		Id:                 id,
		Name:               "reporting-bot",
		FullyQualifiedName: &fqn,
		DisplayName:        &displayName,
		Description:        &description,
		BotUser: ometa.EntityReference{
			Id:                 id,
			Type:               "user",
			FullyQualifiedName: &botUserFQN,
		},
	}, &model)

	if model.ID.ValueString() != id.String() {
		t.Fatalf("ID = %q", model.ID.ValueString())
	}
	if model.Name.ValueString() != "reporting-bot" {
		t.Fatalf("Name = %q", model.Name.ValueString())
	}
	if model.FullyQualifiedName.ValueString() != fqn {
		t.Fatalf("FullyQualifiedName = %q", model.FullyQualifiedName.ValueString())
	}
	if model.DisplayName.ValueString() != displayName {
		t.Fatalf("DisplayName = %q", model.DisplayName.ValueString())
	}
	if model.Description.ValueString() != "Runs scheduled reports." {
		t.Fatalf("Description = %q", model.Description.ValueString())
	}
	if model.BotUser.ValueString() != botUserFQN {
		t.Fatalf("BotUser = %q", model.BotUser.ValueString())
	}
}

func TestApplyBotUserFallsBackToName(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	botUserName := "automation-bot"

	var model botModel
	apply(&ometa.Bot{
		Id:   id,
		Name: "reporting-bot",
		BotUser: ometa.EntityReference{
			Id:   id,
			Type: "user",
			Name: &botUserName,
		},
	}, &model)

	if model.BotUser.ValueString() != botUserName {
		t.Fatalf("BotUser = %q, want %q", model.BotUser.ValueString(), botUserName)
	}
}

func TestApplyDefaultsMissingFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	var model botModel
	apply(&ometa.Bot{Id: id, Name: "bare"}, &model)

	if model.Description.ValueString() != "" {
		t.Fatalf("Description = %q, want the empty string", model.Description.ValueString())
	}
	if !model.DisplayName.IsNull() {
		t.Fatalf("DisplayName = %v, want null", model.DisplayName)
	}
	if !model.BotUser.IsNull() {
		t.Fatalf("BotUser = %v, want null", model.BotUser)
	}
}
