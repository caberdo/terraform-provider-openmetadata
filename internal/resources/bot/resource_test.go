package bot

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

	if resp.TypeName != "openmetadata_bot" {
		t.Fatalf("TypeName = %q, want %q", resp.TypeName, "openmetadata_bot")
	}
}

func TestSchema(t *testing.T) {
	var resp resource.SchemaResponse
	NewResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() diagnostics = %v", resp.Diagnostics.Errors())
	}

	name, ok := resp.Schema.Attributes["name"].(rschema.StringAttribute)
	if !ok || !name.Required {
		t.Fatalf("name attribute is %T, want a required StringAttribute", resp.Schema.Attributes["name"])
	}

	botUser, ok := resp.Schema.Attributes["bot_user"].(rschema.StringAttribute)
	if !ok || !botUser.Required {
		t.Fatalf("bot_user attribute is %T, want a required StringAttribute", resp.Schema.Attributes["bot_user"])
	}

	for _, attribute := range []string{
		"id",
		"fully_qualified_name",
		"updated_at",
		"updated_by",
	} {
		computed, ok := resp.Schema.Attributes[attribute].(rschema.StringAttribute)
		if !ok || !computed.Computed {
			t.Fatalf("%s must be a computed string", attribute)
		}
	}

	version, ok := resp.Schema.Attributes["version"].(rschema.Float64Attribute)
	if !ok || !version.Computed {
		t.Fatalf("version must be a computed float")
	}

	displayName, ok := resp.Schema.Attributes["display_name"].(rschema.StringAttribute)
	if !ok || !displayName.Optional {
		t.Fatalf("display_name must be an optional string")
	}

	description, ok := resp.Schema.Attributes["description"].(rschema.StringAttribute)
	if !ok || !description.Optional || !description.Computed {
		t.Fatalf("description must be optional and computed")
	}
}

func TestApplyMapsAPIFields(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")
	botUserFQN := "automation-bot"
	displayName := "Automation bot"
	fqn := "reporting-bot"
	version := 0.4
	updatedAt := int64(1_700_000_000_000)
	updatedBy := "admin"
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
		Version:   &version,
		UpdatedAt: &updatedAt,
		UpdatedBy: &updatedBy,
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

func TestApplyPreservesConfiguredBotUserWhenServerOmitsIt(t *testing.T) {
	id := uuid.MustParse("1f6a3a52-8d34-4f4b-9d0f-2f1c1c1c1c1c")

	model := botModel{BotUser: types.StringValue("configured-bot-user")}
	apply(&ometa.Bot{
		Id:      id,
		Name:    "reporting-bot",
		BotUser: ometa.EntityReference{Type: "user"},
	}, &model)

	if model.BotUser.ValueString() != "configured-bot-user" {
		t.Fatalf("BotUser = %q, want the configured value preserved", model.BotUser.ValueString())
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
		t.Fatalf("BotUser = %v, want null when both the server and the plan omit it", model.BotUser)
	}
	if !model.UpdatedAt.IsNull() {
		t.Fatalf("UpdatedAt = %v, want null", model.UpdatedAt)
	}
}

func TestPatchOps(t *testing.T) {
	base := botModel{
		Name:        types.StringValue("reporting-bot"),
		BotUser:     types.StringValue("automation-bot"),
		DisplayName: types.StringValue("Display"),
		Description: types.StringValue("before"),
	}

	tests := []struct {
		name      string
		plan      botModel
		state     botModel
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
			plan: botModel{
				Name:        base.Name,
				BotUser:     base.BotUser,
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
			plan: botModel{
				Name:        base.Name,
				BotUser:     base.BotUser,
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
			plan: botModel{
				Name:        base.Name,
				BotUser:     base.BotUser,
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
			name: "name and bot user are never patched",
			plan: botModel{
				Name:        types.StringValue("renamed-bot"),
				BotUser:     types.StringValue("other-bot-user"),
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
