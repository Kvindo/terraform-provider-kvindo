package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var kafkaUserTopicGrantsObjFields = []objField{{TF: "pattern_type", API: "patternType", Kind: "string"}, {TF: "role", API: "role", Kind: "string"}, {TF: "topic_pattern", API: "topicPattern", Kind: "string"}}

type KafkaUserSpecModel struct {
	ConsumerGroupPrefixes types.List   `tfsdk:"consumer_group_prefixes"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	KafkaId               types.String `tfsdk:"kafka_id"`
	Password              types.String `tfsdk:"password"`
	TopicGrants           types.List   `tfsdk:"topic_grants"`
}

type KafkaUserResourceModel struct {
	ID       types.String       `tfsdk:"id"`
	Metadata metadataModel      `tfsdk:"metadata"`
	Spec     KafkaUserSpecModel `tfsdk:"spec"`
	Status   types.Object       `tfsdk:"status"`
}

type KafkaUserResource struct{ client *client.Client }

func NewKafkaUserResource() resource.Resource { return &KafkaUserResource{} }

func (r *KafkaUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_user"
}

func KafkaUserResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"consumer_group_prefixes": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "Consumer-group ACLs, prefix-matched; `\"*\"` is an accepted explicit value for all groups. REQUIRED as soon as any grant is a `Consumer` - a consumer that cannot reach its group cannot commit offsets, so the request is rejected at submit time rather than failing later with an authorization error that points nowhere useful.", PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"enabled":                 schema.BoolAttribute{Optional: true, Computed: true, Description: "Authentication switch. `false` LOCKS the user: the SCRAM credential is overwritten with an unusable random password while every ACL stays in place, so re-enabling is one write instead of a full re-grant.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"kafka_id":                schema.StringAttribute{Optional: true, Computed: true, Description: "The cluster this SCRAM principal belongs to. Immutable - a user cannot move between clusters.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"password":                schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Description: "Write-only: the backend never returns this value on read. If configured, its value is preserved in state rather than overwritten by the always-empty read-back. Unlike valkey_user/postgresql_user there is NO server-side generation: KafkaUserResourceController rejects a CREATE with `enabled` true and an empty password (`KafkaUser requires Password when Enabled is true`), so it must be supplied - a create with `enabled` false and no password is allowed (a pre-locked principal). Omitting it on a MODIFY is legitimate and carries the existing credential forward from the change request.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"topic_grants":            listObjResourceSchema(kafkaUserTopicGrantsObjFields, "Role-based grants, not raw Kafka ACL tuples: `Producer` or `Consumer` on a `Literal` topic name or a `Prefixed` pattern, and the platform expands each into the operations Kafka actually needs (Write+Describe, or Read+Describe plus the group ACLs). There is no Admin role in this release."),
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"acl_count": schema.Int64Attribute{Computed: true}, "credential_present": schema.BoolAttribute{Computed: true}}),
	}
}

func (r *KafkaUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: KafkaUserResourceSchemaAttrs()}
}

func (r *KafkaUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildKafkaUserRequestMap(ctx context.Context, plan KafkaUserResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ConsumerGroupPrefixes.IsNull() && !plan.Spec.ConsumerGroupPrefixes.IsUnknown() {
		spec["consumerGroupPrefixes"] = stringListToInterface(ctx, plan.Spec.ConsumerGroupPrefixes)
	}
	if !plan.Spec.Enabled.IsNull() && !plan.Spec.Enabled.IsUnknown() {
		spec["enabled"] = plan.Spec.Enabled.ValueBool()
	}
	if !plan.Spec.KafkaId.IsNull() && !plan.Spec.KafkaId.IsUnknown() {
		spec["kafkaId"] = plan.Spec.KafkaId.ValueString()
	}
	if !plan.Spec.Password.IsNull() && !plan.Spec.Password.IsUnknown() {
		spec["password"] = plan.Spec.Password.ValueString()
	}
	if !plan.Spec.TopicGrants.IsNull() && !plan.Spec.TopicGrants.IsUnknown() {
		spec["topicGrants"] = listObjToAPI(plan.Spec.TopicGrants, kafkaUserTopicGrantsObjFields)
	}
	return m
}

func populateKafkaUserState(ctx context.Context, data map[string]interface{}, state *KafkaUserResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ConsumerGroupPrefixes = getStringList(ctx, spec, "consumerGroupPrefixes")
	state.Spec.Enabled = getBool(spec, "enabled")
	state.Spec.KafkaId = getString(spec, "kafkaId")
	if state.Spec.Password.IsNull() || state.Spec.Password.IsUnknown() {
		state.Spec.Password = getString(spec, "password")
	}
	state.Spec.TopicGrants = listObjFromAPI(objList(spec, "topicGrants"), kafkaUserTopicGrantsObjFields)
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"acl_count":          types.Int64Type,
			"credential_present": types.BoolType,
		},
		map[string]attr.Value{
			"acl_count":          getInt64FromInfo(data, "aclCount"),
			"credential_present": getBoolFromInfo(data, "credentialPresent"),
		})
	return nil
}

func (r *KafkaUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan KafkaUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildKafkaUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kafka-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka-user", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/kafka-user", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateKafkaUserState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-user", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateKafkaUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KafkaUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KafkaUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateKafkaUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *KafkaUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KafkaUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildKafkaUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/kafka-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-user", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateKafkaUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *KafkaUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KafkaUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/kafka-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/kafka-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *KafkaUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state KafkaUserResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/kafka-user", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateKafkaUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
