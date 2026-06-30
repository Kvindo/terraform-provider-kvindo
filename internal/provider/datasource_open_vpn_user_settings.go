package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type OpenVpnUserSettingsDataSourceModel struct {
	ID       types.String                 `tfsdk:"id"`
	Name     types.String                 `tfsdk:"name"`
	Metadata metadataModel                `tfsdk:"metadata"`
	Spec     OpenVpnUserSettingsSpecModel `tfsdk:"spec"`
	Status   types.Object                 `tfsdk:"status"`
}

type OpenVpnUserSettingsDataSource struct{ client *client.Client }

func NewOpenVpnUserSettingsDataSource() datasource.DataSource {
	return &OpenVpnUserSettingsDataSource{}
}

func (d *OpenVpnUserSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_vpn_user_settings"
}

func (d *OpenVpnUserSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"allowed_domains":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"allowed_ip_v4_cidrs": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"allowed_ip_v6_cidrs": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"denied_domains":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"denied_ip_v4_cidrs":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"denied_ip_v6_cidrs":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true},
		"name":     schema.StringAttribute{Optional: true, Computed: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *OpenVpnUserSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	d.client = pd.Client
}

func (d *OpenVpnUserSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OpenVpnUserSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set")
		return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/open-vpn-user-settings", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/open-vpn-user-settings", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	spec := getSpec(apiData)
	state.Spec.AllowedDomains = getStringList(ctx, spec, "allowedDomains")
	state.Spec.AllowedIpV4Cidrs = getStringList(ctx, spec, "allowedIpV4Cidrs")
	state.Spec.AllowedIpV6Cidrs = getStringList(ctx, spec, "allowedIpV6Cidrs")
	state.Spec.DeniedDomains = getStringList(ctx, spec, "deniedDomains")
	state.Spec.DeniedIpV4Cidrs = getStringList(ctx, spec, "deniedIpV4Cidrs")
	state.Spec.DeniedIpV6Cidrs = getStringList(ctx, spec, "deniedIpV6Cidrs")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
