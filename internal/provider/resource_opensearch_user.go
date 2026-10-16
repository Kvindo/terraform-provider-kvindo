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

var opensearchUserIndexGrantsObjFields = []objField{{TF: "index_pattern", API: "indexPattern", Kind: "string"}, {TF: "pattern_type", API: "patternType", Kind: "string"}, {TF: "role", API: "role", Kind: "string"}}

type OpensearchUserSpecModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	IndexGrants  types.List   `tfsdk:"index_grants"`
	OpenSearchId types.String `tfsdk:"open_search_id"`
	Password     types.String `tfsdk:"password"`
}

type OpensearchUserResourceModel struct {
	ID       types.String            `tfsdk:"id"`
	Metadata metadataModel           `tfsdk:"metadata"`
	Spec     OpensearchUserSpecModel `tfsdk:"spec"`
	Status   types.Object            `tfsdk:"status"`
}

type OpensearchUserResource struct{ client *client.Client }

func NewOpensearchUserResource() resource.Resource { return &OpensearchUserResource{} }

func (r *OpensearchUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_opensearch_user"
}

func OpensearchUserResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"enabled":        schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"index_grants":   listObjResourceSchema(opensearchUserIndexGrantsObjFields),
		"open_search_id": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()}},
		"password":       schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *OpensearchUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: OpensearchUserResourceSchemaAttrs()}
}

func (r *OpensearchUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildOpensearchUserRequestMap(ctx context.Context, plan OpensearchUserResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.Enabled.IsNull() && !plan.Spec.Enabled.IsUnknown() {
		spec["enabled"] = plan.Spec.Enabled.ValueBool()
	}
	if !plan.Spec.IndexGrants.IsNull() && !plan.Spec.IndexGrants.IsUnknown() {
		spec["indexGrants"] = listObjToAPI(plan.Spec.IndexGrants, opensearchUserIndexGrantsObjFields)
	}
	if !plan.Spec.OpenSearchId.IsNull() && !plan.Spec.OpenSearchId.IsUnknown() {
		spec["openSearchId"] = plan.Spec.OpenSearchId.ValueString()
	}
	if !plan.Spec.Password.IsNull() && !plan.Spec.Password.IsUnknown() {
		spec["password"] = plan.Spec.Password.ValueString()
	}
	return m
}

func populateOpensearchUserState(ctx context.Context, data map[string]interface{}, state *OpensearchUserResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.Enabled = getBool(spec, "enabled")
	state.Spec.IndexGrants = listObjFromAPI(objList(spec, "indexGrants"), opensearchUserIndexGrantsObjFields)
	state.Spec.OpenSearchId = getString(spec, "openSearchId")
	if state.Spec.Password.IsNull() || state.Spec.Password.IsUnknown() {
		state.Spec.Password = getString(spec, "password")
	}
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *OpensearchUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OpensearchUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildOpensearchUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/opensearch-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/opensearch-user", modResp.RequestId); err != nil {
		if recoverData, getErr := r.client.Get(ctx, "/api/v1/opensearch-user", resourceId); getErr == nil && recoverData != nil {
			if popErr := populateOpensearchUserState(ctx, recoverData, &plan); popErr == nil {
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
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch-user", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateOpensearchUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpensearchUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OpensearchUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateOpensearchUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *OpensearchUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state OpensearchUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildOpensearchUserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/opensearch-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/opensearch-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch-user", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateOpensearchUserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *OpensearchUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OpensearchUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/opensearch-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/opensearch-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *OpensearchUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state OpensearchUserResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/opensearch-user", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateOpensearchUserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
