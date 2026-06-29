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

var loadbalancerUdpListenerRuleForwardToUdpResponseActionObjFields = []objField{{TF: "port_mapping_type", API: "portMappingType", Kind: "string"}, {TF: "target_group_id", API: "targetGroupId", Kind: "string"}, {TF: "to_ports", API: "toPorts", Kind: "list_string"}}

type LoadbalancerUdpListenerRuleSpecModel struct {
	ForwardToUdpResponseAction types.Object `tfsdk:"forward_to_udp_response_action"`
	Order                      types.Int64  `tfsdk:"order"`
	UdpListenerId              types.String `tfsdk:"udp_listener_id"`
}

type LoadbalancerUdpListenerRuleResourceModel struct {
	ID       types.String                         `tfsdk:"id"`
	Metadata metadataModel                        `tfsdk:"metadata"`
	Spec     LoadbalancerUdpListenerRuleSpecModel `tfsdk:"spec"`
	Status   types.Object                         `tfsdk:"status"`
}

type LoadbalancerUdpListenerRuleResource struct{ client *client.Client }

func NewLoadbalancerUdpListenerRuleResource() resource.Resource {
	return &LoadbalancerUdpListenerRuleResource{}
}

func (r *LoadbalancerUdpListenerRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_udp_listener_rule"
}

func LoadbalancerUdpListenerRuleResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"forward_to_udp_response_action": objResourceSchema(loadbalancerUdpListenerRuleForwardToUdpResponseActionObjFields),
		"order":                          schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"udp_listener_id":                schema.StringAttribute{Required: true},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *LoadbalancerUdpListenerRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: LoadbalancerUdpListenerRuleResourceSchemaAttrs()}
}

func (r *LoadbalancerUdpListenerRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildLoadbalancerUdpListenerRuleRequestMap(ctx context.Context, plan LoadbalancerUdpListenerRuleResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ForwardToUdpResponseAction.IsNull() && !plan.Spec.ForwardToUdpResponseAction.IsUnknown() {
		spec["forwardToUdpResponseAction"] = objToAPI(plan.Spec.ForwardToUdpResponseAction, loadbalancerUdpListenerRuleForwardToUdpResponseActionObjFields)
	}
	if !plan.Spec.Order.IsNull() && !plan.Spec.Order.IsUnknown() {
		spec["order"] = plan.Spec.Order.ValueInt64()
	}
	if !plan.Spec.UdpListenerId.IsNull() && !plan.Spec.UdpListenerId.IsUnknown() {
		spec["udpListenerId"] = plan.Spec.UdpListenerId.ValueString()
	}
	return m
}

func populateLoadbalancerUdpListenerRuleState(ctx context.Context, data map[string]interface{}, state *LoadbalancerUdpListenerRuleResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ForwardToUdpResponseAction = objFromAPI(objMap(spec, "forwardToUdpResponseAction"), loadbalancerUdpListenerRuleForwardToUdpResponseActionObjFields)
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.UdpListenerId = getString(spec, "udpListenerId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerUdpListenerRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerUdpListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildLoadbalancerUdpListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-udp-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-udp-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-udp-listener-rule", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateLoadbalancerUdpListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerUdpListenerRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerUdpListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-udp-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateLoadbalancerUdpListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *LoadbalancerUdpListenerRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadbalancerUdpListenerRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildLoadbalancerUdpListenerRuleRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-udp-listener-rule", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-udp-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-udp-listener-rule", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateLoadbalancerUdpListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerUdpListenerRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerUdpListenerRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/loadbalancer-udp-listener-rule", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-udp-listener-rule", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *LoadbalancerUdpListenerRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state LoadbalancerUdpListenerRuleResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-udp-listener-rule", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateLoadbalancerUdpListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
