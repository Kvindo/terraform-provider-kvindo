package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var vmBootstrapCommandObjFields = []objField{{TF: "command", API: "command", Kind: "string"}}

// vmBootstrapCommandInfoAttrTypes/buildVmBootstrapCommandInfoObj mirror buildUserInfoObj/buildPricingObj
// in resource_common.go: a nested status object, null when the VM has no bootstrap_command set (the
// C# API omits status.bootstrapCommand entirely in that case rather than sending null-valued fields).
var vmBootstrapCommandInfoAttrTypes = map[string]attr.Type{
	"return_code": types.Int64Type, "output": types.StringType, "duration_ms": types.Int64Type,
}

func buildVmBootstrapCommandInfoObj(data map[string]interface{}) types.Object {
	raw, ok := infoFieldRaw(data, "bootstrapCommand")
	m, mapOk := raw.(map[string]interface{})
	if !ok || !mapOk {
		return types.ObjectNull(vmBootstrapCommandInfoAttrTypes)
	}
	obj, _ := types.ObjectValue(vmBootstrapCommandInfoAttrTypes, map[string]attr.Value{
		"return_code": getInt64(m, "returnCode"),
		"output":      getString(m, "output"),
		"duration_ms": getInt64(m, "durationMs"),
	})
	return obj
}

// boot_volume_attachment has no backend/swagger counterpart: the Kvindo API has no such field on
// /api/v1/vm. It is a purely Terraform-side convenience that creates a kvindo_volume_attachment
// behind the scenes (see VmResource.Create/Delete) so a running VM + its boot volume can be
// expressed in one apply, without the vm_id-references-itself dependency cycle a hand-written
// kvindo_volume_attachment resource would otherwise create. Never wired into
// buildVmRequestMap/populateVmState. attachment_id is our own bookkeeping (the attachment we
// create), never user-set.
var bootVolumeAttachmentAttrTypes = map[string]attr.Type{"volume_id": types.StringType, "attachment_id": types.StringType}

