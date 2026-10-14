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

type KafkaDataSourceModel struct {
	ID       types.String    `tfsdk:"id"`
	Name     types.String    `tfsdk:"name"`
	Metadata *metadataModel  `tfsdk:"metadata"`
	Spec     *KafkaSpecModel `tfsdk:"spec"`
	Status   types.Object    `tfsdk:"status"`
}

type KafkaDataSource struct{ client *client.Client }

func NewKafkaDataSource() datasource.DataSource { return &KafkaDataSource{} }

func (d *KafkaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka"
}

func (d *KafkaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"broker_vm_offer_id":         schema.StringAttribute{Computed: true},
		"broker_volume_offer_id":     schema.StringAttribute{Computed: true},
		"broker_volume_size_gib":     schema.Int64Attribute{Computed: true},
		"brokers":                    listObjDatasourceSchema(kafkaBrokersObjFields),
		"controller_vm_offer_id":     schema.StringAttribute{Computed: true},
		"controller_volume_offer_id": schema.StringAttribute{Computed: true},
		"controller_volume_size_gib": schema.Int64Attribute{Computed: true},
		"controllers":                listObjDatasourceSchema(kafkaControllersObjFields),
		"create_public_ipv4":         schema.BoolAttribute{Computed: true},
		"kafka_version":              schema.StringAttribute{Computed: true},
		"parameters_set_id":          schema.StringAttribute{Computed: true},
		"use_fqdn":                   schema.BoolAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"active_controller_count": schema.Int64Attribute{Computed: true}, "anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "bootstrap_servers_plaintext": schema.StringAttribute{Computed: true}, "bootstrap_servers_public": schema.StringAttribute{Computed: true}, "bootstrap_servers_tls": schema.StringAttribute{Computed: true}, "brokers": listObjDatasourceSchema(kafkaStatusBrokersObjFields), "ca_certificate": schema.StringAttribute{Computed: true}, "cluster_id": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "controllers": listObjDatasourceSchema(kafkaStatusControllersObjFields), "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "offline_partitions": schema.Int64Attribute{Computed: true}, "quorum_leader_node_id": schema.Int64Attribute{Computed: true}, "under_replicated_partitions": schema.Int64Attribute{Computed: true}, "unexpected_topics": schema.StringAttribute{Computed: true}}),
	}}
}

func (d *KafkaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KafkaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KafkaDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/kafka", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/kafka", state.Name.ValueString())
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
	state.Spec = &KafkaSpecModel{}
	spec := getSpec(apiData)
	state.Spec.BrokerVmOfferId = getString(spec, "brokerVmOfferId")
	state.Spec.BrokerVolumeOfferId = getString(spec, "brokerVolumeOfferId")
	state.Spec.BrokerVolumeSizeGib = getInt64(spec, "brokerVolumeSizeGiB")
	state.Spec.Brokers = listObjFromAPI(objList(spec, "brokers"), kafkaBrokersObjFields)
	state.Spec.ControllerVmOfferId = getString(spec, "controllerVmOfferId")
	state.Spec.ControllerVolumeOfferId = getString(spec, "controllerVolumeOfferId")
	state.Spec.ControllerVolumeSizeGib = getInt64(spec, "controllerVolumeSizeGiB")
	state.Spec.Controllers = listObjFromAPI(objList(spec, "controllers"), kafkaControllersObjFields)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.KafkaVersion = getString(spec, "kafkaVersion")
	state.Spec.ParametersSetId = getString(spec, "parametersSetId")
	state.Spec.UseFqdn = getBool(spec, "useFqdn")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"active_controller_count":     types.Int64Type,
			"anti_affinity_message":       types.StringType,
			"anti_affinity_ok":            types.BoolType,
			"bootstrap_servers_plaintext": types.StringType,
			"bootstrap_servers_public":    types.StringType,
			"bootstrap_servers_tls":       types.StringType,
			"brokers":                     attrTypeOf("list_object", kafkaStatusBrokersObjFields),
			"ca_certificate":              types.StringType,
			"cluster_id":                  types.StringType,
			"cluster_state":               types.StringType,
			"controllers":                 attrTypeOf("list_object", kafkaStatusControllersObjFields),
			"dns_seed_fqdn":               types.StringType,
			"offline_partitions":          types.Int64Type,
			"quorum_leader_node_id":       types.Int64Type,
			"under_replicated_partitions": types.Int64Type,
			"unexpected_topics":           types.StringType,
		},
		map[string]attr.Value{
			"active_controller_count":     getInt64FromInfo(apiData, "activeControllerCount"),
			"anti_affinity_message":       getStringFromInfo(apiData, "antiAffinityMessage"),
			"anti_affinity_ok":            getBoolFromInfo(apiData, "antiAffinityOk"),
			"bootstrap_servers_plaintext": getStringFromInfo(apiData, "bootstrapServersPlaintext"),
			"bootstrap_servers_public":    getStringFromInfo(apiData, "bootstrapServersPublic"),
			"bootstrap_servers_tls":       getStringFromInfo(apiData, "bootstrapServersTls"),
			"brokers":                     getListObjFromInfo(apiData, "brokers", kafkaStatusBrokersObjFields),
			"ca_certificate":              getStringFromInfo(apiData, "caCertificate"),
			"cluster_id":                  getStringFromInfo(apiData, "clusterId"),
			"cluster_state":               getStringFromInfo(apiData, "clusterState"),
			"controllers":                 getListObjFromInfo(apiData, "controllers", kafkaStatusControllersObjFields),
			"dns_seed_fqdn":               getStringFromInfo(apiData, "dnsSeedFqdn"),
			"offline_partitions":          getInt64FromInfo(apiData, "offlinePartitions"),
			"quorum_leader_node_id":       getInt64FromInfo(apiData, "quorumLeaderNodeId"),
			"under_replicated_partitions": getInt64FromInfo(apiData, "underReplicatedPartitions"),
			"unexpected_topics":           getStringFromInfo(apiData, "unexpectedTopics"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
