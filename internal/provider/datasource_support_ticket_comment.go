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

// SupportTicketCommentDataSourceModel describes the data source data model.
type SupportTicketCommentDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TicketId types.String `tfsdk:"ticket_id"`
	Content types.String `tfsdk:"content"`
	AttachmentsIds types.List `tfsdk:"attachments_ids"`
}

type SupportTicketCommentDataSource struct {
	client *client.Client
}

func NewSupportTicketCommentDataSource() datasource.DataSource {
	return &SupportTicketCommentDataSource{}
}

func (d *SupportTicketCommentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_support_ticket_comment"
}

func (d *SupportTicketCommentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["ticket_id"] = schema.StringAttribute{Computed: true}
	attrs["content"] = schema.StringAttribute{Computed: true}
	attrs["attachments_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *SupportTicketCommentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SupportTicketCommentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SupportTicketCommentDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/support-ticket-comment", state.ID.ValueString())
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
	state.TicketId = getString(apiData, "ticketId")
	state.Content = getString(apiData, "content")
	state.AttachmentsIds = getStringList(ctx, apiData, "attachmentsIds")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
