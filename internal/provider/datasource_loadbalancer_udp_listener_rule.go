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

type LoadbalancerUdpListenerRuleDataSourceModel struct {
	ID       types.String                         `tfsdk:"id"`
	Name     types.String                         `tfsdk:"name"`
	Metadata metadataModel                        `tfsdk:"metadata"`
	Spec     LoadbalancerUdpListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                         `tfsdk:"status"`
}

type LoadbalancerUdpListenerRuleDataSource struct{ client *client.Client }

func NewLoadbalancerUdpListenerRuleDataSource() datasource.DataSource {
	return &LoadbalancerUdpListenerRuleDataSource{}
}

func (d *LoadbalancerUdpListenerRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_udp_listener_rule"
}

func (d *LoadbalancerUdpListenerRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"forward_to_udp_response_action": objDatasourceSchema(loadbalancerUdpListenerRuleForwardToUdpResponseActionObjFields),
		"order":                          schema.Int64Attribute{Computed: true},
		"udp_listener_id":                schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true},
		"name":     schema.StringAttribute{Optional: true, Computed: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerUdpListenerRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerUdpListenerRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerUdpListenerRuleDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/loadbalancer-udp-listener-rule", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/loadbalancer-udp-listener-rule", state.Name.ValueString())
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
	state.Spec.ForwardToUdpResponseAction = objFromAPI(objMap(spec, "forwardToUdpResponseAction"), loadbalancerUdpListenerRuleForwardToUdpResponseActionObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.UdpListenerId = getString(spec, "udpListenerId")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
