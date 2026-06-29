package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type LoadbalancerTargetGroupServiceDiscoveryTargetSpecModel struct {
	LabelSelectors types.Map    `tfsdk:"label_selectors"`
	TargetGroupId  types.String `tfsdk:"target_group_id"`
}

type LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel struct {
	ID       types.String                                           `tfsdk:"id"`
	Metadata metadataModel                                          `tfsdk:"metadata"`
	Spec     LoadbalancerTargetGroupServiceDiscoveryTargetSpecModel `tfsdk:"spec"`
	Status   types.Object                                           `tfsdk:"status"`
}

type LoadbalancerTargetGroupServiceDiscoveryTargetResource struct{ client *client.Client }

func NewLoadbalancerTargetGroupServiceDiscoveryTargetResource() resource.Resource {
	return &LoadbalancerTargetGroupServiceDiscoveryTargetResource{}
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_target_group_service_discovery_target"
}

func LoadbalancerTargetGroupServiceDiscoveryTargetResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"label_selectors": schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"target_group_id": schema.StringAttribute{Required: true},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: LoadbalancerTargetGroupServiceDiscoveryTargetResourceSchemaAttrs()}
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildLoadbalancerTargetGroupServiceDiscoveryTargetRequestMap(ctx context.Context, plan LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.LabelSelectors.IsNull() && !plan.Spec.LabelSelectors.IsUnknown() {
		spec["labelSelectors"] = stringMapToInterface(ctx, plan.Spec.LabelSelectors)
	}
	if !plan.Spec.TargetGroupId.IsNull() && !plan.Spec.TargetGroupId.IsUnknown() {
		spec["targetGroupId"] = plan.Spec.TargetGroupId.ValueString()
	}
	return m
}

func populateLoadbalancerTargetGroupServiceDiscoveryTargetState(ctx context.Context, data map[string]interface{}, state *LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.LabelSelectors = getStringMap(spec, "labelSelectors")
	state.Spec.TargetGroupId = getString(spec, "targetGroupId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildLoadbalancerTargetGroupServiceDiscoveryTargetRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateLoadbalancerTargetGroupServiceDiscoveryTargetState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateLoadbalancerTargetGroupServiceDiscoveryTargetState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildLoadbalancerTargetGroupServiceDiscoveryTargetRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateLoadbalancerTargetGroupServiceDiscoveryTargetState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *LoadbalancerTargetGroupServiceDiscoveryTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-target-group-service-discovery-target", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateLoadbalancerTargetGroupServiceDiscoveryTargetState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
