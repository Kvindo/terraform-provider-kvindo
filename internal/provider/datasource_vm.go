package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// VmDataSourceModel describes the data source data model.
type VmDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VmState types.String `tfsdk:"vm_state"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	ImageId types.String `tfsdk:"image_id"`
	OfferId types.String `tfsdk:"offer_id"`
	ImageBootVolumeDeviceIndex types.Int64 `tfsdk:"image_boot_volume_device_index"`
	SshKeyIds types.List `tfsdk:"ssh_key_ids"`
	ImageScheduleIds types.List `tfsdk:"image_schedule_ids"`
	RecurrentCommandMaintenanceActionIds types.List `tfsdk:"recurrent_command_maintenance_action_ids"`
	OnOffMaintenanceActionIds types.List `tfsdk:"on_off_maintenance_action_ids"`
	BootstrapCommand types.List `tfsdk:"bootstrap_command"`
	InfoState types.String `tfsdk:"info_state"`
	InfoPrivateIpv4 types.String `tfsdk:"info_private_ipv4"`
	InfoPublicIpv4 types.String `tfsdk:"info_public_ipv4"`
	InfoPrivateIpv6 types.String `tfsdk:"info_private_ipv6"`
	InfoPublicIpv6 types.String `tfsdk:"info_public_ipv6"`
}

type VmDataSource struct {
	client *client.Client
}

func NewVmDataSource() datasource.DataSource {
	return &VmDataSource{}
}

func (d *VmDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (d *VmDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["vm_state"] = schema.StringAttribute{Computed: true}
	attrs["vpc_subnet_id"] = schema.StringAttribute{Computed: true}
	attrs["floating_ip_id"] = schema.StringAttribute{Computed: true}
	attrs["image_id"] = schema.StringAttribute{Computed: true}
	attrs["offer_id"] = schema.StringAttribute{Computed: true}
	attrs["image_boot_volume_device_index"] = schema.Int64Attribute{Computed: true}
	attrs["ssh_key_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["image_schedule_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["recurrent_command_maintenance_action_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["on_off_maintenance_action_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["bootstrap_command"] = schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"command": schema.StringAttribute{Computed: true},
					"success_return_code": schema.Int64Attribute{Computed: true},
					"timeout_seconds": schema.Int64Attribute{Computed: true},
				},
			},
		}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ipv4"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ipv4"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ipv6"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ipv6"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *VmDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VmDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VmDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/vm", state.ID.ValueString())
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
	state.VmState = getString(apiData, "vmState")
	state.VpcSubnetId = getString(apiData, "vpcSubnetId")
	state.FloatingIpId = getString(apiData, "floatingIpId")
	state.ImageId = getString(apiData, "imageId")
	state.OfferId = getString(apiData, "offerId")
	state.ImageBootVolumeDeviceIndex = getInt64(apiData, "imageBootVolumeDeviceIndex")
	state.SshKeyIds = getStringList(ctx, apiData, "sshKeyIds")
	state.ImageScheduleIds = getStringList(ctx, apiData, "imageScheduleIds")
	state.RecurrentCommandMaintenanceActionIds = getStringList(ctx, apiData, "recurrentCommandMaintenanceActionIds")
	state.OnOffMaintenanceActionIds = getStringList(ctx, apiData, "onOffMaintenanceActionIds")
	{
		rawBootstrapCommand, _ := apiData["bootstrapCommand"].([]interface{})
		attrTypes := map[string]attr.Type{
			"command": types.StringType,
			"success_return_code": types.Int64Type,
			"timeout_seconds": types.Int64Type,
		}
		objs := make([]attr.Value, 0, len(rawBootstrapCommand))
		for _, item := range rawBootstrapCommand {
			if m, ok := item.(map[string]interface{}); ok {
				attrs := map[string]attr.Value{
					"command": getString(m, "command"),
					"success_return_code": getInt64(m, "successReturnCode"),
					"timeout_seconds": getInt64(m, "timeoutSeconds"),
				}
				obj, _ := types.ObjectValue(attrTypes, attrs)
				objs = append(objs, obj)
			}
		}
		state.BootstrapCommand, _ = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objs)
	}
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoPrivateIpv4 = getStringFromInfo(apiData, "privateipv4")
	state.InfoPublicIpv4 = getStringFromInfo(apiData, "publicipv4")
	state.InfoPrivateIpv6 = getStringFromInfo(apiData, "privateipv6")
	state.InfoPublicIpv6 = getStringFromInfo(apiData, "publicipv6")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
