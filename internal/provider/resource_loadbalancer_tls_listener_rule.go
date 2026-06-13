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
// attr package used for list/object types

// LoadbalancerTlsListenerRuleResourceModel describes the resource data model.
type LoadbalancerTlsListenerRuleResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TlsListenerId types.String `tfsdk:"tls_listener_id"`
	Order types.Int64 `tfsdk:"order"`
	ActionType types.String `tfsdk:"action_type"`
	ActionJson types.String `tfsdk:"action_json"`
	Info types.Object `tfsdk:"info"`
}

// LoadbalancerTlsListenerRuleResource defines the resource implementation.
type LoadbalancerTlsListenerRuleResource struct {
	client *client.Client
}

func NewLoadbalancerTlsListenerRuleResource() resource.Resource {
	return &LoadbalancerTlsListenerRuleResource{}
}

func (r *LoadbalancerTlsListenerRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tls_listener_rule"
}

func (r *LoadbalancerTlsListenerRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["tls_listener_id"] = schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["order"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["action_type"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["action_json"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.TlsListenerId.IsNull() && !plan.TlsListenerId.IsUnknown() {
		m["tlsListenerId"] = plan.TlsListenerId.ValueString()
	}
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		m["order"] = plan.Order.ValueInt64()
	}
	if !plan.ActionType.IsNull() && !plan.ActionType.IsUnknown() {
		m["actionType"] = plan.ActionType.ValueString()
	}
	if !plan.ActionJson.IsNull() && !plan.ActionJson.IsUnknown() {
		m["actionJson"] = plan.ActionJson.ValueString()
	}
	return m
}

func populateLoadbalancerTlsListenerRuleState(ctx context.Context, data map[string]interface{}, state *LoadbalancerTlsListenerRuleResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.TlsListenerId = getString(data, "tlsListenerId")
	state.Order = getInt64(data, "order")
	state.ActionType = getString(data, "actionType")
	state.ActionJson = getString(data, "actionJson")
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerTlsListenerRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerTlsListenerRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LoadbalancerTlsListenerRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerTlsListenerRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *LoadbalancerTlsListenerRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LoadbalancerTlsListenerRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state LoadbalancerTlsListenerRuleResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateLoadbalancerTlsListenerRuleState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LoadbalancerTlsListenerRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerTlsListenerRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	// Import by ID
	var state LoadbalancerTlsListenerRuleResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-tls-listener-rule", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateLoadbalancerTlsListenerRuleState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
