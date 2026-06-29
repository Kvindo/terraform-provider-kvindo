package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var loadbalancerHttpsListenerRuleDeleteRequestHeadersActionObjFields = []objField{{TF: "headers", API: "headers", Kind: "list_string"}}

var loadbalancerHttpsListenerRuleDeleteResponseHeadersActionObjFields = []objField{{TF: "headers", API: "headers", Kind: "list_string"}}

var loadbalancerHttpsListenerRuleForwardToHttpResponseActionObjFields = []objField{{TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

var loadbalancerHttpsListenerRuleForwardToHttpsResponseActionObjFields = []objField{{TF: "pass_as_grpc", API: "passAsGrpc", Kind: "bool"}, {TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "tls", API: "tls", Kind: "object", Obj: []objField{{TF: "ca_certificate_id", API: "caCertificateId", Kind: "string"}, {TF: "m_tls_certificate_id", API: "mTlsCertificateId", Kind: "string"}, {TF: "sni_server_name", API: "sniServerName", Kind: "string"}, {TF: "verify", API: "verify", Kind: "bool"}}}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

var loadbalancerHttpsListenerRuleMatchObjFields = []objField{{TF: "path", API: "path", Kind: "string"}, {TF: "path_match_type", API: "pathMatchType", Kind: "string"}}

var loadbalancerHttpsListenerRulePathRewriteActionObjFields = []objField{{TF: "destination_path", API: "destinationPath", Kind: "string"}, {TF: "path_type", API: "pathType", Kind: "string"}, {TF: "source_path", API: "sourcePath", Kind: "string"}}

var loadbalancerHttpsListenerRuleSetRequestHeadersActionObjFields = []objField{{TF: "headers", API: "headers", Kind: "map_string"}}

var loadbalancerHttpsListenerRuleSetResponseHeadersActionObjFields = []objField{{TF: "headers", API: "headers", Kind: "map_string"}}

var loadbalancerHttpsListenerRuleStaticResponseActionObjFields = []objField{{TF: "body_string", API: "bodyString", Kind: "string"}, {TF: "content_type", API: "contentType", Kind: "string"}, {TF: "headers", API: "headers", Kind: "map_string"}, {TF: "status_code", API: "statusCode", Kind: "int64"}}

type LoadbalancerHttpsListenerRuleSpecModel struct {
	DeleteRequestHeadersAction   types.Object `tfsdk:"delete_request_headers_action"`
	DeleteResponseHeadersAction  types.Object `tfsdk:"delete_response_headers_action"`
	ForwardToHttpResponseAction  types.Object `tfsdk:"forward_to_http_response_action"`
	ForwardToHttpsResponseAction types.Object `tfsdk:"forward_to_https_response_action"`
	HttpsListenerId              types.String `tfsdk:"https_listener_id"`
	Match                        types.Object `tfsdk:"match"`
	Order                        types.Int64  `tfsdk:"order"`
	PathRewriteAction            types.Object `tfsdk:"path_rewrite_action"`
	SetRequestHeadersAction      types.Object `tfsdk:"set_request_headers_action"`
	SetResponseHeadersAction     types.Object `tfsdk:"set_response_headers_action"`
	StaticResponseAction         types.Object `tfsdk:"static_response_action"`
}

type LoadbalancerHttpsListenerRuleResourceModel struct {
	ID       types.String                           `tfsdk:"id"`
	Metadata metadataModel                          `tfsdk:"metadata"`
	Spec     LoadbalancerHttpsListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                           `tfsdk:"status"`
}

type LoadbalancerHttpsListenerRuleResource struct{ client *client.Client }

func NewLoadbalancerHttpsListenerRuleResource() resource.Resource {
	return &LoadbalancerHttpsListenerRuleResource{}
}

func (r *LoadbalancerHttpsListenerRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_https_listener_rule"
}

func LoadbalancerHttpsListenerRuleResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"delete_request_headers_action":    objResourceSchema(loadbalancerHttpsListenerRuleDeleteRequestHeadersActionObjFields),
		"delete_response_headers_action":   objResourceSchema(loadbalancerHttpsListenerRuleDeleteResponseHeadersActionObjFields),
		"forward_to_http_response_action":  objResourceSchema(loadbalancerHttpsListenerRuleForwardToHttpResponseActionObjFields),
		"forward_to_https_response_action": objResourceSchema(loadbalancerHttpsListenerRuleForwardToHttpsResponseActionObjFields),
		"https_listener_id":                schema.StringAttribute{Required: true},
		"match":                            objResourceSchema(loadbalancerHttpsListenerRuleMatchObjFields),
		"order":                            schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"path_rewrite_action":              objResourceSchema(loadbalancerHttpsListenerRulePathRewriteActionObjFields),
		"set_request_headers_action":       objResourceSchema(loadbalancerHttpsListenerRuleSetRequestHeadersActionObjFields),
		"set_response_headers_action":      objResourceSchema(loadbalancerHttpsListenerRuleSetResponseHeadersActionObjFields),
		"static_response_action":           objResourceSchema(loadbalancerHttpsListenerRuleStaticResponseActionObjFields),
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *LoadbalancerHttpsListenerRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: LoadbalancerHttpsListenerRuleResourceSchemaAttrs()}
}

func (r *LoadbalancerHttpsListenerRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	r.client = pd.Client
}

func buildLoadbalancerHttpsListenerRuleRequestMap(ctx context.Context, plan LoadbalancerHttpsListenerRuleResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.DeleteRequestHeadersAction.IsNull() && !plan.Spec.DeleteRequestHeadersAction.IsUnknown() {
		spec["deleteRequestHeadersAction"] = objToAPI(plan.Spec.DeleteRequestHeadersAction, loadbalancerHttpsListenerRuleDeleteRequestHeadersActionObjFields)
	}
	if !plan.Spec.DeleteResponseHeadersAction.IsNull() && !plan.Spec.DeleteResponseHeadersAction.IsUnknown() {
		spec["deleteResponseHeadersAction"] = objToAPI(plan.Spec.DeleteResponseHeadersAction, loadbalancerHttpsListenerRuleDeleteResponseHeadersActionObjFields)
	}
	if !plan.Spec.ForwardToHttpResponseAction.IsNull() && !plan.Spec.ForwardToHttpResponseAction.IsUnknown() {
		spec["forwardToHttpResponseAction"] = objToAPI(plan.Spec.ForwardToHttpResponseAction, loadbalancerHttpsListenerRuleForwardToHttpResponseActionObjFields)
	}
	if !plan.Spec.ForwardToHttpsResponseAction.IsNull() && !plan.Spec.ForwardToHttpsResponseAction.IsUnknown() {
		spec["forwardToHttpsResponseAction"] = objToAPI(plan.Spec.ForwardToHttpsResponseAction, loadbalancerHttpsListenerRuleForwardToHttpsResponseActionObjFields)
	}
	if !plan.Spec.HttpsListenerId.IsNull() && !plan.Spec.HttpsListenerId.IsUnknown() {
		spec["httpsListenerId"] = plan.Spec.HttpsListenerId.ValueString()
	}
	if !plan.Spec.Match.IsNull() && !plan.Spec.Match.IsUnknown() {
		spec["match"] = objToAPI(plan.Spec.Match, loadbalancerHttpsListenerRuleMatchObjFields)
	}
	if !plan.Spec.Order.IsNull() && !plan.Spec.Order.IsUnknown() {
		spec["order"] = plan.Spec.Order.ValueInt64()
	}
	if !plan.Spec.PathRewriteAction.IsNull() && !plan.Spec.PathRewriteAction.IsUnknown() {
		spec["pathRewriteAction"] = objToAPI(plan.Spec.PathRewriteAction, loadbalancerHttpsListenerRulePathRewriteActionObjFields)
	}
	if !plan.Spec.SetRequestHeadersAction.IsNull() && !plan.Spec.SetRequestHeadersAction.IsUnknown() {
		spec["setRequestHeadersAction"] = objToAPI(plan.Spec.SetRequestHeadersAction, loadbalancerHttpsListenerRuleSetRequestHeadersActionObjFields)
	}
	if !plan.Spec.SetResponseHeadersAction.IsNull() && !plan.Spec.SetResponseHeadersAction.IsUnknown() {
		spec["setResponseHeadersAction"] = objToAPI(plan.Spec.SetResponseHeadersAction, loadbalancerHttpsListenerRuleSetResponseHeadersActionObjFields)
	}
	if !plan.Spec.StaticResponseAction.IsNull() && !plan.Spec.StaticResponseAction.IsUnknown() {
		spec["staticResponseAction"] = objToAPI(plan.Spec.StaticResponseAction, loadbalancerHttpsListenerRuleStaticResponseActionObjFields)
	}
	return m
}

func populateLoadbalancerHttpsListenerRuleState(ctx context.Context, data map[string]interface{}, state *LoadbalancerHttpsListenerRuleResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
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
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerHttpsListenerRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerHttpsListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildLoadbalancerHttpsListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-https-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-https-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener-rule", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateLoadbalancerHttpsListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerHttpsListenerRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerHttpsListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateLoadbalancerHttpsListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *LoadbalancerHttpsListenerRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadbalancerHttpsListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildLoadbalancerHttpsListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-https-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-https-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener-rule", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateLoadbalancerHttpsListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerHttpsListenerRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerHttpsListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/loadbalancer-https-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-https-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *LoadbalancerHttpsListenerRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state LoadbalancerHttpsListenerRuleResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener-rule", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateLoadbalancerHttpsListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
