package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type SupportTicketCommentAttachmentDataSourceModel struct {
	ID       types.String                            `tfsdk:"id"`
	Name     types.String                            `tfsdk:"name"`
	Metadata metadataModel                           `tfsdk:"metadata"`
	Spec     SupportTicketCommentAttachmentSpecModel `tfsdk:"spec"`
	Status   types.Object                            `tfsdk:"status"`
}

type SupportTicketCommentAttachmentDataSource struct{ client *client.Client }

func NewSupportTicketCommentAttachmentDataSource() datasource.DataSource {
	return &SupportTicketCommentAttachmentDataSource{}
}

func (d *SupportTicketCommentAttachmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_support_ticket_comment_attachment"
}

func (d *SupportTicketCommentAttachmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"file_content_base64": schema.StringAttribute{Computed: true},
		"file_name":           schema.StringAttribute{Computed: true},
		"file_type":           schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true},
		"name":     schema.StringAttribute{Optional: true, Computed: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"download_url": schema.StringAttribute{Computed: true}}),
	}}
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set")
		return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/support-ticket-comment-attachment", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/support-ticket-comment-attachment", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	spec := getSpec(apiData)
	state.Spec.FileContentBase64 = getString(spec, "fileContentBase64")
	state.Spec.FileName = getString(spec, "fileName")
	state.Spec.FileType = getString(spec, "fileType")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"download_url": types.StringType,
		},
		map[string]attr.Value{
			"download_url": getStringFromInfo(apiData, "downloadUrl"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
