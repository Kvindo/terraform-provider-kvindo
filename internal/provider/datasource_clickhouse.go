package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type ClickhouseDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Metadata *metadataModel `tfsdk:"metadata"`
	Spec     *ClickhouseSpecModel `tfsdk:"spec"`
	Status   types.Object   `tfsdk:"status"`
}

type ClickhouseDataSource struct { client *client.Client }

func NewClickhouseDataSource() datasource.DataSource { return &ClickhouseDataSource{} }

func (d *ClickhouseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse"
}

func (d *ClickhouseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"backup_retention_days": schema.Int64Attribute{Computed: true},
		"click_house_version": schema.StringAttribute{Computed: true},
		"create_public_ipv4": schema.BoolAttribute{Computed: true},
		"full_backup_interval_days": schema.Int64Attribute{Computed: true},
		"keeper_vm_offer_id": schema.StringAttribute{Computed: true},
		"keeper_volume_offer_id": schema.StringAttribute{Computed: true},
		"keeper_volume_size_gib": schema.Int64Attribute{Computed: true},
		"keepers": listObjDatasourceSchema(clickhouseKeepersObjFields),
		"parameters_set_id": schema.StringAttribute{Computed: true},
		"replicas_per_shard_group": schema.Int64Attribute{Computed: true},
		"shard_groups": listObjDatasourceSchema(clickhouseShardGroupsObjFields),
		"use_fqdn": schema.BoolAttribute{Computed: true},
		"vm_offer_id": schema.StringAttribute{Computed: true},
		"volume_offer_id": schema.StringAttribute{Computed: true},
		"volume_size_gib": schema.Int64Attribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name": schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec": schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status": commonInfoDatasourceSchema(map[string]schema.Attribute{"anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "any_replica_readonly": schema.BoolAttribute{Computed: true}, "ca_certificate": schema.StringAttribute{Computed: true}, "cluster_name": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "databases": listObjDatasourceSchema(clickhouseStatusDatabasesObjFields), "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "http_endpoints": schema.StringAttribute{Computed: true}, "https_endpoints": schema.StringAttribute{Computed: true}, "keeper_quorum_size": schema.Int64Attribute{Computed: true}, "keepers": listObjDatasourceSchema(clickhouseStatusKeepersObjFields), "native_endpoints": schema.StringAttribute{Computed: true}, "native_tls_endpoints": schema.StringAttribute{Computed: true}, "nodes": listObjDatasourceSchema(clickhouseStatusNodesObjFields), "public_endpoints": schema.StringAttribute{Computed: true}, "running_version": schema.StringAttribute{Computed: true}, "shard_groups": listObjDatasourceSchema(clickhouseStatusShardGroupsObjFields), }),
	}}
}

func (d *ClickhouseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok { resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData)); return }
	d.client = pd.Client
}

func (d *ClickhouseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ClickhouseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set"); return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/clickhouse", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/clickhouse", state.Name.ValueString())
	}
	if err != nil { resp.Diagnostics.AddError("Read Error", err.Error()); return }
	if apiData == nil { resp.Diagnostics.AddError("Not Found", "resource not found"); return }
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil { resp.Diagnostics.AddError("State Error", err.Error()); return }
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &ClickhouseSpecModel{}
	spec := getSpec(apiData)
	state.Spec.BackupRetentionDays = getInt64(spec, "backupRetentionDays")
	state.Spec.ClickHouseVersion = getString(spec, "clickHouseVersion")
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
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"anti_affinity_message": types.StringType,
			"anti_affinity_ok": types.BoolType,
			"any_replica_readonly": types.BoolType,
			"ca_certificate": types.StringType,
			"cluster_name": types.StringType,
			"cluster_state": types.StringType,
			"databases": attrTypeOf("list_object", clickhouseStatusDatabasesObjFields),
			"dns_seed_fqdn": types.StringType,
			"http_endpoints": types.StringType,
			"https_endpoints": types.StringType,
			"keeper_quorum_size": types.Int64Type,
			"keepers": attrTypeOf("list_object", clickhouseStatusKeepersObjFields),
			"native_endpoints": types.StringType,
			"native_tls_endpoints": types.StringType,
			"nodes": attrTypeOf("list_object", clickhouseStatusNodesObjFields),
			"public_endpoints": types.StringType,
			"running_version": types.StringType,
			"shard_groups": attrTypeOf("list_object", clickhouseStatusShardGroupsObjFields),
		},
		map[string]attr.Value{
			"anti_affinity_message": getStringFromInfo(apiData, "antiAffinityMessage"),
			"anti_affinity_ok": getBoolFromInfo(apiData, "antiAffinityOk"),
			"any_replica_readonly": getBoolFromInfo(apiData, "anyReplicaReadonly"),
			"ca_certificate": getStringFromInfo(apiData, "caCertificate"),
			"cluster_name": getStringFromInfo(apiData, "clusterName"),
			"cluster_state": getStringFromInfo(apiData, "clusterState"),
			"databases": getListObjFromInfo(apiData, "databases", clickhouseStatusDatabasesObjFields),
			"dns_seed_fqdn": getStringFromInfo(apiData, "dnsSeedFqdn"),
			"http_endpoints": getStringFromInfo(apiData, "httpEndpoints"),
			"https_endpoints": getStringFromInfo(apiData, "httpsEndpoints"),
			"keeper_quorum_size": getInt64FromInfo(apiData, "keeperQuorumSize"),
			"keepers": getListObjFromInfo(apiData, "keepers", clickhouseStatusKeepersObjFields),
			"native_endpoints": getStringFromInfo(apiData, "nativeEndpoints"),
			"native_tls_endpoints": getStringFromInfo(apiData, "nativeTlsEndpoints"),
			"nodes": getListObjFromInfo(apiData, "nodes", clickhouseStatusNodesObjFields),
			"public_endpoints": getStringFromInfo(apiData, "publicEndpoints"),
			"running_version": getStringFromInfo(apiData, "runningVersion"),
			"shard_groups": getListObjFromInfo(apiData, "shardGroups", clickhouseStatusShardGroupsObjFields),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
