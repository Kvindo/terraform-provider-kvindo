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

// FloatingIpDataSourceModel describes the data source data model.
type FloatingIpDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	HostingProviderId types.String `tfsdk:"hosting_provider_id"`
	InfoState types.String `tfsdk:"info_state"`
	InfoPublicIpV4 types.String `tfsdk:"info_public_ip_v4"`
}

type FloatingIpDataSource struct {
	client *client.Client
}

func NewFloatingIpDataSource() datasource.DataSource {
	return &FloatingIpDataSource{}
}

func (d *FloatingIpDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}

func (d *FloatingIpDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["hosting_provider_id"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ip_v4"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
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
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	if err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	state.HostingProviderId = getString(apiData, "hostingProviderId")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoPublicIpV4 = getStringFromInfo(apiData, "publicipv4")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
