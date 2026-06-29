package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var vmBootstrapCommandObjFields = []objField{{TF: "command", API: "command", Kind: "string"}, {TF: "success_return_code", API: "successReturnCode", Kind: "int64"}, {TF: "timeout_seconds", API: "timeoutSeconds", Kind: "int64"}}

type VmSpecModel struct {
	BootstrapCommand                     types.Object `tfsdk:"bootstrap_command"`
	FloatingIpId                         types.String `tfsdk:"floating_ip_id"`
	ImageBootVolumeDeviceIndex           types.Int64  `tfsdk:"image_boot_volume_device_index"`
	ImageId                              types.String `tfsdk:"image_id"`
	ImageScheduleIds                     types.List   `tfsdk:"image_schedule_ids"`
	OfferId                              types.String `tfsdk:"offer_id"`
	OnOffMaintenanceActionIds            types.List   `tfsdk:"on_off_maintenance_action_ids"`
	OsType                               types.String `tfsdk:"os_type"`
	RecurrentCommandMaintenanceActionIds types.List   `tfsdk:"recurrent_command_maintenance_action_ids"`
	SecurityGroupIds                     types.List   `tfsdk:"security_group_ids"`
	SshKeyIds                            types.List   `tfsdk:"ssh_key_ids"`
	VmState                              types.String `tfsdk:"vm_state"`
	VpcSubnetId                          types.String `tfsdk:"vpc_subnet_id"`
}

type VmResourceModel struct {
	ID       types.String  `tfsdk:"id"`
	Metadata metadataModel `tfsdk:"metadata"`
	Spec     VmSpecModel   `tfsdk:"spec"`
	Status   types.Object  `tfsdk:"status"`
}

type VmResource struct{ client *client.Client }

func NewVmResource() resource.Resource { return &VmResource{} }

func (r *VmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func VmResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"bootstrap_command":              objResourceSchema(vmBootstrapCommandObjFields),
		"floating_ip_id":                 schema.StringAttribute{Optional: true},
		"image_boot_volume_device_index": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"image_id":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"image_schedule_ids":             schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"offer_id":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"on_off_maintenance_action_ids":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"os_type":                        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"recurrent_command_maintenance_action_ids": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"security_group_ids":                       schema.ListAttribute{Optional: true, ElementType: types.StringType},
		"ssh_key_ids":                              schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"vm_state":                                 schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vpc_subnet_id":                            schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"private_ipv4": schema.StringAttribute{Computed: true}, "private_ipv6": schema.StringAttribute{Computed: true}, "public_ipv4": schema.StringAttribute{Computed: true}, "public_ipv6": schema.StringAttribute{Computed: true}, "windows_administrator_password": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}
}

func (r *VmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: VmResourceSchemaAttrs()}
}

func (r *VmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	r.client = pd.Client
}

func buildVmRequestMap(ctx context.Context, plan VmResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.BootstrapCommand.IsNull() && !plan.Spec.BootstrapCommand.IsUnknown() {
		spec["bootstrapCommand"] = objToAPI(plan.Spec.BootstrapCommand, vmBootstrapCommandObjFields)
	}
	if !plan.Spec.FloatingIpId.IsNull() && !plan.Spec.FloatingIpId.IsUnknown() {
		spec["floatingIpId"] = plan.Spec.FloatingIpId.ValueString()
	}
	if !plan.Spec.ImageBootVolumeDeviceIndex.IsNull() && !plan.Spec.ImageBootVolumeDeviceIndex.IsUnknown() {
		spec["imageBootVolumeDeviceIndex"] = plan.Spec.ImageBootVolumeDeviceIndex.ValueInt64()
	}
	if !plan.Spec.ImageId.IsNull() && !plan.Spec.ImageId.IsUnknown() {
		spec["imageId"] = plan.Spec.ImageId.ValueString()
	}
	if !plan.Spec.ImageScheduleIds.IsNull() && !plan.Spec.ImageScheduleIds.IsUnknown() {
		spec["imageScheduleIds"] = stringListToInterface(ctx, plan.Spec.ImageScheduleIds)
	}
	if !plan.Spec.OfferId.IsNull() && !plan.Spec.OfferId.IsUnknown() {
		spec["offerId"] = plan.Spec.OfferId.ValueString()
	}
	if !plan.Spec.OnOffMaintenanceActionIds.IsNull() && !plan.Spec.OnOffMaintenanceActionIds.IsUnknown() {
		spec["onOffMaintenanceActionIds"] = stringListToInterface(ctx, plan.Spec.OnOffMaintenanceActionIds)
	}
	if !plan.Spec.OsType.IsNull() && !plan.Spec.OsType.IsUnknown() {
		spec["osType"] = plan.Spec.OsType.ValueString()
	}
	if !plan.Spec.RecurrentCommandMaintenanceActionIds.IsNull() && !plan.Spec.RecurrentCommandMaintenanceActionIds.IsUnknown() {
		spec["recurrentCommandMaintenanceActionIds"] = stringListToInterface(ctx, plan.Spec.RecurrentCommandMaintenanceActionIds)
	}
	if !plan.Spec.SecurityGroupIds.IsNull() && !plan.Spec.SecurityGroupIds.IsUnknown() {
		spec["securityGroupIds"] = stringListToInterface(ctx, plan.Spec.SecurityGroupIds)
	}
	if !plan.Spec.SshKeyIds.IsNull() && !plan.Spec.SshKeyIds.IsUnknown() {
		spec["sshKeyIds"] = stringListToInterface(ctx, plan.Spec.SshKeyIds)
	}
	if !plan.Spec.VmState.IsNull() && !plan.Spec.VmState.IsUnknown() {
		spec["vmState"] = plan.Spec.VmState.ValueString()
	}
	if !plan.Spec.VpcSubnetId.IsNull() && !plan.Spec.VpcSubnetId.IsUnknown() {
		spec["vpcSubnetId"] = plan.Spec.VpcSubnetId.ValueString()
	}
	return m
}

func populateVmState(ctx context.Context, data map[string]interface{}, state *VmResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
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
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"private_ipv4":                   types.StringType,
			"private_ipv6":                   types.StringType,
			"public_ipv4":                    types.StringType,
			"public_ipv6":                    types.StringType,
			"windows_administrator_password": types.StringType,
		},
		map[string]attr.Value{
			"private_ipv4":                   getStringFromInfo(data, "privateIpv4"),
			"private_ipv6":                   getStringFromInfo(data, "privateIpv6"),
			"public_ipv4":                    getStringFromInfo(data, "publicIpv4"),
			"public_ipv6":                    getStringFromInfo(data, "publicIpv6"),
			"windows_administrator_password": getStringFromInfo(data, "windowsAdministratorPassword"),
		})
	return nil
}

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildVmRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/vm", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/vm", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/vm", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateVmState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *VmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/vm", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateVmState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *VmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildVmRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/vm", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/vm", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/vm", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateVmState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *VmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VmResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/vm", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/vm", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *VmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state VmResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/vm", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateVmState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
