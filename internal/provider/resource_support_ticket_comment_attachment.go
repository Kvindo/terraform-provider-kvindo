package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

// SupportTicketCommentAttachmentResourceModel describes the resource data model.
type SupportTicketCommentAttachmentResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	FileName types.String `tfsdk:"file_name"`
	FileType types.String `tfsdk:"file_type"`
	FileContentBase64 types.String `tfsdk:"file_content_base64"`
	Info types.Object `tfsdk:"info"`
}

// SupportTicketCommentAttachmentResource defines the resource implementation.
type SupportTicketCommentAttachmentResource struct {
	client *client.Client
}

func NewSupportTicketCommentAttachmentResource() resource.Resource {
	return &SupportTicketCommentAttachmentResource{}
}

func (r *SupportTicketCommentAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_support_ticket_comment_attachment"
}

func (r *SupportTicketCommentAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["file_name"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["file_type"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["file_content_base64"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "download_url": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *SupportTicketCommentAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildSupportTicketCommentAttachmentRequestMap(ctx context.Context, plan SupportTicketCommentAttachmentResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.FileName.IsNull() && !plan.FileName.IsUnknown() {
		m["fileName"] = plan.FileName.ValueString()
	}
	if !plan.FileType.IsNull() && !plan.FileType.IsUnknown() {
		m["fileType"] = plan.FileType.ValueString()
	}
	if !plan.FileContentBase64.IsNull() && !plan.FileContentBase64.IsUnknown() {
		m["fileContentBase64"] = plan.FileContentBase64.ValueString()
	}
	return m
}

func populateSupportTicketCommentAttachmentState(ctx context.Context, data map[string]interface{}, state *SupportTicketCommentAttachmentResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.FileName = getString(data, "fileName")
	state.FileType = getString(data, "fileType")
	state.FileContentBase64 = getString(data, "fileContentBase64")
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "download_url": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "download_url": getStringFromInfo(data, "downloadUrl")})
	return nil
}

func (r *SupportTicketCommentAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SupportTicketCommentAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildSupportTicketCommentAttachmentRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/support-ticket-comment-attachment", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/support-ticket-comment-attachment", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/support-ticket-comment-attachment", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateSupportTicketCommentAttachmentState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SupportTicketCommentAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SupportTicketCommentAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/support-ticket-comment-attachment", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateSupportTicketCommentAttachmentState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *SupportTicketCommentAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SupportTicketCommentAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state SupportTicketCommentAttachmentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildSupportTicketCommentAttachmentRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/support-ticket-comment-attachment", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/support-ticket-comment-attachment", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/support-ticket-comment-attachment", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateSupportTicketCommentAttachmentState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SupportTicketCommentAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SupportTicketCommentAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/support-ticket-comment-attachment", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/support-ticket-comment-attachment", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *SupportTicketCommentAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state SupportTicketCommentAttachmentResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/support-ticket-comment-attachment", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateSupportTicketCommentAttachmentState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
