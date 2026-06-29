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

type LoadbalancerTlsListenerDataSourceModel struct {
	ID       types.String                     `tfsdk:"id"`
	Metadata metadataModel                    `tfsdk:"metadata"`
	Spec     LoadbalancerTlsListenerSpecModel `tfsdk:"spec"`
	Status   types.Object                     `tfsdk:"status"`
}

type LoadbalancerTlsListenerDataSource struct{ client *client.Client }

func NewLoadbalancerTlsListenerDataSource() datasource.DataSource {
	return &LoadbalancerTlsListenerDataSource{}
}

func (d *LoadbalancerTlsListenerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tls_listener"
}

func (d *LoadbalancerTlsListenerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"hosts":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"interface":       schema.StringAttribute{Computed: true},
		"loadbalancer_id": schema.StringAttribute{Computed: true},
		"order":           schema.Int64Attribute{Computed: true},
		"ports":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"security_rules":  listObjDatasourceSchema(loadbalancerTlsListenerSecurityRulesObjFields),
		"tls":             objDatasourceSchema(loadbalancerTlsListenerTlsObjFields),
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerTlsListenerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerTlsListenerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerTlsListenerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-tls-listener", state.ID.ValueString())
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
	state.Spec.Hosts = getStringList(ctx, spec, "hosts")
	state.Spec.Interface = getString(spec, "interface")
	state.Spec.LoadbalancerId = getString(spec, "loadbalancerId")
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.Ports = getStringList(ctx, spec, "ports")
	state.Spec.SecurityRules = listObjFromAPI(objList(spec, "securityRules"), loadbalancerTlsListenerSecurityRulesObjFields)
	state.Spec.Tls = objFromAPI(objMap(spec, "tls"), loadbalancerTlsListenerTlsObjFields)
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
