package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var loadbalancerHttpsListenerSecurityRulesObjFields = []objField{{TF: "action", API: "action", Kind: "string"}, {TF: "description", API: "description", Kind: "string"}, {TF: "ipv4_blocks", API: "ipV4Blocks", Kind: "list_string"}, {TF: "ipv6_blocks", API: "ipV6Blocks", Kind: "list_string"}, {TF: "order", API: "order", Kind: "int64"}}

var loadbalancerHttpsListenerTlsObjFields = []objField{{TF: "autogenerate_certificate", API: "autogenerateCertificate", Kind: "bool"}, {TF: "certificate_id", API: "certificateId", Kind: "string"}, {TF: "protocols", API: "protocols", Kind: "list_string"}}

type LoadbalancerHttpsListenerSpecModel struct {
	EnableHttp2Support types.Bool   `tfsdk:"enable_http2_support"`
	Hosts              types.List   `tfsdk:"hosts"`
	Interface          types.String `tfsdk:"interface"`
	LoadbalancerId     types.String `tfsdk:"loadbalancer_id"`
	Order              types.Int64  `tfsdk:"order"`
	Ports              types.List   `tfsdk:"ports"`
	SecurityRules      types.List   `tfsdk:"security_rules"`
	Tls                types.Object `tfsdk:"tls"`
}

type LoadbalancerHttpsListenerResourceModel struct {
	ID       types.String                       `tfsdk:"id"`
	Metadata metadataModel                      `tfsdk:"metadata"`
	Spec     LoadbalancerHttpsListenerSpecModel `tfsdk:"spec"`
	Status   types.Object                       `tfsdk:"status"`
}

type LoadbalancerHttpsListenerResource struct{ client *client.Client }

func NewLoadbalancerHttpsListenerResource() resource.Resource {
	return &LoadbalancerHttpsListenerResource{}
}

func (r *LoadbalancerHttpsListenerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_https_listener"
}

func LoadbalancerHttpsListenerResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"enable_http2_support": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"hosts":                schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"interface":            schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"loadbalancer_id":      schema.StringAttribute{Required: true},
		"order":                schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"ports":                schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"security_rules":       listObjResourceSchema(loadbalancerHttpsListenerSecurityRulesObjFields),
		"tls":                  objResourceSchema(loadbalancerHttpsListenerTlsObjFields),
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *LoadbalancerHttpsListenerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: LoadbalancerHttpsListenerResourceSchemaAttrs()}
}

func (r *LoadbalancerHttpsListenerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildLoadbalancerHttpsListenerRequestMap(ctx context.Context, plan LoadbalancerHttpsListenerResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.EnableHttp2Support.IsNull() && !plan.Spec.EnableHttp2Support.IsUnknown() {
		spec["enableHttp2Support"] = plan.Spec.EnableHttp2Support.ValueBool()
	}
	if !plan.Spec.Hosts.IsNull() && !plan.Spec.Hosts.IsUnknown() {
		spec["hosts"] = stringListToInterface(ctx, plan.Spec.Hosts)
	}
	if !plan.Spec.Interface.IsNull() && !plan.Spec.Interface.IsUnknown() {
		spec["interface"] = plan.Spec.Interface.ValueString()
	}
	if !plan.Spec.LoadbalancerId.IsNull() && !plan.Spec.LoadbalancerId.IsUnknown() {
		spec["loadbalancerId"] = plan.Spec.LoadbalancerId.ValueString()
	}
	if !plan.Spec.Order.IsNull() && !plan.Spec.Order.IsUnknown() {
		spec["order"] = plan.Spec.Order.ValueInt64()
	}
	if !plan.Spec.Ports.IsNull() && !plan.Spec.Ports.IsUnknown() {
		spec["ports"] = stringListToInterface(ctx, plan.Spec.Ports)
	}
	if !plan.Spec.SecurityRules.IsNull() && !plan.Spec.SecurityRules.IsUnknown() {
		spec["securityRules"] = listObjToAPI(plan.Spec.SecurityRules, loadbalancerHttpsListenerSecurityRulesObjFields)
	}
	if !plan.Spec.Tls.IsNull() && !plan.Spec.Tls.IsUnknown() {
		spec["tls"] = objToAPI(plan.Spec.Tls, loadbalancerHttpsListenerTlsObjFields)
	}
	return m
}

func populateLoadbalancerHttpsListenerState(ctx context.Context, data map[string]interface{}, state *LoadbalancerHttpsListenerResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.EnableHttp2Support = getBool(spec, "enableHttp2Support")
	state.Spec.Hosts = getStringList(ctx, spec, "hosts")
	state.Spec.Interface = getString(spec, "interface")
	state.Spec.LoadbalancerId = getString(spec, "loadbalancerId")
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.Ports = getStringList(ctx, spec, "ports")
	state.Spec.SecurityRules = listObjFromAPI(objList(spec, "securityRules"), loadbalancerHttpsListenerSecurityRulesObjFields)
	state.Spec.Tls = objFromAPI(objMap(spec, "tls"), loadbalancerHttpsListenerTlsObjFields)
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerHttpsListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerHttpsListenerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildLoadbalancerHttpsListenerRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-https-listener", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-https-listener", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateLoadbalancerHttpsListenerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerHttpsListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerHttpsListenerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateLoadbalancerHttpsListenerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *LoadbalancerHttpsListenerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state LoadbalancerHttpsListenerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildLoadbalancerHttpsListenerRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/loadbalancer-https-listener", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-https-listener", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateLoadbalancerHttpsListenerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *LoadbalancerHttpsListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerHttpsListenerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/loadbalancer-https-listener", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/loadbalancer-https-listener", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *LoadbalancerHttpsListenerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state LoadbalancerHttpsListenerResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateLoadbalancerHttpsListenerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
