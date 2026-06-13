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
// attr package used for list/object types

// OpenVpnUserSettingsDataSourceModel describes the data source data model.
type OpenVpnUserSettingsDataSourceModel struct {
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
	InfoState types.String `tfsdk:"info_state"`
}

type OpenVpnUserSettingsDataSource struct {
	client *client.Client
}

func NewOpenVpnUserSettingsDataSource() datasource.DataSource {
	return &OpenVpnUserSettingsDataSource{}
}

func (d *OpenVpnUserSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_vpn_user_settings"
}

func (d *OpenVpnUserSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["allowed_ip_v4_cidrs"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["allowed_ip_v6_cidrs"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["denied_ip_v4_cidrs"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["denied_ip_v6_cidrs"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["allowed_domains"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["denied_domains"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
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
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/open-vpn-user-settings", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	if err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	state.AllowedIpV4Cidrs = getStringList(ctx, apiData, "allowedIpV4Cidrs")
	state.AllowedIpV6Cidrs = getStringList(ctx, apiData, "allowedIpV6Cidrs")
	state.DeniedIpV4Cidrs = getStringList(ctx, apiData, "deniedIpV4Cidrs")
	state.DeniedIpV6Cidrs = getStringList(ctx, apiData, "deniedIpV6Cidrs")
	state.AllowedDomains = getStringList(ctx, apiData, "allowedDomains")
	state.DeniedDomains = getStringList(ctx, apiData, "deniedDomains")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
