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

var valkeyShardsObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string"}}

var valkeyStatusNodesObjFields = []objField{{TF: "announce_fqdn", API: "announceFqdn", Kind: "string"}, {TF: "bus_port", API: "busPort", Kind: "int64"}, {TF: "id", API: "id", Kind: "string"}, {TF: "is_primary", API: "isPrimary", Kind: "bool"}, {TF: "node_id", API: "nodeId", Kind: "string"}, {TF: "observed_role", API: "observedRole", Kind: "string"}, {TF: "port", API: "port", Kind: "int64"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "public_ipv4", API: "publicIpV4", Kind: "string"}, {TF: "shard_index", API: "shardIndex", Kind: "int64"}}

var valkeyStatusShardsObjFields = []objField{{TF: "index", API: "index", Kind: "int64"}, {TF: "primary_endpoint", API: "primaryEndpoint", Kind: "string"}, {TF: "primary_instance_id", API: "primaryInstanceId", Kind: "string"}, {TF: "replica_endpoints", API: "replicaEndpoints", Kind: "list_string"}, {TF: "replica_instance_ids", API: "replicaInstanceIds", Kind: "list_string"}, {TF: "slot_end", API: "slotEnd", Kind: "int64"}, {TF: "slot_start", API: "slotStart", Kind: "int64"}}

type ValkeySpecModel struct {
	CreatePublicIpv4 types.Bool   `tfsdk:"create_public_ipv4"`
	ParametersSetId  types.String `tfsdk:"parameters_set_id"`
	ReplicasPerShard types.Int64  `tfsdk:"replicas_per_shard"`
	RootPassword     types.String `tfsdk:"root_password"`
	Shards           types.List   `tfsdk:"shards"`
	UseFqdn          types.Bool   `tfsdk:"use_fqdn"`
	ValkeyVersion    types.String `tfsdk:"valkey_version"`
	VmOfferId        types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId    types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib    types.Int64  `tfsdk:"volume_size_gib"`
}

type ValkeyResourceModel struct {
	ID       types.String    `tfsdk:"id"`
	Metadata metadataModel   `tfsdk:"metadata"`
	Spec     ValkeySpecModel `tfsdk:"spec"`
	Status   types.Object    `tfsdk:"status"`
}

type ValkeyResource struct{ client *client.Client }

func NewValkeyResource() resource.Resource { return &ValkeyResource{} }

func (r *ValkeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_valkey"
}

func ValkeyResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"create_public_ipv4": schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"parameters_set_id":  schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"replicas_per_shard": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"root_password":      schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"shards":             listObjResourceSchema(valkeyShardsObjFields),
		"use_fqdn":           schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"valkey_version":     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_offer_id":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":    schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "cluster_endpoints": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "connection_uri": schema.StringAttribute{Computed: true}, "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "nodes": listObjStatusSchema(valkeyStatusNodesObjFields), "password": schema.StringAttribute{Computed: true}, "port": schema.Int64Attribute{Computed: true}, "primary_endpoints": schema.StringAttribute{Computed: true}, "shards": listObjStatusSchema(valkeyStatusShardsObjFields)}),
	}
}

func (r *ValkeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: ValkeyResourceSchemaAttrs()}
}

func (r *ValkeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildValkeyRequestMap(ctx context.Context, plan ValkeyResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.ParametersSetId.IsNull() && !plan.Spec.ParametersSetId.IsUnknown() {
		spec["parametersSetId"] = plan.Spec.ParametersSetId.ValueString()
	}
	if !plan.Spec.ReplicasPerShard.IsNull() && !plan.Spec.ReplicasPerShard.IsUnknown() {
		spec["replicasPerShard"] = plan.Spec.ReplicasPerShard.ValueInt64()
	}
	if !plan.Spec.RootPassword.IsNull() && !plan.Spec.RootPassword.IsUnknown() {
		spec["rootPassword"] = plan.Spec.RootPassword.ValueString()
	}
	if !plan.Spec.Shards.IsNull() && !plan.Spec.Shards.IsUnknown() {
		spec["shards"] = listObjToAPI(plan.Spec.Shards, valkeyShardsObjFields)
	}
	if !plan.Spec.UseFqdn.IsNull() && !plan.Spec.UseFqdn.IsUnknown() {
		spec["useFqdn"] = plan.Spec.UseFqdn.ValueBool()
	}
	if !plan.Spec.ValkeyVersion.IsNull() && !plan.Spec.ValkeyVersion.IsUnknown() {
		spec["valkeyVersion"] = plan.Spec.ValkeyVersion.ValueString()
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

func populateValkeyState(ctx context.Context, data map[string]interface{}, state *ValkeyResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.ParametersSetId = getString(spec, "parametersSetId")
	state.Spec.ReplicasPerShard = getInt64(spec, "replicasPerShard")
	state.Spec.RootPassword = getString(spec, "rootPassword")
	state.Spec.Shards = listObjFromAPI(objList(spec, "shards"), valkeyShardsObjFields)
	state.Spec.UseFqdn = getBool(spec, "useFqdn")
	state.Spec.ValkeyVersion = getString(spec, "valkeyVersion")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"anti_affinity_message": types.StringType,
			"anti_affinity_ok":      types.BoolType,
			"cluster_endpoints":     types.StringType,
			"cluster_state":         types.StringType,
			"connection_uri":        types.StringType,
			"dns_seed_fqdn":         types.StringType,
			"nodes":                 attrTypeOf("list_object", valkeyStatusNodesObjFields),
			"password":              types.StringType,
			"port":                  types.Int64Type,
			"primary_endpoints":     types.StringType,
			"shards":                attrTypeOf("list_object", valkeyStatusShardsObjFields),
		},
		map[string]attr.Value{
			"anti_affinity_message": getStringFromInfo(data, "antiAffinityMessage"),
			"anti_affinity_ok":      getBoolFromInfo(data, "antiAffinityOk"),
			"cluster_endpoints":     getStringFromInfo(data, "clusterEndpoints"),
			"cluster_state":         getStringFromInfo(data, "clusterState"),
			"connection_uri":        getStringFromInfo(data, "connectionUri"),
			"dns_seed_fqdn":         getStringFromInfo(data, "dnsSeedFqdn"),
			"nodes":                 getListObjFromInfo(data, "nodes", valkeyStatusNodesObjFields),
			"password":              getStringFromInfo(data, "password"),
			"port":                  getInt64FromInfo(data, "port"),
			"primary_endpoints":     getStringFromInfo(data, "primaryEndpoints"),
			"shards":                getListObjFromInfo(data, "shards", valkeyStatusShardsObjFields),
		})
	return nil
}

func (r *ValkeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ValkeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildValkeyRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/valkey", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/valkey", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/valkey", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateValkeyState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ValkeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ValkeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/valkey", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateValkeyState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ValkeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ValkeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildValkeyRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/valkey", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/valkey", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/valkey", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateValkeyState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ValkeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ValkeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/valkey", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/valkey", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *ValkeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state ValkeyResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/valkey", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateValkeyState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
