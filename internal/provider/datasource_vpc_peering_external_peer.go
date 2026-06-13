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

// VpcPeeringExternalPeerDataSourceModel describes the data source data model.
type VpcPeeringExternalPeerDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VpcPeeringId types.String `tfsdk:"vpc_peering_id"`
	SshUser types.String `tfsdk:"ssh_user"`
	SshPort types.Int64 `tfsdk:"ssh_port"`
	SshIpV4 types.String `tfsdk:"ssh_ip_v4"`
	PrivateIpV4 types.String `tfsdk:"private_ip_v4"`
	IpV4Cidrs types.List `tfsdk:"ip_v4_cidrs"`
	SshPrivateKeyId types.String `tfsdk:"ssh_private_key_id"`
	InfoState types.String `tfsdk:"info_state"`
}

type VpcPeeringExternalPeerDataSource struct {
	client *client.Client
}

func NewVpcPeeringExternalPeerDataSource() datasource.DataSource {
	return &VpcPeeringExternalPeerDataSource{}
}

func (d *VpcPeeringExternalPeerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peering_external_peer"
}

func (d *VpcPeeringExternalPeerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["vpc_peering_id"] = schema.StringAttribute{Computed: true}
	attrs["ssh_user"] = schema.StringAttribute{Computed: true}
	attrs["ssh_port"] = schema.Int64Attribute{Computed: true}
	attrs["ssh_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["private_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["ip_v4_cidrs"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["ssh_private_key_id"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *VpcPeeringExternalPeerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpcPeeringExternalPeerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VpcPeeringExternalPeerDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/vpc-peering-external-peer", state.ID.ValueString())
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
	state.VpcPeeringId = getString(apiData, "vpcPeeringId")
	state.SshUser = getString(apiData, "sshUser")
	state.SshPort = getInt64(apiData, "sshPort")
	state.SshIpV4 = getString(apiData, "sshIpV4")
	state.PrivateIpV4 = getString(apiData, "privateIpV4")
	state.IpV4Cidrs = getStringList(ctx, apiData, "ipV4Cidrs")
	state.SshPrivateKeyId = getString(apiData, "sshPrivateKeyId")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
