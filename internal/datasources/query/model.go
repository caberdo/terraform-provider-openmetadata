package query

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// queryModel is the Terraform representation of a query lookup. The id and
// fully_qualified_name attributes are both selectors and results.
type queryModel struct {
	ID                 types.String  `tfsdk:"id"`
	Service            types.String  `tfsdk:"service"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Query              types.String  `tfsdk:"query"`
	QueryType          types.String  `tfsdk:"query_type"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(query *ometa.Query, model *queryModel) {
	model.ID = types.StringValue(query.Id.String())
	model.Name = types.StringValue(query.Name)
	model.FullyQualifiedName = convert.StringValue(query.FullyQualifiedName)
	model.DisplayName = convert.StringValue(query.DisplayName)
	model.Description = convert.StringValueOr(query.Description, "")
	model.Query = types.StringValue(query.Query)
	model.QueryType = convert.StringValue(query.QueryType)
	model.Version = convert.Float64Value(query.Version)
	model.UpdatedAt = convert.TimestampValue(query.UpdatedAt)
	model.UpdatedBy = convert.StringValue(query.UpdatedBy)

	if fqn := query.Service.FullyQualifiedName; fqn != nil {
		model.Service = types.StringValue(*fqn)
	} else if name := query.Service.Name; name != nil {
		model.Service = types.StringValue(*name)
	}
}
