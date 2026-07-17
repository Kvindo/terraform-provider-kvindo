package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type HostingProviderSpecModel struct {
	City            types.String  `tfsdk:"city"`
	Cloud           types.String  `tfsdk:"cloud"`
	Country         types.String  `tfsdk:"country"`
	CountryIsoCode  types.String  `tfsdk:"country_iso_code"`
	DataCenterIndex types.Int64   `tfsdk:"data_center_index"`
	Disabled        types.Bool    `tfsdk:"disabled"`
	KeyFeatures     types.List    `tfsdk:"key_features"`
	Sla             types.Float64 `tfsdk:"sla"`
}

type HostingProviderResourceModel struct {
	ID       types.String             `tfsdk:"id"`
	Metadata metadataModel            `tfsdk:"metadata"`
	Spec     HostingProviderSpecModel `tfsdk:"spec"`
	Status   types.Object             `tfsdk:"status"`
}

type HostingProviderResource struct{ client *client.Client }

func NewHostingProviderResource() resource.Resource { return &HostingProviderResource{} }

func (r *HostingProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosting_provider"
}

func HostingProviderResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"city":              schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"cloud":             schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"country":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"country_iso_code":  schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"data_center_index": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"disabled":          schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"key_features":      schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"sla":               schema.Float64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *HostingProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: HostingProviderResourceSchemaAttrs()}
}

func (r *HostingProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildHostingProviderRequestMap(ctx context.Context, plan HostingProviderResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.City.IsNull() && !plan.Spec.City.IsUnknown() {
		spec["city"] = plan.Spec.City.ValueString()
	}
	if !plan.Spec.Cloud.IsNull() && !plan.Spec.Cloud.IsUnknown() {
		spec["cloud"] = plan.Spec.Cloud.ValueString()
	}
	if !plan.Spec.Country.IsNull() && !plan.Spec.Country.IsUnknown() {
		spec["country"] = plan.Spec.Country.ValueString()
	}
	if !plan.Spec.CountryIsoCode.IsNull() && !plan.Spec.CountryIsoCode.IsUnknown() {
		spec["countryIsoCode"] = plan.Spec.CountryIsoCode.ValueString()
	}
	if !plan.Spec.DataCenterIndex.IsNull() && !plan.Spec.DataCenterIndex.IsUnknown() {
		spec["dataCenterIndex"] = plan.Spec.DataCenterIndex.ValueInt64()
	}
	if !plan.Spec.Disabled.IsNull() && !plan.Spec.Disabled.IsUnknown() {
		spec["disabled"] = plan.Spec.Disabled.ValueBool()
	}
	if !plan.Spec.KeyFeatures.IsNull() && !plan.Spec.KeyFeatures.IsUnknown() {
		spec["keyFeatures"] = stringListToInterface(ctx, plan.Spec.KeyFeatures)
	}
	if !plan.Spec.Sla.IsNull() && !plan.Spec.Sla.IsUnknown() {
		spec["sla"] = plan.Spec.Sla.ValueFloat64()
	}
	return m
}

func populateHostingProviderState(ctx context.Context, data map[string]interface{}, state *HostingProviderResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.City = getString(spec, "city")
	state.Spec.Cloud = getString(spec, "cloud")
	state.Spec.Country = getString(spec, "country")
	state.Spec.CountryIsoCode = getString(spec, "countryIsoCode")
	state.Spec.DataCenterIndex = getInt64(spec, "dataCenterIndex")
	state.Spec.Disabled = getBool(spec, "disabled")
	state.Spec.KeyFeatures = getStringList(ctx, spec, "keyFeatures")
	state.Spec.Sla = getFloat64(spec, "sla")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *HostingProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HostingProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildHostingProviderRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/hosting-provider", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/hosting-provider", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/hosting-provider", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateHostingProviderState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *HostingProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HostingProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/hosting-provider", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateHostingProviderState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *HostingProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state HostingProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildHostingProviderRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/hosting-provider", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/hosting-provider", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/hosting-provider", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateHostingProviderState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *HostingProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HostingProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/hosting-provider", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/hosting-provider", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *HostingProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state HostingProviderResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/hosting-provider", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateHostingProviderState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
