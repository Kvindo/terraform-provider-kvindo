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

type VpcPeeringPeerDataSourceModel struct {
	ID       types.String             `tfsdk:"id"`
	Name     types.String             `tfsdk:"name"`
	Metadata *metadataModel           `tfsdk:"metadata"`
	Spec     *VpcPeeringPeerSpecModel `tfsdk:"spec"`
	Status   types.Object             `tfsdk:"status"`
}

type VpcPeeringPeerDataSource struct{ client *client.Client }

func NewVpcPeeringPeerDataSource() datasource.DataSource { return &VpcPeeringPeerDataSource{} }

func (d *VpcPeeringPeerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peering_peer"
}

func (d *VpcPeeringPeerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"floating_ip_id": schema.StringAttribute{Computed: true},
		"vpc_peering_id": schema.StringAttribute{Computed: true},
		"vpc_subnet_id":  schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"private_ipv4": schema.StringAttribute{Computed: true}, "private_ipv6": schema.StringAttribute{Computed: true}, "public_ipv4": schema.StringAttribute{Computed: true}, "public_ipv6": schema.StringAttribute{Computed: true}}),
	}}
}

func (d *VpcPeeringPeerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpcPeeringPeerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VpcPeeringPeerDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/vpc-peering-peer", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/vpc-peering-peer", state.Name.ValueString())
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
	state.Spec = &VpcPeeringPeerSpecModel{}
	spec := getSpec(apiData)
	state.Spec.FloatingIpId = getString(spec, "floatingIpId")
	state.Spec.VpcPeeringId = getString(spec, "vpcPeeringId")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"private_ipv4": types.StringType,
			"private_ipv6": types.StringType,
			"public_ipv4":  types.StringType,
			"public_ipv6":  types.StringType,
		},
		map[string]attr.Value{
			"private_ipv4": getStringFromInfo(apiData, "privateIpV4"),
			"private_ipv6": getStringFromInfo(apiData, "privateIpV6"),
			"public_ipv4":  getStringFromInfo(apiData, "publicIpV4"),
			"public_ipv6":  getStringFromInfo(apiData, "publicIpV6"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
