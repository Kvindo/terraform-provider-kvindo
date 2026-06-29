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

var kubernetesControlPlaneLocationsObjFields = []objField{{TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string"}}

type KubernetesSpecModel struct {
	AssignPublicIpV4      types.Bool   `tfsdk:"assign_public_ip_v4"`
	ControlPlaneLocations types.List   `tfsdk:"control_plane_locations"`
	Tier                  types.String `tfsdk:"tier"`
	Version               types.String `tfsdk:"version"`
}

type KubernetesResourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     KubernetesSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type KubernetesResource struct{ client *client.Client }

func NewKubernetesResource() resource.Resource { return &KubernetesResource{} }

func (r *KubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes"
}

func KubernetesResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"assign_public_ip_v4":     schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"control_plane_locations": listObjResourceSchema(kubernetesControlPlaneLocationsObjFields),
		"tier":                    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"version":                 schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"api_server_url": schema.StringAttribute{Computed: true}}),
	}
}

func (r *KubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: KubernetesResourceSchemaAttrs()}
}

func (r *KubernetesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildKubernetesRequestMap(ctx context.Context, plan KubernetesResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.AssignPublicIpV4.IsNull() && !plan.Spec.AssignPublicIpV4.IsUnknown() {
		spec["assignPublicIpV4"] = plan.Spec.AssignPublicIpV4.ValueBool()
	}
	if !plan.Spec.ControlPlaneLocations.IsNull() && !plan.Spec.ControlPlaneLocations.IsUnknown() {
		spec["controlPlaneLocations"] = listObjToAPI(plan.Spec.ControlPlaneLocations, kubernetesControlPlaneLocationsObjFields)
	}
	if !plan.Spec.Tier.IsNull() && !plan.Spec.Tier.IsUnknown() {
		spec["tier"] = plan.Spec.Tier.ValueString()
	}
	if !plan.Spec.Version.IsNull() && !plan.Spec.Version.IsUnknown() {
		spec["version"] = plan.Spec.Version.ValueString()
	}
	return m
}

func populateKubernetesState(ctx context.Context, data map[string]interface{}, state *KubernetesResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.AssignPublicIpV4 = getBool(spec, "assignPublicIpV4")
	state.Spec.ControlPlaneLocations = listObjFromAPI(objList(spec, "controlPlaneLocations"), kubernetesControlPlaneLocationsObjFields)
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.Version = getString(spec, "version")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"api_server_url": types.StringType,
		},
		map[string]attr.Value{
			"api_server_url": getStringFromInfo(data, "apiServerUrl"),
		})
	return nil
}

func (r *KubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildKubernetesRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kubernetes", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kubernetes", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateKubernetesState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateKubernetesState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *KubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KubernetesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildKubernetesRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kubernetes", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kubernetes", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateKubernetesState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KubernetesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/kubernetes", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kubernetes", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *KubernetesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state KubernetesResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateKubernetesState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
