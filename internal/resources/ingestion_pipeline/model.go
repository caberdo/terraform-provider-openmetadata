package ingestion_pipeline

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// ingestionPipelineModel is the Terraform representation of an OpenMetadata
// ingestion pipeline. The source and Airflow configurations are kept as JSON
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
	if pipeline.Id != nil {
		model.ID = types.StringValue(pipeline.Id.String())
	}
	model.Name = types.StringValue(pipeline.Name)
	model.FullyQualifiedName = convert.StringValue(pipeline.FullyQualifiedName)
	model.DisplayName = convert.StringValue(pipeline.DisplayName)
	model.Description = convert.StringValueOr(pipeline.Description, "")

	if pipeline.Service != nil {
		if service := convert.StringValue(pipeline.Service.FullyQualifiedName); !service.IsNull() {
			model.Service = service
		}
		if pipeline.Service.Type != "" {
			model.ServiceType = types.StringValue(pipeline.Service.Type)
		}
	}

	model.PipelineType = types.StringValue(string(pipeline.PipelineType))
	model.SourceConfig = sourceConfigValue(pipeline.SourceConfig, model.SourceConfig)
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

// sourceConfigValue renders the source configuration as canonical JSON. A
// missing configuration never zeroes a value that is already in state.
func sourceConfigValue(source ometa.SourceConfig, existing types.String) types.String {
	if source.Config == nil {
		return keepExisting(existing)
	}
	encoded, err := json.Marshal(source.Config)
	if err != nil {
		return keepExisting(existing)
	}
	return sameOrSuperset(string(encoded), existing)
}

// airflowConfigValue renders the Airflow configuration as canonical JSON. The
// API field is a plain struct, so an empty object is a real value.
func airflowConfigValue(airflow ometa.AirflowConfig, existing types.String) types.String {
	encoded, err := json.Marshal(airflow)
	if err != nil {
		return keepExisting(existing)
	}
	return sameOrSuperset(string(encoded), existing)
}

// loggerLevelValue maps the API logger level to a Terraform string.
func loggerLevelValue(level *ometa.IngestionPipelineLoggerLevel) types.String {
	if level == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*level))
}

// keepExisting keeps a prior attribute value when the API omits the field, so
// state is not blanked by a partial response.
func keepExisting(existing types.String) types.String {
	if existing.IsUnknown() {
		return types.StringNull()
	}
	return existing
}

// sameOrSuperset keeps the prior spelling when it is semantically equal to the
// API document or a subset of it. OpenMetadata expands the Airflow and source
// configs with server defaults; keeping the configured subset avoids a
// perpetual diff and a post-apply consistency error.
func sameOrSuperset(value string, existing types.String) types.String {
	if !existing.IsUnknown() && !existing.IsNull() && jsonSubset(existing.ValueString(), value) {
		return existing
	}
	return types.StringValue(value)
}

// jsonSubset reports whether every value in subset also appears in superset.
func jsonSubset(subset, superset string) bool {
	var sub, super any
	if err := json.Unmarshal([]byte(subset), &sub); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(superset), &super); err != nil {
		return false
	}
	return isSubset(sub, super)
}

func isSubset(subset, superset any) bool {
	switch sub := subset.(type) {
	case map[string]any:
		super, ok := superset.(map[string]any)
		if !ok {
			return false
		}
		for key, value := range sub {
			other, ok := super[key]
			if !ok || !isSubset(value, other) {
				return false
			}
		}
		return true
	case []any:
		super, ok := superset.([]any)
		if !ok || len(sub) != len(super) {
			return false
		}
		for index, value := range sub {
			if !isSubset(value, super[index]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(subset, superset)
	}
}

// createBody builds the create payload from the planned model.
func createBody(plan ingestionPipelineModel, service ometa.EntityReference) (*ometa.CreateIngestionPipeline, error) {
	source, err := parseJSON(plan.SourceConfig.ValueString())
	if err != nil {
		return nil, fmt.Errorf("source_config must be a JSON object: %w", err)
	}

	airflow, err := decodeAirflowConfig(plan.AirflowConfig.ValueString())
	if err != nil {
		return nil, fmt.Errorf("airflow_config must be a JSON object: %w", err)
	}

	body := &ometa.CreateIngestionPipeline{
		Name:          plan.Name.ValueString(),
		Service:       service,
		PipelineType:  ometa.CreateIngestionPipelinePipelineType(plan.PipelineType.ValueString()),
		SourceConfig:  ometa.SourceConfig{Config: source},
		AirflowConfig: *airflow,
		Description:   convert.StringPointer(plan.Description),
		DisplayName:   convert.StringPointer(plan.DisplayName),
		RaiseOnError:  convert.BoolPointer(plan.RaiseOnError),

		EnableStreamableLogs: convert.BoolPointer(plan.EnableStreamable),
	}

	if level := plan.LoggerLevel; !level.IsNull() && !level.IsUnknown() && level.ValueString() != "" {
		cast := ometa.CreateIngestionPipelineLoggerLevel(level.ValueString())
		body.LoggerLevel = &cast
	}

	return body, nil
}

// parseJSON decodes an attribute into a generic JSON value for an API payload.
func parseJSON(value string) (any, error) {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

// decodeAirflowConfig decodes the airflow_config attribute into the API shape.
func decodeAirflowConfig(value string) (*ometa.AirflowConfig, error) {
	airflow := &ometa.AirflowConfig{}
	if strings.TrimSpace(value) == "" {
		return airflow, nil
	}
	if err := json.Unmarshal([]byte(value), airflow); err != nil {
		return nil, err
	}
	return airflow, nil
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
