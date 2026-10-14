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

var kafkaBrokersObjFields = []objField{{TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string", Immutable: true}}

var kafkaControllersObjFields = []objField{{TF: "name", API: "name", Kind: "string"}, {TF: "vpc_subnet_id", API: "vpcSubnetId", Kind: "string", Immutable: true}}

var kafkaStatusBrokersObjFields = []objField{{TF: "announce_fqdn", API: "announceFqdn", Kind: "string"}, {TF: "id", API: "id", Kind: "string"}, {TF: "leader_partition_count", API: "leaderPartitionCount", Kind: "int64"}, {TF: "log_dir_used_bytes", API: "logDirUsedBytes", Kind: "int64"}, {TF: "name", API: "name", Kind: "string"}, {TF: "node_id", API: "nodeId", Kind: "int64"}, {TF: "observed_state", API: "observedState", Kind: "string"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}, {TF: "public_ipv4", API: "publicIpV4", Kind: "string"}, {TF: "under_replicated_partitions", API: "underReplicatedPartitions", Kind: "int64"}}

var kafkaStatusControllersObjFields = []objField{{TF: "id", API: "id", Kind: "string"}, {TF: "is_quorum_leader", API: "isQuorumLeader", Kind: "bool"}, {TF: "lag_offsets", API: "lagOffsets", Kind: "int64"}, {TF: "name", API: "name", Kind: "string"}, {TF: "node_id", API: "nodeId", Kind: "int64"}, {TF: "observed_state", API: "observedState", Kind: "string"}, {TF: "private_ipv4", API: "privateIpV4", Kind: "string"}}

type KafkaSpecModel struct {
	BrokerVmOfferId         types.String `tfsdk:"broker_vm_offer_id"`
	BrokerVolumeOfferId     types.String `tfsdk:"broker_volume_offer_id"`
	BrokerVolumeSizeGib     types.Int64  `tfsdk:"broker_volume_size_gib"`
	Brokers                 types.List   `tfsdk:"brokers"`
	ControllerVmOfferId     types.String `tfsdk:"controller_vm_offer_id"`
	ControllerVolumeOfferId types.String `tfsdk:"controller_volume_offer_id"`
	ControllerVolumeSizeGib types.Int64  `tfsdk:"controller_volume_size_gib"`
	Controllers             types.List   `tfsdk:"controllers"`
	CreatePublicIpv4        types.Bool   `tfsdk:"create_public_ipv4"`
	KafkaVersion            types.String `tfsdk:"kafka_version"`
	ParametersSetId         types.String `tfsdk:"parameters_set_id"`
	UseFqdn                 types.Bool   `tfsdk:"use_fqdn"`
}

type KafkaResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Metadata metadataModel  `tfsdk:"metadata"`
	Spec     KafkaSpecModel `tfsdk:"spec"`
	Status   types.Object   `tfsdk:"status"`
}

type KafkaResource struct{ client *client.Client }

func NewKafkaResource() resource.Resource { return &KafkaResource{} }

func (r *KafkaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka"
}

func KafkaResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"broker_vm_offer_id":         schema.StringAttribute{Optional: true, Computed: true, Description: "Compute offer applied to EVERY broker. One triplet per role rather than per member: a role's nodes are interchangeable by design, so a heterogeneous role cannot be expressed. Omitting it resolves to the role default and the resolved value is written back.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"broker_volume_offer_id":     schema.StringAttribute{Optional: true, Computed: true, Description: "Disk offer applied to every broker's log volume.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"broker_volume_size_gib":     schema.Int64Attribute{Optional: true, Computed: true, Description: "Log volume size per broker, in GiB. IMMUTABLE after create: the volume is the node's log directory and nothing grows the filesystem inside the guest, so plan growth rather than assuming a later resize.", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown(), int64planmodifier.RequiresReplace()}},
		"brokers":                    listObjResourceSchema(kafkaBrokersObjFields, "Broker members, one entry each. 1..N and grow-only: Kafka has no broker shrink, so removing an entry is rejected rather than silently ignored. `name` is the real identity - reusing a name across applies keeps that broker and its VM, while renaming is treated as remove-and-add."),
		"controller_vm_offer_id":     schema.StringAttribute{Optional: true, Computed: true, Description: "Compute offer applied to EVERY controller. A controller only serves the metadata quorum and holds no tenant data, so it is sized far smaller than a broker. Immutable after create - resizing a voter is a dynamic-quorum operation.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"controller_volume_offer_id": schema.StringAttribute{Optional: true, Computed: true, Description: "Disk offer applied to every controller. Immutable after create.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"controller_volume_size_gib": schema.Int64Attribute{Optional: true, Computed: true, Description: "Metadata volume size per controller, in GiB. Immutable after create.", PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown(), int64planmodifier.RequiresReplace()}},
		"controllers":                listObjResourceSchema(kafkaControllersObjFields, "KRaft controller members. EXACTLY 1 or 3 - the metadata quorum needs an odd voter count, and 1 is a supported non-HA tier that can later be grown to 3. Shrinking is rejected, and an existing controller cannot be moved or resized."),
		"create_public_ipv4":         schema.BoolAttribute{Optional: true, Computed: true, Description: "Add the public TLS listener (port 9096) on a floating IP. The plaintext listener stays bound to the private network, so a public cluster is never reachable without TLS. Immutable: it adds or removes a whole advertised listener.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()}},
		"kafka_version":              schema.StringAttribute{Optional: true, Computed: true, Description: "Full patch version, and it must be one the deployed image prefetched - Kafka is installed only from the tarball shipped in that image, so an unprefetched version cannot be installed. IMMUTABLE: there is no rolling upgrade in v1.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"parameters_set_id":          schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"use_fqdn":                   schema.BoolAttribute{Optional: true, Computed: true, Description: "Advertise DNS names instead of raw IPs, so a node's address can change without every client re-seeding. Immutable: it rewrites every broker's advertised listeners, which connected clients have already cached.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"active_controller_count": schema.Int64Attribute{Computed: true}, "anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "bootstrap_servers_plaintext": schema.StringAttribute{Computed: true}, "bootstrap_servers_public": schema.StringAttribute{Computed: true}, "bootstrap_servers_tls": schema.StringAttribute{Computed: true}, "brokers": listObjStatusSchema(kafkaStatusBrokersObjFields), "ca_certificate": schema.StringAttribute{Computed: true}, "cluster_id": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "controllers": listObjStatusSchema(kafkaStatusControllersObjFields), "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "offline_partitions": schema.Int64Attribute{Computed: true}, "quorum_leader_node_id": schema.Int64Attribute{Computed: true}, "under_replicated_partitions": schema.Int64Attribute{Computed: true}, "unexpected_topics": schema.StringAttribute{Computed: true}}),
	}
}

