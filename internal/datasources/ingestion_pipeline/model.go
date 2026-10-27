package ingestion_pipeline

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// ingestionPipelineModel is the Terraform representation of an OpenMetadata
// ingestion pipeline. The source and Airflow configurations are exposed as JSON
// object strings, because their shape depends on the pipeline type.
type ingestionPipelineModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Service            types.String  `tfsdk:"service"`
	ServiceType        types.String  `tfsdk:"service_type"`
	PipelineType       types.String  `tfsdk:"pipeline_type"`
	SourceConfig       types.String  `tfsdk:"source_config"`
	AirflowConfig      types.String  `tfsdk:"airflow_config"`
	LoggerLevel        types.String  `tfsdk:"logger_level"`
	EnableStreamable   types.Bool    `tfsdk:"enable_streamable_logs"`
	RaiseOnError       types.Bool    `tfsdk:"raise_on_error"`
	Deployed           types.Bool    `tfsdk:"deployed"`
	Enabled            types.Bool    `tfsdk:"enabled"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(pipeline *ometa.IngestionPipeline, model *ingestionPipelineModel) {
	model.ID = types.StringNull()
	if pipeline.Id != nil {
		model.ID = types.StringValue(pipeline.Id.String())
	}
	model.Name = types.StringValue(pipeline.Name)
	model.FullyQualifiedName = convert.StringValue(pipeline.FullyQualifiedName)
	model.DisplayName = convert.StringValue(pipeline.DisplayName)
	model.Description = convert.StringValueOr(pipeline.Description, "")

	model.Service = types.StringNull()
	model.ServiceType = types.StringNull()
	if pipeline.Service != nil {
		if service := convert.StringValue(pipeline.Service.FullyQualifiedName); !service.IsNull() {
			model.Service = service
		}
		if pipeline.Service.Type != "" {
			model.ServiceType = types.StringValue(pipeline.Service.Type)
		}
	}

	model.PipelineType = types.StringValue(string(pipeline.PipelineType))
	model.SourceConfig = sourceConfigValue(pipeline.SourceConfig.Config, model.SourceConfig)
	model.AirflowConfig = airflowConfigValue(pipeline.AirflowConfig, model.AirflowConfig)
	model.LoggerLevel = loggerLevelValue(pipeline.LoggerLevel)
	model.EnableStreamable = convert.BoolValueOr(pipeline.EnableStreamableLogs, false)
	model.RaiseOnError = convert.BoolValueOr(pipeline.RaiseOnError, false)
	model.Deployed = convert.BoolValueOr(pipeline.Deployed, false)
	model.Enabled = convert.BoolValueOr(pipeline.Enabled, false)
	model.Version = convert.Float64Value(pipeline.Version)
	model.UpdatedAt = convert.TimestampValue(pipeline.UpdatedAt)
	model.UpdatedBy = convert.StringValue(pipeline.UpdatedBy)
}

// sourceConfigValue renders the source configuration as canonical JSON.
func sourceConfigValue(config any, existing types.String) types.String {
	if config == nil {
		return keepExisting(existing)
	}
	encoded, err := json.Marshal(config)
	if err != nil {
		return keepExisting(existing)
	}
	return sameOrValue(string(encoded), existing)
}

// airflowConfigValue renders the Airflow configuration as canonical JSON.
func airflowConfigValue(airflow ometa.AirflowConfig, existing types.String) types.String {
	encoded, err := json.Marshal(airflow)
	if err != nil {
		return keepExisting(existing)
	}
	return sameOrValue(string(encoded), existing)
}

// loggerLevelValue maps the API logger level to a Terraform string.
func loggerLevelValue(level *ometa.IngestionPipelineLoggerLevel) types.String {
	if level == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*level))
}

// keepExisting keeps a prior attribute value when the API omits the field.
func keepExisting(existing types.String) types.String {
	if existing.IsUnknown() {
		return types.StringNull()
	}
	return existing
}

// sameOrValue keeps the prior spelling when the API document is semantically
// equal, so reformatted JSON does not churn state.
func sameOrValue(value string, existing types.String) types.String {
	if !existing.IsUnknown() && !existing.IsNull() && jsonEqual(existing.ValueString(), value) {
		return existing
	}
	return types.StringValue(value)
}

// jsonEqual reports whether two JSON documents are semantically equal,
// ignoring key order and whitespace.
func jsonEqual(left, right string) bool {
	var decodedLeft any
	if err := json.Unmarshal([]byte(left), &decodedLeft); err != nil {
		return false
	}
	var decodedRight any
	if err := json.Unmarshal([]byte(right), &decodedRight); err != nil {
		return false
	}
	return reflect.DeepEqual(decodedLeft, decodedRight)
}
