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

// LoadbalancerTargetGroupServiceDiscoveryTargetDataSourceModel describes the data source data model.
type LoadbalancerTargetGroupServiceDiscoveryTargetDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TargetGroupId types.String `tfsdk:"target_group_id"`
	LabelSelectors types.Map `tfsdk:"label_selectors"`
	InfoState types.String `tfsdk:"info_state"`
}

type LoadbalancerTargetGroupServiceDiscoveryTargetDataSource struct {
	client *client.Client
}

func NewLoadbalancerTargetGroupServiceDiscoveryTargetDataSource() datasource.DataSource {
	return &LoadbalancerTargetGroupServiceDiscoveryTargetDataSource{}
}

func (d *LoadbalancerTargetGroupServiceDiscoveryTargetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_target_group_service_discovery_target"
}

func (d *LoadbalancerTargetGroupServiceDiscoveryTargetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["target_group_id"] = schema.StringAttribute{Computed: true}
	attrs["label_selectors"] = schema.MapAttribute{Computed: true, ElementType: types.StringType}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
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
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	if err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	state.TargetGroupId = getString(apiData, "targetGroupId")
	state.LabelSelectors = getStringMap(apiData, "labelSelectors")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
