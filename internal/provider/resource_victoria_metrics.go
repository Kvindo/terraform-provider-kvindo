package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var victoriaMetricsConstraintsObjFields = []objField{{TF: "cpu", API: "cpu", Kind: "list_object", Obj: []objField{{TF: "architecture", API: "architecture", Kind: "string"}, {TF: "threads_count", API: "threadsCount", Kind: "string"}}}, {TF: "disk", API: "disk", Kind: "list_object", Obj: []objField{{TF: "read_iops", API: "readIops", Kind: "string"}, {TF: "read_throughput", API: "readThroughput", Kind: "string"}, {TF: "size", API: "size", Kind: "string"}, {TF: "write_iops", API: "writeIops", Kind: "string"}, {TF: "write_throughput", API: "writeThroughput", Kind: "string"}}}, {TF: "memory", API: "memory", Kind: "list_object", Obj: []objField{{TF: "memory", API: "memory", Kind: "string"}}}, {TF: "offer", API: "offer", Kind: "list_object", Obj: []objField{{TF: "offer_id", API: "offerId", Kind: "string"}}}}

var victoriaMetricsScrapTargetsObjFields = []objField{{TF: "job_name", API: "jobName", Kind: "string"}, {TF: "path", API: "path", Kind: "string"}, {TF: "port", API: "port", Kind: "int64"}, {TF: "target_ip_or_hostnames", API: "targetIpOrHostnames", Kind: "list_string"}}

type VictoriaMetricsSpecModel struct {
	Constraints      types.List   `tfsdk:"constraints"`
	CreatePublicIpv4 types.Bool   `tfsdk:"create_public_ipv4"`
	CreatePublicIpv6 types.Bool   `tfsdk:"create_public_ipv6"`
	DnsRecordName    types.String `tfsdk:"dns_record_name"`
	ScrapTargets     types.List   `tfsdk:"scrap_targets"`
	Tier             types.String `tfsdk:"tier"`
	VpcId            types.String `tfsdk:"vpc_id"`
}

type VictoriaMetricsResourceModel struct {
	ID       types.String             `tfsdk:"id"`
	Metadata metadataModel            `tfsdk:"metadata"`
	Spec     VictoriaMetricsSpecModel `tfsdk:"spec"`
	Status   types.Object             `tfsdk:"status"`
}

type VictoriaMetricsResource struct{ client *client.Client }

func NewVictoriaMetricsResource() resource.Resource { return &VictoriaMetricsResource{} }

func (r *VictoriaMetricsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_victoria_metrics"
}

func VictoriaMetricsResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"constraints":        listObjResourceSchema(victoriaMetricsConstraintsObjFields),
		"create_public_ipv4": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"create_public_ipv6": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"dns_record_name":    schema.StringAttribute{Optional: true},
		"scrap_targets":      listObjResourceSchema(victoriaMetricsScrapTargetsObjFields),
		"tier":               schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vpc_id":             schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"discovered_scrap_targets": schema.StringAttribute{Computed: true}, "fqdn": schema.StringAttribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "private_ip_v6": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "public_ip_v6": schema.StringAttribute{Computed: true}}),
	}
}

func (r *VictoriaMetricsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: VictoriaMetricsResourceSchemaAttrs()}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.Constraints.IsNull() && !plan.Spec.Constraints.IsUnknown() {
		spec["constraints"] = listObjToAPI(plan.Spec.Constraints, victoriaMetricsConstraintsObjFields)
	}
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.CreatePublicIpv6.IsNull() && !plan.Spec.CreatePublicIpv6.IsUnknown() {
		spec["createPublicIpv6"] = plan.Spec.CreatePublicIpv6.ValueBool()
	}
	if !plan.Spec.DnsRecordName.IsNull() && !plan.Spec.DnsRecordName.IsUnknown() {
		spec["dnsRecordName"] = plan.Spec.DnsRecordName.ValueString()
	}
	if !plan.Spec.ScrapTargets.IsNull() && !plan.Spec.ScrapTargets.IsUnknown() {
		spec["scrapTargets"] = listObjToAPI(plan.Spec.ScrapTargets, victoriaMetricsScrapTargetsObjFields)
	}
	if !plan.Spec.Tier.IsNull() && !plan.Spec.Tier.IsUnknown() {
		spec["tier"] = plan.Spec.Tier.ValueString()
	}
	if !plan.Spec.VpcId.IsNull() && !plan.Spec.VpcId.IsUnknown() {
		spec["vpcId"] = plan.Spec.VpcId.ValueString()
	}
	return m
}

func populateVictoriaMetricsState(ctx context.Context, data map[string]interface{}, state *VictoriaMetricsResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.Constraints = listObjFromAPI(objList(spec, "constraints"), victoriaMetricsConstraintsObjFields)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.CreatePublicIpv6 = getBool(spec, "createPublicIpv6")
	state.Spec.DnsRecordName = getString(spec, "dnsRecordName")
	state.Spec.ScrapTargets = listObjFromAPI(objList(spec, "scrapTargets"), victoriaMetricsScrapTargetsObjFields)
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.VpcId = getString(spec, "vpcId")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"discovered_scrap_targets": types.StringType,
			"fqdn":                     types.StringType,
			"private_ip_v4":            types.StringType,
			"private_ip_v6":            types.StringType,
			"public_ip_v4":             types.StringType,
			"public_ip_v6":             types.StringType,
		},
		map[string]attr.Value{
			"discovered_scrap_targets": getStringFromInfo(data, "discoveredScrapTargets"),
			"fqdn":                     getStringFromInfo(data, "fqdn"),
			"private_ip_v4":            getStringFromInfo(data, "privateIpV4"),
			"private_ip_v6":            getStringFromInfo(data, "privateIpV6"),
			"public_ip_v4":             getStringFromInfo(data, "publicIpV4"),
			"public_ip_v6":             getStringFromInfo(data, "publicIpV6"),
		})
	return nil
}

func (r *VictoriaMetricsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VictoriaMetricsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *VictoriaMetricsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VictoriaMetricsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *VictoriaMetricsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VictoriaMetricsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateVictoriaMetricsState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *VictoriaMetricsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VictoriaMetricsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
	var state VictoriaMetricsResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/victoria-metrics", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateVictoriaMetricsState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
