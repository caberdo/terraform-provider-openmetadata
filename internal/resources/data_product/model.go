package data_product

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// dataProductModel is the Terraform representation of an OpenMetadata data
// product.
type dataProductModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Domains            types.List    `tfsdk:"domains"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The required
// domains selector keeps its prior state when the server omits the references.
func apply(product *ometa.DataProduct, model *dataProductModel) {
	model.ID = types.StringValue(product.Id.String())
	model.Name = types.StringValue(product.Name)
	model.FullyQualifiedName = convert.StringValue(product.FullyQualifiedName)
	model.DisplayName = convert.StringValue(product.DisplayName)
	model.Description = convert.StringValueOr(&product.Description, "")
	if product.Domains != nil {
		model.Domains = referenceListValue(*product.Domains)
	}
	model.Version = convert.Float64Value(product.Version)
	model.UpdatedAt = convert.TimestampValue(product.UpdatedAt)
	model.UpdatedBy = convert.StringValue(product.UpdatedBy)
}

// referenceListValue maps entity references to the fully qualified names the
// API accepts on write.
func referenceListValue(references []ometa.EntityReference) types.List {
	elements := make([]attr.Value, 0, len(references))
	for _, reference := range references {
		if fqn := reference.FullyQualifiedName; fqn != nil {
			elements = append(elements, types.StringValue(*fqn))
		} else if name := reference.Name; name != nil {
			elements = append(elements, types.StringValue(*name))
		}
	}
	return types.ListValueMust(types.StringType, elements)
}
