package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type BillingAccountResourceModel struct {
	ID       types.String  `tfsdk:"id"`
	Metadata metadataModel `tfsdk:"metadata"`
	Status   types.Object  `tfsdk:"status"`
}

type BillingAccountResource struct{ client *client.Client }

func NewBillingAccountResource() resource.Resource { return &BillingAccountResource{} }

func (r *BillingAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_account"
}

func BillingAccountResourceSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"status":   commonInfoSchema(map[string]schema.Attribute{"rub_balance": schema.StringAttribute{Computed: true}}),
	}
}

func (r *BillingAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: BillingAccountResourceSchemaAttrs()}
}

func (r *BillingAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildBillingAccountRequestMap(ctx context.Context, plan BillingAccountResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	return m
}

func populateBillingAccountState(ctx context.Context, data map[string]interface{}, state *BillingAccountResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"rub_balance": types.StringType,
		},
		map[string]attr.Value{
			"rub_balance": getStringFromInfo(data, "rubBalance"),
		})
	return nil
}

func (r *BillingAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BillingAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildBillingAccountRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/billing-account", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/billing-account", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/billing-account", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateBillingAccountState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *BillingAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BillingAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/billing-account", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateBillingAccountState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *BillingAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state BillingAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildBillingAccountRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/billing-account", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/billing-account", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/billing-account", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateBillingAccountState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *BillingAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BillingAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/billing-account", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/billing-account", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *BillingAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state BillingAccountResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/billing-account", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateBillingAccountState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
