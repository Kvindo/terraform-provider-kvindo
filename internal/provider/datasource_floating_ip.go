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

type FloatingIpDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Metadata metadataModel       `tfsdk:"metadata"`
	Spec     FloatingIpSpecModel `tfsdk:"spec"`
	Status   types.Object        `tfsdk:"status"`
}

type FloatingIpDataSource struct{ client *client.Client }

func NewFloatingIpDataSource() datasource.DataSource { return &FloatingIpDataSource{} }

func (d *FloatingIpDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}

func (d *FloatingIpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"hosting_provider_id": schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"public_ip_v4": schema.StringAttribute{Computed: true}}),
	}}
}

func (d *FloatingIpDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FloatingIpDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state FloatingIpDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/floating-ip", state.ID.ValueString())
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
	spec := getSpec(apiData)
	state.Spec.HostingProviderId = getString(spec, "hostingProviderId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"public_ip_v4": types.StringType,
		},
		map[string]attr.Value{
			"public_ip_v4": getStringFromInfo(apiData, "publicIpV4"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
