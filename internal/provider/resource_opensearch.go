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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var opensearchClusterManagersObjFields = []objField{{TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string", Immutable: true}}

var opensearchDataNodesObjFields = []objField{{TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string", Immutable: true}}

var opensearchStatusClusterManagersObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "is_elected_cluster_manager", API: "isElectedClusterManager", Kind: "bool"}, {TF: "name", API: "name", Kind: "string"}, {TF: "observed_state", API: "observedState", Kind: "string"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}}

var opensearchStatusDataNodesObjFields = []objField{{TF: "announce_fqdn", API: "announceFqdn", Kind: "string"}, {TF: "disk_free_bytes", API: "diskFreeBytes", Kind: "int64"}, {TF: "disk_used_bytes", API: "diskUsedBytes", Kind: "int64"}, {TF: "id", API: "id", Kind: "string"}, {TF: "jvm_heap_used_percent", API: "jvmHeapUsedPercent", Kind: "float64"}, {TF: "name", API: "name", Kind: "string"}, {TF: "observed_state", API: "observedState", Kind: "string"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "public_ipv4", API: "publicIpV4", Kind: "string"}, {TF: "running_version", API: "runningVersion", Kind: "string"}}

type OpensearchSpecModel struct {
	ClusterManagers      types.List   `tfsdk:"cluster_managers"`
	CreatePublicIpv4     types.Bool   `tfsdk:"create_public_ipv4"`
	DataNodes            types.List   `tfsdk:"data_nodes"`
	DataVmOfferId        types.String `tfsdk:"data_vm_offer_id"`
	DataVolumeOfferId    types.String `tfsdk:"data_volume_offer_id"`
	DataVolumeSizeGib    types.Int64  `tfsdk:"data_volume_size_gib"`
	ManagerVmOfferId     types.String `tfsdk:"manager_vm_offer_id"`
	ManagerVolumeOfferId types.String `tfsdk:"manager_volume_offer_id"`
	ManagerVolumeSizeGib types.Int64  `tfsdk:"manager_volume_size_gib"`
	OpenSearchVersion    types.String `tfsdk:"open_search_version"`
	ParametersSetId      types.String `tfsdk:"parameters_set_id"`
	UseFqdn              types.Bool   `tfsdk:"use_fqdn"`
}

type OpensearchResourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     OpensearchSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type OpensearchResource struct{ client *client.Client }

func NewOpensearchResource() resource.Resource { return &OpensearchResource{} }

func (r *OpensearchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_opensearch"
}

func OpensearchResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"cluster_managers":        listObjResourceSchema(opensearchClusterManagersObjFields),
		"create_public_ipv4":      schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()}},
		"data_nodes":              listObjResourceSchema(opensearchDataNodesObjFields),
		"data_vm_offer_id":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"data_volume_offer_id":    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"data_volume_size_gib":    schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown(), int64planmodifier.RequiresReplace()}},
		"manager_vm_offer_id":     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"manager_volume_offer_id": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"manager_volume_size_gib": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown(), int64planmodifier.RequiresReplace()}},
		"open_search_version":     schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"parameters_set_id":       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"use_fqdn":                schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"active_cluster_manager_count": schema.Int64Attribute{Computed: true}, "active_shards_percent": schema.Int64Attribute{Computed: true}, "anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "any_index_readonly": schema.BoolAttribute{Computed: true}, "ca_certificate": schema.StringAttribute{Computed: true}, "cluster_manager_quorum_size": schema.Int64Attribute{Computed: true}, "cluster_managers": listObjStatusSchema(opensearchStatusClusterManagersObjFields), "cluster_name": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "data_nodes": listObjStatusSchema(opensearchStatusDataNodesObjFields), "default_number_of_replicas": schema.Int64Attribute{Computed: true}, "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "health": schema.StringAttribute{Computed: true}, "index_count": schema.Int64Attribute{Computed: true}, "last_probe_utc": schema.Int64Attribute{Computed: true}, "red_index_count": schema.Int64Attribute{Computed: true}, "relocating_shards": schema.Int64Attribute{Computed: true}, "rest_endpoints": schema.StringAttribute{Computed: true}, "rest_public_endpoints": schema.StringAttribute{Computed: true}, "running_version": schema.StringAttribute{Computed: true}, "total_data_bytes": schema.Int64Attribute{Computed: true}, "total_documents": schema.Int64Attribute{Computed: true}, "unassigned_shards": schema.Int64Attribute{Computed: true}, "yellow_index_count": schema.Int64Attribute{Computed: true}}),
	}
}

