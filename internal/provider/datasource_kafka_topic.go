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

type KafkaTopicDataSourceModel struct {
	ID       types.String         `tfsdk:"id"`
	Name     types.String         `tfsdk:"name"`
	Metadata *metadataModel       `tfsdk:"metadata"`
	Spec     *KafkaTopicSpecModel `tfsdk:"spec"`
	Status   types.Object         `tfsdk:"status"`
}

type KafkaTopicDataSource struct{ client *client.Client }

func NewKafkaTopicDataSource() datasource.DataSource { return &KafkaTopicDataSource{} }

func (d *KafkaTopicDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_topic"
}

func (d *KafkaTopicDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"config":             schema.MapAttribute{Computed: true, ElementType: types.StringType},
		"kafka_id":           schema.StringAttribute{Computed: true},
		"partitions":         schema.Int64Attribute{Computed: true},
		"replication_factor": schema.Int64Attribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"offline_partitions": schema.Int64Attribute{Computed: true}, "partition_count": schema.Int64Attribute{Computed: true}, "replication_factor": schema.Int64Attribute{Computed: true}, "size_bytes": schema.Int64Attribute{Computed: true}, "under_replicated_partitions": schema.Int64Attribute{Computed: true}}),
	}}
}

func (d *KafkaTopicDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KafkaTopicDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KafkaTopicDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/kafka-topic", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/kafka-topic", state.Name.ValueString())
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
	state.Spec = &KafkaTopicSpecModel{}
	spec := getSpec(apiData)
	state.Spec.Config = getStringMap(spec, "config")
	state.Spec.KafkaId = getString(spec, "kafkaId")
	state.Spec.Partitions = getInt64(spec, "partitions")
	state.Spec.ReplicationFactor = getInt64(spec, "replicationFactor")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"offline_partitions":          types.Int64Type,
			"partition_count":             types.Int64Type,
			"replication_factor":          types.Int64Type,
			"size_bytes":                  types.Int64Type,
			"under_replicated_partitions": types.Int64Type,
		},
		map[string]attr.Value{
			"offline_partitions":          getInt64FromInfo(apiData, "offlinePartitions"),
			"partition_count":             getInt64FromInfo(apiData, "partitionCount"),
			"replication_factor":          getInt64FromInfo(apiData, "replicationFactor"),
			"size_bytes":                  getInt64FromInfo(apiData, "sizeBytes"),
			"under_replicated_partitions": getInt64FromInfo(apiData, "underReplicatedPartitions"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
