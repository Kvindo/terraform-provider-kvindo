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

type LoadbalancerTcpListenerRuleDataSourceModel struct {
	ID       types.String                         `tfsdk:"id"`
	Name     types.String                         `tfsdk:"name"`
	Metadata metadataModel                        `tfsdk:"metadata"`
	Spec     LoadbalancerTcpListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                         `tfsdk:"status"`
}

type LoadbalancerTcpListenerRuleDataSource struct{ client *client.Client }

func NewLoadbalancerTcpListenerRuleDataSource() datasource.DataSource {
	return &LoadbalancerTcpListenerRuleDataSource{}
}

func (d *LoadbalancerTcpListenerRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tcp_listener_rule"
}

func (d *LoadbalancerTcpListenerRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"forward_to_tcp_response_action": objDatasourceSchema(loadbalancerTcpListenerRuleForwardToTcpResponseActionObjFields),
		"forward_to_tls_response_action": objDatasourceSchema(loadbalancerTcpListenerRuleForwardToTlsResponseActionObjFields),
		"order":                          schema.Int64Attribute{Computed: true},
		"tcp_listener_id":                schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true},
		"name":     schema.StringAttribute{Optional: true, Computed: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerTcpListenerRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerTcpListenerRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerTcpListenerRuleDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener-rule", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/loadbalancer-tcp-listener-rule", state.Name.ValueString())
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
	state.Spec.ForwardToTcpResponseAction = objFromAPI(objMap(spec, "forwardToTcpResponseAction"), loadbalancerTcpListenerRuleForwardToTcpResponseActionObjFields)
	state.Spec.ForwardToTlsResponseAction = objFromAPI(objMap(spec, "forwardToTlsResponseAction"), loadbalancerTcpListenerRuleForwardToTlsResponseActionObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.TcpListenerId = getString(spec, "tcpListenerId")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
