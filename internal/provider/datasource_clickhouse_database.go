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

type ClickhouseDatabaseDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	Metadata *metadataModel `tfsdk:"metadata"`
	Spec     *ClickhouseDatabaseSpecModel `tfsdk:"spec"`
	Status   types.Object   `tfsdk:"status"`
}

type ClickhouseDatabaseDataSource struct { client *client.Client }

func NewClickhouseDatabaseDataSource() datasource.DataSource { return &ClickhouseDatabaseDataSource{} }

func (d *ClickhouseDatabaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clickhouse_database"
}

func (d *ClickhouseDatabaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"click_house_id": schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name": schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec": schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status": commonInfoDatasourceSchema(map[string]schema.Attribute{"data_bytes_on_disk": schema.Int64Attribute{Computed: true}, "declared": schema.BoolAttribute{Computed: true}, "engine": schema.StringAttribute{Computed: true}, "table_count": schema.Int64Attribute{Computed: true}, }),
	}}
}

func (d *ClickhouseDatabaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil { return }
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok { resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData)); return }
	d.client = pd.Client
}

func (d *ClickhouseDatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ClickhouseDatabaseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set"); return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/clickhouse-database", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/clickhouse-database", state.Name.ValueString())
	}
	if err != nil { resp.Diagnostics.AddError("Read Error", err.Error()); return }
	if apiData == nil { resp.Diagnostics.AddError("Not Found", "resource not found"); return }
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil { resp.Diagnostics.AddError("State Error", err.Error()); return }
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &ClickhouseDatabaseSpecModel{}
	spec := getSpec(apiData)
	state.Spec.ClickHouseId = getString(spec, "clickHouseId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"data_bytes_on_disk": types.Int64Type,
			"declared": types.BoolType,
			"engine": types.StringType,
			"table_count": types.Int64Type,
		},
		map[string]attr.Value{
			"data_bytes_on_disk": getInt64FromInfo(apiData, "dataBytesOnDisk"),
			"declared": getBoolFromInfo(apiData, "declared"),
			"engine": getStringFromInfo(apiData, "engine"),
			"table_count": getInt64FromInfo(apiData, "tableCount"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
