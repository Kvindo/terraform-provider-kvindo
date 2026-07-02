package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type OpenVpnUserDataSourceModel struct {
	ID       types.String          `tfsdk:"id"`
	Name     types.String          `tfsdk:"name"`
	Metadata *metadataModel        `tfsdk:"metadata"`
	Spec     *OpenVpnUserSpecModel `tfsdk:"spec"`
	Status   types.Object          `tfsdk:"status"`
}

type OpenVpnUserDataSource struct{ client *client.Client }

func NewOpenVpnUserDataSource() datasource.DataSource { return &OpenVpnUserDataSource{} }

func (d *OpenVpnUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_vpn_user"
}

func (d *OpenVpnUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"open_vpn_id":           schema.StringAttribute{Computed: true},
		"open_vpn_settings_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"config": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}}
}

func (d *OpenVpnUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OpenVpnUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OpenVpnUserDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/open-vpn-user", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/open-vpn-user", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &OpenVpnUserSpecModel{}
	spec := getSpec(apiData)
	state.Spec.OpenVpnId = getString(spec, "openVpnId")
	state.Spec.OpenVpnSettingsIds = getStringList(ctx, spec, "openVpnSettingsIds")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"config": types.StringType,
		},
		map[string]attr.Value{
			"config": getStringFromInfo(apiData, "config"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
