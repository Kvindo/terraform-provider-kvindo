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

// VpcDataSourceModel describes the data source data model.
type VpcDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	HostingProviderId types.String `tfsdk:"hosting_provider_id"`
	Ipv4Cidr types.String `tfsdk:"ipv4_cidr"`
	NatFloatingIpId types.String `tfsdk:"nat_floating_ip_id"`
	SecurityGroupIds types.List `tfsdk:"security_group_ids"`
	ExternallyManaged types.Bool `tfsdk:"externally_managed"`
	InfoState types.String `tfsdk:"info_state"`
	InfoNatPublicIpV4 types.String `tfsdk:"info_nat_public_ip_v4"`
}

type VpcDataSource struct {
	client *client.Client
}

func NewVpcDataSource() datasource.DataSource {
	return &VpcDataSource{}
}

func (d *VpcDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (d *VpcDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["hosting_provider_id"] = schema.StringAttribute{Computed: true}
	attrs["ipv4_cidr"] = schema.StringAttribute{Computed: true}
	attrs["nat_floating_ip_id"] = schema.StringAttribute{Computed: true}
	attrs["security_group_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["externally_managed"] = schema.BoolAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_nat_public_ip_v4"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *VpcDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpcDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VpcDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/vpc", state.ID.ValueString())
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
	state.HostingProviderId = getString(apiData, "hostingProviderId")
	state.Ipv4Cidr = getString(apiData, "ipv4Cidr")
	state.NatFloatingIpId = getString(apiData, "natFloatingIpId")
	state.SecurityGroupIds = getStringList(ctx, apiData, "securityGroupIds")
	state.ExternallyManaged = getBool(apiData, "externallyManaged")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoNatPublicIpV4 = getStringFromInfo(apiData, "natpublicipv4")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
