package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// OpenVpnUserSettingsResourceModel describes the resource data model.
type OpenVpnUserSettingsResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	AllowedIpV4Cidrs types.List `tfsdk:"allowed_ip_v4_cidrs"`
	AllowedIpV6Cidrs types.List `tfsdk:"allowed_ip_v6_cidrs"`
	DeniedIpV4Cidrs types.List `tfsdk:"denied_ip_v4_cidrs"`
	DeniedIpV6Cidrs types.List `tfsdk:"denied_ip_v6_cidrs"`
	AllowedDomains types.List `tfsdk:"allowed_domains"`
	DeniedDomains types.List `tfsdk:"denied_domains"`
	Info types.Object `tfsdk:"info"`
}

// OpenVpnUserSettingsResource defines the resource implementation.
type OpenVpnUserSettingsResource struct {
	client *client.Client
}

func NewOpenVpnUserSettingsResource() resource.Resource {
	return &OpenVpnUserSettingsResource{}
}

func (r *OpenVpnUserSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_vpn_user_settings"
}

func (r *OpenVpnUserSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["allowed_ip_v4_cidrs"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["allowed_ip_v6_cidrs"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["denied_ip_v4_cidrs"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["denied_ip_v6_cidrs"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["allowed_domains"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["denied_domains"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.AllowedIpV4Cidrs.IsNull() && !plan.AllowedIpV4Cidrs.IsUnknown() {
		m["allowedIpV4Cidrs"] = stringListToInterface(ctx, plan.AllowedIpV4Cidrs)
	}
	if !plan.AllowedIpV6Cidrs.IsNull() && !plan.AllowedIpV6Cidrs.IsUnknown() {
		m["allowedIpV6Cidrs"] = stringListToInterface(ctx, plan.AllowedIpV6Cidrs)
	}
	if !plan.DeniedIpV4Cidrs.IsNull() && !plan.DeniedIpV4Cidrs.IsUnknown() {
		m["deniedIpV4Cidrs"] = stringListToInterface(ctx, plan.DeniedIpV4Cidrs)
	}
	if !plan.DeniedIpV6Cidrs.IsNull() && !plan.DeniedIpV6Cidrs.IsUnknown() {
		m["deniedIpV6Cidrs"] = stringListToInterface(ctx, plan.DeniedIpV6Cidrs)
	}
	if !plan.AllowedDomains.IsNull() && !plan.AllowedDomains.IsUnknown() {
		m["allowedDomains"] = stringListToInterface(ctx, plan.AllowedDomains)
	}
	if !plan.DeniedDomains.IsNull() && !plan.DeniedDomains.IsUnknown() {
		m["deniedDomains"] = stringListToInterface(ctx, plan.DeniedDomains)
	}
	return m
}

func populateOpenVpnUserSettingsState(ctx context.Context, data map[string]interface{}, state *OpenVpnUserSettingsResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.AllowedIpV4Cidrs = getStringList(ctx, data, "allowedIpV4Cidrs")
	state.AllowedIpV6Cidrs = getStringList(ctx, data, "allowedIpV6Cidrs")
	state.DeniedIpV4Cidrs = getStringList(ctx, data, "deniedIpV4Cidrs")
	state.DeniedIpV6Cidrs = getStringList(ctx, data, "deniedIpV6Cidrs")
	state.AllowedDomains = getStringList(ctx, data, "allowedDomains")
	state.DeniedDomains = getStringList(ctx, data, "deniedDomains")
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *OpenVpnUserSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OpenVpnUserSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OpenVpnUserSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OpenVpnUserSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *OpenVpnUserSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OpenVpnUserSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state OpenVpnUserSettingsResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateOpenVpnUserSettingsState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OpenVpnUserSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OpenVpnUserSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	// Import by ID
	var state OpenVpnUserSettingsResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user-settings", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateOpenVpnUserSettingsState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
