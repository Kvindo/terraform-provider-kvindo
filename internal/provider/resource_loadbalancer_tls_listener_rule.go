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

var loadbalancerTlsListenerRuleForwardToTcpResponseActionObjFields = []objField{{TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

var loadbalancerTlsListenerRuleForwardToTlsResponseActionObjFields = []objField{{TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "tls", API: "tls", Kind: "object", Obj: []objField{{TF: "ca_certificate_id", API: "caCertificateId", Kind: "string"}, {TF: "m_tls_certificate_id", API: "mTlsCertificateId", Kind: "string"}, {TF: "sni_server_name", API: "sniServerName", Kind: "string"}, {TF: "verify", API: "verify", Kind: "bool"}}}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

type LoadbalancerTlsListenerRuleSpecModel struct {
	ForwardToTcpResponseAction types.Object `tfsdk:"forward_to_tcp_response_action"`
	ForwardToTlsResponseAction types.Object `tfsdk:"forward_to_tls_response_action"`
	Order                      types.Int64  `tfsdk:"order"`
	TlsListenerId              types.String `tfsdk:"tls_listener_id"`
}

type LoadbalancerTlsListenerRuleResourceModel struct {
	ID       types.String                         `tfsdk:"id"`
	Metadata metadataModel                        `tfsdk:"metadata"`
	Spec     LoadbalancerTlsListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                         `tfsdk:"status"`
}

type LoadbalancerTlsListenerRuleResource struct{ client *client.Client }

func NewLoadbalancerTlsListenerRuleResource() resource.Resource {
	return &LoadbalancerTlsListenerRuleResource{}
}

func (r *LoadbalancerTlsListenerRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tls_listener_rule"
}

func LoadbalancerTlsListenerRuleResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"forward_to_tcp_response_action": objResourceSchema(loadbalancerTlsListenerRuleForwardToTcpResponseActionObjFields),
		"forward_to_tls_response_action": objResourceSchema(loadbalancerTlsListenerRuleForwardToTlsResponseActionObjFields),
		"order":                          schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"tls_listener_id":                schema.StringAttribute{Required: true},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *LoadbalancerTlsListenerRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: LoadbalancerTlsListenerRuleResourceSchemaAttrs()}
}

func (r *LoadbalancerTlsListenerRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildLoadbalancerTlsListenerRuleRequestMap(ctx context.Context, plan LoadbalancerTlsListenerRuleResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ForwardToTcpResponseAction.IsNull() && !plan.Spec.ForwardToTcpResponseAction.IsUnknown() {
		spec["forwardToTcpResponseAction"] = objToAPI(plan.Spec.ForwardToTcpResponseAction, loadbalancerTlsListenerRuleForwardToTcpResponseActionObjFields)
	}
	if !plan.Spec.ForwardToTlsResponseAction.IsNull() && !plan.Spec.ForwardToTlsResponseAction.IsUnknown() {
		spec["forwardToTlsResponseAction"] = objToAPI(plan.Spec.ForwardToTlsResponseAction, loadbalancerTlsListenerRuleForwardToTlsResponseActionObjFields)
	}
	if !plan.Spec.Order.IsNull() && !plan.Spec.Order.IsUnknown() {
		spec["order"] = plan.Spec.Order.ValueInt64()
	}
	if !plan.Spec.TlsListenerId.IsNull() && !plan.Spec.TlsListenerId.IsUnknown() {
		spec["tlsListenerId"] = plan.Spec.TlsListenerId.ValueString()
	}
	return m
}

func populateLoadbalancerTlsListenerRuleState(ctx context.Context, data map[string]interface{}, state *LoadbalancerTlsListenerRuleResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ForwardToTcpResponseAction = objFromAPI(objMap(spec, "forwardToTcpResponseAction"), loadbalancerTlsListenerRuleForwardToTcpResponseActionObjFields)
	state.Spec.ForwardToTlsResponseAction = objFromAPI(objMap(spec, "forwardToTlsResponseAction"), loadbalancerTlsListenerRuleForwardToTlsResponseActionObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.TlsListenerId = getString(spec, "tlsListenerId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerTlsListenerRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerTlsListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildLoadbalancerTlsListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-tls-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-tls-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tls-listener-rule", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateLoadbalancerTlsListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerTlsListenerRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerTlsListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tls-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateLoadbalancerTlsListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *LoadbalancerTlsListenerRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadbalancerTlsListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildLoadbalancerTlsListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-tls-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-tls-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tls-listener-rule", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateLoadbalancerTlsListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerTlsListenerRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerTlsListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/loadbalancer-tls-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-tls-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *LoadbalancerTlsListenerRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state LoadbalancerTlsListenerRuleResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tls-listener-rule", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateLoadbalancerTlsListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
