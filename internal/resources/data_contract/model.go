package data_contract

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// dataContractModel is the Terraform representation of an OpenMetadata data
// contract. Effective timestamps are exposed as RFC 3339 strings while the API
// stores epoch milliseconds.
type dataContractModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	Entity             types.String  `tfsdk:"entity"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	TermsOfUse         types.String  `tfsdk:"terms_of_use"`
	EffectiveFrom      types.String  `tfsdk:"effective_from"`
	EffectiveUntil     types.String  `tfsdk:"effective_until"`
	EntityStatus       types.String  `tfsdk:"entity_status"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. Fields the API
// omits keep their prior value so a partial response never zeroes state.
func apply(contract *ometa.DataContract, model *dataContractModel) {
	model.ID = types.StringValue(contract.Id.String())
	model.Name = types.StringValue(contract.Name)
	model.FullyQualifiedName = convert.StringValue(contract.FullyQualifiedName)

	switch {
	case contract.Entity.FullyQualifiedName != nil:
		model.Entity = types.StringValue(*contract.Entity.FullyQualifiedName)
	case contract.Entity.Name != nil:
		model.Entity = types.StringValue(*contract.Entity.Name)
	}

	model.DisplayName = convert.StringValue(contract.DisplayName)
	model.Description = convert.StringValueOr(contract.Description, "")

	if contract.TermsOfUse != nil {
		model.TermsOfUse = convert.StringValue(contract.TermsOfUse.Content)
	}

	if contract.EffectiveFrom != nil {
		model.EffectiveFrom = convert.TimestampValue(contract.EffectiveFrom)
	}
	if contract.EffectiveUntil != nil {
		model.EffectiveUntil = convert.TimestampValue(contract.EffectiveUntil)
	}

	model.EntityStatus = convert.StringValue((*string)(contract.EntityStatus))
	model.Version = convert.Float64Value(contract.Version)
	model.UpdatedAt = convert.TimestampValue(contract.UpdatedAt)
	model.UpdatedBy = convert.StringValue(contract.UpdatedBy)
}

// parseTimestamp converts an RFC 3339 string to epoch milliseconds. Null,
// unknown, and malformed values yield false so the caller can skip the field.
func parseTimestamp(value types.String) (*int64, bool) {
	if value.IsNull() || value.IsUnknown() {
		return nil, false
	}

	parsed, err := time.Parse(time.RFC3339, value.ValueString())
	if err != nil {
		return nil, false
	}

	millis := parsed.UnixMilli()
	return &millis, true
}
