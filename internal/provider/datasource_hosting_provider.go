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

type HostingProviderDataSourceModel struct {
	ID       types.String             `tfsdk:"id"`
	Name     types.String             `tfsdk:"name"`
	Metadata metadataModel            `tfsdk:"metadata"`
	Spec     HostingProviderSpecModel `tfsdk:"spec"`
	Status   types.Object             `tfsdk:"status"`
}

type HostingProviderDataSource struct{ client *client.Client }

func NewHostingProviderDataSource() datasource.DataSource { return &HostingProviderDataSource{} }

func (d *HostingProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosting_provider"
}

func (d *HostingProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"city":              schema.StringAttribute{Computed: true},
		"cloud":             schema.StringAttribute{Computed: true},
		"country":           schema.StringAttribute{Computed: true},
		"country_iso_code":  schema.StringAttribute{Computed: true},
		"data_center_index": schema.Int64Attribute{Computed: true},
		"disabled":          schema.BoolAttribute{Computed: true},
		"key_features":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"sla":               schema.Float64Attribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
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
		apiData, err = d.client.Get(ctx, "/api/v1/hosting-provider", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/hosting-provider", state.Name.ValueString())
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
	state.Spec.City = getString(spec, "city")
	state.Spec.Cloud = getString(spec, "cloud")
	state.Spec.Country = getString(spec, "country")
	state.Spec.CountryIsoCode = getString(spec, "countryIsoCode")
	state.Spec.DataCenterIndex = getInt64(spec, "dataCenterIndex")
	state.Spec.Disabled = getBool(spec, "disabled")
	state.Spec.KeyFeatures = getStringList(ctx, spec, "keyFeatures")
	state.Spec.Sla = getFloat64(spec, "sla")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
