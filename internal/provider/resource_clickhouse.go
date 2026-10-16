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

var clickhouseKeepersObjFields = []objField{{TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string"}}

var clickhouseShardGroupsObjFields = []objField{{TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string"}}

var clickhouseStatusDatabasesObjFields = []objField{{TF: "data_bytes_on_disk", API: "dataBytesOnDisk", Kind: "int64"}, {TF: "declared", API: "declared", Kind: "bool"}, {TF: "engine", API: "engine", Kind: "string"}, {TF: "name", API: "name", Kind: "string"}, {TF: "table_count", API: "tableCount", Kind: "int64"}, {TF: "uuid", API: "uuid", Kind: "string"}}

var clickhouseStatusKeepersObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "is_leader", API: "isLeader", Kind: "bool"}, {TF: "name", API: "name", Kind: "string"}, {TF: "observed_state", API: "observedState", Kind: "string"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "server_id", API: "serverId", Kind: "int64"}, {TF: "znode_count", API: "znodeCount", Kind: "int64"}}

var clickhouseStatusNodesObjFields = []objField{{TF: "absolute_delay_seconds", API: "absoluteDelaySeconds", Kind: "int64"}, {TF: "announce_fqdn", API: "announceFqdn", Kind: "string"}, {TF: "disk_free_bytes", API: "diskFreeBytes", Kind: "int64"}, {TF: "disk_used_bytes", API: "diskUsedBytes", Kind: "int64"}, {TF: "id", API: "id", Kind: "string"}, {TF: "name", API: "name", Kind: "string"}, {TF: "observed_state", API: "observedState", Kind: "string"}, {TF: "parts_count", API: "partsCount", Kind: "int64"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "public_ipv4", API: "publicIpV4", Kind: "string"}, {TF: "running_version", API: "runningVersion", Kind: "string"}, {TF: "shard_index", API: "shardIndex", Kind: "int64"}}

var clickhouseStatusShardGroupsObjFields = []objField{{TF: "any_replica_readonly", API: "anyReplicaReadonly", Kind: "bool"}, {TF: "backup", API: "backup", Kind: "object", Obj: []objField{{TF: "last_backup_name", API: "lastBackupName", Kind: "string"}, {TF: "last_backup_time", API: "lastBackupTime", Kind: "string"}, {TF: "last_backup_type", API: "lastBackupType", Kind: "string"}, {TF: "last_check_ok", API: "lastCheckOk", Kind: "bool"}, {TF: "last_check_time", API: "lastCheckTime", Kind: "string"}, {TF: "message", API: "message", Kind: "string"}}}, {TF: "data_bytes_on_disk", API: "dataBytesOnDisk", Kind: "int64"}, {TF: "max_absolute_delay_seconds", API: "maxAbsoluteDelaySeconds", Kind: "int64"}, {TF: "replica_endpoints", API: "replicaEndpoints", Kind: "list_string"}, {TF: "replica_instance_ids", API: "replicaInstanceIds", Kind: "list_string"}, {TF: "replicated_table_count", API: "replicatedTableCount", Kind: "int64"}, {TF: "replication_queue_size", API: "replicationQueueSize", Kind: "int64"}, {TF: "rows", API: "rows", Kind: "int64"}, {TF: "shard_group_name", API: "shardGroupName", Kind: "string"}, {TF: "shard_index", API: "shardIndex", Kind: "int64"}}

type ClickhouseSpecModel struct {
	BackupRetentionDays    types.Int64  `tfsdk:"backup_retention_days"`
	ClickhouseVersion      types.String `tfsdk:"clickhouse_version"`
	CreatePublicIpv4       types.Bool   `tfsdk:"create_public_ipv4"`
	FullBackupIntervalDays types.Int64  `tfsdk:"full_backup_interval_days"`
	KeeperVmOfferId        types.String `tfsdk:"keeper_vm_offer_id"`
	KeeperVolumeOfferId    types.String `tfsdk:"keeper_volume_offer_id"`
	KeeperVolumeSizeGib    types.Int64  `tfsdk:"keeper_volume_size_gib"`
	Keepers                types.List   `tfsdk:"keepers"`
	ParametersSetId        types.String `tfsdk:"parameters_set_id"`
	ReplicasPerShardGroup  types.Int64  `tfsdk:"replicas_per_shard_group"`
	ShardGroups            types.List   `tfsdk:"shard_groups"`
	UseFqdn                types.Bool   `tfsdk:"use_fqdn"`
	VmOfferId              types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId          types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib          types.Int64  `tfsdk:"volume_size_gib"`
}

type ClickhouseResourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     ClickhouseSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type ClickhouseResource struct{ client *client.Client }

func NewClickhouseResource() resource.Resource { return &ClickhouseResource{} }

func (r *ClickhouseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse"
}

func ClickhouseResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"backup_retention_days":     schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"clickhouse_version":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"create_public_ipv4":        schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"full_backup_interval_days": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"keeper_vm_offer_id":        schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"keeper_volume_offer_id":    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"keeper_volume_size_gib":    schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"keepers":                   listObjResourceSchema(clickhouseKeepersObjFields),
		"parameters_set_id":         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"replicas_per_shard_group":  schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"shard_groups":              listObjResourceSchema(clickhouseShardGroupsObjFields),
		"use_fqdn":                  schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"vm_offer_id":               schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":           schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "any_replica_readonly": schema.BoolAttribute{Computed: true}, "ca_certificate": schema.StringAttribute{Computed: true}, "cluster_name": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "databases": listObjStatusSchema(clickhouseStatusDatabasesObjFields), "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "http_endpoints": schema.StringAttribute{Computed: true}, "https_endpoints": schema.StringAttribute{Computed: true}, "keeper_quorum_size": schema.Int64Attribute{Computed: true}, "keepers": listObjStatusSchema(clickhouseStatusKeepersObjFields), "native_endpoints": schema.StringAttribute{Computed: true}, "native_tls_endpoints": schema.StringAttribute{Computed: true}, "nodes": listObjStatusSchema(clickhouseStatusNodesObjFields), "public_endpoints": schema.StringAttribute{Computed: true}, "running_version": schema.StringAttribute{Computed: true}, "shard_groups": listObjStatusSchema(clickhouseStatusShardGroupsObjFields)}),
	}
}

