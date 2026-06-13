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

// LoadbalancerTcpListenerDataSourceModel describes the data source data model.
type LoadbalancerTcpListenerDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	LoadbalancerId types.String `tfsdk:"loadbalancer_id"`
	Interface types.String `tfsdk:"interface"`
	Order types.Int64 `tfsdk:"order"`
	Ports types.List `tfsdk:"ports"`
	InfoState types.String `tfsdk:"info_state"`
}

type LoadbalancerTcpListenerDataSource struct {
	client *client.Client
}

func NewLoadbalancerTcpListenerDataSource() datasource.DataSource {
	return &LoadbalancerTcpListenerDataSource{}
}

func (d *LoadbalancerTcpListenerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tcp_listener"
}

func (d *LoadbalancerTcpListenerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["loadbalancer_id"] = schema.StringAttribute{Computed: true}
	attrs["interface"] = schema.StringAttribute{Computed: true}
	attrs["order"] = schema.Int64Attribute{Computed: true}
	attrs["ports"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *LoadbalancerTcpListenerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerTcpListenerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerTcpListenerDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener", state.ID.ValueString())
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
	state.LoadbalancerId = getString(apiData, "loadbalancerId")
	state.Interface = getString(apiData, "interface")
	state.Order = getInt64(apiData, "order")
	state.Ports = getStringList(ctx, apiData, "ports")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
