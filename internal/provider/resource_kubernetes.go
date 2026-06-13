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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types
var _ = listplanmodifier.UseStateForUnknown

// KubernetesResourceModel describes the resource data model.
type KubernetesResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	AssignPublicIpV4 types.Bool `tfsdk:"assign_public_ip_v4"`
	Version types.String `tfsdk:"version"`
	ControlPlaneLocations types.List `tfsdk:"control_plane_locations"`
	Info types.Object `tfsdk:"info"`
}

// KubernetesControlPlaneLocationsModel is the nested object model for control_plane_locations.
type KubernetesControlPlaneLocationsModel struct {
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
}

// KubernetesResource defines the resource implementation.
type KubernetesResource struct {
	client *client.Client
}

func NewKubernetesResource() resource.Resource {
	return &KubernetesResource{}
}

func (r *KubernetesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes"
}

func (r *KubernetesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["assign_public_ip_v4"] = schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
		}
	attrs["version"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["control_plane_locations"] = schema.ListNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"vpc_subnet_id": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
				},
			},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "api_server_url": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.Tier.IsNull() && !plan.Tier.IsUnknown() {
		m["tier"] = plan.Tier.ValueString()
	}
	if !plan.AssignPublicIpV4.IsNull() && !plan.AssignPublicIpV4.IsUnknown() {
		m["assignPublicIpV4"] = plan.AssignPublicIpV4.ValueBool()
	}
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		m["version"] = plan.Version.ValueString()
	}
	if !plan.ControlPlaneLocations.IsNull() && !plan.ControlPlaneLocations.IsUnknown() {
		var items []map[string]interface{}
		for _, elem := range plan.ControlPlaneLocations.Elements() {
			if ov, ok := elem.(types.Object); ok {
				item := map[string]interface{}{}
				if v, ok := ov.Attributes()["vpc_subnet_id"]; ok {
					if sv, ok := v.(types.String); ok && !sv.IsNull() {
						item["vpcSubnetId"] = sv.ValueString()
					}
				}
				items = append(items, item)
			}
		}
		m["controlPlaneLocations"] = items
	}
	return m
}

func populateKubernetesState(ctx context.Context, data map[string]interface{}, state *KubernetesResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.Tier = getString(data, "tier")
	state.AssignPublicIpV4 = getBool(data, "assignPublicIpV4")
	state.Version = getString(data, "version")
	{
		rawControlPlaneLocations, _ := data["controlPlaneLocations"].([]interface{})
		attrTypes := map[string]attr.Type{
			"vpc_subnet_id": types.StringType,
		}
		objs := make([]attr.Value, 0, len(rawControlPlaneLocations))
		for _, item := range rawControlPlaneLocations {
			if m, ok := item.(map[string]interface{}); ok {
				attrs := map[string]attr.Value{
					"vpc_subnet_id": getString(m, "vpcSubnetId"),
				}
				obj, _ := types.ObjectValue(attrTypes, attrs)
				objs = append(objs, obj)
			}
		}
		state.ControlPlaneLocations, _ = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objs)
	}
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "api_server_url": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "api_server_url": getStringFromInfo(data, "apiServerUrl")})
	return nil
}

func (r *KubernetesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KubernetesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *KubernetesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KubernetesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *KubernetesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan KubernetesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state KubernetesResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateKubernetesState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *KubernetesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KubernetesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	// Import by ID
	var state KubernetesResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateKubernetesState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
