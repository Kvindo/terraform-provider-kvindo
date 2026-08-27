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

var postgresqlRestoreConfigurationObjFields = []objField{{TF: "postgre_sql_id", API: "postgreSqlId", Kind: "string"}, {TF: "restore_time", API: "restoreTime", Kind: "string"}}

var postgresqlShardGroupsObjFields = []objField{{TF: "is_coordinator", API: "isCoordinator", Kind: "bool"}, {TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string"}}

var postgresqlStatusNodesObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "is_primary", API: "isPrimary", Kind: "bool"}, {TF: "observed_role", API: "observedRole", Kind: "string"}, {TF: "patroni_state", API: "patroniState", Kind: "string"}, {TF: "port", API: "port", Kind: "int64"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "public_ipv4", API: "publicIpV4", Kind: "string"}, {TF: "replication_lag_bytes", API: "replicationLagBytes", Kind: "int64"}, {TF: "shard_group_index", API: "shardGroupIndex", Kind: "int64"}}

var postgresqlStatusShardGroupsObjFields = []objField{{TF: "index", API: "index", Kind: "int64"}, {TF: "is_coordinator", API: "isCoordinator", Kind: "bool"}, {TF: "primary_endpoint", API: "primaryEndpoint", Kind: "string"}, {TF: "primary_instance_id", API: "primaryInstanceId", Kind: "string"}, {TF: "replica_endpoints", API: "replicaEndpoints", Kind: "list_string"}, {TF: "replica_instance_ids", API: "replicaInstanceIds", Kind: "list_string"}, {TF: "shard_count", API: "shardCount", Kind: "int64"}}

type PostgresqlSpecModel struct {
	BackupRetentionDays       types.Int64  `tfsdk:"backup_retention_days"`
	CreatePublicIpv4          types.Bool   `tfsdk:"create_public_ipv4"`
	PostgreSqlParametersSetId types.String `tfsdk:"postgre_sql_parameters_set_id"`
	ReplicasPerShardGroup     types.Int64  `tfsdk:"replicas_per_shard_group"`
	RestoreConfiguration      types.Object `tfsdk:"restore_configuration"`
	ShardGroups               types.List   `tfsdk:"shard_groups"`
	TlsMode                   types.String `tfsdk:"tls_mode"`
	Version                   types.String `tfsdk:"version"`
	VmOfferId                 types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId             types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib             types.Int64  `tfsdk:"volume_size_gib"`
}

type PostgresqlResourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     PostgresqlSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type PostgresqlResource struct{ client *client.Client }

func NewPostgresqlResource() resource.Resource { return &PostgresqlResource{} }

func (r *PostgresqlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql"
}

func PostgresqlResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"backup_retention_days":         schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"create_public_ipv4":            schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"postgre_sql_parameters_set_id": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"replicas_per_shard_group":      schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"restore_configuration":         objResourceSchema(postgresqlRestoreConfigurationObjFields),
		"shard_groups":                  listObjResourceSchema(postgresqlShardGroupsObjFields),
		"tls_mode":                      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"version":                       schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vm_offer_id":                   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_offer_id":               schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"volume_size_gib":               schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "citus_managed_table_counts": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "connection_uri": schema.StringAttribute{Computed: true}, "coordinator_endpoint": schema.StringAttribute{Computed: true}, "nodes": listObjStatusSchema(postgresqlStatusNodesObjFields), "port": schema.Int64Attribute{Computed: true}, "primary_endpoints": schema.StringAttribute{Computed: true}, "read_endpoints": schema.StringAttribute{Computed: true}, "shard_groups": listObjStatusSchema(postgresqlStatusShardGroupsObjFields)}),
	}
}

func (r *PostgresqlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: PostgresqlResourceSchemaAttrs()}
}

