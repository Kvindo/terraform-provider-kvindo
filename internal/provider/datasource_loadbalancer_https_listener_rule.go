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

type LoadbalancerHttpsListenerRuleDataSourceModel struct {
	ID       types.String                            `tfsdk:"id"`
	Name     types.String                            `tfsdk:"name"`
	Metadata *metadataModel                          `tfsdk:"metadata"`
	Spec     *LoadbalancerHttpsListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                            `tfsdk:"status"`
}

type LoadbalancerHttpsListenerRuleDataSource struct{ client *client.Client }

func NewLoadbalancerHttpsListenerRuleDataSource() datasource.DataSource {
	return &LoadbalancerHttpsListenerRuleDataSource{}
}

func (d *LoadbalancerHttpsListenerRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_https_listener_rule"
}

func (d *LoadbalancerHttpsListenerRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"delete_request_headers_action":    objDatasourceSchema(loadbalancerHttpsListenerRuleDeleteRequestHeadersActionObjFields),
		"delete_response_headers_action":   objDatasourceSchema(loadbalancerHttpsListenerRuleDeleteResponseHeadersActionObjFields),
		"forward_to_http_response_action":  objDatasourceSchema(loadbalancerHttpsListenerRuleForwardToHttpResponseActionObjFields),
		"forward_to_https_response_action": objDatasourceSchema(loadbalancerHttpsListenerRuleForwardToHttpsResponseActionObjFields),
		"https_listener_id":                schema.StringAttribute{Computed: true},
		"match":                            objDatasourceSchema(loadbalancerHttpsListenerRuleMatchObjFields),
		"order":                            schema.Int64Attribute{Computed: true},
		"path_rewrite_action":              objDatasourceSchema(loadbalancerHttpsListenerRulePathRewriteActionObjFields),
		"set_request_headers_action":       objDatasourceSchema(loadbalancerHttpsListenerRuleSetRequestHeadersActionObjFields),
		"set_response_headers_action":      objDatasourceSchema(loadbalancerHttpsListenerRuleSetResponseHeadersActionObjFields),
		"static_response_action":           objDatasourceSchema(loadbalancerHttpsListenerRuleStaticResponseActionObjFields),
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerHttpsListenerRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerHttpsListenerRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerHttpsListenerRuleDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/loadbalancer-https-listener-rule", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/loadbalancer-https-listener-rule", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &LoadbalancerHttpsListenerRuleSpecModel{}
	spec := getSpec(apiData)
	state.Spec.DeleteRequestHeadersAction = objFromAPI(objMap(spec, "deleteRequestHeadersAction"), loadbalancerHttpsListenerRuleDeleteRequestHeadersActionObjFields)
	state.Spec.DeleteResponseHeadersAction = objFromAPI(objMap(spec, "deleteResponseHeadersAction"), loadbalancerHttpsListenerRuleDeleteResponseHeadersActionObjFields)
	state.Spec.ForwardToHttpResponseAction = objFromAPI(objMap(spec, "forwardToHttpResponseAction"), loadbalancerHttpsListenerRuleForwardToHttpResponseActionObjFields)
	state.Spec.ForwardToHttpsResponseAction = objFromAPI(objMap(spec, "forwardToHttpsResponseAction"), loadbalancerHttpsListenerRuleForwardToHttpsResponseActionObjFields)
	state.Spec.HttpsListenerId = getString(spec, "httpsListenerId")
	state.Spec.Match = objFromAPI(objMap(spec, "match"), loadbalancerHttpsListenerRuleMatchObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.PathRewriteAction = objFromAPI(objMap(spec, "pathRewriteAction"), loadbalancerHttpsListenerRulePathRewriteActionObjFields)
	state.Spec.SetRequestHeadersAction = objFromAPI(objMap(spec, "setRequestHeadersAction"), loadbalancerHttpsListenerRuleSetRequestHeadersActionObjFields)
	state.Spec.SetResponseHeadersAction = objFromAPI(objMap(spec, "setResponseHeadersAction"), loadbalancerHttpsListenerRuleSetResponseHeadersActionObjFields)
	state.Spec.StaticResponseAction = objFromAPI(objMap(spec, "staticResponseAction"), loadbalancerHttpsListenerRuleStaticResponseActionObjFields)
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
