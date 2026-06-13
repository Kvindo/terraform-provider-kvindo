package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// SupportTicketCommentAttachmentDataSourceModel describes the data source data model.
type SupportTicketCommentAttachmentDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	FileName types.String `tfsdk:"file_name"`
	FileType types.String `tfsdk:"file_type"`
	FileContentBase64 types.String `tfsdk:"file_content_base64"`
	InfoState types.String `tfsdk:"info_state"`
	InfoDownloadUrl types.String `tfsdk:"info_download_url"`
}

type SupportTicketCommentAttachmentDataSource struct {
	client *client.Client
}

func NewSupportTicketCommentAttachmentDataSource() datasource.DataSource {
	return &SupportTicketCommentAttachmentDataSource{}
}

func (d *SupportTicketCommentAttachmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_support_ticket_comment_attachment"
}

func (d *SupportTicketCommentAttachmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["file_name"] = schema.StringAttribute{Computed: true}
	attrs["file_type"] = schema.StringAttribute{Computed: true}
	attrs["file_content_base64"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_download_url"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *SupportTicketCommentAttachmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SupportTicketCommentAttachmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SupportTicketCommentAttachmentDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/support-ticket-comment-attachment", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	if err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	state.FileName = getString(apiData, "fileName")
	state.FileType = getString(apiData, "fileType")
	state.FileContentBase64 = getString(apiData, "fileContentBase64")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoDownloadUrl = getStringFromInfo(apiData, "downloadurl")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