type VmSpecModel struct {
	BootstrapCommand           types.Object `tfsdk:"bootstrap_command"`
	BootVolumeAttachment       types.Object `tfsdk:"boot_volume_attachment"`
	FloatingIpId               types.String `tfsdk:"floating_ip_id"`
	ImageBootVolumeDeviceIndex types.Int64  `tfsdk:"image_boot_volume_device_index"`
	ImageId                    types.String `tfsdk:"image_id"`
	ImageScheduleIds           types.List   `tfsdk:"image_schedule_ids"`
	OfferId                    types.String `tfsdk:"offer_id"`
	OnOffScheduleIds           types.List   `tfsdk:"on_off_schedule_ids"`
	OsType                     types.String `tfsdk:"os_type"`
	CommandScheduleIds         types.List   `tfsdk:"command_schedule_ids"`
	SecurityGroupIds           types.List   `tfsdk:"security_group_ids"`
	SshKeyIds                  types.List   `tfsdk:"ssh_key_ids"`
	VmState                    types.String `tfsdk:"vm_state"`
	VpcSubnetId                types.String `tfsdk:"vpc_subnet_id"`
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
		"bootstrap_command": objResourceSchema(vmBootstrapCommandObjFields),
		"boot_volume_attachment": schema.SingleNestedAttribute{
			Optional:      true,
			Computed:      true,
			PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			Attributes: map[string]schema.Attribute{
				"volume_id":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
				"attachment_id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			},
		},
		"floating_ip_id":                 schema.StringAttribute{Optional: true},
		"image_boot_volume_device_index": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"image_id":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"image_schedule_ids":             schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"offer_id":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"on_off_schedule_ids":            schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"os_type":                        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"command_schedule_ids":           schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"security_group_ids":             schema.ListAttribute{Optional: true, ElementType: types.StringType},
		"ssh_key_ids":                    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"vm_state":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vpc_subnet_id":                  schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status": commonInfoSchema(map[string]schema.Attribute{
			"bootstrap_command": schema.SingleNestedAttribute{Computed: true, Attributes: map[string]schema.Attribute{
				"return_code": schema.Int64Attribute{Computed: true},
				"output":      schema.StringAttribute{Computed: true},
				"duration_ms": schema.Int64Attribute{Computed: true},
			}},
			"private_ipv4": schema.StringAttribute{Computed: true}, "private_ipv6": schema.StringAttribute{Computed: true}, "public_ipv4": schema.StringAttribute{Computed: true}, "public_ipv6": schema.StringAttribute{Computed: true}, "windows_administrator_password": schema.StringAttribute{Computed: true, Sensitive: true}}),
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
	if !plan.Spec.OnOffScheduleIds.IsNull() && !plan.Spec.OnOffScheduleIds.IsUnknown() {
		spec["onOffScheduleIds"] = stringListToInterface(ctx, plan.Spec.OnOffScheduleIds)
	}
	if !plan.Spec.OsType.IsNull() && !plan.Spec.OsType.IsUnknown() {
		spec["osType"] = plan.Spec.OsType.ValueString()
	}
	if !plan.Spec.CommandScheduleIds.IsNull() && !plan.Spec.CommandScheduleIds.IsUnknown() {
		spec["commandScheduleIds"] = stringListToInterface(ctx, plan.Spec.CommandScheduleIds)
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
	state.Spec.OnOffScheduleIds = getStringList(ctx, spec, "onOffScheduleIds")
	state.Spec.OsType = getString(spec, "osType")
	state.Spec.CommandScheduleIds = getStringList(ctx, spec, "commandScheduleIds")
	state.Spec.SecurityGroupIds = getStringList(ctx, spec, "securityGroupIds")
	state.Spec.SshKeyIds = getStringList(ctx, spec, "sshKeyIds")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"bootstrap_command":              types.ObjectType{AttrTypes: vmBootstrapCommandInfoAttrTypes},
			"private_ipv4":                   types.StringType,
			"private_ipv6":                   types.StringType,
			"public_ipv4":                    types.StringType,
			"public_ipv6":                    types.StringType,
			"windows_administrator_password": types.StringType,
		},
		map[string]attr.Value{
			"bootstrap_command":              buildVmBootstrapCommandInfoObj(data),
			"private_ipv4":                   getStringFromInfo(data, "privateIpv4"),
			"private_ipv6":                   getStringFromInfo(data, "privateIpv6"),
			"public_ipv4":                    getStringFromInfo(data, "publicIpv4"),
			"public_ipv6":                    getStringFromInfo(data, "publicIpv6"),
			"windows_administrator_password": getStringFromInfo(data, "windowsAdministratorPassword"),
		})
	return nil
}

// resolvedVmState returns the vm_state the backend will end up applying, mirroring the default
// ("running") that OrganizationVmResourceChangeRequest.CreateFromResourceAsync applies server-side.
func resolvedVmState(spec VmSpecModel) string {
	if spec.VmState.IsNull() || spec.VmState.IsUnknown() || spec.VmState.ValueString() == "" {
		return "running"
	}
	return spec.VmState.ValueString()
}

// vmCreateRequiresBootVolumeAttachment reports whether Create() must fail fast because the VM
// would be created running with no way for the backend to ever attach a boot volume. Mirrors the
// backend's own create-time gate (OrganizationVmReconciler's Queued check): a from-scratch
// running VM with no boot volume attachment would otherwise just poll forever until our own
// client-side timeout. Only applies to Create — an already-existing VM (Update) may already have
// its boot volume attached via a separately managed kvindo_volume_attachment resource.
func vmCreateRequiresBootVolumeAttachment(spec VmSpecModel) bool {
	hasBootVolumeAttachment := !spec.BootVolumeAttachment.IsNull() && !spec.BootVolumeAttachment.IsUnknown()
	return !hasBootVolumeAttachment && resolvedVmState(spec) == "running"
}

