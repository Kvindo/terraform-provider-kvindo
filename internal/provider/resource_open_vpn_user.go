package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type OpenVpnUserSpecModel struct {
	OpenVpnId          types.String `tfsdk:"open_vpn_id"`
	OpenVpnSettingsIds types.List   `tfsdk:"open_vpn_settings_ids"`
}

type OpenVpnUserResourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Metadata metadataModel        `tfsdk:"metadata"`
	Spec     OpenVpnUserSpecModel `tfsdk:"spec"`
	Status   types.Object         `tfsdk:"status"`
}

type OpenVpnUserResource struct{ client *client.Client }

func NewOpenVpnUserResource() resource.Resource { return &OpenVpnUserResource{} }

func (r *OpenVpnUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_vpn_user"
}

func OpenVpnUserResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"open_vpn_id":           schema.StringAttribute{Required: true},
		"open_vpn_settings_ids": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"config": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}
}

func (r *OpenVpnUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: OpenVpnUserResourceSchemaAttrs()}
}

func (r *OpenVpnUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildOpenVpnUserRequestMap(ctx context.Context, plan OpenVpnUserResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.OpenVpnId.IsNull() && !plan.Spec.OpenVpnId.IsUnknown() {
		spec["openVpnId"] = plan.Spec.OpenVpnId.ValueString()
	}
	if !plan.Spec.OpenVpnSettingsIds.IsNull() && !plan.Spec.OpenVpnSettingsIds.IsUnknown() {
		spec["openVpnSettingsIds"] = stringListToInterface(ctx, plan.Spec.OpenVpnSettingsIds)
	}
	return m
}

func populateOpenVpnUserState(ctx context.Context, data map[string]interface{}, state *OpenVpnUserResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.OpenVpnId = getString(spec, "openVpnId")
	state.Spec.OpenVpnSettingsIds = getStringList(ctx, spec, "openVpnSettingsIds")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"config": types.StringType,
		},
		map[string]attr.Value{
			"config": getStringFromInfo(data, "config"),
		})
	return nil
}

func (r *OpenVpnUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OpenVpnUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildOpenVpnUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/open-vpn-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/open-vpn-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateOpenVpnUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpenVpnUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OpenVpnUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateOpenVpnUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *OpenVpnUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state OpenVpnUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildOpenVpnUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/open-vpn-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/open-vpn-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateOpenVpnUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpenVpnUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OpenVpnUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/open-vpn-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/open-vpn-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *OpenVpnUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state OpenVpnUserResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/open-vpn-user", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateOpenVpnUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
