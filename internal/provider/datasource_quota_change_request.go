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

// QuotaChangeRequestDataSourceModel describes the data source data model.
type QuotaChangeRequestDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	QuotaId types.String `tfsdk:"quota_id"`
	NewQuotaLimit types.Int64 `tfsdk:"new_quota_limit"`
	InfoState types.String `tfsdk:"info_state"`
	InfoTicketId types.String `tfsdk:"info_ticket_id"`
}

type QuotaChangeRequestDataSource struct {
	client *client.Client
}

func NewQuotaChangeRequestDataSource() datasource.DataSource {
	return &QuotaChangeRequestDataSource{}
}

func (d *QuotaChangeRequestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota_change_request"
}

func (d *QuotaChangeRequestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["quota_id"] = schema.StringAttribute{Computed: true}
	attrs["new_quota_limit"] = schema.Int64Attribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_ticket_id"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *QuotaChangeRequestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *QuotaChangeRequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state QuotaChangeRequestDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/quota-change-request", state.ID.ValueString())
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
	state.QuotaId = getString(apiData, "quotaId")
	state.NewQuotaLimit = getInt64(apiData, "newQuotaLimit")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoTicketId = getStringFromInfo(apiData, "ticketid")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
