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

// QuotaDataSourceModel describes the data source data model.
type QuotaDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Product types.String `tfsdk:"product"`
	Resource types.String `tfsdk:"resource"`
	Parameter types.String `tfsdk:"parameter"`
	Limit types.Int64 `tfsdk:"limit"`
	InfoState types.String `tfsdk:"info_state"`
	InfoCurrentValue types.Int64 `tfsdk:"info_current_value"`
}

type QuotaDataSource struct {
	client *client.Client
}

func NewQuotaDataSource() datasource.DataSource {
	return &QuotaDataSource{}
}

func (d *QuotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota"
}

func (d *QuotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["product"] = schema.StringAttribute{Computed: true}
	attrs["resource"] = schema.StringAttribute{Computed: true}
	attrs["parameter"] = schema.StringAttribute{Computed: true}
	attrs["limit"] = schema.Int64Attribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_current_value"] = schema.Int64Attribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *QuotaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *QuotaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state QuotaDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/quota", state.ID.ValueString())
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
	state.Product = getString(apiData, "product")
	state.Resource = getString(apiData, "resource")
	state.Parameter = getString(apiData, "parameter")
	state.Limit = getInt64(apiData, "limit")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoCurrentValue = getInt64FromInfo(apiData, "currentvalue")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
