package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// VictoriaMetricsResourceModel describes the resource data model.
type VictoriaMetricsResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	VpcId types.String `tfsdk:"vpc_id"`
	CreatePublicIpv4 types.Bool `tfsdk:"create_public_ipv4"`
	CreatePublicIpv6 types.Bool `tfsdk:"create_public_ipv6"`
	DnsRecordName types.String `tfsdk:"dns_record_name"`
	Info types.Object `tfsdk:"info"`
}

// VictoriaMetricsResource defines the resource implementation.
type VictoriaMetricsResource struct {
	client *client.Client
}

func NewVictoriaMetricsResource() resource.Resource {
	return &VictoriaMetricsResource{}
}

func (r *VictoriaMetricsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_victoria_metrics"
}

func (r *VictoriaMetricsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vpc_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["create_public_ipv4"] = schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	attrs["create_public_ipv6"] = schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	attrs["dns_record_name"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "public_ip_v6": schema.StringAttribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "private_ip_v6": schema.StringAttribute{Computed: true}, "fqdn": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *VictoriaMetricsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildVictoriaMetricsRequestMap(ctx context.Context, plan VictoriaMetricsResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.Tier.IsNull() && !plan.Tier.IsUnknown() {
		m["tier"] = plan.Tier.ValueString()
	}
	if !plan.VpcId.IsNull() && !plan.VpcId.IsUnknown() {
		m["vpcId"] = plan.VpcId.ValueString()
	}
	if !plan.CreatePublicIpv4.IsNull() && !plan.CreatePublicIpv4.IsUnknown() {
		m["createPublicIpv4"] = plan.CreatePublicIpv4.ValueBool()
	}
	if !plan.CreatePublicIpv6.IsNull() && !plan.CreatePublicIpv6.IsUnknown() {
		m["createPublicIpv6"] = plan.CreatePublicIpv6.ValueBool()
	}
	if !plan.DnsRecordName.IsNull() && !plan.DnsRecordName.IsUnknown() {
		m["dnsRecordName"] = plan.DnsRecordName.ValueString()
	}
	return m
}

func populateVictoriaMetricsState(ctx context.Context, data map[string]interface{}, state *VictoriaMetricsResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.Tier = getString(data, "tier")
	state.VpcId = getString(data, "vpcId")
	state.CreatePublicIpv4 = getBool(data, "createPublicIpv4")
	state.CreatePublicIpv6 = getBool(data, "createPublicIpv6")
	state.DnsRecordName = getString(data, "dnsRecordName")
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "public_ip_v4": types.StringType, "public_ip_v6": types.StringType, "private_ip_v4": types.StringType, "private_ip_v6": types.StringType, "fqdn": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "public_ip_v4": getStringFromInfo(data, "publicipv4"), "public_ip_v6": getStringFromInfo(data, "publicipv6"), "private_ip_v4": getStringFromInfo(data, "privateipv4"), "private_ip_v6": getStringFromInfo(data, "privateipv6"), "fqdn": getStringFromInfo(data, "fqdn")})
	return nil
}

func (r *VictoriaMetricsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VictoriaMetricsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildVictoriaMetricsRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/victoria-metrics", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/victoria-metrics", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/victoria-metrics", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateVictoriaMetricsState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VictoriaMetricsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VictoriaMetricsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/victoria-metrics", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateVictoriaMetricsState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VictoriaMetricsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VictoriaMetricsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state VictoriaMetricsResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildVictoriaMetricsRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/victoria-metrics", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/victoria-metrics", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/victoria-metrics", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateVictoriaMetricsState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VictoriaMetricsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VictoriaMetricsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/victoria-metrics", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/victoria-metrics", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *VictoriaMetricsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state VictoriaMetricsResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/victoria-metrics", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateVictoriaMetricsState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
