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

type LoadbalancerHttpListenerRuleDataSourceModel struct {
	ID       types.String                          `tfsdk:"id"`
	Metadata metadataModel                         `tfsdk:"metadata"`
	Spec     LoadbalancerHttpListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                          `tfsdk:"status"`
}

type LoadbalancerHttpListenerRuleDataSource struct{ client *client.Client }

func NewLoadbalancerHttpListenerRuleDataSource() datasource.DataSource {
	return &LoadbalancerHttpListenerRuleDataSource{}
}

func (d *LoadbalancerHttpListenerRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_http_listener_rule"
}

func (d *LoadbalancerHttpListenerRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"delete_request_headers_action":    objDatasourceSchema(loadbalancerHttpListenerRuleDeleteRequestHeadersActionObjFields),
		"delete_response_headers_action":   objDatasourceSchema(loadbalancerHttpListenerRuleDeleteResponseHeadersActionObjFields),
		"forward_to_http_response_action":  objDatasourceSchema(loadbalancerHttpListenerRuleForwardToHttpResponseActionObjFields),
		"forward_to_https_response_action": objDatasourceSchema(loadbalancerHttpListenerRuleForwardToHttpsResponseActionObjFields),
		"http_listener_id":                 schema.StringAttribute{Computed: true},
		"match":                            objDatasourceSchema(loadbalancerHttpListenerRuleMatchObjFields),
		"order":                            schema.Int64Attribute{Computed: true},
		"path_rewrite_action":              objDatasourceSchema(loadbalancerHttpListenerRulePathRewriteActionObjFields),
		"set_request_headers_action":       objDatasourceSchema(loadbalancerHttpListenerRuleSetRequestHeadersActionObjFields),
		"set_response_headers_action":      objDatasourceSchema(loadbalancerHttpListenerRuleSetResponseHeadersActionObjFields),
		"static_response_action":           objDatasourceSchema(loadbalancerHttpListenerRuleStaticResponseActionObjFields),
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerHttpListenerRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerHttpListenerRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerHttpListenerRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-http-listener-rule", state.ID.ValueString())
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
	state.Spec.DeleteRequestHeadersAction = objFromAPI(objMap(spec, "deleteRequestHeadersAction"), loadbalancerHttpListenerRuleDeleteRequestHeadersActionObjFields)
	state.Spec.DeleteResponseHeadersAction = objFromAPI(objMap(spec, "deleteResponseHeadersAction"), loadbalancerHttpListenerRuleDeleteResponseHeadersActionObjFields)
	state.Spec.ForwardToHttpResponseAction = objFromAPI(objMap(spec, "forwardToHttpResponseAction"), loadbalancerHttpListenerRuleForwardToHttpResponseActionObjFields)
	state.Spec.ForwardToHttpsResponseAction = objFromAPI(objMap(spec, "forwardToHttpsResponseAction"), loadbalancerHttpListenerRuleForwardToHttpsResponseActionObjFields)
	state.Spec.HttpListenerId = getString(spec, "httpListenerId")
	state.Spec.Match = objFromAPI(objMap(spec, "match"), loadbalancerHttpListenerRuleMatchObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.PathRewriteAction = objFromAPI(objMap(spec, "pathRewriteAction"), loadbalancerHttpListenerRulePathRewriteActionObjFields)
	state.Spec.SetRequestHeadersAction = objFromAPI(objMap(spec, "setRequestHeadersAction"), loadbalancerHttpListenerRuleSetRequestHeadersActionObjFields)
	state.Spec.SetResponseHeadersAction = objFromAPI(objMap(spec, "setResponseHeadersAction"), loadbalancerHttpListenerRuleSetResponseHeadersActionObjFields)
	state.Spec.StaticResponseAction = objFromAPI(objMap(spec, "staticResponseAction"), loadbalancerHttpListenerRuleStaticResponseActionObjFields)
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
