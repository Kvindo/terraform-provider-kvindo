package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	Endpoint   types.String `tfsdk:"endpoint"`
	Token      types.String `tfsdk:"token"`
	CliProfile types.String `tfsdk:"cli_profile"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &KvindoProvider{version: version} }
}

func (p *KvindoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kvindo"
	resp.Version = p.version
}

func (p *KvindoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Kvindo Cloud provider manages [Kvindo Cloud](https://cloud.kvindo.com) infrastructure as code: VMs, S3 object storage, Kubernetes clusters, load balancers, VPCs, VPNs, managed PostgreSQL, networking, and IAM.",
		Attributes: map[string]schema.Attribute{
			"endpoint":    schema.StringAttribute{Optional: true, Description: "API endpoint, defaults to https://cloud-api.kvindo.com"},
			"token":       schema.StringAttribute{Optional: true, Sensitive: true, Description: "API bearer token"},
			"cli_profile": schema.StringAttribute{Optional: true, Description: "Name of a kc CLI profile (~/.kc/config/<name>.yaml, written by `kc login`/`kc switch`) to source endpoint/token from when they aren't set directly. Lower precedence than endpoint/token and their KVINDO_ENDPOINT/KVINDO_TOKEN env vars. Can also be set via KVINDO_CLI_PROFILE."},
		}}
}

func (p *KvindoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config KvindoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	profileName := ""
	if !config.CliProfile.IsNull() && !config.CliProfile.IsUnknown() && config.CliProfile.ValueString() != "" {
		profileName = config.CliProfile.ValueString()
	} else if v := os.Getenv("KVINDO_CLI_PROFILE"); v != "" {
		profileName = v
	}
	var profile kcProfile
	profileFound := true
	if profileName != "" {
		profile, profileFound = loadKcProfile(profileName)
	}

	endpoint := defaultEndpoint
	if profile.Server != "" {
		endpoint = profile.Server
	}
	if v := os.Getenv("KVINDO_ENDPOINT"); v != "" {
		endpoint = v
	}
	if !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() && config.Endpoint.ValueString() != "" {
		endpoint = config.Endpoint.ValueString()
	}
	endpoint = strings.TrimRight(endpoint, "/")

	token := profile.Token
	if v := os.Getenv("KVINDO_TOKEN"); v != "" {
		token = v
	}
	if !config.Token.IsNull() && !config.Token.IsUnknown() && config.Token.ValueString() != "" {
		token = config.Token.ValueString()
	}
	if token == "" {
		msg := "Set token in provider config, KVINDO_TOKEN env var, or cli_profile (~/.kc/config/<name>.yaml)"
		if profileName != "" && !profileFound {
			msg = fmt.Sprintf("cli_profile %q was set, but ~/.kc/config/%s.yaml could not be read (missing or unreadable) - check the profile name, or set token directly / KVINDO_TOKEN instead", profileName, profileName)
		}
		resp.Diagnostics.AddError("Missing API Token", msg)
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
		NewEtcdResource,
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
		NewOllamaResource,
		NewOnOffScheduleResource,
		NewOpenVpnResource,
		NewOpenVpnUserResource,
		NewOpenVpnUserSettingsResource,
		NewPostgresqlResource,
		NewPostgresqlDatabaseResource,
		NewPostgresqlParametersSetResource,
		NewPostgresqlUserResource,
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
		NewValkeyResource,
		NewValkeyParametersSetResource,
		NewValkeyUserResource,
		NewVmResource,
		NewVmCommandScheduleResource,
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
		NewEtcdDataSource,
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
		NewOllamaDataSource,
		NewOnOffScheduleDataSource,
		NewOpenVpnDataSource,
		NewOpenVpnUserDataSource,
		NewOpenVpnUserSettingsDataSource,
		NewPostgresqlDataSource,
		NewPostgresqlDatabaseDataSource,
		NewPostgresqlParametersSetDataSource,
		NewPostgresqlUserDataSource,
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
		NewValkeyDataSource,
		NewValkeyParametersSetDataSource,
		NewValkeyUserDataSource,
		NewVmDataSource,
		NewVmCommandScheduleDataSource,
		NewVolumeDataSource,
		NewVolumeAttachmentDataSource,
		NewVpcDataSource,
		NewVpcPeeringDataSource,
		NewVpcPeeringExternalPeerDataSource,
		NewVpcPeeringPeerDataSource,
		NewVpcSubnetDataSource,
	}
}
