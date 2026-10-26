package domain

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// domainModel is the Terraform representation of an OpenMetadata domain.
type domainModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	DomainType         types.String  `tfsdk:"domain_type"`
	Parent             types.String  `tfsdk:"parent"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The required
// domain type and the parent selector keep their prior value when omitted.
func apply(domain *ometa.Domain, model *domainModel) {
	model.ID = types.StringValue(domain.Id.String())
	model.Name = types.StringValue(domain.Name)
	model.FullyQualifiedName = convert.StringValue(domain.FullyQualifiedName)
	model.DisplayName = convert.StringValue(domain.DisplayName)
	model.Description = convert.StringValueOr(&domain.Description, "")
	if domain.DomainType != "" {
		model.DomainType = types.StringValue(string(domain.DomainType))
	}
	model.Version = convert.Float64Value(domain.Version)
	model.UpdatedAt = convert.TimestampValue(domain.UpdatedAt)
	model.UpdatedBy = convert.StringValue(domain.UpdatedBy)

	if domain.Parent != nil {
		if fqn := domain.Parent.FullyQualifiedName; fqn != nil {
			model.Parent = types.StringValue(*fqn)
		} else if name := domain.Parent.Name; name != nil {
			model.Parent = types.StringValue(*name)
		}
	}
}
