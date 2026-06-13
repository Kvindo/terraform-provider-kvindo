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

// KubernetesNodeGroupDataSourceModel describes the data source data model.
type KubernetesNodeGroupDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	KubernetesId types.String `tfsdk:"kubernetes_id"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	VmOfferId types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib types.Int64 `tfsdk:"volume_size_gib"`
	DesiredNodeCount types.Int64 `tfsdk:"desired_node_count"`
	VmState types.String `tfsdk:"vm_state"`
	CreatePublicIpv4 types.Bool `tfsdk:"create_public_ipv4"`
	InfoState types.String `tfsdk:"info_state"`
}

type KubernetesNodeGroupDataSource struct {
	client *client.Client
}

func NewKubernetesNodeGroupDataSource() datasource.DataSource {
	return &KubernetesNodeGroupDataSource{}
}

func (d *KubernetesNodeGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_node_group"
}

func (d *KubernetesNodeGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["kubernetes_id"] = schema.StringAttribute{Computed: true}
	attrs["vpc_subnet_id"] = schema.StringAttribute{Computed: true}
	attrs["vm_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_size_gib"] = schema.Int64Attribute{Computed: true}
	attrs["desired_node_count"] = schema.Int64Attribute{Computed: true}
	attrs["vm_state"] = schema.StringAttribute{Computed: true}
	attrs["create_public_ipv4"] = schema.BoolAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
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
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/kubernetes-node-group", state.ID.ValueString())
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
	state.KubernetesId = getString(apiData, "kubernetesId")
	state.VpcSubnetId = getString(apiData, "vpcSubnetId")
	state.VmOfferId = getString(apiData, "vmOfferId")
	state.VolumeOfferId = getString(apiData, "volumeOfferId")
	state.VolumeSizeGib = getInt64(apiData, "volumeSizeGib")
	state.DesiredNodeCount = getInt64(apiData, "desiredNodeCount")
	state.VmState = getString(apiData, "vmState")
	state.CreatePublicIpv4 = getBool(apiData, "createPublicIpv4")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
