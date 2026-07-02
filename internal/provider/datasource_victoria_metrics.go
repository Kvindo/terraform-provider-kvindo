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

type VictoriaMetricsDataSourceModel struct {
	ID       types.String              `tfsdk:"id"`
	Name     types.String              `tfsdk:"name"`
	Metadata *metadataModel            `tfsdk:"metadata"`
	Spec     *VictoriaMetricsSpecModel `tfsdk:"spec"`
	Status   types.Object              `tfsdk:"status"`
}

type VictoriaMetricsDataSource struct{ client *client.Client }

func NewVictoriaMetricsDataSource() datasource.DataSource { return &VictoriaMetricsDataSource{} }

func (d *VictoriaMetricsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_victoria_metrics"
}

func (d *VictoriaMetricsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"constraints":        listObjDatasourceSchema(victoriaMetricsConstraintsObjFields),
		"create_public_ipv4": schema.BoolAttribute{Computed: true},
		"create_public_ipv6": schema.BoolAttribute{Computed: true},
		"dns_record_name":    schema.StringAttribute{Computed: true},
		"scrap_targets":      listObjDatasourceSchema(victoriaMetricsScrapTargetsObjFields),
		"tier":               schema.StringAttribute{Computed: true},
		"vpc_id":             schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"discovered_scrap_targets": schema.StringAttribute{Computed: true}, "fqdn": schema.StringAttribute{Computed: true}, "private_ipv4": schema.StringAttribute{Computed: true}, "private_ipv6": schema.StringAttribute{Computed: true}, "public_ipv4": schema.StringAttribute{Computed: true}, "public_ipv6": schema.StringAttribute{Computed: true}}),
	}}
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
		apiData, err = d.client.Get(ctx, "/api/v1/victoria-metrics", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/victoria-metrics", state.Name.ValueString())
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
	state.Spec = &VictoriaMetricsSpecModel{}
	spec := getSpec(apiData)
	state.Spec.Constraints = listObjFromAPI(objList(spec, "constraints"), victoriaMetricsConstraintsObjFields)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.CreatePublicIpv6 = getBool(spec, "createPublicIpv6")
	state.Spec.DnsRecordName = getString(spec, "dnsRecordName")
	state.Spec.ScrapTargets = listObjFromAPI(objList(spec, "scrapTargets"), victoriaMetricsScrapTargetsObjFields)
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.VpcId = getString(spec, "vpcId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"discovered_scrap_targets": types.StringType,
			"fqdn":                     types.StringType,
			"private_ipv4":             types.StringType,
			"private_ipv6":             types.StringType,
			"public_ipv4":              types.StringType,
			"public_ipv6":              types.StringType,
		},
		map[string]attr.Value{
			"discovered_scrap_targets": getStringFromInfo(apiData, "discoveredScrapTargets"),
			"fqdn":                     getStringFromInfo(apiData, "fqdn"),
			"private_ipv4":             getStringFromInfo(apiData, "privateIpV4"),
			"private_ipv6":             getStringFromInfo(apiData, "privateIpV6"),
			"public_ipv4":              getStringFromInfo(apiData, "publicIpV4"),
			"public_ipv6":              getStringFromInfo(apiData, "publicIpV6"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
