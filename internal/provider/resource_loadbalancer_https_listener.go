package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// LoadbalancerHttpsListenerResourceModel describes the resource data model.
type LoadbalancerHttpsListenerResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	LoadbalancerId types.String `tfsdk:"loadbalancer_id"`
	Interface types.String `tfsdk:"interface"`
	Order types.Int64 `tfsdk:"order"`
	Ports types.List `tfsdk:"ports"`
	Hosts types.List `tfsdk:"hosts"`
	EnableHttp2Support types.Bool `tfsdk:"enable_http2_support"`
	TlsCertificateId types.String `tfsdk:"tls_certificate_id"`
	TlsProtocols types.List `tfsdk:"tls_protocols"`
	TlsAutogenerateCertificate types.Bool `tfsdk:"tls_autogenerate_certificate"`
	Info types.Object `tfsdk:"info"`
}

// LoadbalancerHttpsListenerResource defines the resource implementation.
type LoadbalancerHttpsListenerResource struct {
	client *client.Client
}

func NewLoadbalancerHttpsListenerResource() resource.Resource {
	return &LoadbalancerHttpsListenerResource{}
}

func (r *LoadbalancerHttpsListenerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_https_listener"
}

func (r *LoadbalancerHttpsListenerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["loadbalancer_id"] = schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["interface"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["order"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["ports"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["hosts"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["enable_http2_support"] = schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	attrs["tls_certificate_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["tls_protocols"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["tls_autogenerate_certificate"] = schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.LoadbalancerId.IsNull() && !plan.LoadbalancerId.IsUnknown() {
		m["loadbalancerId"] = plan.LoadbalancerId.ValueString()
	}
	if !plan.Interface.IsNull() && !plan.Interface.IsUnknown() {
		m["interface"] = plan.Interface.ValueString()
	}
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		m["order"] = plan.Order.ValueInt64()
	}
	if !plan.Ports.IsNull() && !plan.Ports.IsUnknown() {
		m["ports"] = stringListToInterface(ctx, plan.Ports)
	}
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		m["hosts"] = stringListToInterface(ctx, plan.Hosts)
	}
	if !plan.EnableHttp2Support.IsNull() && !plan.EnableHttp2Support.IsUnknown() {
		m["enableHttp2Support"] = plan.EnableHttp2Support.ValueBool()
	}
	if !plan.TlsCertificateId.IsNull() && !plan.TlsCertificateId.IsUnknown() {
		m["tlsCertificateId"] = plan.TlsCertificateId.ValueString()
	}
	if !plan.TlsProtocols.IsNull() && !plan.TlsProtocols.IsUnknown() {
		m["tlsProtocols"] = stringListToInterface(ctx, plan.TlsProtocols)
	}
	if !plan.TlsAutogenerateCertificate.IsNull() && !plan.TlsAutogenerateCertificate.IsUnknown() {
		m["tlsAutogenerateCertificate"] = plan.TlsAutogenerateCertificate.ValueBool()
	}
	return m
}

func populateLoadbalancerHttpsListenerState(ctx context.Context, data map[string]interface{}, state *LoadbalancerHttpsListenerResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.LoadbalancerId = getString(data, "loadbalancerId")
	state.Interface = getString(data, "interface")
	state.Order = getInt64(data, "order")
	state.Ports = getStringList(ctx, data, "ports")
	state.Hosts = getStringList(ctx, data, "hosts")
	state.EnableHttp2Support = getBool(data, "enableHttp2Support")
	state.TlsCertificateId = getString(data, "tlsCertificateId")
	state.TlsProtocols = getStringList(ctx, data, "tlsProtocols")
	state.TlsAutogenerateCertificate = getBool(data, "tlsAutogenerateCertificate")
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *LoadbalancerHttpsListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadbalancerHttpsListenerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LoadbalancerHttpsListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadbalancerHttpsListenerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *LoadbalancerHttpsListenerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LoadbalancerHttpsListenerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state LoadbalancerHttpsListenerResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateLoadbalancerHttpsListenerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LoadbalancerHttpsListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadbalancerHttpsListenerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	// Import by ID
	var state LoadbalancerHttpsListenerResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/loadbalancer-https-listener", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateLoadbalancerHttpsListenerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
