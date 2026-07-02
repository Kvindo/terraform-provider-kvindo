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

type VmDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Metadata *metadataModel `tfsdk:"metadata"`
	Spec     *VmSpecModel   `tfsdk:"spec"`
	Status   types.Object   `tfsdk:"status"`
}

type VmDataSource struct{ client *client.Client }

func NewVmDataSource() datasource.DataSource { return &VmDataSource{} }

func (d *VmDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (d *VmDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"bootstrap_command":              objDatasourceSchema(vmBootstrapCommandObjFields),
		"floating_ip_id":                 schema.StringAttribute{Computed: true},
		"image_boot_volume_device_index": schema.Int64Attribute{Computed: true},
		"image_id":                       schema.StringAttribute{Computed: true},
		"image_schedule_ids":             schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"offer_id":                       schema.StringAttribute{Computed: true},
		"on_off_maintenance_action_ids":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"os_type":                        schema.StringAttribute{Computed: true},
		"recurrent_command_maintenance_action_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"security_group_ids":                       schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"ssh_key_ids":                              schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"vm_state":                                 schema.StringAttribute{Computed: true},
		"vpc_subnet_id":                            schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"private_ipv4": schema.StringAttribute{Computed: true}, "private_ipv6": schema.StringAttribute{Computed: true}, "public_ipv4": schema.StringAttribute{Computed: true}, "public_ipv6": schema.StringAttribute{Computed: true}, "windows_administrator_password": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}}
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
		apiData, err = d.client.Get(ctx, "/api/v1/vm", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/vm", state.Name.ValueString())
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
	state.Spec = &VmSpecModel{}
	spec := getSpec(apiData)
	state.Spec.BootstrapCommand = objFromAPI(objMap(spec, "bootstrapCommand"), vmBootstrapCommandObjFields)
	state.Spec.FloatingIpId = getString(spec, "floatingIpId")
	state.Spec.ImageBootVolumeDeviceIndex = getInt64(spec, "imageBootVolumeDeviceIndex")
	state.Spec.ImageId = getString(spec, "imageId")
	state.Spec.ImageScheduleIds = getStringList(ctx, spec, "imageScheduleIds")
	state.Spec.OfferId = getString(spec, "offerId")
	state.Spec.OnOffMaintenanceActionIds = getStringList(ctx, spec, "onOffMaintenanceActionIds")
	state.Spec.OsType = getString(spec, "osType")
	state.Spec.RecurrentCommandMaintenanceActionIds = getStringList(ctx, spec, "recurrentCommandMaintenanceActionIds")
	state.Spec.SecurityGroupIds = getStringList(ctx, spec, "securityGroupIds")
	state.Spec.SshKeyIds = getStringList(ctx, spec, "sshKeyIds")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"private_ipv4":                   types.StringType,
			"private_ipv6":                   types.StringType,
			"public_ipv4":                    types.StringType,
			"public_ipv6":                    types.StringType,
			"windows_administrator_password": types.StringType,
		},
		map[string]attr.Value{
			"private_ipv4":                   getStringFromInfo(apiData, "privateIpv4"),
			"private_ipv6":                   getStringFromInfo(apiData, "privateIpv6"),
			"public_ipv4":                    getStringFromInfo(apiData, "publicIpv4"),
			"public_ipv6":                    getStringFromInfo(apiData, "publicIpv6"),
			"windows_administrator_password": getStringFromInfo(apiData, "windowsAdministratorPassword"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
