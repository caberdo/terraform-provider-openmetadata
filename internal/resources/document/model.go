package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// documentModel is the Terraform representation of an OpenMetadata document.
// The document payload is kept as a JSON string so callers can build it with
// jsonencode() instead of a deeply nested schema.
type documentModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	EntityType         types.String  `tfsdk:"entity_type"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Data               types.String  `tfsdk:"data"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model. The payload is
// canonicalized so equivalent JSON documents produce the same state string, and
// an omitted payload leaves the prior value untouched.
func apply(document *ometa.Document, model *documentModel) {
	if document.Id != nil {
		model.ID = types.StringValue(document.Id.String())
	}
	model.Name = types.StringValue(document.Name)
	model.EntityType = types.StringValue(document.EntityType)
	model.FullyQualifiedName = types.StringValue(document.FullyQualifiedName)
	model.DisplayName = convert.StringValue(document.DisplayName)
	model.Description = convert.StringValueOr(document.Description, "")
	if document.Data != nil {
		if data, err := canonicalJSON(document.Data); err == nil {
			model.Data = types.StringValue(data)
		}
	}
	model.Version = convert.Float64Value(document.Version)
	model.UpdatedAt = convert.TimestampValue(document.UpdatedAt)
	model.UpdatedBy = convert.StringValue(document.UpdatedBy)
}

// parseData decodes the data attribute into the map the API expects. A JSON
// array or scalar is rejected because the document payload must be an object.
func parseData(raw string) (map[string]any, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("value must be a JSON object")
	}
	return data, nil
}

// canonicalJSON renders a value as compact JSON with sorted object keys and
// without HTML escaping. The trailing newline the encoder appends is trimmed.
func canonicalJSON(value any) (string, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

// canonicalData validates raw JSON and returns its canonical form.
func canonicalData(raw string) (string, error) {
	data, err := parseData(raw)
	if err != nil {
		return "", err
	}
	return canonicalJSON(data)
}
