package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// GitlabResourceModel describes the resource data model.
type GitlabResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	Version types.String `tfsdk:"version"`
	RootPassword types.String `tfsdk:"root_password"`
	VmState types.String `tfsdk:"vm_state"`
	VmOfferId types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib types.Int64 `tfsdk:"volume_size_gib"`
	Edition types.String `tfsdk:"edition"`
	RecordName types.String `tfsdk:"record_name"`
	Info types.Object `tfsdk:"info"`
}

// GitlabResource defines the resource implementation.
type GitlabResource struct {
	client *client.Client
}

func NewGitlabResource() resource.Resource {
	return &GitlabResource{}
}

func (r *GitlabResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab"
}

func (r *GitlabResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["floating_ip_id"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["vpc_subnet_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["version"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["root_password"] = schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
		}
	attrs["vm_state"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vm_offer_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["volume_offer_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["volume_size_gib"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["edition"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["record_name"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "public_ip_v6": schema.StringAttribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "private_ip_v6": schema.StringAttribute{Computed: true}, "fqdn": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *GitlabResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildGitlabRequestMap(ctx context.Context, plan GitlabResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.Tier.IsNull() && !plan.Tier.IsUnknown() {
		m["tier"] = plan.Tier.ValueString()
	}
	if !plan.FloatingIpId.IsNull() && !plan.FloatingIpId.IsUnknown() {
		m["floatingIpId"] = plan.FloatingIpId.ValueString()
	}
	if !plan.VpcSubnetId.IsNull() && !plan.VpcSubnetId.IsUnknown() {
		m["vpcSubnetId"] = plan.VpcSubnetId.ValueString()
	}
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		m["version"] = plan.Version.ValueString()
	}
	if !plan.RootPassword.IsNull() && !plan.RootPassword.IsUnknown() {
		m["rootPassword"] = plan.RootPassword.ValueString()
	}
	if !plan.VmState.IsNull() && !plan.VmState.IsUnknown() {
		m["vmState"] = plan.VmState.ValueString()
	}
	if !plan.VmOfferId.IsNull() && !plan.VmOfferId.IsUnknown() {
		m["vmOfferId"] = plan.VmOfferId.ValueString()
	}
	if !plan.VolumeOfferId.IsNull() && !plan.VolumeOfferId.IsUnknown() {
		m["volumeOfferId"] = plan.VolumeOfferId.ValueString()
	}
	if !plan.VolumeSizeGib.IsNull() && !plan.VolumeSizeGib.IsUnknown() {
		m["volumeSizeGiB"] = plan.VolumeSizeGib.ValueInt64()
	}
	if !plan.Edition.IsNull() && !plan.Edition.IsUnknown() {
		m["edition"] = plan.Edition.ValueString()
	}
	if !plan.RecordName.IsNull() && !plan.RecordName.IsUnknown() {
		m["recordName"] = plan.RecordName.ValueString()
	}
	return m
}

func populateGitlabState(ctx context.Context, data map[string]interface{}, state *GitlabResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.Tier = getString(data, "tier")
	state.FloatingIpId = getString(data, "floatingIpId")
	state.VpcSubnetId = getString(data, "vpcSubnetId")
	state.Version = getString(data, "version")
	state.VmState = getString(data, "vmState")
	state.VmOfferId = getString(data, "vmOfferId")
	state.VolumeOfferId = getString(data, "volumeOfferId")
	state.VolumeSizeGib = getInt64(data, "volumeSizeGiB")
	state.Edition = getString(data, "edition")
	state.RecordName = getString(data, "recordName")
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "public_ip_v4": types.StringType, "public_ip_v6": types.StringType, "private_ip_v4": types.StringType, "private_ip_v6": types.StringType, "fqdn": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "public_ip_v4": getStringFromInfo(data, "publicipv4"), "public_ip_v6": getStringFromInfo(data, "publicipv6"), "private_ip_v4": getStringFromInfo(data, "privateipv4"), "private_ip_v6": getStringFromInfo(data, "privateipv6"), "fqdn": getStringFromInfo(data, "fqdn")})
	return nil
}

func (r *GitlabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GitlabResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildGitlabRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/gitlab", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/gitlab", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/gitlab", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateGitlabState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GitlabResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/gitlab", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateGitlabState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GitlabResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state GitlabResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildGitlabRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/gitlab", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/gitlab", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/gitlab", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateGitlabState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GitlabResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/gitlab", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/gitlab", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *GitlabResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state GitlabResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/gitlab", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateGitlabState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
