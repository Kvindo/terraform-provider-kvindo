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

type KubernetesNodeGroupDataSourceModel struct {
	ID       types.String                 `tfsdk:"id"`
	Name     types.String                 `tfsdk:"name"`
	Metadata metadataModel                `tfsdk:"metadata"`
	Spec     KubernetesNodeGroupSpecModel `tfsdk:"spec"`
	Status   types.Object                 `tfsdk:"status"`
}

type KubernetesNodeGroupDataSource struct{ client *client.Client }

func NewKubernetesNodeGroupDataSource() datasource.DataSource {
	return &KubernetesNodeGroupDataSource{}
}

func (d *KubernetesNodeGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_node_group"
}

func (d *KubernetesNodeGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"create_public_ipv4": schema.BoolAttribute{Computed: true},
		"desired_node_count": schema.Int64Attribute{Computed: true},
		"kubernetes_id":      schema.StringAttribute{Computed: true},
		"vm_offer_id":        schema.StringAttribute{Computed: true},
		"vm_state":           schema.StringAttribute{Computed: true},
		"volume_offer_id":    schema.StringAttribute{Computed: true},
		"volume_size_gib":    schema.Int64Attribute{Computed: true},
		"vpc_subnet_id":      schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"nodes": schema.StringAttribute{Computed: true}}),
	}}
}

func (d *KubernetesNodeGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KubernetesNodeGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KubernetesNodeGroupDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/kubernetes-node-group", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/kubernetes-node-group", state.Name.ValueString())
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
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.DesiredNodeCount = getInt64(spec, "desiredNodeCount")
	state.Spec.KubernetesId = getString(spec, "kubernetesId")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"nodes": types.StringType,
		},
		map[string]attr.Value{
			"nodes": getStringFromInfo(apiData, "nodes"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
