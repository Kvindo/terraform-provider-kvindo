package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type ValkeyUserSpecModel struct {
	Categories  types.List   `tfsdk:"categories"`
	Channels    types.List   `tfsdk:"channels"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	KeyPatterns types.List   `tfsdk:"key_patterns"`
	Password    types.String `tfsdk:"password"`
	ValkeyId    types.String `tfsdk:"valkey_id"`
}

type ValkeyUserResourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     ValkeyUserSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type ValkeyUserResource struct{ client *client.Client }

func NewValkeyUserResource() resource.Resource { return &ValkeyUserResource{} }

func (r *ValkeyUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_valkey_user"
}

func ValkeyUserResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"categories":   schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: "Valkey ACL command categories, e.g. `read`, `write`. Entered without the leading `+@`. Not validated by Kvindo Cloud - an invalid category is rejected by Valkey's own ACL SETUSER at apply time."},
		"channels":     schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: "Valkey ACL pub/sub channel globs, e.g. `notify:*`. Entered without the leading `&`. Empty/null denies all pub/sub access."},
		"enabled":      schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"key_patterns": schema.ListAttribute{Optional: true, ElementType: types.StringType, Description: "Valkey ACL key-pattern globs, e.g. `cache:*`. Entered without the leading `~` - Kvindo Cloud adds it when applying the ACL. Empty/null denies all key access."},
		"password":     schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, Description: "Write-only: the backend never returns this value on read. If configured, its value is preserved in state rather than overwritten by the always-empty read-back. If left unset, the platform generates a random password on create, which will never appear in state or plan output.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"valkey_id":    schema.StringAttribute{Required: true},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *ValkeyUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: ValkeyUserResourceSchemaAttrs()}
}

func (r *ValkeyUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildValkeyUserRequestMap(ctx context.Context, plan ValkeyUserResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.Categories.IsNull() && !plan.Spec.Categories.IsUnknown() {
		spec["categories"] = stringListToInterface(ctx, plan.Spec.Categories)
	}
	if !plan.Spec.Channels.IsNull() && !plan.Spec.Channels.IsUnknown() {
		spec["channels"] = stringListToInterface(ctx, plan.Spec.Channels)
	}
	if !plan.Spec.Enabled.IsNull() && !plan.Spec.Enabled.IsUnknown() {
		spec["enabled"] = plan.Spec.Enabled.ValueBool()
	}
	if !plan.Spec.KeyPatterns.IsNull() && !plan.Spec.KeyPatterns.IsUnknown() {
		spec["keyPatterns"] = stringListToInterface(ctx, plan.Spec.KeyPatterns)
	}
	if !plan.Spec.Password.IsNull() && !plan.Spec.Password.IsUnknown() {
		spec["password"] = plan.Spec.Password.ValueString()
	}
	if !plan.Spec.ValkeyId.IsNull() && !plan.Spec.ValkeyId.IsUnknown() {
		spec["valkeyId"] = plan.Spec.ValkeyId.ValueString()
	}
	return m
}

func populateValkeyUserState(ctx context.Context, data map[string]interface{}, state *ValkeyUserResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.Categories = getStringList(ctx, spec, "categories")
	state.Spec.Channels = getStringList(ctx, spec, "channels")
	state.Spec.Enabled = getBool(spec, "enabled")
	state.Spec.KeyPatterns = getStringList(ctx, spec, "keyPatterns")
	if state.Spec.Password.IsNull() || state.Spec.Password.IsUnknown() {
		state.Spec.Password = getString(spec, "password")
	}
	state.Spec.ValkeyId = getString(spec, "valkeyId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *ValkeyUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ValkeyUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildValkeyUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/valkey-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/valkey-user", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/valkey-user", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateValkeyUserState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/valkey-user", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateValkeyUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ValkeyUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ValkeyUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/valkey-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateValkeyUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *ValkeyUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ValkeyUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildValkeyUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/valkey-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/valkey-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/valkey-user", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateValkeyUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ValkeyUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ValkeyUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/valkey-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/valkey-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *ValkeyUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state ValkeyUserResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/valkey-user", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateValkeyUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
