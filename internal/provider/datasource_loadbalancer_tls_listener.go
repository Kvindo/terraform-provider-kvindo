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

// LoadbalancerTlsListenerDataSourceModel describes the data source data model.
type LoadbalancerTlsListenerDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	LoadbalancerId types.String `tfsdk:"loadbalancer_id"`
	Interface types.String `tfsdk:"interface"`
	Order types.Int64 `tfsdk:"order"`
	Ports types.List `tfsdk:"ports"`
	Hosts types.List `tfsdk:"hosts"`
	TlsCertificateId types.String `tfsdk:"tls_certificate_id"`
	TlsProtocols types.List `tfsdk:"tls_protocols"`
	TlsAutogenerateCertificate types.Bool `tfsdk:"tls_autogenerate_certificate"`
	InfoState types.String `tfsdk:"info_state"`
}

type LoadbalancerTlsListenerDataSource struct {
	client *client.Client
}

func NewLoadbalancerTlsListenerDataSource() datasource.DataSource {
	return &LoadbalancerTlsListenerDataSource{}
}

func (d *LoadbalancerTlsListenerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_tls_listener"
}

func (d *LoadbalancerTlsListenerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["loadbalancer_id"] = schema.StringAttribute{Computed: true}
	attrs["interface"] = schema.StringAttribute{Computed: true}
	attrs["order"] = schema.Int64Attribute{Computed: true}
	attrs["ports"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["hosts"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["tls_certificate_id"] = schema.StringAttribute{Computed: true}
	attrs["tls_protocols"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["tls_autogenerate_certificate"] = schema.BoolAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
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
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
	if err := setCommonFields(ctx, apiData, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	state.LoadbalancerId = getString(apiData, "loadbalancerId")
	state.Interface = getString(apiData, "interface")
	state.Order = getInt64(apiData, "order")
	state.Ports = getStringList(ctx, apiData, "ports")
	state.Hosts = getStringList(ctx, apiData, "hosts")
	state.TlsCertificateId = getString(apiData, "tlsCertificateId")
	state.TlsProtocols = getStringList(ctx, apiData, "tlsProtocols")
	state.TlsAutogenerateCertificate = getBool(apiData, "tlsAutogenerateCertificate")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
