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

var gitlabCustomIngressConfigurationObjFields = []objField{{TF: "certificate_id", API: "certificateId", Kind: "string"}, {TF: "hostname", API: "hostname", Kind: "string"}}

var gitlabSecretPushProtectionObjFields = []objField{{TF: "enabled", API: "enabled", Kind: "bool"}, {TF: "ignored_repositories_paths", API: "ignoredRepositoriesPaths", Kind: "list_string"}, {TF: "regexp_patterns", API: "regexpPatterns", Kind: "list_string"}, {TF: "string_contains_patterns", API: "stringContainsPatterns", Kind: "list_string"}}

type GitlabSpecModel struct {
	CustomIngressConfiguration types.Object `tfsdk:"custom_ingress_configuration"`
	Edition                    types.String `tfsdk:"edition"`
	FloatingIpId               types.String `tfsdk:"floating_ip_id"`
	RecordName                 types.String `tfsdk:"record_name"`
	RootPassword               types.String `tfsdk:"root_password"`
	SecretPushProtection       types.Object `tfsdk:"secret_push_protection"`
	Tier                       types.String `tfsdk:"tier"`
	Version                    types.String `tfsdk:"version"`
	VmOfferId                  types.String `tfsdk:"vm_offer_id"`
	VmState                    types.String `tfsdk:"vm_state"`
	VolumeOfferId              types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib              types.Int64  `tfsdk:"volume_size_gib"`
	VpcSubnetId                types.String `tfsdk:"vpc_subnet_id"`
}

type GitlabResourceModel struct {
	ID       types.String    `tfsdk:"id"`
	Metadata metadataModel   `tfsdk:"metadata"`
	Spec     GitlabSpecModel `tfsdk:"spec"`
	Status   types.Object    `tfsdk:"status"`
}

type GitlabResource struct{ client *client.Client }

func NewGitlabResource() resource.Resource { return &GitlabResource{} }

func (r *GitlabResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab"
}

func GitlabResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"custom_ingress_configuration": objResourceSchema(gitlabCustomIngressConfigurationObjFields),
		"edition":                      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"floating_ip_id":               schema.StringAttribute{Optional: true},
		"record_name":                  schema.StringAttribute{Optional: true},
		"root_password":                schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"secret_push_protection":       objResourceSchema(gitlabSecretPushProtectionObjFields),
		"tier":                         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"version":                      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_offer_id":                  schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_state":                     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":              schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":              schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"vpc_subnet_id":                schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"fqdn": schema.StringAttribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "private_ip_v6": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "public_ip_v6": schema.StringAttribute{Computed: true}}),
	}
}

func (r *GitlabResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: GitlabResourceSchemaAttrs()}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.CustomIngressConfiguration.IsNull() && !plan.Spec.CustomIngressConfiguration.IsUnknown() {
		spec["customIngressConfiguration"] = objToAPI(plan.Spec.CustomIngressConfiguration, gitlabCustomIngressConfigurationObjFields)
	}
	if !plan.Spec.Edition.IsNull() && !plan.Spec.Edition.IsUnknown() {
		spec["edition"] = plan.Spec.Edition.ValueString()
	}
	if !plan.Spec.FloatingIpId.IsNull() && !plan.Spec.FloatingIpId.IsUnknown() {
		spec["floatingIpId"] = plan.Spec.FloatingIpId.ValueString()
	}
	if !plan.Spec.RecordName.IsNull() && !plan.Spec.RecordName.IsUnknown() {
		spec["recordName"] = plan.Spec.RecordName.ValueString()
	}
	if !plan.Spec.RootPassword.IsNull() && !plan.Spec.RootPassword.IsUnknown() {
		spec["rootPassword"] = plan.Spec.RootPassword.ValueString()
	}
	if !plan.Spec.SecretPushProtection.IsNull() && !plan.Spec.SecretPushProtection.IsUnknown() {
		spec["secretPushProtection"] = objToAPI(plan.Spec.SecretPushProtection, gitlabSecretPushProtectionObjFields)
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

func populateGitlabState(ctx context.Context, data map[string]interface{}, state *GitlabResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.CustomIngressConfiguration = objFromAPI(objMap(spec, "customIngressConfiguration"), gitlabCustomIngressConfigurationObjFields)
	state.Spec.Edition = getString(spec, "edition")
	state.Spec.FloatingIpId = getString(spec, "floatingIpId")
	state.Spec.RecordName = getString(spec, "recordName")
	state.Spec.RootPassword = getString(spec, "rootPassword")
	state.Spec.SecretPushProtection = objFromAPI(objMap(spec, "secretPushProtection"), gitlabSecretPushProtectionObjFields)
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.Version = getString(spec, "version")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"fqdn":          types.StringType,
			"private_ip_v4": types.StringType,
			"private_ip_v6": types.StringType,
			"public_ip_v4":  types.StringType,
			"public_ip_v6":  types.StringType,
		},
		map[string]attr.Value{
			"fqdn":          getStringFromInfo(data, "fqdn"),
			"private_ip_v4": getStringFromInfo(data, "privateIpV4"),
			"private_ip_v6": getStringFromInfo(data, "privateIpV6"),
			"public_ip_v4":  getStringFromInfo(data, "publicIpV4"),
			"public_ip_v6":  getStringFromInfo(data, "publicIpV6"),
		})
	return nil
}

func (r *GitlabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GitlabResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GitlabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GitlabResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *GitlabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GitlabResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateGitlabState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *GitlabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GitlabResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
	var state GitlabResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/gitlab", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateGitlabState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