func (r *KafkaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: KafkaResourceSchemaAttrs()}
}

func (r *KafkaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildKafkaRequestMap(ctx context.Context, plan KafkaResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.BrokerVmOfferId.IsNull() && !plan.Spec.BrokerVmOfferId.IsUnknown() {
		spec["brokerVmOfferId"] = plan.Spec.BrokerVmOfferId.ValueString()
	}
	if !plan.Spec.BrokerVolumeOfferId.IsNull() && !plan.Spec.BrokerVolumeOfferId.IsUnknown() {
		spec["brokerVolumeOfferId"] = plan.Spec.BrokerVolumeOfferId.ValueString()
	}
	if !plan.Spec.BrokerVolumeSizeGib.IsNull() && !plan.Spec.BrokerVolumeSizeGib.IsUnknown() {
		spec["brokerVolumeSizeGiB"] = plan.Spec.BrokerVolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.Brokers.IsNull() && !plan.Spec.Brokers.IsUnknown() {
		spec["brokers"] = listObjToAPI(plan.Spec.Brokers, kafkaBrokersObjFields)
	}
	if !plan.Spec.ControllerVmOfferId.IsNull() && !plan.Spec.ControllerVmOfferId.IsUnknown() {
		spec["controllerVmOfferId"] = plan.Spec.ControllerVmOfferId.ValueString()
	}
	if !plan.Spec.ControllerVolumeOfferId.IsNull() && !plan.Spec.ControllerVolumeOfferId.IsUnknown() {
		spec["controllerVolumeOfferId"] = plan.Spec.ControllerVolumeOfferId.ValueString()
	}
	if !plan.Spec.ControllerVolumeSizeGib.IsNull() && !plan.Spec.ControllerVolumeSizeGib.IsUnknown() {
		spec["controllerVolumeSizeGiB"] = plan.Spec.ControllerVolumeSizeGib.ValueInt64()
	}
	if !plan.Spec.Controllers.IsNull() && !plan.Spec.Controllers.IsUnknown() {
		spec["controllers"] = listObjToAPI(plan.Spec.Controllers, kafkaControllersObjFields)
	}
	if !plan.Spec.CreatePublicIpv4.IsNull() && !plan.Spec.CreatePublicIpv4.IsUnknown() {
		spec["createPublicIpv4"] = plan.Spec.CreatePublicIpv4.ValueBool()
	}
	if !plan.Spec.KafkaVersion.IsNull() && !plan.Spec.KafkaVersion.IsUnknown() {
		spec["kafkaVersion"] = plan.Spec.KafkaVersion.ValueString()
	}
	if !plan.Spec.ParametersSetId.IsNull() && !plan.Spec.ParametersSetId.IsUnknown() {
		spec["parametersSetId"] = plan.Spec.ParametersSetId.ValueString()
	}
	if !plan.Spec.UseFqdn.IsNull() && !plan.Spec.UseFqdn.IsUnknown() {
		spec["useFqdn"] = plan.Spec.UseFqdn.ValueBool()
	}
	return m
}

func populateKafkaState(ctx context.Context, data map[string]interface{}, state *KafkaResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
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
	state.Status = buildInfoObj(data,
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
			"active_controller_count":     getInt64FromInfo(data, "activeControllerCount"),
			"anti_affinity_message":       getStringFromInfo(data, "antiAffinityMessage"),
			"anti_affinity_ok":            getBoolFromInfo(data, "antiAffinityOk"),
			"bootstrap_servers_plaintext": getStringFromInfo(data, "bootstrapServersPlaintext"),
			"bootstrap_servers_public":    getStringFromInfo(data, "bootstrapServersPublic"),
			"bootstrap_servers_tls":       getStringFromInfo(data, "bootstrapServersTls"),
			"brokers":                     getListObjFromInfo(data, "brokers", kafkaStatusBrokersObjFields),
			"ca_certificate":              getStringFromInfo(data, "caCertificate"),
			"cluster_id":                  getStringFromInfo(data, "clusterId"),
			"cluster_state":               getStringFromInfo(data, "clusterState"),
			"controllers":                 getListObjFromInfo(data, "controllers", kafkaStatusControllersObjFields),
			"dns_seed_fqdn":               getStringFromInfo(data, "dnsSeedFqdn"),
			"offline_partitions":          getInt64FromInfo(data, "offlinePartitions"),
			"quorum_leader_node_id":       getInt64FromInfo(data, "quorumLeaderNodeId"),
			"under_replicated_partitions": getInt64FromInfo(data, "underReplicatedPartitions"),
			"unexpected_topics":           getStringFromInfo(data, "unexpectedTopics"),
		})
	return nil
}

func (r *KafkaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KafkaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildKafkaRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kafka", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/kafka", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateKafkaState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/kafka", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateKafkaState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KafkaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KafkaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kafka", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateKafkaState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *KafkaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KafkaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildKafkaRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kafka", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kafka", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateKafkaState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KafkaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KafkaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/kafka", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *KafkaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state KafkaResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/kafka", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateKafkaState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
