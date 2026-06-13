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

// HostingProviderDataSourceModel describes the data source data model.
type HostingProviderDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Country types.String `tfsdk:"country"`
	CountryIsoCode types.String `tfsdk:"country_iso_code"`
	City types.String `tfsdk:"city"`
	Cloud types.String `tfsdk:"cloud"`
	Sla types.Float64 `tfsdk:"sla"`
	DataCenterIndex types.Int64 `tfsdk:"data_center_index"`
	KeyFeatures types.List `tfsdk:"key_features"`
	Disabled types.Bool `tfsdk:"disabled"`
	InfoState types.String `tfsdk:"info_state"`
}

type HostingProviderDataSource struct {
	client *client.Client
}

func NewHostingProviderDataSource() datasource.DataSource {
	return &HostingProviderDataSource{}
}

func (d *HostingProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosting_provider"
}

func (d *HostingProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["country"] = schema.StringAttribute{Computed: true}
	attrs["country_iso_code"] = schema.StringAttribute{Computed: true}
	attrs["city"] = schema.StringAttribute{Computed: true}
	attrs["cloud"] = schema.StringAttribute{Computed: true}
	attrs["sla"] = schema.Float64Attribute{Computed: true}
	attrs["data_center_index"] = schema.Int64Attribute{Computed: true}
	attrs["key_features"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["disabled"] = schema.BoolAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *HostingProviderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostingProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HostingProviderDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/hosting-provider", state.ID.ValueString())
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
	state.Country = getString(apiData, "country")
	state.CountryIsoCode = getString(apiData, "countryIsoCode")
	state.City = getString(apiData, "city")
	state.Cloud = getString(apiData, "cloud")
	state.Sla = getFloat64(apiData, "sla")
	state.DataCenterIndex = getInt64(apiData, "dataCenterIndex")
	state.KeyFeatures = getStringList(ctx, apiData, "keyFeatures")
	state.Disabled = getBool(apiData, "disabled")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