// buildBootVolumeAttachmentPlan constructs the kvindo_volume_attachment created behind the scenes
// by VmResource.Create so a running VM + its boot volume can be expressed in one apply.
func buildBootVolumeAttachmentPlan(vmPlan VmResourceModel, vmId, attachmentId, volumeId string) VolumeAttachmentResourceModel {
	return VolumeAttachmentResourceModel{
		ID: types.StringValue(attachmentId),
		Metadata: metadataModel{
			Name:        types.StringValue(vmPlan.Metadata.Name.ValueString() + "-boot"),
			Description: types.StringNull(),
			FolderID:    vmPlan.Metadata.FolderID,
			Labels:      types.MapNull(types.StringType),
		},
		Spec: VolumeAttachmentSpecModel{
			VmId:          types.StringValue(vmId),
			VolumeId:      types.StringValue(volumeId),
			VmDeviceIndex: types.Int64Value(0),
		},
	}
}

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VmResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasBootVolumeAttachment := !plan.Spec.BootVolumeAttachment.IsNull() && !plan.Spec.BootVolumeAttachment.IsUnknown()
	if vmCreateRequiresBootVolumeAttachment(plan.Spec) {
		resp.Diagnostics.AddError(
			"Missing boot_volume_attachment",
			"spec.boot_volume_attachment.volume_id must be set to create a VM in state \"running\", "+
				"unless it is already attached via a separately managed kvindo_volume_attachment resource.",
		)
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildVmRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/vm", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}

	var attachmentId string
	if hasBootVolumeAttachment {
		bootVolAttrs := plan.Spec.BootVolumeAttachment.Attributes()
		volumeId := bootVolAttrs["volume_id"].(types.String).ValueString()
		attachmentId = newULID()
		// Create the boot volume_attachment behind the scenes, right after the VM's DB row is
		// committed (the PUT above already returned, so it exists) but before polling the VM to
		// done — the VM reconciler's own Queued gate waits for exactly this row instead of
		// failing, so ordering here just needs "soon", not "before".
		attPlan := buildBootVolumeAttachmentPlan(plan, resourceId, attachmentId, volumeId)
		attBody := buildVolumeAttachmentRequestMap(ctx, attPlan)
		if _, err := r.client.Put(ctx, "/api/v1/volume-attachment", attBody); err != nil {
			resp.Diagnostics.AddError("Boot Volume Attachment Create Error", err.Error())
			return
		}
	}

	if err := r.client.PollUntilDone(ctx, "/api/v1/vm", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
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
	if hasBootVolumeAttachment {
		bootVolAttrs := plan.Spec.BootVolumeAttachment.Attributes()
		obj, diags := types.ObjectValue(bootVolumeAttachmentAttrTypes, map[string]attr.Value{
			"volume_id":     bootVolAttrs["volume_id"],
			"attachment_id": types.StringValue(attachmentId),
		})
		resp.Diagnostics.Append(diags...)
		plan.Spec.BootVolumeAttachment = obj
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

	// Clean up the boot volume_attachment created behind the scenes at Create() time, if any.
	// Safe to do after the VM delete: the attachment reconciler's delete path explicitly
	// tolerates the VM already being gone.
	if !state.Spec.BootVolumeAttachment.IsNull() && !state.Spec.BootVolumeAttachment.IsUnknown() {
		if attachmentIdVal, ok := state.Spec.BootVolumeAttachment.Attributes()["attachment_id"].(types.String); ok && !attachmentIdVal.IsNull() && attachmentIdVal.ValueString() != "" {
			attModResp, err := r.client.Delete(ctx, "/api/v1/volume-attachment", attachmentIdVal.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Boot Volume Attachment Delete Error", err.Error())
				return
			}
			if err := r.client.PollUntilDone(ctx, "/api/v1/volume-attachment", attModResp.RequestId); err != nil {
				resp.Diagnostics.AddError("Boot Volume Attachment Delete Poll Error", err.Error())
				return
			}
		}
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
