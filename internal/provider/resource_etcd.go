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

var etcdInstancesObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string"}}

var etcdStatusInstancesObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "public_ipv4", API: "publicIpV4", Kind: "string"}}

type EtcdSpecModel struct {
	CreatePublicIpv4 types.Bool   `tfsdk:"create_public_ipv4"`
	EtcdVersion      types.String `tfsdk:"etcd_version"`
	Instances        types.List   `tfsdk:"instances"`
	RootPassword     types.String `tfsdk:"root_password"`
	Tier             types.String `tfsdk:"tier"`
	VmOfferId        types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId    types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib    types.Int64  `tfsdk:"volume_size_gib"`
}

type EtcdResourceModel struct {
	ID       types.String  `tfsdk:"id"`
	Metadata metadataModel `tfsdk:"metadata"`
	Spec     EtcdSpecModel `tfsdk:"spec"`
	Status   types.Object  `tfsdk:"status"`
}

type EtcdResource struct{ client *client.Client }

func NewEtcdResource() resource.Resource { return &EtcdResource{} }

func (r *EtcdResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_etcd"
}

func EtcdResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"create_public_ipv4": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"etcd_version":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"instances":          listObjResourceSchema(etcdInstancesObjFields),
		"root_password":      schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"tier":               schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_offer_id":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":    schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"ca_cert": schema.StringAttribute{Computed: true}, "instances": listObjStatusSchema(etcdStatusInstancesObjFields), "port": schema.Int64Attribute{Computed: true}}),
	}
}

func (r *EtcdResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: EtcdResourceSchemaAttrs()}
}

func (r *EtcdResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildEtcdRequestMap(ctx context.Context, plan EtcdResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.EtcdVersion.IsNull() && !plan.Spec.EtcdVersion.IsUnknown() {
		spec["etcdVersion"] = plan.Spec.EtcdVersion.ValueString()
	}
	if !plan.Spec.Instances.IsNull() && !plan.Spec.Instances.IsUnknown() {
		spec["instances"] = listObjToAPI(plan.Spec.Instances, etcdInstancesObjFields)
	}
	if !plan.Spec.RootPassword.IsNull() && !plan.Spec.RootPassword.IsUnknown() {
		spec["rootPassword"] = plan.Spec.RootPassword.ValueString()
	}
	if !plan.Spec.Tier.IsNull() && !plan.Spec.Tier.IsUnknown() {
		spec["tier"] = plan.Spec.Tier.ValueString()
	}
	if !plan.Spec.VmOfferId.IsNull() && !plan.Spec.VmOfferId.IsUnknown() {
		spec["vmOfferId"] = plan.Spec.VmOfferId.ValueString()
	}
	if !plan.Spec.VolumeOfferId.IsNull() && !plan.Spec.VolumeOfferId.IsUnknown() {
		spec["volumeOfferId"] = plan.Spec.VolumeOfferId.ValueString()
	}
	if !plan.Spec.VolumeSizeGib.IsNull() && !plan.Spec.VolumeSizeGib.IsUnknown() {
		spec["volumeSizeGiB"] = plan.Spec.VolumeSizeGib.ValueInt64()
	}
	return m
}

func populateEtcdState(ctx context.Context, data map[string]interface{}, state *EtcdResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.EtcdVersion = getString(spec, "etcdVersion")
	state.Spec.Instances = listObjFromAPI(objList(spec, "instances"), etcdInstancesObjFields)
	state.Spec.RootPassword = getString(spec, "rootPassword")
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"ca_cert":   types.StringType,
			"instances": attrTypeOf("list_object", etcdStatusInstancesObjFields),
			"port":      types.Int64Type,
		},
		map[string]attr.Value{
			"ca_cert":   getStringFromInfo(data, "caCert"),
			"instances": getListObjFromInfo(data, "instances", etcdStatusInstancesObjFields),
			"port":      getInt64FromInfo(data, "port"),
		})
	return nil
}

func (r *EtcdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EtcdResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildEtcdRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/etcd", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/etcd", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/etcd", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateEtcdState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EtcdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EtcdResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/etcd", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateEtcdState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *EtcdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EtcdResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildEtcdRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/etcd", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/etcd", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/etcd", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateEtcdState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *EtcdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EtcdResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/etcd", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/etcd", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *EtcdResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state EtcdResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/etcd", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateEtcdState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
