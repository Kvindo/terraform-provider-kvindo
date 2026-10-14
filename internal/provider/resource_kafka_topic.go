package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type KafkaTopicSpecModel struct {
	Config            types.Map    `tfsdk:"config"`
	KafkaId           types.String `tfsdk:"kafka_id"`
	Partitions        types.Int64  `tfsdk:"partitions"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
}

type KafkaTopicResourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     KafkaTopicSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type KafkaTopicResource struct{ client *client.Client }

func NewKafkaTopicResource() resource.Resource { return &KafkaTopicResource{} }

func (r *KafkaTopicResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_topic"
}

func KafkaTopicResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"config":             schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()}},
		"kafka_id":           schema.StringAttribute{Optional: true, Computed: true, Description: "The cluster this topic lives in. Immutable - a topic cannot move between clusters.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"partitions":         schema.Int64Attribute{Optional: true, Computed: true, Description: "Partition count. Increase-only: Kafka has no partition-decrease operation, so a reduction is rejected rather than silently ignored.", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"replication_factor": schema.Int64Attribute{Optional: true, Computed: true, Description: "OMIT this to let the platform resolve it - on create it takes the cluster's effective `default.replication.factor` (`min(3, broker count)` unless overridden), and on a later apply it keeps the topic's current factor. Sending `1` is NOT a way to say unset: it is an explicit request that pins the topic to a single replica, which on a 3-broker cluster silently removes all fault tolerance. An explicit value must be 1..broker count; exceeding it is rejected ASYNCHRONOUSLY by the service when it applies the change, so the failure appears on the request, not as a plan-time or submit-time error. Immutable once set.", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"offline_partitions": schema.Int64Attribute{Computed: true}, "partition_count": schema.Int64Attribute{Computed: true}, "replication_factor": schema.Int64Attribute{Computed: true}, "size_bytes": schema.Int64Attribute{Computed: true}, "under_replicated_partitions": schema.Int64Attribute{Computed: true}}),
	}
}

func (r *KafkaTopicResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: KafkaTopicResourceSchemaAttrs()}
}

func (r *KafkaTopicResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildKafkaTopicRequestMap(ctx context.Context, plan KafkaTopicResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.Config.IsNull() && !plan.Spec.Config.IsUnknown() {
		spec["config"] = stringMapToInterface(ctx, plan.Spec.Config)
	}
	if !plan.Spec.KafkaId.IsNull() && !plan.Spec.KafkaId.IsUnknown() {
		spec["kafkaId"] = plan.Spec.KafkaId.ValueString()
	}
	if !plan.Spec.Partitions.IsNull() && !plan.Spec.Partitions.IsUnknown() {
		spec["partitions"] = plan.Spec.Partitions.ValueInt64()
	}
	if !plan.Spec.ReplicationFactor.IsNull() && !plan.Spec.ReplicationFactor.IsUnknown() {
		spec["replicationFactor"] = plan.Spec.ReplicationFactor.ValueInt64()
	}
	return m
}

func populateKafkaTopicState(ctx context.Context, data map[string]interface{}, state *KafkaTopicResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.Config = getStringMap(spec, "config")
	state.Spec.KafkaId = getString(spec, "kafkaId")
	state.Spec.Partitions = getInt64(spec, "partitions")
	state.Spec.ReplicationFactor = getInt64(spec, "replicationFactor")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"offline_partitions":          types.Int64Type,
			"partition_count":             types.Int64Type,
			"replication_factor":          types.Int64Type,
			"size_bytes":                  types.Int64Type,
			"under_replicated_partitions": types.Int64Type,
		},
		map[string]attr.Value{
			"offline_partitions":          getInt64FromInfo(data, "offlinePartitions"),
			"partition_count":             getInt64FromInfo(data, "partitionCount"),
			"replication_factor":          getInt64FromInfo(data, "replicationFactor"),
			"size_bytes":                  getInt64FromInfo(data, "sizeBytes"),
			"under_replicated_partitions": getInt64FromInfo(data, "underReplicatedPartitions"),
		})
	return nil
}

func (r *KafkaTopicResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KafkaTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildKafkaTopicRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kafka-topic", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka-topic", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/kafka-topic", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateKafkaTopicState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-topic", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateKafkaTopicState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KafkaTopicResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KafkaTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-topic", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateKafkaTopicState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *KafkaTopicResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KafkaTopicResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildKafkaTopicRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kafka-topic", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka-topic", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-topic", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateKafkaTopicState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KafkaTopicResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KafkaTopicResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/kafka-topic", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka-topic", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *KafkaTopicResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state KafkaTopicResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-topic", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateKafkaTopicState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
