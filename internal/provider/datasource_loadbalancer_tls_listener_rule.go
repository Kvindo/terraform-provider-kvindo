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

type LoadbalancerTlsListenerRuleDataSourceModel struct {
	ID       types.String                         `tfsdk:"id"`
	Metadata metadataModel                        `tfsdk:"metadata"`
	Spec     LoadbalancerTlsListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                         `tfsdk:"status"`
}

type LoadbalancerTlsListenerRuleDataSource struct{ client *client.Client }

func NewLoadbalancerTlsListenerRuleDataSource() datasource.DataSource {
	return &LoadbalancerTlsListenerRuleDataSource{}
}

func (d *LoadbalancerTlsListenerRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tls_listener_rule"
}

func (d *LoadbalancerTlsListenerRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"forward_to_tcp_response_action": objDatasourceSchema(loadbalancerTlsListenerRuleForwardToTcpResponseActionObjFields),
		"forward_to_tls_response_action": objDatasourceSchema(loadbalancerTlsListenerRuleForwardToTlsResponseActionObjFields),
		"order":                          schema.Int64Attribute{Computed: true},
		"tls_listener_id":                schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerTlsListenerRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerTlsListenerRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerTlsListenerRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-tls-listener-rule", state.ID.ValueString())
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
	state.Spec.ForwardToTcpResponseAction = objFromAPI(objMap(spec, "forwardToTcpResponseAction"), loadbalancerTlsListenerRuleForwardToTcpResponseActionObjFields)
	state.Spec.ForwardToTlsResponseAction = objFromAPI(objMap(spec, "forwardToTlsResponseAction"), loadbalancerTlsListenerRuleForwardToTlsResponseActionObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.TlsListenerId = getString(spec, "tlsListenerId")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
