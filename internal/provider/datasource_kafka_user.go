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

type KafkaUserDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Name     types.String        `tfsdk:"name"`
	Metadata *metadataModel      `tfsdk:"metadata"`
	Spec     *KafkaUserSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type KafkaUserDataSource struct{ client *client.Client }

func NewKafkaUserDataSource() datasource.DataSource { return &KafkaUserDataSource{} }

func (d *KafkaUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_user"
}

func (d *KafkaUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"cluster_operations":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"consumer_group_prefixes": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"enabled":                 schema.BoolAttribute{Computed: true},
		"kafka_id":                schema.StringAttribute{Computed: true},
		"password":                schema.StringAttribute{Computed: true, Sensitive: true},
		"topic_grants":            listObjDatasourceSchema(kafkaUserTopicGrantsObjFields),
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"acl_count": schema.Int64Attribute{Computed: true}, "credential_present": schema.BoolAttribute{Computed: true}}),
	}}
}

func (d *KafkaUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KafkaUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KafkaUserDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/kafka-user", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/kafka-user", state.Name.ValueString())
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
	state.Spec = &KafkaUserSpecModel{}
	spec := getSpec(apiData)
	state.Spec.ClusterOperations = getStringList(ctx, spec, "clusterOperations")
	state.Spec.ConsumerGroupPrefixes = getStringList(ctx, spec, "consumerGroupPrefixes")
	state.Spec.Enabled = getBool(spec, "enabled")
	state.Spec.KafkaId = getString(spec, "kafkaId")
	state.Spec.Password = getString(spec, "password")
	state.Spec.TopicGrants = listObjFromAPI(objList(spec, "topicGrants"), kafkaUserTopicGrantsObjFields)
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"acl_count":          types.Int64Type,
			"credential_present": types.BoolType,
		},
		map[string]attr.Value{
			"acl_count":          getInt64FromInfo(apiData, "aclCount"),
			"credential_present": getBoolFromInfo(apiData, "credentialPresent"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
