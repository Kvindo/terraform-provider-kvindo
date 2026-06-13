package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// ImageScheduleDataSourceModel describes the data source data model.
type ImageScheduleDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Enabled types.Bool `tfsdk:"enabled"`
	ScheduleFormat types.String `tfsdk:"schedule_format"`
	Schedule types.String `tfsdk:"schedule"`
	RetentionCount types.Int64 `tfsdk:"retention_count"`
	InfoState types.String `tfsdk:"info_state"`
}

type ImageScheduleDataSource struct {
	client *client.Client
}

func NewImageScheduleDataSource() datasource.DataSource {
	return &ImageScheduleDataSource{}
}

func (d *ImageScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_schedule"
}

func (d *ImageScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["enabled"] = schema.BoolAttribute{Computed: true}
	attrs["schedule_format"] = schema.StringAttribute{Computed: true}
	attrs["schedule"] = schema.StringAttribute{Computed: true}
	attrs["retention_count"] = schema.Int64Attribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *ImageScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	d.client = pd.Client
}

func (d *ImageScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ImageScheduleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/image-schedule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	if err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	state.Enabled = getBool(apiData, "enabled")
	state.ScheduleFormat = getString(apiData, "scheduleFormat")
	state.Schedule = getString(apiData, "schedule")
	state.RetentionCount = getInt64(apiData, "retentionCount")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
