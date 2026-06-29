package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type KubernetesNodeGroupSpecModel struct {
	CreatePublicIpv4 types.Bool   `tfsdk:"create_public_ipv4"`
	DesiredNodeCount types.Int64  `tfsdk:"desired_node_count"`
	KubernetesId     types.String `tfsdk:"kubernetes_id"`
	VmOfferId        types.String `tfsdk:"vm_offer_id"`
	VmState          types.String `tfsdk:"vm_state"`
	VolumeOfferId    types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib    types.Int64  `tfsdk:"volume_size_gib"`
	VpcSubnetId      types.String `tfsdk:"vpc_subnet_id"`
}

type KubernetesNodeGroupResourceModel struct {
	ID       types.String                 `tfsdk:"id"`
	Metadata metadataModel                `tfsdk:"metadata"`
	Spec     KubernetesNodeGroupSpecModel `tfsdk:"spec"`
	Status   types.Object                 `tfsdk:"status"`
}

type KubernetesNodeGroupResource struct{ client *client.Client }

func NewKubernetesNodeGroupResource() resource.Resource { return &KubernetesNodeGroupResource{} }

func (r *KubernetesNodeGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_node_group"
}

func KubernetesNodeGroupResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"create_public_ipv4": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"desired_node_count": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"kubernetes_id":      schema.StringAttribute{Required: true},
		"vm_offer_id":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_state":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":    schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"vpc_subnet_id":      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"nodes": schema.StringAttribute{Computed: true}}),
	}
}

func (r *KubernetesNodeGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: KubernetesNodeGroupResourceSchemaAttrs()}
}

func (r *KubernetesNodeGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildKubernetesNodeGroupRequestMap(ctx context.Context, plan KubernetesNodeGroupResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.DesiredNodeCount.IsNull() && !plan.Spec.DesiredNodeCount.IsUnknown() {
		spec["desiredNodeCount"] = plan.Spec.DesiredNodeCount.ValueInt64()
	}
	if !plan.Spec.KubernetesId.IsNull() && !plan.Spec.KubernetesId.IsUnknown() {
		spec["kubernetesId"] = plan.Spec.KubernetesId.ValueString()
	}
	if !plan.Spec.VmOfferId.IsNull() && !plan.Spec.VmOfferId.IsUnknown() {
		spec["vmOfferId"] = plan.Spec.VmOfferId.ValueString()
	}
	if !plan.Spec.VmState.IsNull() && !plan.Spec.VmState.IsUnknown() {
		spec["vmState"] = plan.Spec.VmState.ValueString()
	}
	if !plan.Spec.VolumeOfferId.IsNull() && !plan.Spec.VolumeOfferId.IsUnknown() {
		spec["volumeOfferId"] = plan.Spec.VolumeOfferId.ValueString()
	}
	if !plan.Spec.VolumeSizeGib.IsNull() && !plan.Spec.VolumeSizeGib.IsUnknown() {
		spec["volumeSizeGiB"] = plan.Spec.VolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.VpcSubnetId.IsNull() && !plan.Spec.VpcSubnetId.IsUnknown() {
		spec["vpcSubnetId"] = plan.Spec.VpcSubnetId.ValueString()
	}
	return m
}

func populateKubernetesNodeGroupState(ctx context.Context, data map[string]interface{}, state *KubernetesNodeGroupResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.DesiredNodeCount = getInt64(spec, "desiredNodeCount")
	state.Spec.KubernetesId = getString(spec, "kubernetesId")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"nodes": types.StringType,
		},
		map[string]attr.Value{
			"nodes": getStringFromInfo(data, "nodes"),
		})
	return nil
}

func (r *KubernetesNodeGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KubernetesNodeGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildKubernetesNodeGroupRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kubernetes-node-group", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kubernetes-node-group", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes-node-group", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateKubernetesNodeGroupState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KubernetesNodeGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KubernetesNodeGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes-node-group", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateKubernetesNodeGroupState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *KubernetesNodeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KubernetesNodeGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildKubernetesNodeGroupRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kubernetes-node-group", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kubernetes-node-group", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes-node-group", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateKubernetesNodeGroupState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KubernetesNodeGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KubernetesNodeGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/kubernetes-node-group", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kubernetes-node-group", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *KubernetesNodeGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state KubernetesNodeGroupResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/kubernetes-node-group", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateKubernetesNodeGroupState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
