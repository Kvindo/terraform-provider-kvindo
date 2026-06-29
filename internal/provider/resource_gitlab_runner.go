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

var gitlabRunnerGitlabInstancesObjFields = []objField{{TF: "runner_token", API: "runnerToken", Kind: "string"}, {TF: "url", API: "url", Kind: "string"}}

type GitlabRunnerSpecModel struct {
	Concurrency             types.Int64  `tfsdk:"concurrency"`
	DockerOptionsJsonString types.String `tfsdk:"docker_options_json_string"`
	FloatingIpId            types.String `tfsdk:"floating_ip_id"`
	GitlabInstances         types.List   `tfsdk:"gitlab_instances"`
	Tier                    types.String `tfsdk:"tier"`
	Version                 types.String `tfsdk:"version"`
	VmOfferId               types.String `tfsdk:"vm_offer_id"`
	VmState                 types.String `tfsdk:"vm_state"`
	VolumeOfferId           types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib           types.Int64  `tfsdk:"volume_size_gib"`
	VpcSubnetId             types.String `tfsdk:"vpc_subnet_id"`
}

type GitlabRunnerResourceModel struct {
	ID       types.String          `tfsdk:"id"`
	Metadata metadataModel         `tfsdk:"metadata"`
	Spec     GitlabRunnerSpecModel `tfsdk:"spec"`
	Status   types.Object          `tfsdk:"status"`
}

type GitlabRunnerResource struct{ client *client.Client }

func NewGitlabRunnerResource() resource.Resource { return &GitlabRunnerResource{} }

func (r *GitlabRunnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab_runner"
}

func GitlabRunnerResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"concurrency":                schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"docker_options_json_string": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"floating_ip_id":             schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"gitlab_instances":           listObjResourceSchema(gitlabRunnerGitlabInstancesObjFields),
		"tier":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"version":                    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_offer_id":                schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_state":                   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":            schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":            schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"vpc_subnet_id":              schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *GitlabRunnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: GitlabRunnerResourceSchemaAttrs()}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.Concurrency.IsNull() && !plan.Spec.Concurrency.IsUnknown() {
		spec["concurrency"] = plan.Spec.Concurrency.ValueInt64()
	}
	if !plan.Spec.DockerOptionsJsonString.IsNull() && !plan.Spec.DockerOptionsJsonString.IsUnknown() {
		spec["dockerOptionsJsonString"] = plan.Spec.DockerOptionsJsonString.ValueString()
	}
	if !plan.Spec.FloatingIpId.IsNull() && !plan.Spec.FloatingIpId.IsUnknown() {
		spec["floatingIpId"] = plan.Spec.FloatingIpId.ValueString()
	}
	if !plan.Spec.GitlabInstances.IsNull() && !plan.Spec.GitlabInstances.IsUnknown() {
		spec["gitlabInstances"] = listObjToAPI(plan.Spec.GitlabInstances, gitlabRunnerGitlabInstancesObjFields)
	}
	if !plan.Spec.Tier.IsNull() && !plan.Spec.Tier.IsUnknown() {
		spec["tier"] = plan.Spec.Tier.ValueString()
	}
	if !plan.Spec.Version.IsNull() && !plan.Spec.Version.IsUnknown() {
		spec["version"] = plan.Spec.Version.ValueString()
	}
	if !plan.Spec.VmOfferId.IsNull() && !plan.Spec.VmOfferId.IsUnknown() {
		spec["vmOfferId"] = plan.Spec.VmOfferId.ValueString()
	}
	if !plan.Spec.VmState.IsNull() && !plan.Spec.VmState.IsUnknown() {
		spec["vmState"] = plan.Spec.VmState.ValueString()
	}
	if !plan.Spec.VolumeOfferId.IsNull() && !plan.Spec.VolumeOfferId.IsUnknown() {
		spec["volumeOfferId"] = plan.Spec.VolumeOfferId.ValueString()
	}
	if !plan.Spec.VolumeSizeGib.IsNull() && !plan.Spec.VolumeSizeGib.IsUnknown() {
		spec["volumeSizeGiB"] = plan.Spec.VolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.VpcSubnetId.IsNull() && !plan.Spec.VpcSubnetId.IsUnknown() {
		spec["vpcSubnetId"] = plan.Spec.VpcSubnetId.ValueString()
	}
	return m
}

func populateGitlabRunnerState(ctx context.Context, data map[string]interface{}, state *GitlabRunnerResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.Concurrency = getInt64(spec, "concurrency")
	state.Spec.DockerOptionsJsonString = getString(spec, "dockerOptionsJsonString")
	state.Spec.FloatingIpId = getString(spec, "floatingIpId")
	state.Spec.GitlabInstances = listObjFromAPI(objList(spec, "gitlabInstances"), gitlabRunnerGitlabInstancesObjFields)
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.Version = getString(spec, "version")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *GitlabRunnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GitlabRunnerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GitlabRunnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GitlabRunnerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GitlabRunnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GitlabRunnerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateGitlabRunnerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GitlabRunnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GitlabRunnerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
	var state GitlabRunnerResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/gitlab-runner", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateGitlabRunnerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
