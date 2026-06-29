package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

const defaultEndpoint = "https://cloud-api.kvindo.com"

var _ provider.Provider = &KvindoProvider{}

type KvindoProvider struct{ version string }
type KvindoProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &KvindoProvider{version: version} }
}

func (p *KvindoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kvindo"
	resp.Version = p.version
}

func (p *KvindoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"endpoint": schema.StringAttribute{Optional: true, Description: "API endpoint, defaults to https://cloud-api.kvindo.com"},
		"token":    schema.StringAttribute{Optional: true, Sensitive: true, Description: "API bearer token"},
	}}
}

func (p *KvindoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config KvindoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := defaultEndpoint
	if !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() && config.Endpoint.ValueString() != "" {
		endpoint = config.Endpoint.ValueString()
	} else if v := os.Getenv("KVINDO_ENDPOINT"); v != "" {
		endpoint = v
	}
	token := ""
	if !config.Token.IsNull() && !config.Token.IsUnknown() {
		token = config.Token.ValueString()
	}
	if token == "" {
		token = os.Getenv("KVINDO_TOKEN")
	}
	if token == "" {
		resp.Diagnostics.AddError("Missing API Token", "Set token in provider config or KVINDO_TOKEN env var")
		return
	}
	pd := &KvindoProviderData{Client: client.New(endpoint, token, p.version)}
	resp.DataSourceData = pd
	resp.ResourceData = pd
}

func (p *KvindoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAccessPolicyResource,
		NewBillingAccountResource,
		NewCertificateResource,
		NewFloatingIpResource,
		NewFolderResource,
		NewGitlabResource,
		NewGitlabRunnerResource,
		NewHostingProviderResource,
		NewImageResource,
		NewImageScheduleResource,
		NewKubernetesResource,
		NewKubernetesNodeGroupResource,
		NewKubernetesUserResource,
		NewKubernetesUserRoleResource,
		NewLoadbalancerResource,
		NewLoadbalancerHttpListenerResource,
		NewLoadbalancerHttpListenerRuleResource,
		NewLoadbalancerHttpsListenerResource,
		NewLoadbalancerHttpsListenerRuleResource,
		NewLoadbalancerTargetGroupResource,
		NewLoadbalancerTargetGroupServiceDiscoveryTargetResource,
		NewLoadbalancerTargetGroupStaticTargetResource,
		NewLoadbalancerTcpListenerResource,
		NewLoadbalancerTcpListenerRuleResource,
		NewLoadbalancerTlsListenerResource,
		NewLoadbalancerTlsListenerRuleResource,
		NewLoadbalancerUdpListenerResource,
		NewLoadbalancerUdpListenerRuleResource,
		NewOpenVpnResource,
		NewOpenVpnUserResource,
		NewOpenVpnUserSettingsResource,
		NewPostgresqlParametersSetResource,
		NewPostgresqlStandaloneResource,
		NewQuotaResource,
		NewQuotaChangeRequestResource,
		NewRouteTableResource,
		NewRouteTableAttachmentResource,
		NewRouteTableRouteResource,
		NewS3BucketResource,
		NewS3UserResource,
		NewS3UserAccessPolicyResource,
		NewSecurityGroupResource,
		NewSshKeyResource,
		NewSshPrivateKeyResource,
		NewSupportPlanResource,
		NewSupportTicketResource,
		NewSupportTicketCommentResource,
		NewSupportTicketCommentAttachmentResource,
		NewUserResource,
		NewUserTokenResource,
		NewVictoriaMetricsResource,
		NewVmResource,
		NewVolumeResource,
		NewVolumeAttachmentResource,
		NewVpcResource,
		NewVpcPeeringResource,
		NewVpcPeeringExternalPeerResource,
		NewVpcPeeringPeerResource,
		NewVpcSubnetResource,
		NewTransactionResource,
	}
}

func (p *KvindoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccessPolicyDataSource,
		NewBillingAccountDataSource,
		NewCertificateDataSource,
		NewFloatingIpDataSource,
		NewFolderDataSource,
		NewGitlabDataSource,
		NewGitlabRunnerDataSource,
		NewHostingProviderDataSource,
		NewImageDataSource,
		NewImageScheduleDataSource,
		NewKubernetesDataSource,
		NewKubernetesNodeGroupDataSource,
		NewKubernetesUserDataSource,
		NewKubernetesUserRoleDataSource,
		NewLoadbalancerDataSource,
		NewLoadbalancerHttpListenerDataSource,
		NewLoadbalancerHttpListenerRuleDataSource,
		NewLoadbalancerHttpsListenerDataSource,
		NewLoadbalancerHttpsListenerRuleDataSource,
		NewLoadbalancerTargetGroupDataSource,
		NewLoadbalancerTargetGroupServiceDiscoveryTargetDataSource,
		NewLoadbalancerTargetGroupStaticTargetDataSource,
		NewLoadbalancerTcpListenerDataSource,
		NewLoadbalancerTcpListenerRuleDataSource,
		NewLoadbalancerTlsListenerDataSource,
		NewLoadbalancerTlsListenerRuleDataSource,
		NewLoadbalancerUdpListenerDataSource,
		NewLoadbalancerUdpListenerRuleDataSource,
		NewOpenVpnDataSource,
		NewOpenVpnUserDataSource,
		NewOpenVpnUserSettingsDataSource,
		NewPostgresqlParametersSetDataSource,
		NewPostgresqlStandaloneDataSource,
		NewQuotaDataSource,
		NewQuotaChangeRequestDataSource,
		NewRouteTableDataSource,
		NewRouteTableAttachmentDataSource,
		NewRouteTableRouteDataSource,
		NewS3BucketDataSource,
		NewS3UserDataSource,
		NewS3UserAccessPolicyDataSource,
		NewSecurityGroupDataSource,
		NewSshKeyDataSource,
		NewSshPrivateKeyDataSource,
		NewSupportPlanDataSource,
		NewSupportTicketDataSource,
		NewSupportTicketCommentDataSource,
		NewSupportTicketCommentAttachmentDataSource,
		NewUserDataSource,
		NewUserTokenDataSource,
		NewVictoriaMetricsDataSource,
		NewVmDataSource,
		NewVolumeDataSource,
		NewVolumeAttachmentDataSource,
		NewVpcDataSource,
		NewVpcPeeringDataSource,
		NewVpcPeeringExternalPeerDataSource,
		NewVpcPeeringPeerDataSource,
		NewVpcSubnetDataSource,
	}
}
