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

type OpensearchDataSourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Name     types.String         `tfsdk:"name"`
	Metadata *metadataModel       `tfsdk:"metadata"`
	Spec     *OpensearchSpecModel `tfsdk:"spec"`
	Status   types.Object         `tfsdk:"status"`
}

type OpensearchDataSource struct{ client *client.Client }

func NewOpensearchDataSource() datasource.DataSource { return &OpensearchDataSource{} }

func (d *OpensearchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_opensearch"
}

func (d *OpensearchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"cluster_managers":        listObjDatasourceSchema(opensearchClusterManagersObjFields),
		"create_public_ipv4":      schema.BoolAttribute{Computed: true},
		"data_nodes":              listObjDatasourceSchema(opensearchDataNodesObjFields),
		"data_vm_offer_id":        schema.StringAttribute{Computed: true},
		"data_volume_offer_id":    schema.StringAttribute{Computed: true},
		"data_volume_size_gib":    schema.Int64Attribute{Computed: true},
		"manager_vm_offer_id":     schema.StringAttribute{Computed: true},
		"manager_volume_offer_id": schema.StringAttribute{Computed: true},
		"manager_volume_size_gib": schema.Int64Attribute{Computed: true},
		"open_search_version":     schema.StringAttribute{Computed: true},
		"parameters_set_id":       schema.StringAttribute{Computed: true},
		"use_fqdn":                schema.BoolAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"active_cluster_manager_count": schema.Int64Attribute{Computed: true}, "active_shards_percent": schema.Int64Attribute{Computed: true}, "anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "any_index_readonly": schema.BoolAttribute{Computed: true}, "ca_certificate": schema.StringAttribute{Computed: true}, "cluster_manager_quorum_size": schema.Int64Attribute{Computed: true}, "cluster_managers": listObjDatasourceSchema(opensearchStatusClusterManagersObjFields), "cluster_name": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "data_nodes": listObjDatasourceSchema(opensearchStatusDataNodesObjFields), "default_number_of_replicas": schema.Int64Attribute{Computed: true}, "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "health": schema.StringAttribute{Computed: true}, "index_count": schema.Int64Attribute{Computed: true}, "last_probe_utc": schema.Int64Attribute{Computed: true}, "red_index_count": schema.Int64Attribute{Computed: true}, "relocating_shards": schema.Int64Attribute{Computed: true}, "rest_endpoints": schema.StringAttribute{Computed: true}, "rest_public_endpoints": schema.StringAttribute{Computed: true}, "running_version": schema.StringAttribute{Computed: true}, "total_data_bytes": schema.Int64Attribute{Computed: true}, "total_documents": schema.Int64Attribute{Computed: true}, "unassigned_shards": schema.Int64Attribute{Computed: true}, "yellow_index_count": schema.Int64Attribute{Computed: true}}),
	}}
}

func (d *OpensearchDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OpensearchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OpensearchDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/opensearch", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/opensearch", state.Name.ValueString())
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
	state.Spec = &OpensearchSpecModel{}
	spec := getSpec(apiData)
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
	state.Status = buildInfoObj(apiData,
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
			"active_cluster_manager_count": getInt64FromInfo(apiData, "activeClusterManagerCount"),
			"active_shards_percent":        getInt64FromInfo(apiData, "activeShardsPercent"),
			"anti_affinity_message":        getStringFromInfo(apiData, "antiAffinityMessage"),
			"anti_affinity_ok":             getBoolFromInfo(apiData, "antiAffinityOk"),
			"any_index_readonly":           getBoolFromInfo(apiData, "anyIndexReadonly"),
			"ca_certificate":               getStringFromInfo(apiData, "caCertificate"),
			"cluster_manager_quorum_size":  getInt64FromInfo(apiData, "clusterManagerQuorumSize"),
			"cluster_managers":             getListObjFromInfo(apiData, "clusterManagers", opensearchStatusClusterManagersObjFields),
			"cluster_name":                 getStringFromInfo(apiData, "clusterName"),
			"cluster_state":                getStringFromInfo(apiData, "clusterState"),
			"data_nodes":                   getListObjFromInfo(apiData, "dataNodes", opensearchStatusDataNodesObjFields),
			"default_number_of_replicas":   getInt64FromInfo(apiData, "defaultNumberOfReplicas"),
			"dns_seed_fqdn":                getStringFromInfo(apiData, "dnsSeedFqdn"),
			"health":                       getStringFromInfo(apiData, "health"),
			"index_count":                  getInt64FromInfo(apiData, "indexCount"),
			"last_probe_utc":               getInt64FromInfo(apiData, "lastProbeUtc"),
			"red_index_count":              getInt64FromInfo(apiData, "redIndexCount"),
			"relocating_shards":            getInt64FromInfo(apiData, "relocatingShards"),
			"rest_endpoints":               getStringFromInfo(apiData, "restEndpoints"),
			"rest_public_endpoints":        getStringFromInfo(apiData, "restPublicEndpoints"),
			"running_version":              getStringFromInfo(apiData, "runningVersion"),
			"total_data_bytes":             getInt64FromInfo(apiData, "totalDataBytes"),
			"total_documents":              getInt64FromInfo(apiData, "totalDocuments"),
			"unassigned_shards":            getInt64FromInfo(apiData, "unassignedShards"),
			"yellow_index_count":           getInt64FromInfo(apiData, "yellowIndexCount"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