func (r *PostgresqlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildPostgresqlRequestMap(ctx context.Context, plan PostgresqlResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.BackupRetentionDays.IsNull() && !plan.Spec.BackupRetentionDays.IsUnknown() {
		spec["backupRetentionDays"] = plan.Spec.BackupRetentionDays.ValueInt64()
	}
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.PostgreSqlParametersSetId.IsNull() && !plan.Spec.PostgreSqlParametersSetId.IsUnknown() {
		spec["postgreSqlParametersSetId"] = plan.Spec.PostgreSqlParametersSetId.ValueString()
	}
	if !plan.Spec.ReplicasPerShardGroup.IsNull() && !plan.Spec.ReplicasPerShardGroup.IsUnknown() {
		spec["replicasPerShardGroup"] = plan.Spec.ReplicasPerShardGroup.ValueInt64()
	}
	if !plan.Spec.RestoreConfiguration.IsNull() && !plan.Spec.RestoreConfiguration.IsUnknown() {
		spec["restoreConfiguration"] = objToAPI(plan.Spec.RestoreConfiguration, postgresqlRestoreConfigurationObjFields)
	}
	if !plan.Spec.ShardGroups.IsNull() && !plan.Spec.ShardGroups.IsUnknown() {
		spec["shardGroups"] = listObjToAPI(plan.Spec.ShardGroups, postgresqlShardGroupsObjFields)
	}
	if !plan.Spec.TlsMode.IsNull() && !plan.Spec.TlsMode.IsUnknown() {
		spec["tlsMode"] = plan.Spec.TlsMode.ValueString()
	}
	if !plan.Spec.Version.IsNull() && !plan.Spec.Version.IsUnknown() {
		spec["version"] = plan.Spec.Version.ValueString()
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

func populatePostgresqlState(ctx context.Context, data map[string]interface{}, state *PostgresqlResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.BackupRetentionDays = getInt64(spec, "backupRetentionDays")
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.PostgreSqlParametersSetId = getString(spec, "postgreSqlParametersSetId")
	state.Spec.ReplicasPerShardGroup = getInt64(spec, "replicasPerShardGroup")
	state.Spec.RestoreConfiguration = objFromAPI(objMap(spec, "restoreConfiguration"), postgresqlRestoreConfigurationObjFields)
	state.Spec.ShardGroups = listObjFromAPI(objList(spec, "shardGroups"), postgresqlShardGroupsObjFields)
	state.Spec.TlsMode = getString(spec, "tlsMode")
	state.Spec.Version = getString(spec, "version")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"anti_affinity_message":      types.StringType,
			"anti_affinity_ok":           types.BoolType,
			"citus_managed_table_counts": types.StringType,
			"cluster_state":              types.StringType,
			"connection_uri":             types.StringType,
			"coordinator_endpoint":       types.StringType,
			"nodes":                      attrTypeOf("list_object", postgresqlStatusNodesObjFields),
			"port":                       types.Int64Type,
			"primary_endpoints":          types.StringType,
			"read_endpoints":             types.StringType,
			"shard_groups":               attrTypeOf("list_object", postgresqlStatusShardGroupsObjFields),
		},
		map[string]attr.Value{
			"anti_affinity_message":      getStringFromInfo(data, "antiAffinityMessage"),
			"anti_affinity_ok":           getBoolFromInfo(data, "antiAffinityOk"),
			"citus_managed_table_counts": getStringFromInfo(data, "citusManagedTableCounts"),
			"cluster_state":              getStringFromInfo(data, "clusterState"),
			"connection_uri":             getStringFromInfo(data, "connectionUri"),
			"coordinator_endpoint":       getStringFromInfo(data, "coordinatorEndpoint"),
			"nodes":                      getListObjFromInfo(data, "nodes", postgresqlStatusNodesObjFields),
			"port":                       getInt64FromInfo(data, "port"),
			"primary_endpoints":          getStringFromInfo(data, "primaryEndpoints"),
			"read_endpoints":             getStringFromInfo(data, "readEndpoints"),
			"shard_groups":               getListObjFromInfo(data, "shardGroups", postgresqlStatusShardGroupsObjFields),
		})
	return nil
}

func (r *PostgresqlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PostgresqlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildPostgresqlRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/postgresql", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/postgresql", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/postgresql", resourceId); getErr == nil && recoverData != nil {
			if popErr := populatePostgresqlState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/postgresql", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populatePostgresqlState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PostgresqlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PostgresqlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/postgresql", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populatePostgresqlState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *PostgresqlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PostgresqlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildPostgresqlRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/postgresql", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/postgresql", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/postgresql", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populatePostgresqlState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PostgresqlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PostgresqlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/postgresql", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/postgresql", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *PostgresqlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state PostgresqlResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/postgresql", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populatePostgresqlState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
