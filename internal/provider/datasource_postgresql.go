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

type PostgresqlDataSourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Name     types.String         `tfsdk:"name"`
	Metadata *metadataModel       `tfsdk:"metadata"`
	Spec     *PostgresqlSpecModel `tfsdk:"spec"`
	Status   types.Object         `tfsdk:"status"`
}

type PostgresqlDataSource struct{ client *client.Client }

func NewPostgresqlDataSource() datasource.DataSource { return &PostgresqlDataSource{} }

func (d *PostgresqlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql"
}

func (d *PostgresqlDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"backup_retention_days":         schema.Int64Attribute{Computed: true},
		"create_public_ipv4":            schema.BoolAttribute{Computed: true},
		"postgre_sql_parameters_set_id": schema.StringAttribute{Computed: true},
		"replicas_per_shard_group":      schema.Int64Attribute{Computed: true},
		"restore_configuration":         objDatasourceSchema(postgresqlRestoreConfigurationObjFields),
		"shard_groups":                  listObjDatasourceSchema(postgresqlShardGroupsObjFields),
		"tls_mode":                      schema.StringAttribute{Computed: true},
		"version":                       schema.StringAttribute{Computed: true},
		"vm_offer_id":                   schema.StringAttribute{Computed: true},
		"volume_offer_id":               schema.StringAttribute{Computed: true},
		"volume_size_gib":               schema.Int64Attribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "citus_managed_table_counts": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "connection_uri": schema.StringAttribute{Computed: true}, "coordinator_endpoint": schema.StringAttribute{Computed: true}, "nodes": listObjDatasourceSchema(postgresqlStatusNodesObjFields), "port": schema.Int64Attribute{Computed: true}, "primary_endpoints": schema.StringAttribute{Computed: true}, "read_endpoints": schema.StringAttribute{Computed: true}, "shard_groups": listObjDatasourceSchema(postgresqlStatusShardGroupsObjFields)}),
	}}
}

func (d *PostgresqlDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	d.client = pd.Client
}

func (d *PostgresqlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PostgresqlDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set")
		return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/postgresql", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/postgresql", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &PostgresqlSpecModel{}
	spec := getSpec(apiData)
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
	state.Status = buildInfoObj(apiData,
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
			"anti_affinity_message":      getStringFromInfo(apiData, "antiAffinityMessage"),
			"anti_affinity_ok":           getBoolFromInfo(apiData, "antiAffinityOk"),
			"citus_managed_table_counts": getStringFromInfo(apiData, "citusManagedTableCounts"),
			"cluster_state":              getStringFromInfo(apiData, "clusterState"),
			"connection_uri":             getStringFromInfo(apiData, "connectionUri"),
			"coordinator_endpoint":       getStringFromInfo(apiData, "coordinatorEndpoint"),
			"nodes":                      getListObjFromInfo(apiData, "nodes", postgresqlStatusNodesObjFields),
			"port":                       getInt64FromInfo(apiData, "port"),
			"primary_endpoints":          getStringFromInfo(apiData, "primaryEndpoints"),
			"read_endpoints":             getStringFromInfo(apiData, "readEndpoints"),
			"shard_groups":               getListObjFromInfo(apiData, "shardGroups", postgresqlStatusShardGroupsObjFields),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
