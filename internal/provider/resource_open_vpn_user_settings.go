package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type OpenVpnUserSettingsSpecModel struct {
	AllowedDomains   types.List `tfsdk:"allowed_domains"`
	AllowedIpV4Cidrs types.List `tfsdk:"allowed_ipv4_cidrs"`
	AllowedIpV6Cidrs types.List `tfsdk:"allowed_ipv6_cidrs"`
	DeniedDomains    types.List `tfsdk:"denied_domains"`
	DeniedIpV4Cidrs  types.List `tfsdk:"denied_ipv4_cidrs"`
	DeniedIpV6Cidrs  types.List `tfsdk:"denied_ipv6_cidrs"`
}

type OpenVpnUserSettingsResourceModel struct {
	ID       types.String                 `tfsdk:"id"`
	Metadata metadataModel                `tfsdk:"metadata"`
	Spec     OpenVpnUserSettingsSpecModel `tfsdk:"spec"`
	Status   types.Object                 `tfsdk:"status"`
}

type OpenVpnUserSettingsResource struct{ client *client.Client }

func NewOpenVpnUserSettingsResource() resource.Resource { return &OpenVpnUserSettingsResource{} }

func (r *OpenVpnUserSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_vpn_user_settings"
}

func OpenVpnUserSettingsResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"allowed_domains":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"allowed_ipv4_cidrs": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"allowed_ipv6_cidrs": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"denied_domains":     schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"denied_ipv4_cidrs":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"denied_ipv6_cidrs":  schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *OpenVpnUserSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: OpenVpnUserSettingsResourceSchemaAttrs()}
}

func (r *OpenVpnUserSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildOpenVpnUserSettingsRequestMap(ctx context.Context, plan OpenVpnUserSettingsResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.AllowedDomains.IsNull() && !plan.Spec.AllowedDomains.IsUnknown() {
		spec["allowedDomains"] = stringListToInterface(ctx, plan.Spec.AllowedDomains)
	}
	if !plan.Spec.AllowedIpV4Cidrs.IsNull() && !plan.Spec.AllowedIpV4Cidrs.IsUnknown() {
		spec["allowedIpV4Cidrs"] = stringListToInterface(ctx, plan.Spec.AllowedIpV4Cidrs)
	}
	if !plan.Spec.AllowedIpV6Cidrs.IsNull() && !plan.Spec.AllowedIpV6Cidrs.IsUnknown() {
		spec["allowedIpV6Cidrs"] = stringListToInterface(ctx, plan.Spec.AllowedIpV6Cidrs)
	}
	if !plan.Spec.DeniedDomains.IsNull() && !plan.Spec.DeniedDomains.IsUnknown() {
		spec["deniedDomains"] = stringListToInterface(ctx, plan.Spec.DeniedDomains)
	}
	if !plan.Spec.DeniedIpV4Cidrs.IsNull() && !plan.Spec.DeniedIpV4Cidrs.IsUnknown() {
		spec["deniedIpV4Cidrs"] = stringListToInterface(ctx, plan.Spec.DeniedIpV4Cidrs)
	}
	if !plan.Spec.DeniedIpV6Cidrs.IsNull() && !plan.Spec.DeniedIpV6Cidrs.IsUnknown() {
		spec["deniedIpV6Cidrs"] = stringListToInterface(ctx, plan.Spec.DeniedIpV6Cidrs)
	}
	return m
}

func populateOpenVpnUserSettingsState(ctx context.Context, data map[string]interface{}, state *OpenVpnUserSettingsResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.AllowedDomains = getStringList(ctx, spec, "allowedDomains")
	state.Spec.AllowedIpV4Cidrs = getStringList(ctx, spec, "allowedIpV4Cidrs")
	state.Spec.AllowedIpV6Cidrs = getStringList(ctx, spec, "allowedIpV6Cidrs")
	state.Spec.DeniedDomains = getStringList(ctx, spec, "deniedDomains")
	state.Spec.DeniedIpV4Cidrs = getStringList(ctx, spec, "deniedIpV4Cidrs")
	state.Spec.DeniedIpV6Cidrs = getStringList(ctx, spec, "deniedIpV6Cidrs")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *OpenVpnUserSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OpenVpnUserSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildOpenVpnUserSettingsRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/open-vpn-user-settings", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/open-vpn-user-settings", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user-settings", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateOpenVpnUserSettingsState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpenVpnUserSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OpenVpnUserSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user-settings", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateOpenVpnUserSettingsState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *OpenVpnUserSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state OpenVpnUserSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildOpenVpnUserSettingsRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/open-vpn-user-settings", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/open-vpn-user-settings", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user-settings", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateOpenVpnUserSettingsState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpenVpnUserSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OpenVpnUserSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/open-vpn-user-settings", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/open-vpn-user-settings", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *OpenVpnUserSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state OpenVpnUserSettingsResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user-settings", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateOpenVpnUserSettingsState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
