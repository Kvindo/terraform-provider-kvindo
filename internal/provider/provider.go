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

// Ensure KvindoProvider satisfies the provider.Provider interface.
var _ provider.Provider = &KvindoProvider{}

// KvindoProvider defines the provider implementation.
type KvindoProvider struct {
	version string
}

// KvindoProviderModel describes the provider data model.
type KvindoProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

// New returns a provider.Provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KvindoProvider{version: version}
	}
}

func (p *KvindoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kvindo"
	resp.Version = p.version
}

func (p *KvindoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "The Kvindo Cloud API endpoint. Defaults to https://cloud-api.kvindo.com. Can also be set via KVINDO_ENDPOINT env var.",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API bearer token for authentication. Can also be set via KVINDO_TOKEN env var.",
			},
		},
	}
}

func (p *KvindoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config KvindoProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
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
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider requires a token. Set the token attribute in the provider block or the KVINDO_TOKEN environment variable.",
		)
		return
	}

	c := client.New(endpoint, token)
	providerData := &KvindoProviderData{Client: c}
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *KvindoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Compute
		NewVmResource,
		NewVolumeResource,
		NewVolumeAttachmentResource,
		NewImageResource,
		NewImageScheduleResource,
		NewSshKeyResource,
		NewSshPrivateKeyResource,
		NewCertificateResource,
		// Networking
		NewVpcResource,
		NewVpcSubnetResource,
		NewFloatingIpResource,
		NewSecurityGroupResource,
		NewRouteTableResource,
		NewRouteTableRouteResource,
		NewRouteTableAttachmentResource,
		NewVpcPeeringResource,
		NewVpcPeeringPeerResource,
		NewVpcPeeringExternalPeerResource,
		// IAM / Organization
		NewFolderResource,
		NewUserResource,
		NewUserTokenResource,
		NewAccessPolicyResource,
		NewBillingAccountResource,
		NewQuotaResource,
		NewQuotaChangeRequestResource,
		NewHostingProviderResource,
		// Kubernetes
		NewKubernetesResource,
		NewKubernetesNodeGroupResource,
		NewKubernetesUserResource,
		NewKubernetesUserRoleResource,
		// Load Balancer
		NewLoadbalancerResource,
		NewLoadbalancerTargetGroupResource,
		NewLoadbalancerTargetGroupStaticTargetResource,
		NewLoadbalancerTargetGroupServiceDiscoveryTargetResource,
		NewLoadbalancerHttpListenerResource,
		NewLoadbalancerHttpsListenerResource,
		NewLoadbalancerHttpListenerRuleResource,
		NewLoadbalancerHttpsListenerRuleResource,
		NewLoadbalancerTcpListenerResource,
		NewLoadbalancerTlsListenerResource,
		NewLoadbalancerTcpListenerRuleResource,
		NewLoadbalancerTlsListenerRuleResource,
		NewLoadbalancerUdpListenerResource,
		NewLoadbalancerUdpListenerRuleResource,
		// Databases
		NewPostgresqlParametersSetResource,
		NewPostgresqlStandaloneResource,
		// Object Storage
		NewS3BucketResource,
		NewS3UserResource,
		NewS3UserAccessPolicyResource,
		// Transaction
		NewTransactionResource,
		// Monitoring
		NewVictoriaMetricsResource,
		// VPN
		NewOpenVpnResource,
		NewOpenVpnUserResource,
		NewOpenVpnUserSettingsResource,
		// Dev Tools
		NewGitlabResource,
		NewGitlabRunnerResource,
		// Support
		NewSupportPlanResource,
		NewSupportTicketResource,
		NewSupportTicketCommentResource,
		NewSupportTicketCommentAttachmentResource,
	}
}

func (p *KvindoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Compute
		NewVmDataSource,
		NewVolumeDataSource,
		NewVolumeAttachmentDataSource,
		NewImageDataSource,
		NewImageScheduleDataSource,
		NewSshKeyDataSource,
		NewSshPrivateKeyDataSource,
		NewCertificateDataSource,
		// Networking
		NewVpcDataSource,
		NewVpcSubnetDataSource,
		NewFloatingIpDataSource,
		NewSecurityGroupDataSource,
		NewRouteTableDataSource,
		NewRouteTableRouteDataSource,
		NewRouteTableAttachmentDataSource,
		NewVpcPeeringDataSource,
		NewVpcPeeringPeerDataSource,
		NewVpcPeeringExternalPeerDataSource,
		// IAM / Organization
		NewFolderDataSource,
		NewUserDataSource,
		NewUserTokenDataSource,
		NewAccessPolicyDataSource,
		NewBillingAccountDataSource,
		NewQuotaDataSource,
		NewQuotaChangeRequestDataSource,
		NewHostingProviderDataSource,
		// Kubernetes
		NewKubernetesDataSource,
		NewKubernetesNodeGroupDataSource,
		NewKubernetesUserDataSource,
		NewKubernetesUserRoleDataSource,
		// Load Balancer
		NewLoadbalancerDataSource,
		NewLoadbalancerTargetGroupDataSource,
		NewLoadbalancerTargetGroupStaticTargetDataSource,
		NewLoadbalancerTargetGroupServiceDiscoveryTargetDataSource,
		NewLoadbalancerHttpListenerDataSource,
		NewLoadbalancerHttpsListenerDataSource,
		NewLoadbalancerHttpListenerRuleDataSource,
		NewLoadbalancerHttpsListenerRuleDataSource,
		NewLoadbalancerTcpListenerDataSource,
		NewLoadbalancerTlsListenerDataSource,
		NewLoadbalancerTcpListenerRuleDataSource,
		NewLoadbalancerTlsListenerRuleDataSource,
		NewLoadbalancerUdpListenerDataSource,
		NewLoadbalancerUdpListenerRuleDataSource,
		// Databases
		NewPostgresqlParametersSetDataSource,
		NewPostgresqlStandaloneDataSource,
		// Object Storage
		NewS3BucketDataSource,
		NewS3UserDataSource,
		NewS3UserAccessPolicyDataSource,
		// Monitoring
		NewVictoriaMetricsDataSource,
		// VPN
		NewOpenVpnDataSource,
		NewOpenVpnUserDataSource,
		NewOpenVpnUserSettingsDataSource,
		// Dev Tools
		NewGitlabDataSource,
		NewGitlabRunnerDataSource,
		// Support
		NewSupportPlanDataSource,
		NewSupportTicketDataSource,
		NewSupportTicketCommentDataSource,
		NewSupportTicketCommentAttachmentDataSource,
	}
}
