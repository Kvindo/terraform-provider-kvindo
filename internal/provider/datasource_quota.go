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

type QuotaDataSourceModel struct {
	ID       types.String    `tfsdk:"id"`
	Name     types.String    `tfsdk:"name"`
	Metadata *metadataModel  `tfsdk:"metadata"`
	Spec     *QuotaSpecModel `tfsdk:"spec"`
	Status   types.Object    `tfsdk:"status"`
}

type QuotaDataSource struct{ client *client.Client }

func NewQuotaDataSource() datasource.DataSource { return &QuotaDataSource{} }

func (d *QuotaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quota"
}

func (d *QuotaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"limit":     schema.Int64Attribute{Computed: true},
		"parameter": schema.StringAttribute{Computed: true},
		"product":   schema.StringAttribute{Computed: true},
		"resource":  schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"current_value": schema.Int64Attribute{Computed: true}}),
	}}
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
		apiData, err = d.client.Get(ctx, "/api/v1/quota", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/quota", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &QuotaSpecModel{}
	spec := getSpec(apiData)
	state.Spec.Limit = getInt64(spec, "limit")
	state.Spec.Parameter = getString(spec, "parameter")
	state.Spec.Product = getString(spec, "product")
	state.Spec.Resource = getString(spec, "resource")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"current_value": types.Int64Type,
		},
		map[string]attr.Value{
			"current_value": getInt64FromInfo(apiData, "currentValue"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