func (r *ClickhouseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: ClickhouseResourceSchemaAttrs()}
}

func (r *ClickhouseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildClickhouseRequestMap(ctx context.Context, plan ClickhouseResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.BackupRetentionDays.IsNull() && !plan.Spec.BackupRetentionDays.IsUnknown() {
		spec["backupRetentionDays"] = plan.Spec.BackupRetentionDays.ValueInt64()
	}
	if !plan.Spec.ClickhouseVersion.IsNull() && !plan.Spec.ClickhouseVersion.IsUnknown() {
		spec["clickHouseVersion"] = plan.Spec.ClickhouseVersion.ValueString()
	}
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.FullBackupIntervalDays.IsNull() && !plan.Spec.FullBackupIntervalDays.IsUnknown() {
		spec["fullBackupIntervalDays"] = plan.Spec.FullBackupIntervalDays.ValueInt64()
	}
	if !plan.Spec.KeeperVmOfferId.IsNull() && !plan.Spec.KeeperVmOfferId.IsUnknown() {
		spec["keeperVmOfferId"] = plan.Spec.KeeperVmOfferId.ValueString()
	}
	if !plan.Spec.KeeperVolumeOfferId.IsNull() && !plan.Spec.KeeperVolumeOfferId.IsUnknown() {
		spec["keeperVolumeOfferId"] = plan.Spec.KeeperVolumeOfferId.ValueString()
	}
	if !plan.Spec.KeeperVolumeSizeGib.IsNull() && !plan.Spec.KeeperVolumeSizeGib.IsUnknown() {
		spec["keeperVolumeSizeGiB"] = plan.Spec.KeeperVolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.Keepers.IsNull() && !plan.Spec.Keepers.IsUnknown() {
		spec["keepers"] = listObjToAPI(plan.Spec.Keepers, clickhouseKeepersObjFields)
	}
	if !plan.Spec.ParametersSetId.IsNull() && !plan.Spec.ParametersSetId.IsUnknown() {
		spec["parametersSetId"] = plan.Spec.ParametersSetId.ValueString()
	}
	if !plan.Spec.ReplicasPerShardGroup.IsNull() && !plan.Spec.ReplicasPerShardGroup.IsUnknown() {
		spec["replicasPerShardGroup"] = plan.Spec.ReplicasPerShardGroup.ValueInt64()
	}
	if !plan.Spec.ShardGroups.IsNull() && !plan.Spec.ShardGroups.IsUnknown() {
		spec["shardGroups"] = listObjToAPI(plan.Spec.ShardGroups, clickhouseShardGroupsObjFields)
	}
	if !plan.Spec.UseFqdn.IsNull() && !plan.Spec.UseFqdn.IsUnknown() {
		spec["useFqdn"] = plan.Spec.UseFqdn.ValueBool()
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

func populateClickhouseState(ctx context.Context, data map[string]interface{}, state *ClickhouseResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.BackupRetentionDays = getInt64(spec, "backupRetentionDays")
	state.Spec.ClickhouseVersion = getString(spec, "clickHouseVersion")
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.FullBackupIntervalDays = getInt64(spec, "fullBackupIntervalDays")
	state.Spec.KeeperVmOfferId = getString(spec, "keeperVmOfferId")
	state.Spec.KeeperVolumeOfferId = getString(spec, "keeperVolumeOfferId")
	state.Spec.KeeperVolumeSizeGib = getInt64(spec, "keeperVolumeSizeGiB")
	state.Spec.Keepers = listObjFromAPI(objList(spec, "keepers"), clickhouseKeepersObjFields)
	state.Spec.ParametersSetId = getString(spec, "parametersSetId")
	state.Spec.ReplicasPerShardGroup = getInt64(spec, "replicasPerShardGroup")
	state.Spec.ShardGroups = listObjFromAPI(objList(spec, "shardGroups"), clickhouseShardGroupsObjFields)
	state.Spec.UseFqdn = getBool(spec, "useFqdn")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"anti_affinity_message": types.StringType,
			"anti_affinity_ok":      types.BoolType,
			"any_replica_readonly":  types.BoolType,
			"ca_certificate":        types.StringType,
			"cluster_name":          types.StringType,
			"cluster_state":         types.StringType,
			"databases":             attrTypeOf("list_object", clickhouseStatusDatabasesObjFields),
			"dns_seed_fqdn":         types.StringType,
			"http_endpoints":        types.StringType,
			"https_endpoints":       types.StringType,
			"keeper_quorum_size":    types.Int64Type,
			"keepers":               attrTypeOf("list_object", clickhouseStatusKeepersObjFields),
			"native_endpoints":      types.StringType,
			"native_tls_endpoints":  types.StringType,
			"nodes":                 attrTypeOf("list_object", clickhouseStatusNodesObjFields),
			"public_endpoints":      types.StringType,
			"running_version":       types.StringType,
			"shard_groups":          attrTypeOf("list_object", clickhouseStatusShardGroupsObjFields),
		},
		map[string]attr.Value{
			"anti_affinity_message": getStringFromInfo(data, "antiAffinityMessage"),
			"anti_affinity_ok":      getBoolFromInfo(data, "antiAffinityOk"),
			"any_replica_readonly":  getBoolFromInfo(data, "anyReplicaReadonly"),
			"ca_certificate":        getStringFromInfo(data, "caCertificate"),
			"cluster_name":          getStringFromInfo(data, "clusterName"),
			"cluster_state":         getStringFromInfo(data, "clusterState"),
			"databases":             getListObjFromInfo(data, "databases", clickhouseStatusDatabasesObjFields),
			"dns_seed_fqdn":         getStringFromInfo(data, "dnsSeedFqdn"),
			"http_endpoints":        getStringFromInfo(data, "httpEndpoints"),
			"https_endpoints":       getStringFromInfo(data, "httpsEndpoints"),
			"keeper_quorum_size":    getInt64FromInfo(data, "keeperQuorumSize"),
			"keepers":               getListObjFromInfo(data, "keepers", clickhouseStatusKeepersObjFields),
			"native_endpoints":      getStringFromInfo(data, "nativeEndpoints"),
			"native_tls_endpoints":  getStringFromInfo(data, "nativeTlsEndpoints"),
			"nodes":                 getListObjFromInfo(data, "nodes", clickhouseStatusNodesObjFields),
			"public_endpoints":      getStringFromInfo(data, "publicEndpoints"),
			"running_version":       getStringFromInfo(data, "runningVersion"),
			"shard_groups":          getListObjFromInfo(data, "shardGroups", clickhouseStatusShardGroupsObjFields),
		})
	return nil
}

func (r *ClickhouseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClickhouseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildClickhouseRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/clickhouse", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/clickhouse", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/clickhouse", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateClickhouseState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateClickhouseState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClickhouseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClickhouseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateClickhouseState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ClickhouseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ClickhouseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildClickhouseRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/clickhouse", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/clickhouse", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateClickhouseState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClickhouseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClickhouseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/clickhouse", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/clickhouse", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *ClickhouseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state ClickhouseResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateClickhouseState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
