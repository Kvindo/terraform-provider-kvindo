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

type LoadbalancerTcpListenerDataSourceModel struct {
	ID       types.String                     `tfsdk:"id"`
	Metadata metadataModel                    `tfsdk:"metadata"`
	Spec     LoadbalancerTcpListenerSpecModel `tfsdk:"spec"`
	Status   types.Object                     `tfsdk:"status"`
}

type LoadbalancerTcpListenerDataSource struct{ client *client.Client }

func NewLoadbalancerTcpListenerDataSource() datasource.DataSource {
	return &LoadbalancerTcpListenerDataSource{}
}

func (d *LoadbalancerTcpListenerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tcp_listener"
}

func (d *LoadbalancerTcpListenerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"interface":       schema.StringAttribute{Computed: true},
		"loadbalancer_id": schema.StringAttribute{Computed: true},
		"order":           schema.Int64Attribute{Computed: true},
		"ports":           schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"security_rules":  listObjDatasourceSchema(loadbalancerTcpListenerSecurityRulesObjFields),
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *LoadbalancerTcpListenerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerTcpListenerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerTcpListenerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-tcp-listener", state.ID.ValueString())
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
	state.Spec.Interface = getString(spec, "interface")
	state.Spec.LoadbalancerId = getString(spec, "loadbalancerId")
	state.Spec.Order = getInt64(spec, "order")
	state.Spec.Ports = getStringList(ctx, spec, "ports")
	state.Spec.SecurityRules = listObjFromAPI(objList(spec, "securityRules"), loadbalancerTcpListenerSecurityRulesObjFields)
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
