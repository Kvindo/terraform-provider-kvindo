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

// GitlabRunnerResourceModel describes the resource data model.
type GitlabRunnerResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	VmState types.String `tfsdk:"vm_state"`
	VmOfferId types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib types.Int64 `tfsdk:"volume_size_gib"`
	Concurrency types.Int64 `tfsdk:"concurrency"`
	Version types.String `tfsdk:"version"`
	DockerOptionsJsonString types.String `tfsdk:"docker_options_json_string"`
	Info types.Object `tfsdk:"info"`
}

// GitlabRunnerResource defines the resource implementation.
type GitlabRunnerResource struct {
	client *client.Client
}

func NewGitlabRunnerResource() resource.Resource {
	return &GitlabRunnerResource{}
}

func (r *GitlabRunnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab_runner"
}

func (r *GitlabRunnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{
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
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
	attrs["concurrency"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["version"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["docker_options_json_string"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *GitlabRunnerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildGitlabRunnerRequestMap(ctx context.Context, plan GitlabRunnerResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.Tier.IsNull() && !plan.Tier.IsUnknown() {
		m["tier"] = plan.Tier.ValueString()
	}
	if !plan.VpcSubnetId.IsNull() && !plan.VpcSubnetId.IsUnknown() {
		m["vpcSubnetId"] = plan.VpcSubnetId.ValueString()
	}
	if !plan.FloatingIpId.IsNull() && !plan.FloatingIpId.IsUnknown() {
		m["floatingIpId"] = plan.FloatingIpId.ValueString()
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
	if !plan.Concurrency.IsNull() && !plan.Concurrency.IsUnknown() {
		m["concurrency"] = plan.Concurrency.ValueInt64()
	}
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		m["version"] = plan.Version.ValueString()
	}
	if !plan.DockerOptionsJsonString.IsNull() && !plan.DockerOptionsJsonString.IsUnknown() {
		m["dockerOptionsJsonString"] = plan.DockerOptionsJsonString.ValueString()
	}
	return m
}

func populateGitlabRunnerState(ctx context.Context, data map[string]interface{}, state *GitlabRunnerResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.Tier = getString(data, "tier")
	state.VpcSubnetId = getString(data, "vpcSubnetId")
	state.FloatingIpId = getString(data, "floatingIpId")
	state.VmState = getString(data, "vmState")
	state.VmOfferId = getString(data, "vmOfferId")
	state.VolumeOfferId = getString(data, "volumeOfferId")
	state.VolumeSizeGib = getInt64(data, "volumeSizeGiB")
	state.Concurrency = getInt64(data, "concurrency")
	state.Version = getString(data, "version")
	state.DockerOptionsJsonString = getString(data, "dockerOptionsJsonString")
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *GitlabRunnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GitlabRunnerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildGitlabRunnerRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/gitlab-runner", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/gitlab-runner", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/gitlab-runner", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateGitlabRunnerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabRunnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GitlabRunnerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/gitlab-runner", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateGitlabRunnerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabRunnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GitlabRunnerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state GitlabRunnerResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildGitlabRunnerRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/gitlab-runner", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/gitlab-runner", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/gitlab-runner", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateGitlabRunnerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabRunnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GitlabRunnerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/gitlab-runner", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/gitlab-runner", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *GitlabRunnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state GitlabRunnerResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/gitlab-runner", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateGitlabRunnerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
