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

type LoadbalancerTargetGroupServiceDiscoveryTargetDataSourceModel struct {
	ID       types.String                                           `tfsdk:"id"`
	Metadata metadataModel                                          `tfsdk:"metadata"`
	Spec     LoadbalancerTargetGroupServiceDiscoveryTargetSpecModel `tfsdk:"spec"`
	Status   types.Object                                           `tfsdk:"status"`
}

type LoadbalancerTargetGroupServiceDiscoveryTargetDataSource struct{ client *client.Client }

func NewLoadbalancerTargetGroupServiceDiscoveryTargetDataSource() datasource.DataSource {
	return &LoadbalancerTargetGroupServiceDiscoveryTargetDataSource{}
}

func (d *LoadbalancerTargetGroupServiceDiscoveryTargetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_target_group_service_discovery_target"
}

func (d *LoadbalancerTargetGroupServiceDiscoveryTargetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"label_selectors": schema.MapAttribute{Computed: true, ElementType: types.StringType},
		"target_group_id": schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerTargetGroupServiceDiscoveryTargetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerTargetGroupServiceDiscoveryTargetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerTargetGroupServiceDiscoveryTargetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", state.ID.ValueString())
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
	spec := getSpec(apiData)
	state.Spec.LabelSelectors = getStringMap(spec, "labelSelectors")
	state.Spec.TargetGroupId = getString(spec, "targetGroupId")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
