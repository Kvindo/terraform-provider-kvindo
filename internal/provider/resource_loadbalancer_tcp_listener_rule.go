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

var loadbalancerTcpListenerRuleForwardToTcpResponseActionObjFields = []objField{{TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

var loadbalancerTcpListenerRuleForwardToTlsResponseActionObjFields = []objField{{TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "tls", API: "tls", Kind: "object", Obj: []objField{{TF: "ca_certificate_id", API: "caCertificateId", Kind: "string"}, {TF: "m_tls_certificate_id", API: "mTlsCertificateId", Kind: "string"}, {TF: "sni_server_name", API: "sniServerName", Kind: "string"}, {TF: "verify", API: "verify", Kind: "bool"}}}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

type LoadbalancerTcpListenerRuleSpecModel struct {
	ForwardToTcpResponseAction types.Object `tfsdk:"forward_to_tcp_response_action"`
	ForwardToTlsResponseAction types.Object `tfsdk:"forward_to_tls_response_action"`
	Order                      types.Int64  `tfsdk:"order"`
	TcpListenerId              types.String `tfsdk:"tcp_listener_id"`
}

type LoadbalancerTcpListenerRuleResourceModel struct {
	ID       types.String                         `tfsdk:"id"`
	Metadata metadataModel                        `tfsdk:"metadata"`
	Spec     LoadbalancerTcpListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                         `tfsdk:"status"`
}

type LoadbalancerTcpListenerRuleResource struct{ client *client.Client }

func NewLoadbalancerTcpListenerRuleResource() resource.Resource {
	return &LoadbalancerTcpListenerRuleResource{}
}

func (r *LoadbalancerTcpListenerRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tcp_listener_rule"
}

func LoadbalancerTcpListenerRuleResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"forward_to_tcp_response_action": objResourceSchema(loadbalancerTcpListenerRuleForwardToTcpResponseActionObjFields),
		"forward_to_tls_response_action": objResourceSchema(loadbalancerTcpListenerRuleForwardToTlsResponseActionObjFields),
		"order":                          schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"tcp_listener_id":                schema.StringAttribute{Required: true},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *LoadbalancerTcpListenerRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: LoadbalancerTcpListenerRuleResourceSchemaAttrs()}
}

func (r *LoadbalancerTcpListenerRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildLoadbalancerTcpListenerRuleRequestMap(ctx context.Context, plan LoadbalancerTcpListenerRuleResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ForwardToTcpResponseAction.IsNull() && !plan.Spec.ForwardToTcpResponseAction.IsUnknown() {
		spec["forwardToTcpResponseAction"] = objToAPI(plan.Spec.ForwardToTcpResponseAction, loadbalancerTcpListenerRuleForwardToTcpResponseActionObjFields)
	}
	if !plan.Spec.ForwardToTlsResponseAction.IsNull() && !plan.Spec.ForwardToTlsResponseAction.IsUnknown() {
		spec["forwardToTlsResponseAction"] = objToAPI(plan.Spec.ForwardToTlsResponseAction, loadbalancerTcpListenerRuleForwardToTlsResponseActionObjFields)
	}
	if !plan.Spec.Order.IsNull() && !plan.Spec.Order.IsUnknown() {
		spec["order"] = plan.Spec.Order.ValueInt64()
	}
	if !plan.Spec.TcpListenerId.IsNull() && !plan.Spec.TcpListenerId.IsUnknown() {
		spec["tcpListenerId"] = plan.Spec.TcpListenerId.ValueString()
	}
	return m
}

func populateLoadbalancerTcpListenerRuleState(ctx context.Context, data map[string]interface{}, state *LoadbalancerTcpListenerRuleResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ForwardToTcpResponseAction = objFromAPI(objMap(spec, "forwardToTcpResponseAction"), loadbalancerTcpListenerRuleForwardToTcpResponseActionObjFields)
	state.Spec.ForwardToTlsResponseAction = objFromAPI(objMap(spec, "forwardToTlsResponseAction"), loadbalancerTcpListenerRuleForwardToTlsResponseActionObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.TcpListenerId = getString(spec, "tcpListenerId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerTcpListenerRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerTcpListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildLoadbalancerTcpListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-tcp-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-tcp-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener-rule", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateLoadbalancerTcpListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerTcpListenerRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerTcpListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateLoadbalancerTcpListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *LoadbalancerTcpListenerRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadbalancerTcpListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildLoadbalancerTcpListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-tcp-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-tcp-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener-rule", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateLoadbalancerTcpListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerTcpListenerRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerTcpListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/loadbalancer-tcp-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-tcp-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *LoadbalancerTcpListenerRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state LoadbalancerTcpListenerRuleResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener-rule", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateLoadbalancerTcpListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
