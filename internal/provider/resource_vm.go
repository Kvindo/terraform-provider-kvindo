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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types
var _ = listplanmodifier.UseStateForUnknown

// VmResourceModel describes the resource data model.
type VmResourceModel struct {
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
	Info types.Object `tfsdk:"info"`
}

// VmBootstrapCommandModel is the nested object model for bootstrap_command.
type VmBootstrapCommandModel struct {
	Command types.String `tfsdk:"command"`
	SuccessReturnCode types.Int64 `tfsdk:"success_return_code"`
	TimeoutSeconds types.Int64 `tfsdk:"timeout_seconds"`
}

// VmResource defines the resource implementation.
type VmResource struct {
	client *client.Client
}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

func (r *VmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["vm_state"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vpc_subnet_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["floating_ip_id"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["image_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["offer_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["image_boot_volume_device_index"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["ssh_key_ids"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["image_schedule_ids"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["recurrent_command_maintenance_action_ids"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["on_off_maintenance_action_ids"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["bootstrap_command"] = schema.ListNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"command": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
					"success_return_code": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
					"timeout_seconds": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
				},
			},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "private_ipv4": schema.StringAttribute{Computed: true}, "public_ipv4": schema.StringAttribute{Computed: true}, "private_ipv6": schema.StringAttribute{Computed: true}, "public_ipv6": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.VmState.IsNull() && !plan.VmState.IsUnknown() {
		m["vmState"] = plan.VmState.ValueString()
	}
	if !plan.VpcSubnetId.IsNull() && !plan.VpcSubnetId.IsUnknown() {
		m["vpcSubnetId"] = plan.VpcSubnetId.ValueString()
	}
	if !plan.FloatingIpId.IsNull() && !plan.FloatingIpId.IsUnknown() {
		m["floatingIpId"] = plan.FloatingIpId.ValueString()
	}
	if !plan.ImageId.IsNull() && !plan.ImageId.IsUnknown() {
		m["imageId"] = plan.ImageId.ValueString()
	}
	if !plan.OfferId.IsNull() && !plan.OfferId.IsUnknown() {
		m["offerId"] = plan.OfferId.ValueString()
	}
	if !plan.ImageBootVolumeDeviceIndex.IsNull() && !plan.ImageBootVolumeDeviceIndex.IsUnknown() {
		m["imageBootVolumeDeviceIndex"] = plan.ImageBootVolumeDeviceIndex.ValueInt64()
	}
	if !plan.SshKeyIds.IsNull() && !plan.SshKeyIds.IsUnknown() {
		m["sshKeyIds"] = stringListToInterface(ctx, plan.SshKeyIds)
	}
	if !plan.ImageScheduleIds.IsNull() && !plan.ImageScheduleIds.IsUnknown() {
		m["imageScheduleIds"] = stringListToInterface(ctx, plan.ImageScheduleIds)
	}
	if !plan.RecurrentCommandMaintenanceActionIds.IsNull() && !plan.RecurrentCommandMaintenanceActionIds.IsUnknown() {
		m["recurrentCommandMaintenanceActionIds"] = stringListToInterface(ctx, plan.RecurrentCommandMaintenanceActionIds)
	}
	if !plan.OnOffMaintenanceActionIds.IsNull() && !plan.OnOffMaintenanceActionIds.IsUnknown() {
		m["onOffMaintenanceActionIds"] = stringListToInterface(ctx, plan.OnOffMaintenanceActionIds)
	}
	if !plan.BootstrapCommand.IsNull() && !plan.BootstrapCommand.IsUnknown() {
		var items []map[string]interface{}
		for _, elem := range plan.BootstrapCommand.Elements() {
			if ov, ok := elem.(types.Object); ok {
				item := map[string]interface{}{}
				if v, ok := ov.Attributes()["command"]; ok {
					if sv, ok := v.(types.String); ok && !sv.IsNull() {
						item["command"] = sv.ValueString()
					}
				}
				if v, ok := ov.Attributes()["success_return_code"]; ok {
					if iv, ok := v.(types.Int64); ok && !iv.IsNull() {
						item["successReturnCode"] = iv.ValueInt64()
					}
				}
				if v, ok := ov.Attributes()["timeout_seconds"]; ok {
					if iv, ok := v.(types.Int64); ok && !iv.IsNull() {
						item["timeoutSeconds"] = iv.ValueInt64()
					}
				}
				items = append(items, item)
			}
		}
		m["bootstrapCommand"] = items
	}
	return m
}

func populateVmState(ctx context.Context, data map[string]interface{}, state *VmResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.VmState = getString(data, "vmState")
	state.VpcSubnetId = getString(data, "vpcSubnetId")
	state.FloatingIpId = getString(data, "floatingIpId")
	state.ImageId = getString(data, "imageId")
	state.OfferId = getString(data, "offerId")
	state.ImageBootVolumeDeviceIndex = getInt64(data, "imageBootVolumeDeviceIndex")
	state.SshKeyIds = getStringList(ctx, data, "sshKeyIds")
	state.ImageScheduleIds = getStringList(ctx, data, "imageScheduleIds")
	state.RecurrentCommandMaintenanceActionIds = getStringList(ctx, data, "recurrentCommandMaintenanceActionIds")
	state.OnOffMaintenanceActionIds = getStringList(ctx, data, "onOffMaintenanceActionIds")
	{
		rawBootstrapCommand, _ := data["bootstrapCommand"].([]interface{})
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
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "private_ipv4": types.StringType, "public_ipv4": types.StringType, "private_ipv6": types.StringType, "public_ipv6": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "private_ipv4": getStringFromInfo(data, "privateipv4"), "public_ipv4": getStringFromInfo(data, "publicipv4"), "private_ipv6": getStringFromInfo(data, "privateipv6"), "public_ipv6": getStringFromInfo(data, "publicipv6")})
	return nil
}

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VmResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state VmResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateVmState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VmResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	// Import by ID
	var state VmResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/vm", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateVmState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
