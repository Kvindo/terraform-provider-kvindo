package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

var clickhouseUserDatabaseGrantsObjFields = []objField{{TF: "database_name", API: "databaseName", Kind: "string"}, {TF: "role", API: "role", Kind: "string"}}

type ClickhouseUserSpecModel struct {
	ClickhouseId   types.String `tfsdk:"clickhouse_id"`
	DatabaseGrants types.List   `tfsdk:"database_grants"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Password       types.String `tfsdk:"password"`
	Settings       types.Map    `tfsdk:"settings"`
}

type ClickhouseUserResourceModel struct {
	ID       types.String            `tfsdk:"id"`
	Metadata metadataModel           `tfsdk:"metadata"`
	Spec     ClickhouseUserSpecModel `tfsdk:"spec"`
	Status   types.Object            `tfsdk:"status"`
}

type ClickhouseUserResource struct{ client *client.Client }

func NewClickhouseUserResource() resource.Resource { return &ClickhouseUserResource{} }

func (r *ClickhouseUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse_user"
}

func ClickhouseUserResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"clickhouse_id":   schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"database_grants": listObjResourceSchema(clickhouseUserDatabaseGrantsObjFields),
		"enabled":         schema.BoolAttribute{Optional: true, Computed: true, Description: "Authentication switch. `false` locks the user; re-enabling is one write rather than a full re-grant.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"password":        schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Description: "Write-only: the backend never returns this value on read. If configured, its value is preserved in state rather than overwritten by the always-empty read-back. Unlike valkey_user/postgresql_user there is NO server-side generation: ClickHouseUserResourceController rejects a CREATE with `enabled` true and an empty password (`ClickHouseUser requires Password when Enabled is true`), so it must be supplied - a create with `enabled` false and no password is allowed (a pre-locked principal).", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"settings":        schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *ClickhouseUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: ClickhouseUserResourceSchemaAttrs()}
}

func (r *ClickhouseUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildClickhouseUserRequestMap(ctx context.Context, plan ClickhouseUserResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ClickhouseId.IsNull() && !plan.Spec.ClickhouseId.IsUnknown() {
		spec["clickHouseId"] = plan.Spec.ClickhouseId.ValueString()
	}
	if !plan.Spec.DatabaseGrants.IsNull() && !plan.Spec.DatabaseGrants.IsUnknown() {
		spec["databaseGrants"] = listObjToAPI(plan.Spec.DatabaseGrants, clickhouseUserDatabaseGrantsObjFields)
	}
	if !plan.Spec.Enabled.IsNull() && !plan.Spec.Enabled.IsUnknown() {
		spec["enabled"] = plan.Spec.Enabled.ValueBool()
	}
	if !plan.Spec.Password.IsNull() && !plan.Spec.Password.IsUnknown() {
		spec["password"] = plan.Spec.Password.ValueString()
	}
	if !plan.Spec.Settings.IsNull() && !plan.Spec.Settings.IsUnknown() {
		spec["settings"] = stringMapToInterface(ctx, plan.Spec.Settings)
	}
	return m
}

func populateClickhouseUserState(ctx context.Context, data map[string]interface{}, state *ClickhouseUserResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ClickhouseId = getString(spec, "clickHouseId")
	state.Spec.DatabaseGrants = listObjFromAPI(objList(spec, "databaseGrants"), clickhouseUserDatabaseGrantsObjFields)
	state.Spec.Enabled = getBool(spec, "enabled")
	if state.Spec.Password.IsNull() || state.Spec.Password.IsUnknown() {
		state.Spec.Password = getString(spec, "password")
	}
	state.Spec.Settings = getStringMap(spec, "settings")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *ClickhouseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClickhouseUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildClickhouseUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/clickhouse-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/clickhouse-user", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/clickhouse-user", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateClickhouseUserState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse-user", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateClickhouseUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClickhouseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClickhouseUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateClickhouseUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ClickhouseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ClickhouseUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildClickhouseUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/clickhouse-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/clickhouse-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse-user", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateClickhouseUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ClickhouseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClickhouseUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/clickhouse-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/clickhouse-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *ClickhouseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state ClickhouseUserResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/clickhouse-user", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateClickhouseUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
