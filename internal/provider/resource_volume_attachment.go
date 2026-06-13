package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// VolumeAttachmentResourceModel describes the resource data model.
type VolumeAttachmentResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VolumeId types.String `tfsdk:"volume_id"`
	VmId types.String `tfsdk:"vm_id"`
	VmDeviceIndex types.Int64 `tfsdk:"vm_device_index"`
	Info types.Object `tfsdk:"info"`
}

// VolumeAttachmentResource defines the resource implementation.
type VolumeAttachmentResource struct {
	client *client.Client
}

func NewVolumeAttachmentResource() resource.Resource {
	return &VolumeAttachmentResource{}
}

func (r *VolumeAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_attachment"
}

func (r *VolumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["volume_id"] = schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vm_id"] = schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vm_device_index"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *VolumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildVolumeAttachmentRequestMap(ctx context.Context, plan VolumeAttachmentResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.VolumeId.IsNull() && !plan.VolumeId.IsUnknown() {
		m["volumeId"] = plan.VolumeId.ValueString()
	}
	if !plan.VmId.IsNull() && !plan.VmId.IsUnknown() {
		m["vmId"] = plan.VmId.ValueString()
	}
	if !plan.VmDeviceIndex.IsNull() && !plan.VmDeviceIndex.IsUnknown() {
		m["vmDeviceIndex"] = plan.VmDeviceIndex.ValueInt64()
	}
	return m
}

func populateVolumeAttachmentState(ctx context.Context, data map[string]interface{}, state *VolumeAttachmentResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.VolumeId = getString(data, "volumeId")
	state.VmId = getString(data, "vmId")
	state.VmDeviceIndex = getInt64(data, "vmDeviceIndex")
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *VolumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VolumeAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildVolumeAttachmentRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/volume-attachment", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/volume-attachment", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/volume-attachment", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateVolumeAttachmentState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VolumeAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/volume-attachment", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateVolumeAttachmentState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VolumeAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state VolumeAttachmentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildVolumeAttachmentRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/volume-attachment", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/volume-attachment", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/volume-attachment", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateVolumeAttachmentState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VolumeAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/volume-attachment", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/volume-attachment", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *VolumeAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state VolumeAttachmentResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/volume-attachment", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateVolumeAttachmentState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
