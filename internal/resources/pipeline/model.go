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

// pipelineModel is the Terraform representation of an OpenMetadata pipeline.
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

// apply copies the API representation into the Terraform model. The parent
// service selector is left untouched when the server omits the reference, and
// tasks already in state are kept when the API returns none.
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
	if pipeline.State != nil {
		model.State = types.StringValue(string(*pipeline.State))
	} else if model.State.IsUnknown() {
		model.State = types.StringNull()
	}
	model.Version = convert.Float64Value(pipeline.Version)
	model.UpdatedAt = convert.TimestampValue(pipeline.UpdatedAt)
	model.UpdatedBy = convert.StringValue(pipeline.UpdatedBy)

	if pipeline.Tasks != nil {
		tasks, diags := tasksValue(ctx, *pipeline.Tasks)
		model.Tasks = tasks
		return diags
	}
	if model.Tasks.IsUnknown() {
		model.Tasks = types.ListNull(pipelineTaskObjectType)
	}
	return nil
}

// tasksValue maps API tasks to a Terraform list in the server's order.
func tasksValue(ctx context.Context, tasks []ometa.Task) (types.List, diag.Diagnostics) {
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
	return types.ListValueFrom(ctx, pipelineTaskObjectType, models)
}

// downstreamTasksValue maps an optional slice of task names to a list. An
// absent slice becomes null so an omitted configuration does not cause a diff.
func downstreamTasksValue(ctx context.Context, values *[]string) types.List {
	if values == nil || len(*values) == 0 {
		return types.ListNull(types.StringType)
	}
	list, _ := convert.StringListValue(ctx, *values)
	return list
}

// tasksToAPI maps planned tasks to the SDK shape.
func tasksToAPI(ctx context.Context, value types.List) ([]ometa.Task, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	var models []pipelineTaskModel
	diags := value.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}

	tasks := make([]ometa.Task, 0, len(models))
	for _, model := range models {
		task := ometa.Task{
			Name:        model.Name.ValueString(),
			DisplayName: convert.StringPointer(model.DisplayName),
			Description: convert.StringPointer(model.Description),
			TaskType:    convert.StringPointer(model.TaskType),
			TaskSQL:     convert.StringPointer(model.TaskSQL),
			SourceUrl:   convert.StringPointer(model.SourceURL),
			StartDate:   convert.StringPointer(model.StartDate),
			EndDate:     convert.StringPointer(model.EndDate),
		}
		downstream, downstreamDiags := convert.StringList(ctx, model.DownstreamTasks)
		diags.Append(downstreamDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if downstream != nil {
			task.DownstreamTasks = &downstream
		}
		tasks = append(tasks, task)
	}
	return tasks, diags
}

// int32Pointer maps an optional Terraform number to the SDK's int32 pointer.
// Null and unknown values are omitted from the request.
func int32Pointer(value types.Int64) *int32 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := int32(value.ValueInt64())
	return &v
}

// int64Pointer maps an optional Terraform number to an int64 pointer. Null and
// unknown values are omitted from the request.
func int64Pointer(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueInt64()
	return &v
}
