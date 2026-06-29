package provider

// This file is the registry that drives the kvindo_transaction resource. Each transactable
// sub-resource reuses the standalone resource's model, build, populate and schema (generated
// by tools/generator), so the transaction stays in lockstep with the resources instead of
// duplicating all 60 of them. Generic helpers (txnBuild/txnPop) and the nested element type
// (derived from the schema via Attribute.GetType) live in transaction_helpers.go.

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// txnSub describes one transactable sub-resource type.
type txnSub struct {
	tfKey    string // schema/state map attribute name (snake_case)
	apiKey   string // API spec key (camelCase, irregular)
	gate     string // ""|"folder"|"bucket"|"policy"|"user" for the two-phase S3 dance
	field    func(*TransactionResourceModel) *types.Map
	attrs    func() map[string]schema.Attribute
	build    func(context.Context, types.Object) map[string]interface{}
	populate func(context.Context, map[string]interface{}) (types.Object, string)
}

var txnSubs = []txnSub{
	{
		tfKey: "folders", apiKey: "folders", gate: "folder",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Folders },
		attrs:    FolderResourceSchemaAttrs,
		build:    txnBuild(buildFolderRequestMap),
		populate: txnPop[FolderResourceModel](populateFolderState, FolderResourceSchemaAttrs),
	},
	{
		tfKey: "ssh_keys", apiKey: "sshKeys", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SshKeys },
		attrs:    SshKeyResourceSchemaAttrs,
		build:    txnBuild(buildSshKeyRequestMap),
		populate: txnPop[SshKeyResourceModel](populateSshKeyState, SshKeyResourceSchemaAttrs),
	},
	{
		tfKey: "s3_buckets", apiKey: "s3Buckets", gate: "bucket",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.S3Buckets },
		attrs:    S3BucketResourceSchemaAttrs,
		build:    txnBuild(buildS3BucketRequestMap),
		populate: txnPop[S3BucketResourceModel](populateS3BucketState, S3BucketResourceSchemaAttrs),
	},
	{
		tfKey: "s3_user_access_policies", apiKey: "s3UserAccessPolicies", gate: "policy",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.S3UserAccessPolicies },
		attrs:    S3UserAccessPolicyResourceSchemaAttrs,
		build:    txnBuild(buildS3UserAccessPolicyRequestMap),
		populate: txnPop[S3UserAccessPolicyResourceModel](populateS3UserAccessPolicyState, S3UserAccessPolicyResourceSchemaAttrs),
	},
	{
		tfKey: "s3_users", apiKey: "s3Users", gate: "user",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.S3Users },
		attrs:    S3UserResourceSchemaAttrs,
		build:    txnBuild(buildS3UserRequestMap),
		populate: txnPop[S3UserResourceModel](populateS3UserState, S3UserResourceSchemaAttrs),
	},
	{
		tfKey: "volumes", apiKey: "volumes", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Volumes },
		attrs:    VolumeResourceSchemaAttrs,
		build:    txnBuild(buildVolumeRequestMap),
		populate: txnPop[VolumeResourceModel](populateVolumeState, VolumeResourceSchemaAttrs),
	},
	{
		tfKey: "volume_attachments", apiKey: "volumeAttachments", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.VolumeAttachments },
		attrs:    VolumeAttachmentResourceSchemaAttrs,
		build:    txnBuild(buildVolumeAttachmentRequestMap),
		populate: txnPop[VolumeAttachmentResourceModel](populateVolumeAttachmentState, VolumeAttachmentResourceSchemaAttrs),
	},
	{
		tfKey: "access_policies", apiKey: "accessPolicies", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.AccessPolicies },
		attrs:    AccessPolicyResourceSchemaAttrs,
		build:    txnBuild(buildAccessPolicyRequestMap),
		populate: txnPop[AccessPolicyResourceModel](populateAccessPolicyState, AccessPolicyResourceSchemaAttrs),
	},
	{
		tfKey: "hosting_providers", apiKey: "hostingProviders", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.HostingProviders },
		attrs:    HostingProviderResourceSchemaAttrs,
		build:    txnBuild(buildHostingProviderRequestMap),
		populate: txnPop[HostingProviderResourceModel](populateHostingProviderState, HostingProviderResourceSchemaAttrs),
	},
	{
		tfKey: "ssh_private_keys", apiKey: "sshPrivateKeys", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SshPrivateKeys },
		attrs:    SshPrivateKeyResourceSchemaAttrs,
		build:    txnBuild(buildSshPrivateKeyRequestMap),
		populate: txnPop[SshPrivateKeyResourceModel](populateSshPrivateKeyState, SshPrivateKeyResourceSchemaAttrs),
	},
	{
		tfKey: "certificates", apiKey: "certificates", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Certificates },
		attrs:    CertificateResourceSchemaAttrs,
		build:    txnBuild(buildCertificateRequestMap),
		populate: txnPop[CertificateResourceModel](populateCertificateState, CertificateResourceSchemaAttrs),
	},
	{
		tfKey: "vpc_subnets", apiKey: "vpcSubnets", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.VpcSubnets },
		attrs:    VpcSubnetResourceSchemaAttrs,
		build:    txnBuild(buildVpcSubnetRequestMap),
		populate: txnPop[VpcSubnetResourceModel](populateVpcSubnetState, VpcSubnetResourceSchemaAttrs),
	},
	{
		tfKey: "vpc_peerings", apiKey: "vpcPeerings", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.VpcPeerings },
		attrs:    VpcPeeringResourceSchemaAttrs,
		build:    txnBuild(buildVpcPeeringRequestMap),
		populate: txnPop[VpcPeeringResourceModel](populateVpcPeeringState, VpcPeeringResourceSchemaAttrs),
	},
	{
		tfKey: "vpc_peering_external_peers", apiKey: "vpcPeeringExternalPeers", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.VpcPeeringExternalPeers },
		attrs:    VpcPeeringExternalPeerResourceSchemaAttrs,
		build:    txnBuild(buildVpcPeeringExternalPeerRequestMap),
		populate: txnPop[VpcPeeringExternalPeerResourceModel](populateVpcPeeringExternalPeerState, VpcPeeringExternalPeerResourceSchemaAttrs),
	},
	{
		tfKey: "route_tables", apiKey: "routeTables", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.RouteTables },
		attrs:    RouteTableResourceSchemaAttrs,
		build:    txnBuild(buildRouteTableRequestMap),
		populate: txnPop[RouteTableResourceModel](populateRouteTableState, RouteTableResourceSchemaAttrs),
	},
	{
		tfKey: "route_table_routes", apiKey: "routeTableRoutes", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.RouteTableRoutes },
		attrs:    RouteTableRouteResourceSchemaAttrs,
		build:    txnBuild(buildRouteTableRouteRequestMap),
		populate: txnPop[RouteTableRouteResourceModel](populateRouteTableRouteState, RouteTableRouteResourceSchemaAttrs),
	},
	{
		tfKey: "route_table_attachments", apiKey: "routeTableAttachments", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.RouteTableAttachments },
		attrs:    RouteTableAttachmentResourceSchemaAttrs,
		build:    txnBuild(buildRouteTableAttachmentRequestMap),
		populate: txnPop[RouteTableAttachmentResourceModel](populateRouteTableAttachmentState, RouteTableAttachmentResourceSchemaAttrs),
	},
	{
		tfKey: "image_schedules", apiKey: "imageSchedules", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.ImageSchedules },
		attrs:    ImageScheduleResourceSchemaAttrs,
		build:    txnBuild(buildImageScheduleRequestMap),
		populate: txnPop[ImageScheduleResourceModel](populateImageScheduleState, ImageScheduleResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_target_groups", apiKey: "loadbalancerTargetGroups", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerTargetGroups },
		attrs:    LoadbalancerTargetGroupResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTargetGroupRequestMap),
		populate: txnPop[LoadbalancerTargetGroupResourceModel](populateLoadbalancerTargetGroupState, LoadbalancerTargetGroupResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_target_group_static_targets", apiKey: "loadbalancerTargetGroupStaticTargets", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerTargetGroupStaticTargets },
		attrs:    LoadbalancerTargetGroupStaticTargetResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTargetGroupStaticTargetRequestMap),
		populate: txnPop[LoadbalancerTargetGroupStaticTargetResourceModel](populateLoadbalancerTargetGroupStaticTargetState, LoadbalancerTargetGroupStaticTargetResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_target_group_service_discovery_targets", apiKey: "loadbalancerTargetGroupServiceDiscoveryTargets", gate: "",
		field: func(m *TransactionResourceModel) *types.Map {
			return &m.Spec.LoadbalancerTargetGroupServiceDiscoveryTargets
		},
		attrs:    LoadbalancerTargetGroupServiceDiscoveryTargetResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTargetGroupServiceDiscoveryTargetRequestMap),
		populate: txnPop[LoadbalancerTargetGroupServiceDiscoveryTargetResourceModel](populateLoadbalancerTargetGroupServiceDiscoveryTargetState, LoadbalancerTargetGroupServiceDiscoveryTargetResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_http_listeners", apiKey: "loadbalancerHttpListeners", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerHttpListeners },
		attrs:    LoadbalancerHttpListenerResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerHttpListenerRequestMap),
		populate: txnPop[LoadbalancerHttpListenerResourceModel](populateLoadbalancerHttpListenerState, LoadbalancerHttpListenerResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_https_listeners", apiKey: "loadbalancerHttpsListeners", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerHttpsListeners },
		attrs:    LoadbalancerHttpsListenerResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerHttpsListenerRequestMap),
		populate: txnPop[LoadbalancerHttpsListenerResourceModel](populateLoadbalancerHttpsListenerState, LoadbalancerHttpsListenerResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_tls_listeners", apiKey: "loadbalancerTlsListeners", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerTlsListeners },
		attrs:    LoadbalancerTlsListenerResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTlsListenerRequestMap),
		populate: txnPop[LoadbalancerTlsListenerResourceModel](populateLoadbalancerTlsListenerState, LoadbalancerTlsListenerResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_tcp_listeners", apiKey: "loadbalancerTcpListeners", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerTcpListeners },
		attrs:    LoadbalancerTcpListenerResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTcpListenerRequestMap),
		populate: txnPop[LoadbalancerTcpListenerResourceModel](populateLoadbalancerTcpListenerState, LoadbalancerTcpListenerResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_udp_listeners", apiKey: "loadbalancerUdpListeners", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerUdpListeners },
		attrs:    LoadbalancerUdpListenerResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerUdpListenerRequestMap),
		populate: txnPop[LoadbalancerUdpListenerResourceModel](populateLoadbalancerUdpListenerState, LoadbalancerUdpListenerResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_http_listener_rules", apiKey: "loadbalancerHttpListenerRules", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerHttpListenerRules },
		attrs:    LoadbalancerHttpListenerRuleResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerHttpListenerRuleRequestMap),
		populate: txnPop[LoadbalancerHttpListenerRuleResourceModel](populateLoadbalancerHttpListenerRuleState, LoadbalancerHttpListenerRuleResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_https_listener_rules", apiKey: "loadbalancerHttpsListenerRules", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerHttpsListenerRules },
		attrs:    LoadbalancerHttpsListenerRuleResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerHttpsListenerRuleRequestMap),
		populate: txnPop[LoadbalancerHttpsListenerRuleResourceModel](populateLoadbalancerHttpsListenerRuleState, LoadbalancerHttpsListenerRuleResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_tls_listener_rules", apiKey: "loadbalancerTlsListenerRules", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerTlsListenerRules },
		attrs:    LoadbalancerTlsListenerRuleResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTlsListenerRuleRequestMap),
		populate: txnPop[LoadbalancerTlsListenerRuleResourceModel](populateLoadbalancerTlsListenerRuleState, LoadbalancerTlsListenerRuleResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_tcp_listener_rules", apiKey: "loadbalancerTcpListenerRules", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerTcpListenerRules },
		attrs:    LoadbalancerTcpListenerRuleResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerTcpListenerRuleRequestMap),
		populate: txnPop[LoadbalancerTcpListenerRuleResourceModel](populateLoadbalancerTcpListenerRuleState, LoadbalancerTcpListenerRuleResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancer_udp_listener_rules", apiKey: "loadbalancerUdpListenerRules", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.LoadbalancerUdpListenerRules },
		attrs:    LoadbalancerUdpListenerRuleResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerUdpListenerRuleRequestMap),
		populate: txnPop[LoadbalancerUdpListenerRuleResourceModel](populateLoadbalancerUdpListenerRuleState, LoadbalancerUdpListenerRuleResourceSchemaAttrs),
	},
	{
		tfKey: "kubernetes_node_groups", apiKey: "kubernetesNodeGroups", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.KubernetesNodeGroups },
		attrs:    KubernetesNodeGroupResourceSchemaAttrs,
		build:    txnBuild(buildKubernetesNodeGroupRequestMap),
		populate: txnPop[KubernetesNodeGroupResourceModel](populateKubernetesNodeGroupState, KubernetesNodeGroupResourceSchemaAttrs),
	},
	{
		tfKey: "kubernetes_user_roles", apiKey: "kubernetesUserRoles", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.KubernetesUserRoles },
		attrs:    KubernetesUserRoleResourceSchemaAttrs,
		build:    txnBuild(buildKubernetesUserRoleRequestMap),
		populate: txnPop[KubernetesUserRoleResourceModel](populateKubernetesUserRoleState, KubernetesUserRoleResourceSchemaAttrs),
	},
	{
		tfKey: "open_vpns", apiKey: "openVpns", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.OpenVpns },
		attrs:    OpenVpnResourceSchemaAttrs,
		build:    txnBuild(buildOpenVpnRequestMap),
		populate: txnPop[OpenVpnResourceModel](populateOpenVpnState, OpenVpnResourceSchemaAttrs),
	},
	{
		tfKey: "postgresql_parameters_sets", apiKey: "postgreSqlParametersSets", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.PostgresqlParametersSets },
		attrs:    PostgresqlParametersSetResourceSchemaAttrs,
		build:    txnBuild(buildPostgresqlParametersSetRequestMap),
		populate: txnPop[PostgresqlParametersSetResourceModel](populatePostgresqlParametersSetState, PostgresqlParametersSetResourceSchemaAttrs),
	},
	{
		tfKey: "support_plans", apiKey: "supportPlans", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SupportPlans },
		attrs:    SupportPlanResourceSchemaAttrs,
		build:    txnBuild(buildSupportPlanRequestMap),
		populate: txnPop[SupportPlanResourceModel](populateSupportPlanState, SupportPlanResourceSchemaAttrs),
	},
	{
		tfKey: "support_tickets", apiKey: "supportTickets", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SupportTickets },
		attrs:    SupportTicketResourceSchemaAttrs,
		build:    txnBuild(buildSupportTicketRequestMap),
		populate: txnPop[SupportTicketResourceModel](populateSupportTicketState, SupportTicketResourceSchemaAttrs),
	},
	{
		tfKey: "support_ticket_comments", apiKey: "supportTicketComments", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SupportTicketComments },
		attrs:    SupportTicketCommentResourceSchemaAttrs,
		build:    txnBuild(buildSupportTicketCommentRequestMap),
		populate: txnPop[SupportTicketCommentResourceModel](populateSupportTicketCommentState, SupportTicketCommentResourceSchemaAttrs),
	},
	{
		tfKey: "gitlab_runners", apiKey: "gitlabRunners", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.GitlabRunners },
		attrs:    GitlabRunnerResourceSchemaAttrs,
		build:    txnBuild(buildGitlabRunnerRequestMap),
		populate: txnPop[GitlabRunnerResourceModel](populateGitlabRunnerState, GitlabRunnerResourceSchemaAttrs),
	},
	{
		tfKey: "open_vpn_user_settings", apiKey: "openVpnUserSettings", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.OpenVpnUserSettings },
		attrs:    OpenVpnUserSettingsResourceSchemaAttrs,
		build:    txnBuild(buildOpenVpnUserSettingsRequestMap),
		populate: txnPop[OpenVpnUserSettingsResourceModel](populateOpenVpnUserSettingsState, OpenVpnUserSettingsResourceSchemaAttrs),
	},
	{
		tfKey: "users", apiKey: "users", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Users },
		attrs:    UserResourceSchemaAttrs,
		build:    txnBuild(buildUserRequestMap),
		populate: txnPop[UserResourceModel](populateUserState, UserResourceSchemaAttrs),
	},
	{
		tfKey: "user_tokens", apiKey: "userTokens", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.UserTokens },
		attrs:    UserTokenResourceSchemaAttrs,
		build:    txnBuild(buildUserTokenRequestMap),
		populate: txnPop[UserTokenResourceModel](populateUserTokenState, UserTokenResourceSchemaAttrs),
	},
	{
		tfKey: "floating_ips", apiKey: "floatingIps", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.FloatingIps },
		attrs:    FloatingIpResourceSchemaAttrs,
		build:    txnBuild(buildFloatingIpRequestMap),
		populate: txnPop[FloatingIpResourceModel](populateFloatingIpState, FloatingIpResourceSchemaAttrs),
	},
	{
		tfKey: "vpcs", apiKey: "vpcs", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Vpcs },
		attrs:    VpcResourceSchemaAttrs,
		build:    txnBuild(buildVpcRequestMap),
		populate: txnPop[VpcResourceModel](populateVpcState, VpcResourceSchemaAttrs),
	},
	{
		tfKey: "vpc_peering_peers", apiKey: "vpcPeeringPeers", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.VpcPeeringPeers },
		attrs:    VpcPeeringPeerResourceSchemaAttrs,
		build:    txnBuild(buildVpcPeeringPeerRequestMap),
		populate: txnPop[VpcPeeringPeerResourceModel](populateVpcPeeringPeerState, VpcPeeringPeerResourceSchemaAttrs),
	},
	{
		tfKey: "loadbalancers", apiKey: "loadbalancers", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Loadbalancers },
		attrs:    LoadbalancerResourceSchemaAttrs,
		build:    txnBuild(buildLoadbalancerRequestMap),
		populate: txnPop[LoadbalancerResourceModel](populateLoadbalancerState, LoadbalancerResourceSchemaAttrs),
	},
	{
		tfKey: "kubernetes", apiKey: "kuberneteses", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Kubernetes },
		attrs:    KubernetesResourceSchemaAttrs,
		build:    txnBuild(buildKubernetesRequestMap),
		populate: txnPop[KubernetesResourceModel](populateKubernetesState, KubernetesResourceSchemaAttrs),
	},
	{
		tfKey: "kubernetes_users", apiKey: "kubernetesUsers", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.KubernetesUsers },
		attrs:    KubernetesUserResourceSchemaAttrs,
		build:    txnBuild(buildKubernetesUserRequestMap),
		populate: txnPop[KubernetesUserResourceModel](populateKubernetesUserState, KubernetesUserResourceSchemaAttrs),
	},
	{
		tfKey: "postgresql_standalones", apiKey: "postgreSqlStandalones", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.PostgresqlStandalones },
		attrs:    PostgresqlStandaloneResourceSchemaAttrs,
		build:    txnBuild(buildPostgresqlStandaloneRequestMap),
		populate: txnPop[PostgresqlStandaloneResourceModel](populatePostgresqlStandaloneState, PostgresqlStandaloneResourceSchemaAttrs),
	},
	{
		tfKey: "open_vpn_users", apiKey: "openVpnUsers", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.OpenVpnUsers },
		attrs:    OpenVpnUserResourceSchemaAttrs,
		build:    txnBuild(buildOpenVpnUserRequestMap),
		populate: txnPop[OpenVpnUserResourceModel](populateOpenVpnUserState, OpenVpnUserResourceSchemaAttrs),
	},
	{
		tfKey: "billing_accounts", apiKey: "billingAccounts", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.BillingAccounts },
		attrs:    BillingAccountResourceSchemaAttrs,
		build:    txnBuild(buildBillingAccountRequestMap),
		populate: txnPop[BillingAccountResourceModel](populateBillingAccountState, BillingAccountResourceSchemaAttrs),
	},
	{
		tfKey: "quotas", apiKey: "quotas", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Quotas },
		attrs:    QuotaResourceSchemaAttrs,
		build:    txnBuild(buildQuotaRequestMap),
		populate: txnPop[QuotaResourceModel](populateQuotaState, QuotaResourceSchemaAttrs),
	},
	{
		tfKey: "quota_change_requests", apiKey: "quotaChangeRequests", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.QuotaChangeRequests },
		attrs:    QuotaChangeRequestResourceSchemaAttrs,
		build:    txnBuild(buildQuotaChangeRequestRequestMap),
		populate: txnPop[QuotaChangeRequestResourceModel](populateQuotaChangeRequestState, QuotaChangeRequestResourceSchemaAttrs),
	},
	{
		tfKey: "victoria_metrics", apiKey: "victoriaMetricss", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.VictoriaMetrics },
		attrs:    VictoriaMetricsResourceSchemaAttrs,
		build:    txnBuild(buildVictoriaMetricsRequestMap),
		populate: txnPop[VictoriaMetricsResourceModel](populateVictoriaMetricsState, VictoriaMetricsResourceSchemaAttrs),
	},
	{
		tfKey: "gitlabs", apiKey: "gitlabs", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Gitlabs },
		attrs:    GitlabResourceSchemaAttrs,
		build:    txnBuild(buildGitlabRequestMap),
		populate: txnPop[GitlabResourceModel](populateGitlabState, GitlabResourceSchemaAttrs),
	},
	{
		tfKey: "support_ticket_comment_attachments", apiKey: "supportTicketCommentAttachments", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SupportTicketCommentAttachments },
		attrs:    SupportTicketCommentAttachmentResourceSchemaAttrs,
		build:    txnBuild(buildSupportTicketCommentAttachmentRequestMap),
		populate: txnPop[SupportTicketCommentAttachmentResourceModel](populateSupportTicketCommentAttachmentState, SupportTicketCommentAttachmentResourceSchemaAttrs),
	},
	{
		tfKey: "images", apiKey: "images", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Images },
		attrs:    ImageResourceSchemaAttrs,
		build:    txnBuild(buildImageRequestMap),
		populate: txnPop[ImageResourceModel](populateImageState, ImageResourceSchemaAttrs),
	},
	{
		tfKey: "vms", apiKey: "vms", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.Vms },
		attrs:    VmResourceSchemaAttrs,
		build:    txnBuild(buildVmRequestMap),
		populate: txnPop[VmResourceModel](populateVmState, VmResourceSchemaAttrs),
	},
	{
		tfKey: "security_groups", apiKey: "securityGroups", gate: "",
		field:    func(m *TransactionResourceModel) *types.Map { return &m.Spec.SecurityGroups },
		attrs:    SecurityGroupResourceSchemaAttrs,
		build:    txnBuild(buildSecurityGroupRequestMap),
		populate: txnPop[SecurityGroupResourceModel](populateSecurityGroupState, SecurityGroupResourceSchemaAttrs),
	},
}
