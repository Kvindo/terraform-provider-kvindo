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

// VictoriaMetricsDataSourceModel describes the data source data model.
type VictoriaMetricsDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	VpcId types.String `tfsdk:"vpc_id"`
	CreatePublicIpv4 types.Bool `tfsdk:"create_public_ipv4"`
	CreatePublicIpv6 types.Bool `tfsdk:"create_public_ipv6"`
	DnsRecordName types.String `tfsdk:"dns_record_name"`
	InfoState types.String `tfsdk:"info_state"`
	InfoPublicIpV4 types.String `tfsdk:"info_public_ip_v4"`
	InfoPublicIpV6 types.String `tfsdk:"info_public_ip_v6"`
	InfoPrivateIpV4 types.String `tfsdk:"info_private_ip_v4"`
	InfoPrivateIpV6 types.String `tfsdk:"info_private_ip_v6"`
	InfoFqdn types.String `tfsdk:"info_fqdn"`
}

type VictoriaMetricsDataSource struct {
	client *client.Client
}

func NewVictoriaMetricsDataSource() datasource.DataSource {
	return &VictoriaMetricsDataSource{}
}

func (d *VictoriaMetricsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_victoria_metrics"
}

func (d *VictoriaMetricsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{Computed: true}
	attrs["vpc_id"] = schema.StringAttribute{Computed: true}
	attrs["create_public_ipv4"] = schema.BoolAttribute{Computed: true}
	attrs["create_public_ipv6"] = schema.BoolAttribute{Computed: true}
	attrs["dns_record_name"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ip_v6"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ip_v6"] = schema.StringAttribute{Computed: true}
	attrs["info_fqdn"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *VictoriaMetricsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VictoriaMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state VictoriaMetricsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/victoria-metrics", state.ID.ValueString())
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
	state.Tier = getString(apiData, "tier")
	state.VpcId = getString(apiData, "vpcId")
	state.CreatePublicIpv4 = getBool(apiData, "createPublicIpv4")
	state.CreatePublicIpv6 = getBool(apiData, "createPublicIpv6")
	state.DnsRecordName = getString(apiData, "dnsRecordName")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoPublicIpV4 = getStringFromInfo(apiData, "publicipv4")
	state.InfoPublicIpV6 = getStringFromInfo(apiData, "publicipv6")
	state.InfoPrivateIpV4 = getStringFromInfo(apiData, "privateipv4")
	state.InfoPrivateIpV6 = getStringFromInfo(apiData, "privateipv6")
	state.InfoFqdn = getStringFromInfo(apiData, "fqdn")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