func (r *OpensearchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: OpensearchResourceSchemaAttrs()}
}

func (r *OpensearchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildOpensearchRequestMap(ctx context.Context, plan OpensearchResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ClusterManagers.IsNull() && !plan.Spec.ClusterManagers.IsUnknown() {
		spec["clusterManagers"] = listObjToAPI(plan.Spec.ClusterManagers, opensearchClusterManagersObjFields)
	}
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.DataNodes.IsNull() && !plan.Spec.DataNodes.IsUnknown() {
		spec["dataNodes"] = listObjToAPI(plan.Spec.DataNodes, opensearchDataNodesObjFields)
	}
	if !plan.Spec.DataVmOfferId.IsNull() && !plan.Spec.DataVmOfferId.IsUnknown() {
		spec["dataVmOfferId"] = plan.Spec.DataVmOfferId.ValueString()
	}
	if !plan.Spec.DataVolumeOfferId.IsNull() && !plan.Spec.DataVolumeOfferId.IsUnknown() {
		spec["dataVolumeOfferId"] = plan.Spec.DataVolumeOfferId.ValueString()
	}
	if !plan.Spec.DataVolumeSizeGib.IsNull() && !plan.Spec.DataVolumeSizeGib.IsUnknown() {
		spec["dataVolumeSizeGiB"] = plan.Spec.DataVolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.ManagerVmOfferId.IsNull() && !plan.Spec.ManagerVmOfferId.IsUnknown() {
		spec["managerVmOfferId"] = plan.Spec.ManagerVmOfferId.ValueString()
	}
	if !plan.Spec.ManagerVolumeOfferId.IsNull() && !plan.Spec.ManagerVolumeOfferId.IsUnknown() {
		spec["managerVolumeOfferId"] = plan.Spec.ManagerVolumeOfferId.ValueString()
	}
	if !plan.Spec.ManagerVolumeSizeGib.IsNull() && !plan.Spec.ManagerVolumeSizeGib.IsUnknown() {
		spec["managerVolumeSizeGiB"] = plan.Spec.ManagerVolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.OpenSearchVersion.IsNull() && !plan.Spec.OpenSearchVersion.IsUnknown() {
		spec["openSearchVersion"] = plan.Spec.OpenSearchVersion.ValueString()
	}
	if !plan.Spec.ParametersSetId.IsNull() && !plan.Spec.ParametersSetId.IsUnknown() {
		spec["parametersSetId"] = plan.Spec.ParametersSetId.ValueString()
	}
	if !plan.Spec.UseFqdn.IsNull() && !plan.Spec.UseFqdn.IsUnknown() {
		spec["useFqdn"] = plan.Spec.UseFqdn.ValueBool()
	}
	return m
}

func populateOpensearchState(ctx context.Context, data map[string]interface{}, state *OpensearchResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ClusterManagers = listObjFromAPI(objList(spec, "clusterManagers"), opensearchClusterManagersObjFields)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.DataNodes = listObjFromAPI(objList(spec, "dataNodes"), opensearchDataNodesObjFields)
	state.Spec.DataVmOfferId = getString(spec, "dataVmOfferId")
	state.Spec.DataVolumeOfferId = getString(spec, "dataVolumeOfferId")
	state.Spec.DataVolumeSizeGib = getInt64(spec, "dataVolumeSizeGiB")
	state.Spec.ManagerVmOfferId = getString(spec, "managerVmOfferId")
	state.Spec.ManagerVolumeOfferId = getString(spec, "managerVolumeOfferId")
	state.Spec.ManagerVolumeSizeGib = getInt64(spec, "managerVolumeSizeGiB")
	state.Spec.OpenSearchVersion = getString(spec, "openSearchVersion")
	state.Spec.ParametersSetId = getString(spec, "parametersSetId")
	state.Spec.UseFqdn = getBool(spec, "useFqdn")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"active_cluster_manager_count": types.Int64Type,
			"active_shards_percent":        types.Int64Type,
			"anti_affinity_message":        types.StringType,
			"anti_affinity_ok":             types.BoolType,
			"any_index_readonly":           types.BoolType,
			"ca_certificate":               types.StringType,
			"cluster_manager_quorum_size":  types.Int64Type,
			"cluster_managers":             attrTypeOf("list_object", opensearchStatusClusterManagersObjFields),
			"cluster_name":                 types.StringType,
			"cluster_state":                types.StringType,
			"data_nodes":                   attrTypeOf("list_object", opensearchStatusDataNodesObjFields),
			"default_number_of_replicas":   types.Int64Type,
			"dns_seed_fqdn":                types.StringType,
			"health":                       types.StringType,
			"index_count":                  types.Int64Type,
			"last_probe_utc":               types.Int64Type,
			"red_index_count":              types.Int64Type,
			"relocating_shards":            types.Int64Type,
			"rest_endpoints":               types.StringType,
			"rest_public_endpoints":        types.StringType,
			"running_version":              types.StringType,
			"total_data_bytes":             types.Int64Type,
			"total_documents":              types.Int64Type,
			"unassigned_shards":            types.Int64Type,
			"yellow_index_count":           types.Int64Type,
		},
		map[string]attr.Value{
			"active_cluster_manager_count": getInt64FromInfo(data, "activeClusterManagerCount"),
			"active_shards_percent":        getInt64FromInfo(data, "activeShardsPercent"),
			"anti_affinity_message":        getStringFromInfo(data, "antiAffinityMessage"),
			"anti_affinity_ok":             getBoolFromInfo(data, "antiAffinityOk"),
			"any_index_readonly":           getBoolFromInfo(data, "anyIndexReadonly"),
			"ca_certificate":               getStringFromInfo(data, "caCertificate"),
			"cluster_manager_quorum_size":  getInt64FromInfo(data, "clusterManagerQuorumSize"),
			"cluster_managers":             getListObjFromInfo(data, "clusterManagers", opensearchStatusClusterManagersObjFields),
			"cluster_name":                 getStringFromInfo(data, "clusterName"),
			"cluster_state":                getStringFromInfo(data, "clusterState"),
			"data_nodes":                   getListObjFromInfo(data, "dataNodes", opensearchStatusDataNodesObjFields),
			"default_number_of_replicas":   getInt64FromInfo(data, "defaultNumberOfReplicas"),
			"dns_seed_fqdn":                getStringFromInfo(data, "dnsSeedFqdn"),
			"health":                       getStringFromInfo(data, "health"),
			"index_count":                  getInt64FromInfo(data, "indexCount"),
			"last_probe_utc":               getInt64FromInfo(data, "lastProbeUtc"),
			"red_index_count":              getInt64FromInfo(data, "redIndexCount"),
			"relocating_shards":            getInt64FromInfo(data, "relocatingShards"),
			"rest_endpoints":               getStringFromInfo(data, "restEndpoints"),
			"rest_public_endpoints":        getStringFromInfo(data, "restPublicEndpoints"),
			"running_version":              getStringFromInfo(data, "runningVersion"),
			"total_data_bytes":             getInt64FromInfo(data, "totalDataBytes"),
			"total_documents":              getInt64FromInfo(data, "totalDocuments"),
			"unassigned_shards":            getInt64FromInfo(data, "unassignedShards"),
			"yellow_index_count":           getInt64FromInfo(data, "yellowIndexCount"),
		})
	return nil
}

func (r *OpensearchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OpensearchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildOpensearchRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/opensearch", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/opensearch", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/opensearch", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateOpensearchState(ctx, recoverData, &plan); popErr == nil {
				resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
			} else {
				tflog.Warn(ctx, "Create Poll Error: recovery state population also failed", map[string]interface{}{"error": popErr.Error()})
			}
		} else if getErr != nil {
			tflog.Warn(ctx, "Create Poll Error: recovery Get also failed", map[string]interface{}{"error": getErr.Error()})
		}
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateOpensearchState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpensearchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OpensearchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateOpensearchState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *OpensearchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state OpensearchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildOpensearchRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/opensearch", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/opensearch", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateOpensearchState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpensearchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OpensearchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/opensearch", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/opensearch", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *OpensearchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state OpensearchResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateOpensearchState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
