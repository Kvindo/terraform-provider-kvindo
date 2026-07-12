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

type VmCommandScheduleDataSourceModel struct {
	ID       types.String                `tfsdk:"id"`
	Name     types.String                `tfsdk:"name"`
	Metadata *metadataModel              `tfsdk:"metadata"`
	Spec     *VmCommandScheduleSpecModel `tfsdk:"spec"`
	Status   types.Object                `tfsdk:"status"`
}

type VmCommandScheduleDataSource struct{ client *client.Client }

func NewVmCommandScheduleDataSource() datasource.DataSource { return &VmCommandScheduleDataSource{} }

func (d *VmCommandScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_command_schedule"
}

func (d *VmCommandScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"command":                 schema.StringAttribute{Computed: true},
		"command_timeout_seconds": schema.Int64Attribute{Computed: true},
		"enabled":                 schema.BoolAttribute{Computed: true},
		"schedule":                schema.StringAttribute{Computed: true},
		"schedule_format":         schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *VmCommandScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VmCommandScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VmCommandScheduleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set")
		return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/vm-command-schedule", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/vm-command-schedule", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &VmCommandScheduleSpecModel{}
	spec := getSpec(apiData)
	state.Spec.Command = getString(spec, "command")
	state.Spec.CommandTimeoutSeconds = getInt64(spec, "commandTimeoutSeconds")
	state.Spec.Enabled = getBool(spec, "enabled")
	state.Spec.Schedule = getString(spec, "schedule")
	state.Spec.ScheduleFormat = getString(spec, "scheduleFormat")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
