package pipeline

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/open-metadata/openmetadata-sdk/openmetadata-go-client/pkg/ometa"

	"github.com/caberdo/terraform-provider-openmetadata/internal/convert"
)

// pipelineTaskObjectType describes one pipeline task. The resource and the data
// source share this shape so a task read from the API can be written straight
// back into state.
var pipelineTaskObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"name":             types.StringType,
	"display_name":     types.StringType,
	"description":      types.StringType,
	"task_type":        types.StringType,
	"task_sql":         types.StringType,
	"source_url":       types.StringType,
	"start_date":       types.StringType,
	"end_date":         types.StringType,
	"downstream_tasks": types.ListType{ElemType: types.StringType},
}}

// pipelineTaskModel is the Terraform representation of one OpenMetadata task.
type pipelineTaskModel struct {
	Name            types.String `tfsdk:"name"`
	DisplayName     types.String `tfsdk:"display_name"`
	Description     types.String `tfsdk:"description"`
	TaskType        types.String `tfsdk:"task_type"`
	TaskSQL         types.String `tfsdk:"task_sql"`
	SourceURL       types.String `tfsdk:"source_url"`
	StartDate       types.String `tfsdk:"start_date"`
	EndDate         types.String `tfsdk:"end_date"`
	DownstreamTasks types.List   `tfsdk:"downstream_tasks"`
}

// pipelineModel is the Terraform representation of a pipeline lookup. The id
// and fully_qualified_name attributes are both selectors and results.
type pipelineModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Service            types.String  `tfsdk:"service"`
	FullyQualifiedName types.String  `tfsdk:"fully_qualified_name"`
	DisplayName        types.String  `tfsdk:"display_name"`
	Description        types.String  `tfsdk:"description"`
	Concurrency        types.Int64   `tfsdk:"concurrency"`
	PipelineLocation   types.String  `tfsdk:"pipeline_location"`
	ScheduleInterval   types.String  `tfsdk:"schedule_interval"`
	SourceURL          types.String  `tfsdk:"source_url"`
	StartDate          types.Int64   `tfsdk:"start_date"`
	State              types.String  `tfsdk:"state"`
	Tasks              types.List    `tfsdk:"tasks"`
	Version            types.Float64 `tfsdk:"version"`
	UpdatedAt          types.String  `tfsdk:"updated_at"`
	UpdatedBy          types.String  `tfsdk:"updated_by"`
}

// apply copies the API representation into the Terraform model.
func apply(ctx context.Context, pipeline *ometa.Pipeline, model *pipelineModel) diag.Diagnostics {
	model.ID = types.StringValue(pipeline.Id.String())
	model.Name = types.StringValue(pipeline.Name)
	model.FullyQualifiedName = convert.StringValue(pipeline.FullyQualifiedName)
	model.DisplayName = convert.StringValue(pipeline.DisplayName)
	model.Description = convert.StringValueOr(pipeline.Description, "")
	if fqn := pipeline.Service.FullyQualifiedName; fqn != nil && *fqn != "" {
		model.Service = types.StringValue(*fqn)
	} else if name := pipeline.Service.Name; name != nil && *name != "" {
		model.Service = types.StringValue(*name)
	}
	model.Concurrency = convert.Int32Value(pipeline.Concurrency)
	model.PipelineLocation = convert.StringValue(pipeline.PipelineLocation)
	model.ScheduleInterval = convert.StringValue(pipeline.ScheduleInterval)
	model.SourceURL = convert.StringValue(pipeline.SourceUrl)
	model.StartDate = convert.Int64Value(pipeline.StartDate)
	model.State = convert.StringValue(pipelineState(pipeline.State))
	model.Version = convert.Float64Value(pipeline.Version)
	model.UpdatedAt = convert.TimestampValue(pipeline.UpdatedAt)
	model.UpdatedBy = convert.StringValue(pipeline.UpdatedBy)

	if pipeline.Tasks != nil {
		return tasksValue(ctx, *pipeline.Tasks, &model.Tasks)
	}
	if model.Tasks.IsUnknown() {
		model.Tasks = types.ListNull(pipelineTaskObjectType)
	}
	return nil
}

// pipelineState maps the SDK enum to a string pointer for convert.
func pipelineState(state *ometa.PipelineState) *string {
	if state == nil {
		return nil
	}
	value := string(*state)
	return &value
}

// tasksValue maps API tasks to a Terraform list in the server's order.
func tasksValue(ctx context.Context, tasks []ometa.Task, target *types.List) diag.Diagnostics {
	models := make([]pipelineTaskModel, 0, len(tasks))
	for _, task := range tasks {
		models = append(models, pipelineTaskModel{
			Name:            types.StringValue(task.Name),
			DisplayName:     convert.StringValue(task.DisplayName),
			Description:     convert.StringValue(task.Description),
			TaskType:        convert.StringValue(task.TaskType),
			TaskSQL:         convert.StringValue(task.TaskSQL),
			SourceURL:       convert.StringValue(task.SourceUrl),
			StartDate:       convert.StringValue(task.StartDate),
			EndDate:         convert.StringValue(task.EndDate),
			DownstreamTasks: downstreamTasksValue(ctx, task.DownstreamTasks),
		})
	}
	list, diags := types.ListValueFrom(ctx, pipelineTaskObjectType, models)
	*target = list
	return diags
}

// downstreamTasksValue maps an optional slice of task names to a list. An
// absent slice becomes null so it matches an omitted configuration.
func downstreamTasksValue(ctx context.Context, values *[]string) types.List {
	if values == nil || len(*values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, _ := convert.StringListValue(ctx, *values)
	return list
}
