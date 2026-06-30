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

type VpcDataSourceModel struct {
	ID       types.String  `tfsdk:"id"`
	Name     types.String  `tfsdk:"name"`
	Metadata metadataModel `tfsdk:"metadata"`
	Spec     VpcSpecModel  `tfsdk:"spec"`
	Status   types.Object  `tfsdk:"status"`
}

type VpcDataSource struct{ client *client.Client }

func NewVpcDataSource() datasource.DataSource { return &VpcDataSource{} }

func (d *VpcDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (d *VpcDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"externally_managed":  schema.BoolAttribute{Computed: true},
		"hosting_provider_id": schema.StringAttribute{Computed: true},
		"ipv4_cidr":           schema.StringAttribute{Computed: true},
		"nat_floating_ip_id":  schema.StringAttribute{Computed: true},
		"security_group_ids":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"nat_public_ip_v4": schema.StringAttribute{Computed: true}}),
	}}
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
		apiData, err = d.client.Get(ctx, "/api/v1/vpc", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/vpc", state.Name.ValueString())
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
	state.Spec.ExternallyManaged = getBool(spec, "externallyManaged")
	state.Spec.HostingProviderId = getString(spec, "hostingProviderId")
	state.Spec.Ipv4Cidr = getString(spec, "ipv4Cidr")
	state.Spec.NatFloatingIpId = getString(spec, "natFloatingIpId")
	state.Spec.SecurityGroupIds = getStringList(ctx, spec, "securityGroupIds")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"nat_public_ip_v4": types.StringType,
		},
		map[string]attr.Value{
			"nat_public_ip_v4": getStringFromInfo(apiData, "natPublicIpV4"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
