package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

// ---------------------------------------------------------------------------
// Models
// ---------------------------------------------------------------------------

type TxnFolderModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Info types.Object `tfsdk:"info"`
}

type TxnS3BucketModel struct {
	ID                      types.String         `tfsdk:"id"`
	Name                    types.String         `tfsdk:"name"`
	Description             types.String         `tfsdk:"description"`
	FolderID                types.String         `tfsdk:"folder_id"`
	DeleteProtection        types.Bool           `tfsdk:"delete_protection"`
	Labels                  types.Map            `tfsdk:"labels"`
	Tier                    types.String         `tfsdk:"tier"`
	Region                  types.String         `tfsdk:"region"`
	IsPublic                types.Bool           `tfsdk:"is_public"`
	IsVersioned             types.Bool           `tfsdk:"is_versioned"`
	IsLockEnabled           types.Bool           `tfsdk:"is_lock_enabled"`
	QuotaGib                types.Int64          `tfsdk:"quota_gib"`
	ObjectExpirationDays    types.Int64          `tfsdk:"object_expiration_days"`
	ComplianceRetentionDays types.Int64          `tfsdk:"compliance_retention_days"`
	Info types.Object `tfsdk:"info"`
}

type TxnS3PolicyModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	PolicyJson       types.String `tfsdk:"policy_json"`
	Info types.Object `tfsdk:"info"`
}

type TxnS3UserModel struct {
	ID               types.String       `tfsdk:"id"`
	Name             types.String       `tfsdk:"name"`
	Description      types.String       `tfsdk:"description"`
	FolderID         types.String       `tfsdk:"folder_id"`
	DeleteProtection types.Bool         `tfsdk:"delete_protection"`
	Labels           types.Map          `tfsdk:"labels"`
	BucketID         types.String       `tfsdk:"bucket_id"`
	AccessPolicyIDs  types.List         `tfsdk:"access_policy_ids"`
	Info types.Object `tfsdk:"info"`
}

type TxnSshKeyModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	PublicKey        types.String `tfsdk:"public_key"`
	ResourceName     types.String `tfsdk:"resource_name"`
	Info types.Object `tfsdk:"info"`
}

type TxnVolumeModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	FolderID          types.String `tfsdk:"folder_id"`
	DeleteProtection  types.Bool   `tfsdk:"delete_protection"`
	Labels            types.Map    `tfsdk:"labels"`
	HostingProviderId types.String `tfsdk:"hosting_provider_id"`
	OfferId           types.String `tfsdk:"offer_id"`
	SizeGib           types.Int64  `tfsdk:"size_gib"`
	OsImageId         types.String `tfsdk:"os_image_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnVolumeAttachmentModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VolumeId         types.String `tfsdk:"volume_id"`
	VmId             types.String `tfsdk:"vm_id"`
	VmDeviceIndex    types.Int64  `tfsdk:"vm_device_index"`
	Info types.Object `tfsdk:"info"`
}

type TxnAccessPolicyModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Content          types.String `tfsdk:"content"`
	Info types.Object `tfsdk:"info"`
}

type TxnHostingProviderModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	FolderID         types.String  `tfsdk:"folder_id"`
	DeleteProtection types.Bool    `tfsdk:"delete_protection"`
	Labels           types.Map     `tfsdk:"labels"`
	Country          types.String  `tfsdk:"country"`
	CountryIsoCode   types.String  `tfsdk:"country_iso_code"`
	City             types.String  `tfsdk:"city"`
	Cloud            types.String  `tfsdk:"cloud"`
	Sla              types.Float64 `tfsdk:"sla"`
	DataCenterIndex  types.Int64   `tfsdk:"data_center_index"`
	KeyFeatures      types.List    `tfsdk:"key_features"`
	Disabled         types.Bool    `tfsdk:"disabled"`
	Info types.Object `tfsdk:"info"`
}

type TxnSshPrivateKeyModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	PrivateKey       types.String `tfsdk:"private_key"`
	Info types.Object `tfsdk:"info"`
}

type TxnCertificateModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	CertificatePem   types.String `tfsdk:"certificate_pem"`
	PrivateKeyPem    types.String `tfsdk:"private_key_pem"`
	ResourceName     types.String `tfsdk:"resource_name"`
	Info types.Object `tfsdk:"info"`
}

type TxnVpcSubnetModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VpcId            types.String `tfsdk:"vpc_id"`
	Ipv4Cidr         types.String `tfsdk:"ipv4_cidr"`
	Info types.Object `tfsdk:"info"`
}

type TxnVpcPeeringModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Info types.Object `tfsdk:"info"`
}

type TxnVpcPeeringExternalPeerModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VpcPeeringId     types.String `tfsdk:"vpc_peering_id"`
	SshUser          types.String `tfsdk:"ssh_user"`
	SshPort          types.Int64  `tfsdk:"ssh_port"`
	SshIpV4          types.String `tfsdk:"ssh_ip_v4"`
	PrivateIpV4      types.String `tfsdk:"private_ip_v4"`
	IpV4Cidrs        types.List   `tfsdk:"ip_v4_cidrs"`
	SshPrivateKeyId  types.String `tfsdk:"ssh_private_key_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnRouteTableModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Info types.Object `tfsdk:"info"`
}

type TxnRouteTableRouteModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	RouteTableId     types.String `tfsdk:"route_table_id"`
	DestinationCidr  types.String `tfsdk:"destination_cidr"`
	TargetIp         types.String `tfsdk:"target_ip"`
	Info types.Object `tfsdk:"info"`
}

type TxnRouteTableAttachmentModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	RouteTableId     types.String `tfsdk:"route_table_id"`
	VpcId            types.String `tfsdk:"vpc_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnImageScheduleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	ScheduleFormat   types.String `tfsdk:"schedule_format"`
	Schedule         types.String `tfsdk:"schedule"`
	RetentionCount   types.Int64  `tfsdk:"retention_count"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTargetGroupModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTargetGroupStaticTargetModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TargetGroupId    types.String `tfsdk:"target_group_id"`
	IpOrHostname     types.String `tfsdk:"ip_or_hostname"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTargetGroupServiceDiscoveryTargetModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TargetGroupId    types.String `tfsdk:"target_group_id"`
	LabelSelectors   types.Map    `tfsdk:"label_selectors"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerHttpListenerModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	LoadbalancerId   types.String `tfsdk:"loadbalancer_id"`
	Interface        types.String `tfsdk:"interface"`
	Order            types.Int64  `tfsdk:"order"`
	Ports            types.List   `tfsdk:"ports"`
	Hosts            types.List   `tfsdk:"hosts"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerHttpsListenerModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	FolderID                   types.String `tfsdk:"folder_id"`
	DeleteProtection           types.Bool   `tfsdk:"delete_protection"`
	Labels                     types.Map    `tfsdk:"labels"`
	LoadbalancerId             types.String `tfsdk:"loadbalancer_id"`
	Interface                  types.String `tfsdk:"interface"`
	Order                      types.Int64  `tfsdk:"order"`
	Ports                      types.List   `tfsdk:"ports"`
	Hosts                      types.List   `tfsdk:"hosts"`
	EnableHttp2Support         types.Bool   `tfsdk:"enable_http2_support"`
	TlsCertificateId           types.String `tfsdk:"tls_certificate_id"`
	TlsProtocols               types.List   `tfsdk:"tls_protocols"`
	TlsAutogenerateCertificate types.Bool   `tfsdk:"tls_autogenerate_certificate"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTlsListenerModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Description                types.String `tfsdk:"description"`
	FolderID                   types.String `tfsdk:"folder_id"`
	DeleteProtection           types.Bool   `tfsdk:"delete_protection"`
	Labels                     types.Map    `tfsdk:"labels"`
	LoadbalancerId             types.String `tfsdk:"loadbalancer_id"`
	Interface                  types.String `tfsdk:"interface"`
	Order                      types.Int64  `tfsdk:"order"`
	Ports                      types.List   `tfsdk:"ports"`
	Hosts                      types.List   `tfsdk:"hosts"`
	TlsCertificateId           types.String `tfsdk:"tls_certificate_id"`
	TlsProtocols               types.List   `tfsdk:"tls_protocols"`
	TlsAutogenerateCertificate types.Bool   `tfsdk:"tls_autogenerate_certificate"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTcpListenerModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	LoadbalancerId   types.String `tfsdk:"loadbalancer_id"`
	Interface        types.String `tfsdk:"interface"`
	Order            types.Int64  `tfsdk:"order"`
	Ports            types.List   `tfsdk:"ports"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerUdpListenerModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	LoadbalancerId   types.String `tfsdk:"loadbalancer_id"`
	Interface        types.String `tfsdk:"interface"`
	Order            types.Int64  `tfsdk:"order"`
	Ports            types.List   `tfsdk:"ports"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerHttpListenerRuleModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	FolderID           types.String `tfsdk:"folder_id"`
	DeleteProtection   types.Bool   `tfsdk:"delete_protection"`
	Labels             types.Map    `tfsdk:"labels"`
	HttpListenerId     types.String `tfsdk:"http_listener_id"`
	Order              types.Int64  `tfsdk:"order"`
	MatchPath          types.String `tfsdk:"match_path"`
	MatchPathMatchType types.String `tfsdk:"match_path_match_type"`
	ActionType         types.String `tfsdk:"action_type"`
	ActionJson         types.String `tfsdk:"action_json"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerHttpsListenerRuleModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	FolderID           types.String `tfsdk:"folder_id"`
	DeleteProtection   types.Bool   `tfsdk:"delete_protection"`
	Labels             types.Map    `tfsdk:"labels"`
	HttpsListenerId    types.String `tfsdk:"https_listener_id"`
	Order              types.Int64  `tfsdk:"order"`
	MatchPath          types.String `tfsdk:"match_path"`
	MatchPathMatchType types.String `tfsdk:"match_path_match_type"`
	ActionType         types.String `tfsdk:"action_type"`
	ActionJson         types.String `tfsdk:"action_json"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTlsListenerRuleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TlsListenerId    types.String `tfsdk:"tls_listener_id"`
	Order            types.Int64  `tfsdk:"order"`
	ActionType       types.String `tfsdk:"action_type"`
	ActionJson       types.String `tfsdk:"action_json"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerTcpListenerRuleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TcpListenerId    types.String `tfsdk:"tcp_listener_id"`
	Order            types.Int64  `tfsdk:"order"`
	ActionType       types.String `tfsdk:"action_type"`
	ActionJson       types.String `tfsdk:"action_json"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerUdpListenerRuleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	UdpListenerId    types.String `tfsdk:"udp_listener_id"`
	Order            types.Int64  `tfsdk:"order"`
	ActionJson       types.String `tfsdk:"action_json"`
	Info types.Object `tfsdk:"info"`
}

type TxnKubernetesNodeGroupModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	KubernetesId     types.String `tfsdk:"kubernetes_id"`
	VpcSubnetId      types.String `tfsdk:"vpc_subnet_id"`
	VmOfferId        types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId    types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib    types.Int64  `tfsdk:"volume_size_gib"`
	DesiredNodeCount types.Int64  `tfsdk:"desired_node_count"`
	VmState          types.String `tfsdk:"vm_state"`
	CreatePublicIpv4 types.Bool   `tfsdk:"create_public_ipv4"`
	Info types.Object `tfsdk:"info"`
}

type TxnKubernetesUserRoleModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	ApiGroups        types.List   `tfsdk:"api_groups"`
	Resources        types.List   `tfsdk:"resources"`
	Verbs            types.List   `tfsdk:"verbs"`
	Namespaces       types.List   `tfsdk:"namespaces"`
	Info types.Object `tfsdk:"info"`
}

type TxnOpenVpnModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier             types.String `tfsdk:"tier"`
	VpcSubnetId      types.String `tfsdk:"vpc_subnet_id"`
	FloatingIpId     types.String `tfsdk:"floating_ip_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnPostgresqlParametersSetModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Parameters       types.Map    `tfsdk:"parameters"`
	Info types.Object `tfsdk:"info"`
}

type TxnSupportPlanModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier             types.String `tfsdk:"tier"`
	Info types.Object `tfsdk:"info"`
}

type TxnSupportTicketModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Kind             types.String `tfsdk:"kind"`
	Severity         types.String `tfsdk:"severity"`
	Status           types.String `tfsdk:"status"`
	Info types.Object `tfsdk:"info"`
}

type TxnSupportTicketCommentModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	TicketId         types.String `tfsdk:"ticket_id"`
	Content          types.String `tfsdk:"content"`
	AttachmentsIds   types.List   `tfsdk:"attachments_ids"`
}

type TxnGitlabRunnerModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Description             types.String `tfsdk:"description"`
	FolderID                types.String `tfsdk:"folder_id"`
	DeleteProtection        types.Bool   `tfsdk:"delete_protection"`
	Labels                  types.Map    `tfsdk:"labels"`
	Tier                    types.String `tfsdk:"tier"`
	VpcSubnetId             types.String `tfsdk:"vpc_subnet_id"`
	FloatingIpId            types.String `tfsdk:"floating_ip_id"`
	VmState                 types.String `tfsdk:"vm_state"`
	VmOfferId               types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId           types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib           types.Int64  `tfsdk:"volume_size_gib"`
	Concurrency             types.Int64  `tfsdk:"concurrency"`
	Version                 types.String `tfsdk:"version"`
	DockerOptionsJsonString types.String `tfsdk:"docker_options_json_string"`
	Info types.Object `tfsdk:"info"`
}

type TxnOpenVpnUserSettingsModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	AllowedIpV4Cidrs types.List   `tfsdk:"allowed_ip_v4_cidrs"`
	AllowedIpV6Cidrs types.List   `tfsdk:"allowed_ip_v6_cidrs"`
	DeniedIpV4Cidrs  types.List   `tfsdk:"denied_ip_v4_cidrs"`
	DeniedIpV6Cidrs  types.List   `tfsdk:"denied_ip_v6_cidrs"`
	AllowedDomains   types.List   `tfsdk:"allowed_domains"`
	DeniedDomains    types.List   `tfsdk:"denied_domains"`
	Info types.Object `tfsdk:"info"`
}

type TxnIamUserModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Email            types.String `tfsdk:"email"`
	AccessPolicyIds  types.List   `tfsdk:"access_policy_ids"`
	Info types.Object `tfsdk:"info"`
}

type TxnUserTokenModel struct {
	ID               types.String          `tfsdk:"id"`
	Name             types.String          `tfsdk:"name"`
	Description      types.String          `tfsdk:"description"`
	FolderID         types.String          `tfsdk:"folder_id"`
	DeleteProtection types.Bool            `tfsdk:"delete_protection"`
	Labels           types.Map             `tfsdk:"labels"`
	UserId           types.String          `tfsdk:"user_id"`
	SendToEmail      types.Bool            `tfsdk:"send_to_email"`
	Info types.Object `tfsdk:"info"`
}

type TxnFloatingIpModel struct {
	ID                types.String           `tfsdk:"id"`
	Name              types.String           `tfsdk:"name"`
	Description       types.String           `tfsdk:"description"`
	FolderID          types.String           `tfsdk:"folder_id"`
	DeleteProtection  types.Bool             `tfsdk:"delete_protection"`
	Labels            types.Map              `tfsdk:"labels"`
	HostingProviderId types.String           `tfsdk:"hosting_provider_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnVpcModel struct {
	ID                types.String    `tfsdk:"id"`
	Name              types.String    `tfsdk:"name"`
	Description       types.String    `tfsdk:"description"`
	FolderID          types.String    `tfsdk:"folder_id"`
	DeleteProtection  types.Bool      `tfsdk:"delete_protection"`
	Labels            types.Map       `tfsdk:"labels"`
	HostingProviderId types.String    `tfsdk:"hosting_provider_id"`
	Ipv4Cidr          types.String    `tfsdk:"ipv4_cidr"`
	NatFloatingIpId   types.String    `tfsdk:"nat_floating_ip_id"`
	SecurityGroupIds  types.List      `tfsdk:"security_group_ids"`
	ExternallyManaged types.Bool      `tfsdk:"externally_managed"`
	Info types.Object `tfsdk:"info"`
}

type TxnVpcPeeringPeerModel struct {
	ID               types.String               `tfsdk:"id"`
	Name             types.String               `tfsdk:"name"`
	Description      types.String               `tfsdk:"description"`
	FolderID         types.String               `tfsdk:"folder_id"`
	DeleteProtection types.Bool                 `tfsdk:"delete_protection"`
	Labels           types.Map                  `tfsdk:"labels"`
	VpcPeeringId     types.String               `tfsdk:"vpc_peering_id"`
	VpcSubnetId      types.String               `tfsdk:"vpc_subnet_id"`
	FloatingIpId     types.String               `tfsdk:"floating_ip_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnLoadbalancerModel struct {
	ID               types.String             `tfsdk:"id"`
	Name             types.String             `tfsdk:"name"`
	Description      types.String             `tfsdk:"description"`
	FolderID         types.String             `tfsdk:"folder_id"`
	DeleteProtection types.Bool               `tfsdk:"delete_protection"`
	Labels           types.Map                `tfsdk:"labels"`
	Tier             types.String             `tfsdk:"tier"`
	VpcSubnetId      types.String             `tfsdk:"vpc_subnet_id"`
	FloatingIpId     types.String             `tfsdk:"floating_ip_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnKubernetesModel struct {
	ID                    types.String           `tfsdk:"id"`
	Name                  types.String           `tfsdk:"name"`
	Description           types.String           `tfsdk:"description"`
	FolderID              types.String           `tfsdk:"folder_id"`
	DeleteProtection      types.Bool             `tfsdk:"delete_protection"`
	Labels                types.Map              `tfsdk:"labels"`
	Tier                  types.String           `tfsdk:"tier"`
	AssignPublicIpV4      types.Bool             `tfsdk:"assign_public_ip_v4"`
	Version               types.String           `tfsdk:"version"`
	ControlPlaneLocations types.List             `tfsdk:"control_plane_locations"`
	Info types.Object `tfsdk:"info"`
}

type TxnKubernetesUserModel struct {
	ID               types.String               `tfsdk:"id"`
	Name             types.String               `tfsdk:"name"`
	Description      types.String               `tfsdk:"description"`
	FolderID         types.String               `tfsdk:"folder_id"`
	DeleteProtection types.Bool                 `tfsdk:"delete_protection"`
	Labels           types.Map                  `tfsdk:"labels"`
	KubernetesId     types.String               `tfsdk:"kubernetes_id"`
	RoleIds          types.List                 `tfsdk:"role_ids"`
	Info types.Object `tfsdk:"info"`
}

type TxnPostgresqlStandaloneModel struct {
	ID                  types.String                     `tfsdk:"id"`
	Name                types.String                     `tfsdk:"name"`
	Description         types.String                     `tfsdk:"description"`
	FolderID            types.String                     `tfsdk:"folder_id"`
	DeleteProtection    types.Bool                       `tfsdk:"delete_protection"`
	Labels              types.Map                        `tfsdk:"labels"`
	Tier                types.String                     `tfsdk:"tier"`
	Version             types.String                     `tfsdk:"version"`
	RootPassword        types.String                     `tfsdk:"root_password"`
	ParametersSetId     types.String                     `tfsdk:"parameters_set_id"`
	BackupRetentionDays types.Int64                      `tfsdk:"backup_retention_days"`
	FloatingIpId        types.String                     `tfsdk:"floating_ip_id"`
	VpcSubnetId         types.String                     `tfsdk:"vpc_subnet_id"`
	VmState             types.String                     `tfsdk:"vm_state"`
	VmOfferId           types.String                     `tfsdk:"vm_offer_id"`
	VolumeOfferId       types.String                     `tfsdk:"volume_offer_id"`
	VolumeSizeGib       types.Int64                      `tfsdk:"volume_size_gib"`
	Info types.Object `tfsdk:"info"`
}

type TxnOpenVpnUserModel struct {
	ID                 types.String            `tfsdk:"id"`
	Name               types.String            `tfsdk:"name"`
	Description        types.String            `tfsdk:"description"`
	FolderID           types.String            `tfsdk:"folder_id"`
	DeleteProtection   types.Bool              `tfsdk:"delete_protection"`
	Labels             types.Map               `tfsdk:"labels"`
	OpenVpnId          types.String            `tfsdk:"open_vpn_id"`
	OpenVpnSettingsIds types.List              `tfsdk:"open_vpn_settings_ids"`
	Info types.Object `tfsdk:"info"`
}

type TxnBillingAccountModel struct {
	ID               types.String               `tfsdk:"id"`
	Name             types.String               `tfsdk:"name"`
	Description      types.String               `tfsdk:"description"`
	FolderID         types.String               `tfsdk:"folder_id"`
	DeleteProtection types.Bool                 `tfsdk:"delete_protection"`
	Labels           types.Map                  `tfsdk:"labels"`
	ResourceName     types.String               `tfsdk:"resource_name"`
	Info types.Object `tfsdk:"info"`
}

type TxnQuotaModel struct {
	ID               types.String      `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	Description      types.String      `tfsdk:"description"`
	FolderID         types.String      `tfsdk:"folder_id"`
	DeleteProtection types.Bool        `tfsdk:"delete_protection"`
	Labels           types.Map         `tfsdk:"labels"`
	Product          types.String      `tfsdk:"product"`
	Resource         types.String      `tfsdk:"resource"`
	Parameter        types.String      `tfsdk:"parameter"`
	Limit            types.Int64       `tfsdk:"limit"`
	Info types.Object `tfsdk:"info"`
}

type TxnQuotaChangeRequestModel struct {
	ID               types.String                   `tfsdk:"id"`
	Name             types.String                   `tfsdk:"name"`
	Description      types.String                   `tfsdk:"description"`
	FolderID         types.String                   `tfsdk:"folder_id"`
	DeleteProtection types.Bool                     `tfsdk:"delete_protection"`
	Labels           types.Map                      `tfsdk:"labels"`
	QuotaId          types.String                   `tfsdk:"quota_id"`
	NewQuotaLimit    types.Int64                    `tfsdk:"new_quota_limit"`
	Info types.Object `tfsdk:"info"`
}

type TxnVictoriaMetricsModel struct {
	ID               types.String                `tfsdk:"id"`
	Name             types.String                `tfsdk:"name"`
	Description      types.String                `tfsdk:"description"`
	FolderID         types.String                `tfsdk:"folder_id"`
	DeleteProtection types.Bool                  `tfsdk:"delete_protection"`
	Labels           types.Map                   `tfsdk:"labels"`
	Tier             types.String                `tfsdk:"tier"`
	VpcId            types.String                `tfsdk:"vpc_id"`
	CreatePublicIpv4 types.Bool                  `tfsdk:"create_public_ipv4"`
	CreatePublicIpv6 types.Bool                  `tfsdk:"create_public_ipv6"`
	DnsRecordName    types.String                `tfsdk:"dns_record_name"`
	Info types.Object `tfsdk:"info"`
}

type TxnGitlabModel struct {
	ID               types.String       `tfsdk:"id"`
	Name             types.String       `tfsdk:"name"`
	Description      types.String       `tfsdk:"description"`
	FolderID         types.String       `tfsdk:"folder_id"`
	DeleteProtection types.Bool         `tfsdk:"delete_protection"`
	Labels           types.Map          `tfsdk:"labels"`
	Tier             types.String       `tfsdk:"tier"`
	FloatingIpId     types.String       `tfsdk:"floating_ip_id"`
	VpcSubnetId      types.String       `tfsdk:"vpc_subnet_id"`
	Version          types.String       `tfsdk:"version"`
	RootPassword     types.String       `tfsdk:"root_password"`
	VmState          types.String       `tfsdk:"vm_state"`
	VmOfferId        types.String       `tfsdk:"vm_offer_id"`
	VolumeOfferId    types.String       `tfsdk:"volume_offer_id"`
	VolumeSizeGib    types.Int64        `tfsdk:"volume_size_gib"`
	Edition          types.String       `tfsdk:"edition"`
	RecordName       types.String       `tfsdk:"record_name"`
	Info types.Object `tfsdk:"info"`
}

type TxnSupportTicketCommentAttachmentModel struct {
	ID                types.String                               `tfsdk:"id"`
	Name              types.String                               `tfsdk:"name"`
	Description       types.String                               `tfsdk:"description"`
	FolderID          types.String                               `tfsdk:"folder_id"`
	DeleteProtection  types.Bool                                 `tfsdk:"delete_protection"`
	Labels            types.Map                                  `tfsdk:"labels"`
	FileName          types.String                               `tfsdk:"file_name"`
	FileType          types.String                               `tfsdk:"file_type"`
	FileContentBase64 types.String                               `tfsdk:"file_content_base64"`
	Info types.Object `tfsdk:"info"`
}

type TxnImageModel struct {
	ID               types.String      `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	Description      types.String      `tfsdk:"description"`
	FolderID         types.String      `tfsdk:"folder_id"`
	DeleteProtection types.Bool        `tfsdk:"delete_protection"`
	Labels           types.Map         `tfsdk:"labels"`
	VmId             types.String      `tfsdk:"vm_id"`
	Info types.Object `tfsdk:"info"`
}

type TxnVmModel struct {
	ID                         types.String   `tfsdk:"id"`
	Name                       types.String   `tfsdk:"name"`
	Description                types.String   `tfsdk:"description"`
	FolderID                   types.String   `tfsdk:"folder_id"`
	DeleteProtection           types.Bool     `tfsdk:"delete_protection"`
	Labels                     types.Map      `tfsdk:"labels"`
	VmState                    types.String   `tfsdk:"vm_state"`
	VpcSubnetId                types.String   `tfsdk:"vpc_subnet_id"`
	FloatingIpId               types.String   `tfsdk:"floating_ip_id"`
	ImageId                    types.String   `tfsdk:"image_id"`
	OfferId                    types.String   `tfsdk:"offer_id"`
	ImageBootVolumeDeviceIndex types.Int64    `tfsdk:"image_boot_volume_device_index"`
	SshKeyIds                  types.List     `tfsdk:"ssh_key_ids"`
	ImageScheduleIds           types.List     `tfsdk:"image_schedule_ids"`
	BootstrapCommand           types.List     `tfsdk:"bootstrap_command"`
	Info types.Object `tfsdk:"info"`
}

type TxnSecurityGroupModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Ingress          types.List   `tfsdk:"ingress"`
	Egress           types.List   `tfsdk:"egress"`
	Info types.Object `tfsdk:"info"`
}

type TransactionResourceModel struct {
	ID                                             types.String `tfsdk:"id"`
	Name                                           types.String `tfsdk:"name"`
	Description                                    types.String `tfsdk:"description"`
	FolderID                                       types.String `tfsdk:"folder_id"`
	DeleteProtection                               types.Bool   `tfsdk:"delete_protection"`
	Labels                                         types.Map    `tfsdk:"labels"`
	DeleteResourcesOnTransactionDelete             types.Bool   `tfsdk:"delete_resources_on_transaction_delete"`
	State types.String `tfsdk:"state"`
	Folders                                        types.Map    `tfsdk:"folders"`
	SshKeys                                        types.Map    `tfsdk:"ssh_keys"`
	S3Buckets                                      types.Map    `tfsdk:"s3_buckets"`
	S3UserAccessPolicies                           types.Map    `tfsdk:"s3_user_access_policies"`
	S3Users                                        types.Map    `tfsdk:"s3_users"`
	Volumes                                        types.Map    `tfsdk:"volumes"`
	VolumeAttachments                              types.Map    `tfsdk:"volume_attachments"`
	AccessPolicies                                 types.Map    `tfsdk:"access_policies"`
	HostingProviders                               types.Map    `tfsdk:"hosting_providers"`
	SshPrivateKeys                                 types.Map    `tfsdk:"ssh_private_keys"`
	Certificates                                   types.Map    `tfsdk:"certificates"`
	VpcSubnets                                     types.Map    `tfsdk:"vpc_subnets"`
	VpcPeerings                                    types.Map    `tfsdk:"vpc_peerings"`
	VpcPeeringExternalPeers                        types.Map    `tfsdk:"vpc_peering_external_peers"`
	RouteTables                                    types.Map    `tfsdk:"route_tables"`
	RouteTableRoutes                               types.Map    `tfsdk:"route_table_routes"`
	RouteTableAttachments                          types.Map    `tfsdk:"route_table_attachments"`
	ImageSchedules                                 types.Map    `tfsdk:"image_schedules"`
	LoadbalancerTargetGroups                       types.Map    `tfsdk:"loadbalancer_target_groups"`
	LoadbalancerTargetGroupStaticTargets           types.Map    `tfsdk:"loadbalancer_target_group_static_targets"`
	LoadbalancerTargetGroupServiceDiscoveryTargets types.Map    `tfsdk:"loadbalancer_target_group_service_discovery_targets"`
	LoadbalancerHttpListeners                      types.Map    `tfsdk:"loadbalancer_http_listeners"`
	LoadbalancerHttpsListeners                     types.Map    `tfsdk:"loadbalancer_https_listeners"`
	LoadbalancerTlsListeners                       types.Map    `tfsdk:"loadbalancer_tls_listeners"`
	LoadbalancerTcpListeners                       types.Map    `tfsdk:"loadbalancer_tcp_listeners"`
	LoadbalancerUdpListeners                       types.Map    `tfsdk:"loadbalancer_udp_listeners"`
	LoadbalancerHttpListenerRules                  types.Map    `tfsdk:"loadbalancer_http_listener_rules"`
	LoadbalancerHttpsListenerRules                 types.Map    `tfsdk:"loadbalancer_https_listener_rules"`
	LoadbalancerTlsListenerRules                   types.Map    `tfsdk:"loadbalancer_tls_listener_rules"`
	LoadbalancerTcpListenerRules                   types.Map    `tfsdk:"loadbalancer_tcp_listener_rules"`
	LoadbalancerUdpListenerRules                   types.Map    `tfsdk:"loadbalancer_udp_listener_rules"`
	KubernetesNodeGroups                           types.Map    `tfsdk:"kubernetes_node_groups"`
	KubernetesUserRoles                            types.Map    `tfsdk:"kubernetes_user_roles"`
	OpenVpns                                       types.Map    `tfsdk:"open_vpns"`
	PostgresqlParametersSets                       types.Map    `tfsdk:"postgresql_parameters_sets"`
	SupportPlans                                   types.Map    `tfsdk:"support_plans"`
	SupportTickets                                 types.Map    `tfsdk:"support_tickets"`
	SupportTicketComments                          types.Map    `tfsdk:"support_ticket_comments"`
	GitlabRunners                                  types.Map    `tfsdk:"gitlab_runners"`
	OpenVpnUserSettings                            types.Map    `tfsdk:"open_vpn_user_settings"`
	Users                                          types.Map    `tfsdk:"users"`
	UserTokens                                     types.Map    `tfsdk:"user_tokens"`
	FloatingIps                                    types.Map    `tfsdk:"floating_ips"`
	Vpcs                                           types.Map    `tfsdk:"vpcs"`
	VpcPeeringPeers                                types.Map    `tfsdk:"vpc_peering_peers"`
	Loadbalancers                                  types.Map    `tfsdk:"loadbalancers"`
	Kubernetes                                     types.Map    `tfsdk:"kubernetes"`
	KubernetesUsers                                types.Map    `tfsdk:"kubernetes_users"`
	PostgresqlStandalones                          types.Map    `tfsdk:"postgresql_standalones"`
	OpenVpnUsers                                   types.Map    `tfsdk:"open_vpn_users"`
	BillingAccounts                                types.Map    `tfsdk:"billing_accounts"`
	Quotas                                         types.Map    `tfsdk:"quotas"`
	QuotaChangeRequests                            types.Map    `tfsdk:"quota_change_requests"`
	VictoriaMetrics                                types.Map    `tfsdk:"victoria_metrics"`
	Gitlabs                                        types.Map    `tfsdk:"gitlabs"`
	SupportTicketCommentAttachments                types.Map    `tfsdk:"support_ticket_comment_attachments"`
	Images                                         types.Map    `tfsdk:"images"`
	Vms                                            types.Map    `tfsdk:"vms"`
	SecurityGroups                                 types.Map    `tfsdk:"security_groups"`
}

// ---------------------------------------------------------------------------
// Attr types
// ---------------------------------------------------------------------------

var txnFolderAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnS3BucketAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"tier": types.StringType, "region": types.StringType, "is_public": types.BoolType,
	"is_versioned": types.BoolType, "is_lock_enabled": types.BoolType,
	"quota_gib": types.Int64Type, "object_expiration_days": types.Int64Type,
	"compliance_retention_days": types.Int64Type,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType, "endpoint_url": types.StringType}},
}

var txnS3PolicyAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"policy_json": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSshKeyAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"public_key": types.StringType, "resource_name": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnS3UserAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"bucket_id": types.StringType, "access_policy_ids": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType, "access_key": types.StringType, "secret_key": types.StringType}},
}

var txnVolumeAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"hosting_provider_id": types.StringType, "offer_id": types.StringType,
	"size_gib": types.Int64Type, "os_image_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVolumeAttachmentAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"volume_id": types.StringType, "vm_id": types.StringType, "vm_device_index": types.Int64Type,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnAccessPolicyAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"content": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnHostingProviderAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"country": types.StringType, "country_iso_code": types.StringType, "city": types.StringType,
	"cloud": types.StringType, "sla": types.Float64Type, "data_center_index": types.Int64Type,
	"key_features": types.ListType{ElemType: types.StringType}, "disabled": types.BoolType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSshPrivateKeyAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"private_key": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnCertificateAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"certificate_pem": types.StringType, "private_key_pem": types.StringType, "resource_name": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVpcSubnetAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"vpc_id": types.StringType, "ipv4_cidr": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVpcPeeringAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVpcPeeringExternalPeerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"vpc_peering_id": types.StringType, "ssh_user": types.StringType, "ssh_port": types.Int64Type,
	"ssh_ip_v4": types.StringType, "private_ip_v4": types.StringType,
	"ip_v4_cidrs": types.ListType{ElemType: types.StringType}, "ssh_private_key_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnRouteTableAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnRouteTableRouteAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"route_table_id": types.StringType, "destination_cidr": types.StringType, "target_ip": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnRouteTableAttachmentAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"route_table_id": types.StringType, "vpc_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnImageScheduleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"enabled": types.BoolType, "schedule_format": types.StringType,
	"schedule": types.StringType, "retention_count": types.Int64Type,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTargetGroupAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTargetGroupStaticTargetAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"target_group_id": types.StringType, "ip_or_hostname": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"target_group_id": types.StringType,
	"label_selectors": types.MapType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerHttpListenerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"loadbalancer_id": types.StringType, "interface": types.StringType, "order": types.Int64Type,
	"ports": types.ListType{ElemType: types.StringType}, "hosts": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerHttpsListenerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"loadbalancer_id": types.StringType, "interface": types.StringType, "order": types.Int64Type,
	"ports": types.ListType{ElemType: types.StringType}, "hosts": types.ListType{ElemType: types.StringType},
	"enable_http2_support": types.BoolType, "tls_certificate_id": types.StringType,
	"tls_protocols": types.ListType{ElemType: types.StringType}, "tls_autogenerate_certificate": types.BoolType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTlsListenerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"loadbalancer_id": types.StringType, "interface": types.StringType, "order": types.Int64Type,
	"ports": types.ListType{ElemType: types.StringType}, "hosts": types.ListType{ElemType: types.StringType},
	"tls_certificate_id": types.StringType,
	"tls_protocols": types.ListType{ElemType: types.StringType}, "tls_autogenerate_certificate": types.BoolType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTcpListenerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"loadbalancer_id": types.StringType, "interface": types.StringType, "order": types.Int64Type,
	"ports": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerUdpListenerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"loadbalancer_id": types.StringType, "interface": types.StringType, "order": types.Int64Type,
	"ports": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerHttpListenerRuleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"http_listener_id": types.StringType, "order": types.Int64Type,
	"match_path": types.StringType, "match_path_match_type": types.StringType,
	"action_type": types.StringType, "action_json": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerHttpsListenerRuleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"https_listener_id": types.StringType, "order": types.Int64Type,
	"match_path": types.StringType, "match_path_match_type": types.StringType,
	"action_type": types.StringType, "action_json": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTlsListenerRuleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"tls_listener_id": types.StringType, "order": types.Int64Type,
	"action_type": types.StringType, "action_json": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerTcpListenerRuleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"tcp_listener_id": types.StringType, "order": types.Int64Type,
	"action_type": types.StringType, "action_json": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerUdpListenerRuleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"udp_listener_id": types.StringType, "order": types.Int64Type,
	"action_json": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnKubernetesNodeGroupAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"kubernetes_id": types.StringType, "vpc_subnet_id": types.StringType,
	"vm_offer_id": types.StringType, "volume_offer_id": types.StringType,
	"volume_size_gib": types.Int64Type, "desired_node_count": types.Int64Type,
	"vm_state": types.StringType, "create_public_ipv4": types.BoolType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnKubernetesUserRoleAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"api_groups": types.ListType{ElemType: types.StringType},
	"resources":  types.ListType{ElemType: types.StringType},
	"verbs":      types.ListType{ElemType: types.StringType},
	"namespaces": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnOpenVpnAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"tier": types.StringType, "vpc_subnet_id": types.StringType, "floating_ip_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnPostgresqlParametersSetAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":     types.MapType{ElemType: types.StringType},
	"parameters": types.MapType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSupportPlanAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"tier":   types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSupportTicketAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":   types.MapType{ElemType: types.StringType},
	"kind":     types.StringType, "severity": types.StringType, "status": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSupportTicketCommentAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":          types.MapType{ElemType: types.StringType},
	"ticket_id":       types.StringType, "content": types.StringType,
	"attachments_ids": types.ListType{ElemType: types.StringType},
}

var txnGitlabRunnerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"tier": types.StringType, "vpc_subnet_id": types.StringType, "floating_ip_id": types.StringType,
	"vm_state": types.StringType, "vm_offer_id": types.StringType, "volume_offer_id": types.StringType,
	"volume_size_gib": types.Int64Type, "concurrency": types.Int64Type,
	"version": types.StringType, "docker_options_json_string": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnOpenVpnUserSettingsAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":              types.MapType{ElemType: types.StringType},
	"allowed_ip_v4_cidrs": types.ListType{ElemType: types.StringType},
	"allowed_ip_v6_cidrs": types.ListType{ElemType: types.StringType},
	"denied_ip_v4_cidrs":  types.ListType{ElemType: types.StringType},
	"denied_ip_v6_cidrs":  types.ListType{ElemType: types.StringType},
	"allowed_domains":     types.ListType{ElemType: types.StringType},
	"denied_domains":      types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnIamUserAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":            types.MapType{ElemType: types.StringType},
	"email":             types.StringType,
	"access_policy_ids": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnUserTokenAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels": types.MapType{ElemType: types.StringType},
	"user_id": types.StringType, "send_to_email": types.BoolType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnFloatingIpAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":              types.MapType{ElemType: types.StringType},
	"hosting_provider_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVpcAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":              types.MapType{ElemType: types.StringType},
	"hosting_provider_id": types.StringType, "ipv4_cidr": types.StringType,
	"nat_floating_ip_id":  types.StringType,
	"security_group_ids":  types.ListType{ElemType: types.StringType},
	"externally_managed":  types.BoolType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVpcPeeringPeerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":         types.MapType{ElemType: types.StringType},
	"vpc_peering_id": types.StringType, "vpc_subnet_id": types.StringType, "floating_ip_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnLoadbalancerAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":         types.MapType{ElemType: types.StringType},
	"tier":           types.StringType, "vpc_subnet_id": types.StringType, "floating_ip_id": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnControlPlaneLocationAttrTypes = map[string]attr.Type{
	"vpc_subnet_id": types.StringType,
}

var txnKubernetesAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":               types.MapType{ElemType: types.StringType},
	"tier":                 types.StringType, "assign_public_ip_v4": types.BoolType, "version": types.StringType,
	"control_plane_locations": types.ListType{ElemType: types.ObjectType{AttrTypes: txnControlPlaneLocationAttrTypes}},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnKubernetesUserAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":        types.MapType{ElemType: types.StringType},
	"kubernetes_id": types.StringType, "role_ids": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnPostgresqlStandaloneAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":               types.MapType{ElemType: types.StringType},
	"tier":                 types.StringType, "version": types.StringType, "root_password": types.StringType,
	"parameters_set_id":    types.StringType, "backup_retention_days": types.Int64Type,
	"floating_ip_id":       types.StringType, "vpc_subnet_id": types.StringType,
	"vm_state":             types.StringType, "vm_offer_id": types.StringType, "volume_offer_id": types.StringType,
	"volume_size_gib":      types.Int64Type,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnOpenVpnUserAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":                types.MapType{ElemType: types.StringType},
	"open_vpn_id":           types.StringType,
	"open_vpn_settings_ids": types.ListType{ElemType: types.StringType},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnBillingAccountAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":        types.MapType{ElemType: types.StringType},
	"resource_name": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnQuotaAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":    types.MapType{ElemType: types.StringType},
	"product":   types.StringType, "resource": types.StringType, "parameter": types.StringType, "limit": types.Int64Type,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnQuotaChangeRequestAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":          types.MapType{ElemType: types.StringType},
	"quota_id":        types.StringType, "new_quota_limit": types.Int64Type,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnVictoriaMetricsAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":             types.MapType{ElemType: types.StringType},
	"tier":               types.StringType, "vpc_id": types.StringType,
	"create_public_ipv4": types.BoolType, "create_public_ipv6": types.BoolType,
	"dns_record_name":    types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnGitlabAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":          types.MapType{ElemType: types.StringType},
	"tier":            types.StringType, "floating_ip_id": types.StringType, "vpc_subnet_id": types.StringType,
	"version":         types.StringType, "root_password": types.StringType, "vm_state": types.StringType,
	"vm_offer_id":     types.StringType, "volume_offer_id": types.StringType, "volume_size_gib": types.Int64Type,
	"edition":         types.StringType, "record_name": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSupportTicketCommentAttachmentAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":              types.MapType{ElemType: types.StringType},
	"file_name":           types.StringType, "file_type": types.StringType, "file_content_base64": types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnImageAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":  types.MapType{ElemType: types.StringType},
	"vm_id":   types.StringType,
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnBootstrapCmdAttrTypes = map[string]attr.Type{
	"command": types.StringType, "success_return_code": types.Int64Type, "timeout_seconds": types.Int64Type,
}

var txnVmAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":                         types.MapType{ElemType: types.StringType},
	"vm_state":                       types.StringType, "vpc_subnet_id": types.StringType, "floating_ip_id": types.StringType,
	"image_id":                       types.StringType, "offer_id": types.StringType,
	"image_boot_volume_device_index": types.Int64Type,
	"ssh_key_ids":                    types.ListType{ElemType: types.StringType},
	"image_schedule_ids":             types.ListType{ElemType: types.StringType},
	"bootstrap_command":              types.ListType{ElemType: types.ObjectType{AttrTypes: txnBootstrapCmdAttrTypes}},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

var txnSgRuleAttrTypes = map[string]attr.Type{
	"ports":       types.ListType{ElemType: types.StringType},
	"ipv4_blocks": types.ListType{ElemType: types.StringType},
	"ipv6_blocks": types.ListType{ElemType: types.StringType},
	"action":      types.StringType,
}

var txnSecurityGroupAttrTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "description": types.StringType,
	"folder_id": types.StringType, "delete_protection": types.BoolType,
	"labels":  types.MapType{ElemType: types.StringType},
	"ingress": types.ListType{ElemType: types.ObjectType{AttrTypes: txnSgRuleAttrTypes}},
	"egress":  types.ListType{ElemType: types.ObjectType{AttrTypes: txnSgRuleAttrTypes}},
	"info": types.ObjectType{AttrTypes: map[string]attr.Type{"state": types.StringType}},
}

// ---------------------------------------------------------------------------
// Resource
// ---------------------------------------------------------------------------

type TransactionResource struct{ client *client.Client }

func NewTransactionResource() resource.Resource { return &TransactionResource{} }

func (r *TransactionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_transaction"
}

func (r *TransactionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	ss := func() schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
	}
	sb := func() schema.BoolAttribute {
		return schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}}
	}
	si := func() schema.Int64Attribute {
		return schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}}
	}
	sf := func() schema.Float64Attribute {
		return schema.Float64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Float64{float64planmodifier.UseStateForUnknown()}}
	}
	sl := func() schema.ListAttribute {
		return schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}}
	}
	sm := func() schema.MapAttribute {
		return schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()}}
	}
	stateInfo := func() schema.Attribute {
		return commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})
	}
	ipQuartetInfo := func() schema.Attribute {
		return commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})
	}
	ipQuartetFqdnInfo := func() schema.Attribute {
		return commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})
	}
	base := func(extra map[string]schema.Attribute) map[string]schema.Attribute {
		m := map[string]schema.Attribute{
			"id": ss(), "name": schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{Optional: true},
			"folder_id": schema.StringAttribute{Optional: true}, "delete_protection": sb(),
			"labels": sm(),
		}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	mapAttr := func(attrs map[string]schema.Attribute) schema.MapNestedAttribute {
		return schema.MapNestedAttribute{
			Optional: true, Computed: true,
			PlanModifiers: []planmodifier.Map{mapplanmodifier.UseStateForUnknown()},
			NestedObject:  schema.NestedAttributeObject{Attributes: attrs},
		}
	}
	listNestedAttr := func(attrs map[string]schema.Attribute) schema.ListNestedAttribute {
		return schema.ListNestedAttribute{
			Optional: true, Computed: true,
			PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			NestedObject:  schema.NestedAttributeObject{Attributes: attrs},
		}
	}

	attrs := commonSchemaAttributes()
	attrs["delete_resources_on_transaction_delete"] = schema.BoolAttribute{
		Optional: true, Computed: true,
		PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
	// "state" is volatile (the server transitions it during reconciliation). volatileStateModifier
	// freezes it to prior only when prior == "stable" (no idle diff, safe to re-read); while
	// in-flight it stays "(known after apply)" so the apply re-resolves it without tripping
	// "Provider produced inconsistent result after apply".
	attrs["state"] = schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{volatileStateModifier{}}}

	// Original 5 types
	attrs["folders"] = mapAttr(base(map[string]schema.Attribute{"info": stateInfo()}))
	attrs["ssh_keys"] = mapAttr(base(map[string]schema.Attribute{
		"public_key": ss(), "resource_name": ss(), "info": stateInfo(),
	}))
	attrs["s3_buckets"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "region": ss(), "is_public": sb(), "is_versioned": sb(), "is_lock_enabled": sb(),
		"quota_gib": si(), "object_expiration_days": schema.Int64Attribute{Optional: true},
		"compliance_retention_days": schema.Int64Attribute{Optional: true},
		"info": commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "endpoint_url": schema.StringAttribute{Computed: true}}),
	}))
	attrs["s3_user_access_policies"] = mapAttr(base(map[string]schema.Attribute{
		"policy_json": ss(), "info": stateInfo(),
	}))
	attrs["s3_users"] = mapAttr(base(map[string]schema.Attribute{
		"bucket_id": ss(),
		"access_policy_ids": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"info": commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "access_key": schema.StringAttribute{Computed: true}, "secret_key": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}))

	// New 54 types
	attrs["volumes"] = mapAttr(base(map[string]schema.Attribute{
		"hosting_provider_id": ss(), "offer_id": ss(), "size_gib": si(), "os_image_id": schema.StringAttribute{Optional: true},
		"info": stateInfo(),
	}))
	attrs["volume_attachments"] = mapAttr(base(map[string]schema.Attribute{
		"volume_id": ss(), "vm_id": ss(), "vm_device_index": si(), "info": stateInfo(),
	}))
	attrs["access_policies"] = mapAttr(base(map[string]schema.Attribute{
		"content": ss(), "info": stateInfo(),
	}))
	attrs["hosting_providers"] = mapAttr(base(map[string]schema.Attribute{
		"country": ss(), "country_iso_code": ss(), "city": ss(), "cloud": ss(),
		"sla": sf(), "data_center_index": si(), "key_features": sl(), "disabled": sb(),
		"info": stateInfo(),
	}))
	attrs["ssh_private_keys"] = mapAttr(base(map[string]schema.Attribute{
		"private_key": schema.StringAttribute{Optional: true, Sensitive: true},
		"info": stateInfo(),
	}))
	attrs["certificates"] = mapAttr(base(map[string]schema.Attribute{
		"certificate_pem": ss(), "private_key_pem": schema.StringAttribute{Optional: true, Sensitive: true}, "resource_name": ss(), "info": stateInfo(),
	}))
	attrs["vpc_subnets"] = mapAttr(base(map[string]schema.Attribute{
		"vpc_id": ss(), "ipv4_cidr": ss(), "info": stateInfo(),
	}))
	attrs["vpc_peerings"] = mapAttr(base(map[string]schema.Attribute{"info": stateInfo()}))
	attrs["vpc_peering_external_peers"] = mapAttr(base(map[string]schema.Attribute{
		"vpc_peering_id": ss(), "ssh_user": ss(), "ssh_port": si(), "ssh_ip_v4": ss(),
		"private_ip_v4": ss(), "ip_v4_cidrs": sl(), "ssh_private_key_id": schema.StringAttribute{Optional: true},
		"info": stateInfo(),
	}))
	attrs["route_tables"] = mapAttr(base(map[string]schema.Attribute{"info": stateInfo()}))
	attrs["route_table_routes"] = mapAttr(base(map[string]schema.Attribute{
		"route_table_id": ss(), "destination_cidr": ss(), "target_ip": ss(), "info": stateInfo(),
	}))
	attrs["route_table_attachments"] = mapAttr(base(map[string]schema.Attribute{
		"route_table_id": ss(), "vpc_id": ss(), "info": stateInfo(),
	}))
	attrs["image_schedules"] = mapAttr(base(map[string]schema.Attribute{
		"enabled": sb(), "schedule_format": ss(), "schedule": ss(), "retention_count": si(),
		"info": stateInfo(),
	}))
	attrs["loadbalancer_target_groups"] = mapAttr(base(map[string]schema.Attribute{"info": stateInfo()}))
	attrs["loadbalancer_target_group_static_targets"] = mapAttr(base(map[string]schema.Attribute{
		"target_group_id": ss(), "ip_or_hostname": ss(), "info": stateInfo(),
	}))
	attrs["loadbalancer_target_group_service_discovery_targets"] = mapAttr(base(map[string]schema.Attribute{
		"target_group_id": ss(), "label_selectors": sm(), "info": stateInfo(),
	}))
	attrs["loadbalancer_http_listeners"] = mapAttr(base(map[string]schema.Attribute{
		"loadbalancer_id": ss(), "interface": ss(), "order": si(), "ports": sl(), "hosts": sl(),
		"info": stateInfo(),
	}))
	attrs["loadbalancer_https_listeners"] = mapAttr(base(map[string]schema.Attribute{
		"loadbalancer_id": ss(), "interface": ss(), "order": si(), "ports": sl(), "hosts": sl(),
		"enable_http2_support": sb(), "tls_certificate_id": schema.StringAttribute{Optional: true}, "tls_protocols": sl(),
		"tls_autogenerate_certificate": sb(), "info": stateInfo(),
	}))
	attrs["loadbalancer_tls_listeners"] = mapAttr(base(map[string]schema.Attribute{
		"loadbalancer_id": ss(), "interface": ss(), "order": si(), "ports": sl(), "hosts": sl(),
		"tls_certificate_id": schema.StringAttribute{Optional: true}, "tls_protocols": sl(), "tls_autogenerate_certificate": sb(),
		"info": stateInfo(),
	}))
	attrs["loadbalancer_tcp_listeners"] = mapAttr(base(map[string]schema.Attribute{
		"loadbalancer_id": ss(), "interface": ss(), "order": si(), "ports": sl(), "info": stateInfo(),
	}))
	attrs["loadbalancer_udp_listeners"] = mapAttr(base(map[string]schema.Attribute{
		"loadbalancer_id": ss(), "interface": ss(), "order": si(), "ports": sl(), "info": stateInfo(),
	}))
	attrs["loadbalancer_http_listener_rules"] = mapAttr(base(map[string]schema.Attribute{
		"http_listener_id": ss(), "order": si(), "match_path": ss(), "match_path_match_type": ss(),
		"action_type": ss(), "action_json": ss(), "info": stateInfo(),
	}))
	attrs["loadbalancer_https_listener_rules"] = mapAttr(base(map[string]schema.Attribute{
		"https_listener_id": ss(), "order": si(), "match_path": ss(), "match_path_match_type": ss(),
		"action_type": ss(), "action_json": ss(), "info": stateInfo(),
	}))
	attrs["loadbalancer_tls_listener_rules"] = mapAttr(base(map[string]schema.Attribute{
		"tls_listener_id": ss(), "order": si(), "action_type": ss(), "action_json": ss(),
		"info": stateInfo(),
	}))
	attrs["loadbalancer_tcp_listener_rules"] = mapAttr(base(map[string]schema.Attribute{
		"tcp_listener_id": ss(), "order": si(), "action_type": ss(), "action_json": ss(),
		"info": stateInfo(),
	}))
	attrs["loadbalancer_udp_listener_rules"] = mapAttr(base(map[string]schema.Attribute{
		"udp_listener_id": ss(), "order": si(), "action_json": ss(), "info": stateInfo(),
	}))
	attrs["kubernetes_node_groups"] = mapAttr(base(map[string]schema.Attribute{
		"kubernetes_id": ss(), "vpc_subnet_id": ss(), "vm_offer_id": ss(), "volume_offer_id": ss(),
		"volume_size_gib": si(), "desired_node_count": si(), "vm_state": ss(), "create_public_ipv4": sb(),
		"info": stateInfo(),
	}))
	attrs["kubernetes_user_roles"] = mapAttr(base(map[string]schema.Attribute{
		"api_groups": sl(), "resources": sl(), "verbs": sl(), "namespaces": sl(), "info": stateInfo(),
	}))
	attrs["open_vpns"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "vpc_subnet_id": ss(), "floating_ip_id": schema.StringAttribute{Optional: true}, "info": stateInfo(),
	}))
	attrs["postgresql_parameters_sets"] = mapAttr(base(map[string]schema.Attribute{
		"parameters": sm(), "info": stateInfo(),
	}))
	attrs["support_plans"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "info": stateInfo(),
	}))
	attrs["support_tickets"] = mapAttr(base(map[string]schema.Attribute{
		"kind": ss(), "severity": ss(), "status": ss(), "info": stateInfo(),
	}))
	// support_ticket_comments has no info field
	attrs["support_ticket_comments"] = mapAttr(base(map[string]schema.Attribute{
		"ticket_id": ss(), "content": ss(),
		"attachments_ids": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
	}))
	attrs["gitlab_runners"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "vpc_subnet_id": ss(), "floating_ip_id": schema.StringAttribute{Optional: true}, "vm_state": ss(),
		"vm_offer_id": ss(), "volume_offer_id": ss(), "volume_size_gib": si(), "concurrency": si(),
		"version": ss(), "docker_options_json_string": ss(), "info": stateInfo(),
	}))
	attrs["open_vpn_user_settings"] = mapAttr(base(map[string]schema.Attribute{
		"allowed_ip_v4_cidrs": sl(), "allowed_ip_v6_cidrs": sl(),
		"denied_ip_v4_cidrs": sl(), "denied_ip_v6_cidrs": sl(),
		"allowed_domains": sl(), "denied_domains": sl(), "info": stateInfo(),
	}))
	attrs["users"] = mapAttr(base(map[string]schema.Attribute{
		"email":             ss(),
		"access_policy_ids": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()}},
		"info": stateInfo(),
	}))
	attrs["user_tokens"] = mapAttr(base(map[string]schema.Attribute{
		"user_id": ss(), "send_to_email": sb(),
		"info": stateInfo(),
	}))
	attrs["floating_ips"] = mapAttr(base(map[string]schema.Attribute{
		"hosting_provider_id": ss(),
		"info": stateInfo(),
	}))
	attrs["vpcs"] = mapAttr(base(map[string]schema.Attribute{
		"hosting_provider_id": ss(), "ipv4_cidr": ss(), "nat_floating_ip_id": schema.StringAttribute{Optional: true},
		"security_group_ids": sl(), "externally_managed": sb(),
		"info": stateInfo(),
	}))
	attrs["vpc_peering_peers"] = mapAttr(base(map[string]schema.Attribute{
		"vpc_peering_id": ss(), "vpc_subnet_id": ss(), "floating_ip_id": schema.StringAttribute{Optional: true},
		"info": ipQuartetInfo(),
	}))
	attrs["loadbalancers"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "vpc_subnet_id": ss(), "floating_ip_id": schema.StringAttribute{Optional: true},
		"info": ipQuartetInfo(),
	}))
	attrs["kubernetes"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "assign_public_ip_v4": sb(), "version": ss(),
		"control_plane_locations": listNestedAttr(map[string]schema.Attribute{
			"vpc_subnet_id": ss(),
		}),
		"info": stateInfo(),
	}))
	attrs["kubernetes_users"] = mapAttr(base(map[string]schema.Attribute{
		"kubernetes_id": ss(), "role_ids": sl(),
		"info": stateInfo(),
	}))
	attrs["postgresql_standalones"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "version": ss(), "root_password": schema.StringAttribute{Optional: true, Sensitive: true}, "parameters_set_id": schema.StringAttribute{Optional: true},
		"backup_retention_days": si(), "floating_ip_id": schema.StringAttribute{Optional: true}, "vpc_subnet_id": ss(),
		"vm_state": ss(), "vm_offer_id": ss(), "volume_offer_id": ss(), "volume_size_gib": si(),
		"info": stateInfo(),
	}))
	attrs["open_vpn_users"] = mapAttr(base(map[string]schema.Attribute{
		"open_vpn_id": ss(), "open_vpn_settings_ids": sl(),
		"info": stateInfo(),
	}))
	attrs["billing_accounts"] = mapAttr(base(map[string]schema.Attribute{
		"resource_name": ss(),
		"info": stateInfo(),
	}))
	attrs["quotas"] = mapAttr(base(map[string]schema.Attribute{
		"product": ss(), "resource": ss(), "parameter": ss(), "limit": si(),
		"info": stateInfo(),
	}))
	attrs["quota_change_requests"] = mapAttr(base(map[string]schema.Attribute{
		"quota_id": ss(), "new_quota_limit": si(),
		"info": stateInfo(),
	}))
	attrs["victoria_metrics"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "vpc_id": ss(), "create_public_ipv4": sb(), "create_public_ipv6": sb(),
		"dns_record_name": schema.StringAttribute{Optional: true}, "info": ipQuartetFqdnInfo(),
	}))
	attrs["gitlabs"] = mapAttr(base(map[string]schema.Attribute{
		"tier": ss(), "floating_ip_id": schema.StringAttribute{Optional: true}, "vpc_subnet_id": ss(), "version": ss(),
		"root_password": schema.StringAttribute{Optional: true, Sensitive: true}, "vm_state": ss(), "vm_offer_id": ss(), "volume_offer_id": ss(),
		"volume_size_gib": si(), "edition": ss(), "record_name": schema.StringAttribute{Optional: true},
		"info": ipQuartetFqdnInfo(),
	}))
	attrs["support_ticket_comment_attachments"] = mapAttr(base(map[string]schema.Attribute{
		"file_name": ss(), "file_type": ss(), "file_content_base64": schema.StringAttribute{Optional: true, Sensitive: true},
		"info": stateInfo(),
	}))
	attrs["images"] = mapAttr(base(map[string]schema.Attribute{
		"vm_id": ss(),
		"info": stateInfo(),
	}))
	attrs["vms"] = mapAttr(base(map[string]schema.Attribute{
		"vm_state": ss(), "vpc_subnet_id": ss(), "floating_ip_id": ss(),
		"image_id": ss(), "offer_id": ss(), "image_boot_volume_device_index": si(),
		"ssh_key_ids": sl(), "image_schedule_ids": sl(),
		"bootstrap_command": listNestedAttr(map[string]schema.Attribute{
			"command":             schema.StringAttribute{Required: true},
			"success_return_code": si(),
			"timeout_seconds":     si(),
		}),
		"info": stateInfo(),
	}))
	attrs["security_groups"] = mapAttr(base(map[string]schema.Attribute{
		"ingress": listNestedAttr(map[string]schema.Attribute{
			"ports":       sl(),
			"ipv4_blocks": sl(),
			"ipv6_blocks": sl(),
			"action":      ss(),
		}),
		"egress": listNestedAttr(map[string]schema.Attribute{
			"ports":       sl(),
			"ipv4_blocks": sl(),
			"ipv6_blocks": sl(),
			"action":      ss(),
		}),
		"info": stateInfo(),
	}))

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *TransactionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	r.client = pd.Client
}

// ---------------------------------------------------------------------------
// Request body builder
// ---------------------------------------------------------------------------

func buildTxnBody(ctx context.Context, plan TransactionResourceModel, includeFolders, includeBuckets, includePolicies, includeUsers bool) map[string]interface{} {
	m := map[string]interface{}{
		"id":   plan.ID.ValueString(),
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() { m["description"] = plan.Description.ValueString() }
	if !plan.FolderID.IsNull() && !plan.FolderID.IsUnknown() && plan.FolderID.ValueString() != "" { m["folderId"] = plan.FolderID.ValueString() }
	if !plan.DeleteProtection.IsNull() && !plan.DeleteProtection.IsUnknown() { m["deleteProtection"] = plan.DeleteProtection.ValueBool() }
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() { m["labels"] = strMapFromTF(ctx, plan.Labels) }
	if !plan.DeleteResourcesOnTransactionDelete.IsNull() && !plan.DeleteResourcesOnTransactionDelete.IsUnknown() {
		m["deleteResourcesOnTransactionDelete"] = plan.DeleteResourcesOnTransactionDelete.ValueBool()
	}

	addStrMap := func(src types.Map, apiKey string) {
		if src.IsNull() || src.IsUnknown() { return }
		m[apiKey] = strMapFromTF(ctx, src)
	}
	addStrList := func(src types.List, apiKey string) {
		if src.IsNull() || src.IsUnknown() { return }
		m[apiKey] = stringListToInterface(ctx, src)
	}
	subBase := func(id, name types.String, desc, folderID types.String, dp types.Bool, labels types.Map) map[string]interface{} {
		sub := map[string]interface{}{"id": ulidOrNew(id), "name": name.ValueString()}
		if !desc.IsNull() && !desc.IsUnknown() { sub["description"] = desc.ValueString() }
		if !folderID.IsNull() && !folderID.IsUnknown() { sub["folderId"] = folderID.ValueString() }
		if !dp.IsNull() && !dp.IsUnknown() { sub["deleteProtection"] = dp.ValueBool() }
		if !labels.IsNull() && !labels.IsUnknown() { sub["labels"] = strMapFromTF(ctx, labels) }
		return sub
	}

	// --- Original 5 types ---
	if includeFolders && !plan.Folders.IsNull() && !plan.Folders.IsUnknown() {
		var mp map[string]TxnFolderModel
		plan.Folders.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, f := range mp {
			items = append(items, subBase(f.ID, f.Name, f.Description, f.FolderID, f.DeleteProtection, f.Labels))
		}
		m["folders"] = items
	}
	if !plan.SshKeys.IsNull() && !plan.SshKeys.IsUnknown() {
		var mp map[string]TxnSshKeyModel
		plan.SshKeys.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, k := range mp {
			sub := subBase(k.ID, k.Name, k.Description, k.FolderID, k.DeleteProtection, k.Labels)
			if !k.PublicKey.IsNull() && !k.PublicKey.IsUnknown() { sub["publicKey"] = k.PublicKey.ValueString() }
			if !k.ResourceName.IsNull() && !k.ResourceName.IsUnknown() { sub["resourceName"] = k.ResourceName.ValueString() }
			items = append(items, sub)
		}
		m["sshKeys"] = items
	}
	if includeBuckets && !plan.S3Buckets.IsNull() && !plan.S3Buckets.IsUnknown() {
		var mp map[string]TxnS3BucketModel
		plan.S3Buckets.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, b := range mp {
			sub := subBase(b.ID, b.Name, b.Description, b.FolderID, b.DeleteProtection, b.Labels)
			if !b.Tier.IsNull() && !b.Tier.IsUnknown() { sub["tier"] = b.Tier.ValueString() }
			if !b.Region.IsNull() && !b.Region.IsUnknown() { sub["region"] = b.Region.ValueString() }
			if !b.IsPublic.IsNull() && !b.IsPublic.IsUnknown() { sub["isPublic"] = b.IsPublic.ValueBool() }
			if !b.IsVersioned.IsNull() && !b.IsVersioned.IsUnknown() { sub["isVersioned"] = b.IsVersioned.ValueBool() }
			if !b.IsLockEnabled.IsNull() && !b.IsLockEnabled.IsUnknown() { sub["isLockEnabled"] = b.IsLockEnabled.ValueBool() }
			if !b.QuotaGib.IsNull() && !b.QuotaGib.IsUnknown() { sub["quotaGiB"] = b.QuotaGib.ValueInt64() }
			if !b.ObjectExpirationDays.IsNull() && !b.ObjectExpirationDays.IsUnknown() { sub["objectExpirationDays"] = b.ObjectExpirationDays.ValueInt64() }
			if !b.ComplianceRetentionDays.IsNull() && !b.ComplianceRetentionDays.IsUnknown() { sub["complianceRetentionDays"] = b.ComplianceRetentionDays.ValueInt64() }
			items = append(items, sub)
		}
		m["s3Buckets"] = items
	}
	if includePolicies && !plan.S3UserAccessPolicies.IsNull() && !plan.S3UserAccessPolicies.IsUnknown() {
		var mp map[string]TxnS3PolicyModel
		plan.S3UserAccessPolicies.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, p := range mp {
			sub := subBase(p.ID, p.Name, p.Description, p.FolderID, p.DeleteProtection, p.Labels)
			if !p.PolicyJson.IsNull() && !p.PolicyJson.IsUnknown() { sub["policyJson"] = p.PolicyJson.ValueString() }
			items = append(items, sub)
		}
		m["s3UserAccessPolicies"] = items
	}
	if includeUsers && !plan.S3Users.IsNull() && !plan.S3Users.IsUnknown() {
		var mp map[string]TxnS3UserModel
		plan.S3Users.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, u := range mp {
			sub := subBase(u.ID, u.Name, u.Description, u.FolderID, u.DeleteProtection, u.Labels)
			if !u.BucketID.IsNull() && !u.BucketID.IsUnknown() { sub["bucketId"] = u.BucketID.ValueString() }
			if !u.AccessPolicyIDs.IsNull() && !u.AccessPolicyIDs.IsUnknown() { sub["accessPolicyIds"] = stringListToInterface(ctx, u.AccessPolicyIDs) }
			items = append(items, sub)
		}
		m["s3Users"] = items
	}

	// --- New 54 types (always included) ---
	_ = addStrMap
	_ = addStrList

	if !plan.Volumes.IsNull() && !plan.Volumes.IsUnknown() {
		var mp map[string]TxnVolumeModel
		plan.Volumes.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.HostingProviderId.IsNull() && !v.HostingProviderId.IsUnknown() { sub["hostingProviderId"] = v.HostingProviderId.ValueString() }
			if !v.OfferId.IsNull() && !v.OfferId.IsUnknown() { sub["offerId"] = v.OfferId.ValueString() }
			if !v.SizeGib.IsNull() && !v.SizeGib.IsUnknown() { sub["sizeGib"] = v.SizeGib.ValueInt64() }
			if !v.OsImageId.IsNull() && !v.OsImageId.IsUnknown() { sub["osImageId"] = v.OsImageId.ValueString() }
			items = append(items, sub)
		}
		m["volumes"] = items
	}
	if !plan.VolumeAttachments.IsNull() && !plan.VolumeAttachments.IsUnknown() {
		var mp map[string]TxnVolumeAttachmentModel
		plan.VolumeAttachments.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.VolumeId.IsNull() && !v.VolumeId.IsUnknown() { sub["volumeId"] = v.VolumeId.ValueString() }
			if !v.VmId.IsNull() && !v.VmId.IsUnknown() { sub["vmId"] = v.VmId.ValueString() }
			if !v.VmDeviceIndex.IsNull() && !v.VmDeviceIndex.IsUnknown() { sub["vmDeviceIndex"] = v.VmDeviceIndex.ValueInt64() }
			items = append(items, sub)
		}
		m["volumeAttachments"] = items
	}
	if !plan.AccessPolicies.IsNull() && !plan.AccessPolicies.IsUnknown() {
		var mp map[string]TxnAccessPolicyModel
		plan.AccessPolicies.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Content.IsNull() && !v.Content.IsUnknown() { sub["content"] = v.Content.ValueString() }
			items = append(items, sub)
		}
		m["accessPolicies"] = items
	}
	if !plan.HostingProviders.IsNull() && !plan.HostingProviders.IsUnknown() {
		var mp map[string]TxnHostingProviderModel
		plan.HostingProviders.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Country.IsNull() && !v.Country.IsUnknown() { sub["country"] = v.Country.ValueString() }
			if !v.CountryIsoCode.IsNull() && !v.CountryIsoCode.IsUnknown() { sub["countryIsoCode"] = v.CountryIsoCode.ValueString() }
			if !v.City.IsNull() && !v.City.IsUnknown() { sub["city"] = v.City.ValueString() }
			if !v.Cloud.IsNull() && !v.Cloud.IsUnknown() { sub["cloud"] = v.Cloud.ValueString() }
			if !v.Sla.IsNull() && !v.Sla.IsUnknown() { sub["sla"] = v.Sla.ValueFloat64() }
			if !v.DataCenterIndex.IsNull() && !v.DataCenterIndex.IsUnknown() { sub["dataCenterIndex"] = v.DataCenterIndex.ValueInt64() }
			if !v.KeyFeatures.IsNull() && !v.KeyFeatures.IsUnknown() { sub["keyFeatures"] = stringListToInterface(ctx, v.KeyFeatures) }
			if !v.Disabled.IsNull() && !v.Disabled.IsUnknown() { sub["disabled"] = v.Disabled.ValueBool() }
			items = append(items, sub)
		}
		m["hostingProviders"] = items
	}
	if !plan.SshPrivateKeys.IsNull() && !plan.SshPrivateKeys.IsUnknown() {
		var mp map[string]TxnSshPrivateKeyModel
		plan.SshPrivateKeys.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.PrivateKey.IsNull() && !v.PrivateKey.IsUnknown() { sub["privateKey"] = v.PrivateKey.ValueString() }
			items = append(items, sub)
		}
		m["sshPrivateKeys"] = items
	}
	if !plan.Certificates.IsNull() && !plan.Certificates.IsUnknown() {
		var mp map[string]TxnCertificateModel
		plan.Certificates.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.CertificatePem.IsNull() && !v.CertificatePem.IsUnknown() { sub["certificatePem"] = v.CertificatePem.ValueString() }
			if !v.PrivateKeyPem.IsNull() && !v.PrivateKeyPem.IsUnknown() { sub["privateKeyPem"] = v.PrivateKeyPem.ValueString() }
			if !v.ResourceName.IsNull() && !v.ResourceName.IsUnknown() { sub["resourceName"] = v.ResourceName.ValueString() }
			items = append(items, sub)
		}
		m["certificates"] = items
	}
	if !plan.VpcSubnets.IsNull() && !plan.VpcSubnets.IsUnknown() {
		var mp map[string]TxnVpcSubnetModel
		plan.VpcSubnets.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.VpcId.IsNull() && !v.VpcId.IsUnknown() { sub["vpcId"] = v.VpcId.ValueString() }
			if !v.Ipv4Cidr.IsNull() && !v.Ipv4Cidr.IsUnknown() { sub["ipv4Cidr"] = v.Ipv4Cidr.ValueString() }
			items = append(items, sub)
		}
		m["vpcSubnets"] = items
	}
	if !plan.VpcPeerings.IsNull() && !plan.VpcPeerings.IsUnknown() {
		var mp map[string]TxnVpcPeeringModel
		plan.VpcPeerings.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			items = append(items, subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels))
		}
		m["vpcPeerings"] = items
	}
	if !plan.VpcPeeringExternalPeers.IsNull() && !plan.VpcPeeringExternalPeers.IsUnknown() {
		var mp map[string]TxnVpcPeeringExternalPeerModel
		plan.VpcPeeringExternalPeers.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.VpcPeeringId.IsNull() && !v.VpcPeeringId.IsUnknown() { sub["vpcPeeringId"] = v.VpcPeeringId.ValueString() }
			if !v.SshUser.IsNull() && !v.SshUser.IsUnknown() { sub["sshUser"] = v.SshUser.ValueString() }
			if !v.SshPort.IsNull() && !v.SshPort.IsUnknown() { sub["sshPort"] = v.SshPort.ValueInt64() }
			if !v.SshIpV4.IsNull() && !v.SshIpV4.IsUnknown() { sub["sshIpV4"] = v.SshIpV4.ValueString() }
			if !v.PrivateIpV4.IsNull() && !v.PrivateIpV4.IsUnknown() { sub["privateIpV4"] = v.PrivateIpV4.ValueString() }
			if !v.IpV4Cidrs.IsNull() && !v.IpV4Cidrs.IsUnknown() { sub["ipV4Cidrs"] = stringListToInterface(ctx, v.IpV4Cidrs) }
			if !v.SshPrivateKeyId.IsNull() && !v.SshPrivateKeyId.IsUnknown() { sub["sshPrivateKeyId"] = v.SshPrivateKeyId.ValueString() }
			items = append(items, sub)
		}
		m["vpcPeeringExternalPeers"] = items
	}
	if !plan.RouteTables.IsNull() && !plan.RouteTables.IsUnknown() {
		var mp map[string]TxnRouteTableModel
		plan.RouteTables.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			items = append(items, subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels))
		}
		m["routeTables"] = items
	}
	if !plan.RouteTableRoutes.IsNull() && !plan.RouteTableRoutes.IsUnknown() {
		var mp map[string]TxnRouteTableRouteModel
		plan.RouteTableRoutes.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.RouteTableId.IsNull() && !v.RouteTableId.IsUnknown() { sub["routeTableId"] = v.RouteTableId.ValueString() }
			if !v.DestinationCidr.IsNull() && !v.DestinationCidr.IsUnknown() { sub["destinationCidr"] = v.DestinationCidr.ValueString() }
			if !v.TargetIp.IsNull() && !v.TargetIp.IsUnknown() { sub["targetIp"] = v.TargetIp.ValueString() }
			items = append(items, sub)
		}
		m["routeTableRoutes"] = items
	}
	if !plan.RouteTableAttachments.IsNull() && !plan.RouteTableAttachments.IsUnknown() {
		var mp map[string]TxnRouteTableAttachmentModel
		plan.RouteTableAttachments.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.RouteTableId.IsNull() && !v.RouteTableId.IsUnknown() { sub["routeTableId"] = v.RouteTableId.ValueString() }
			if !v.VpcId.IsNull() && !v.VpcId.IsUnknown() { sub["vpcId"] = v.VpcId.ValueString() }
			items = append(items, sub)
		}
		m["routeTableAttachments"] = items
	}
	if !plan.ImageSchedules.IsNull() && !plan.ImageSchedules.IsUnknown() {
		var mp map[string]TxnImageScheduleModel
		plan.ImageSchedules.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Enabled.IsNull() && !v.Enabled.IsUnknown() { sub["enabled"] = v.Enabled.ValueBool() }
			if !v.ScheduleFormat.IsNull() && !v.ScheduleFormat.IsUnknown() { sub["scheduleFormat"] = v.ScheduleFormat.ValueString() }
			if !v.Schedule.IsNull() && !v.Schedule.IsUnknown() { sub["schedule"] = v.Schedule.ValueString() }
			if !v.RetentionCount.IsNull() && !v.RetentionCount.IsUnknown() { sub["retentionCount"] = v.RetentionCount.ValueInt64() }
			items = append(items, sub)
		}
		m["imageSchedules"] = items
	}
	if !plan.LoadbalancerTargetGroups.IsNull() && !plan.LoadbalancerTargetGroups.IsUnknown() {
		var mp map[string]TxnLoadbalancerTargetGroupModel
		plan.LoadbalancerTargetGroups.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			items = append(items, subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels))
		}
		m["loadbalancerTargetGroups"] = items
	}
	if !plan.LoadbalancerTargetGroupStaticTargets.IsNull() && !plan.LoadbalancerTargetGroupStaticTargets.IsUnknown() {
		var mp map[string]TxnLoadbalancerTargetGroupStaticTargetModel
		plan.LoadbalancerTargetGroupStaticTargets.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.TargetGroupId.IsNull() && !v.TargetGroupId.IsUnknown() { sub["targetGroupId"] = v.TargetGroupId.ValueString() }
			if !v.IpOrHostname.IsNull() && !v.IpOrHostname.IsUnknown() { sub["ipOrHostname"] = v.IpOrHostname.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancerTargetGroupStaticTargets"] = items
	}
	if !plan.LoadbalancerTargetGroupServiceDiscoveryTargets.IsNull() && !plan.LoadbalancerTargetGroupServiceDiscoveryTargets.IsUnknown() {
		var mp map[string]TxnLoadbalancerTargetGroupServiceDiscoveryTargetModel
		plan.LoadbalancerTargetGroupServiceDiscoveryTargets.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.TargetGroupId.IsNull() && !v.TargetGroupId.IsUnknown() { sub["targetGroupId"] = v.TargetGroupId.ValueString() }
			if !v.LabelSelectors.IsNull() && !v.LabelSelectors.IsUnknown() { sub["labelSelectors"] = strMapFromTF(ctx, v.LabelSelectors) }
			items = append(items, sub)
		}
		m["loadbalancerTargetGroupServiceDiscoveryTargets"] = items
	}
	if !plan.LoadbalancerHttpListeners.IsNull() && !plan.LoadbalancerHttpListeners.IsUnknown() {
		var mp map[string]TxnLoadbalancerHttpListenerModel
		plan.LoadbalancerHttpListeners.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.LoadbalancerId.IsNull() && !v.LoadbalancerId.IsUnknown() { sub["loadbalancerId"] = v.LoadbalancerId.ValueString() }
			if !v.Interface.IsNull() && !v.Interface.IsUnknown() { sub["interface"] = v.Interface.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.Ports.IsNull() && !v.Ports.IsUnknown() { sub["ports"] = stringListToInterface(ctx, v.Ports) }
			if !v.Hosts.IsNull() && !v.Hosts.IsUnknown() { sub["hosts"] = stringListToInterface(ctx, v.Hosts) }
			items = append(items, sub)
		}
		m["loadbalancerHttpListeners"] = items
	}
	if !plan.LoadbalancerHttpsListeners.IsNull() && !plan.LoadbalancerHttpsListeners.IsUnknown() {
		var mp map[string]TxnLoadbalancerHttpsListenerModel
		plan.LoadbalancerHttpsListeners.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.LoadbalancerId.IsNull() && !v.LoadbalancerId.IsUnknown() { sub["loadbalancerId"] = v.LoadbalancerId.ValueString() }
			if !v.Interface.IsNull() && !v.Interface.IsUnknown() { sub["interface"] = v.Interface.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.Ports.IsNull() && !v.Ports.IsUnknown() { sub["ports"] = stringListToInterface(ctx, v.Ports) }
			if !v.Hosts.IsNull() && !v.Hosts.IsUnknown() { sub["hosts"] = stringListToInterface(ctx, v.Hosts) }
			if !v.EnableHttp2Support.IsNull() && !v.EnableHttp2Support.IsUnknown() { sub["enableHttp2Support"] = v.EnableHttp2Support.ValueBool() }
			if !v.TlsCertificateId.IsNull() && !v.TlsCertificateId.IsUnknown() { sub["tlsCertificateId"] = v.TlsCertificateId.ValueString() }
			if !v.TlsProtocols.IsNull() && !v.TlsProtocols.IsUnknown() { sub["tlsProtocols"] = stringListToInterface(ctx, v.TlsProtocols) }
			if !v.TlsAutogenerateCertificate.IsNull() && !v.TlsAutogenerateCertificate.IsUnknown() { sub["tlsAutogenerateCertificate"] = v.TlsAutogenerateCertificate.ValueBool() }
			items = append(items, sub)
		}
		m["loadbalancerHttpsListeners"] = items
	}
	if !plan.LoadbalancerTlsListeners.IsNull() && !plan.LoadbalancerTlsListeners.IsUnknown() {
		var mp map[string]TxnLoadbalancerTlsListenerModel
		plan.LoadbalancerTlsListeners.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.LoadbalancerId.IsNull() && !v.LoadbalancerId.IsUnknown() { sub["loadbalancerId"] = v.LoadbalancerId.ValueString() }
			if !v.Interface.IsNull() && !v.Interface.IsUnknown() { sub["interface"] = v.Interface.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.Ports.IsNull() && !v.Ports.IsUnknown() { sub["ports"] = stringListToInterface(ctx, v.Ports) }
			if !v.Hosts.IsNull() && !v.Hosts.IsUnknown() { sub["hosts"] = stringListToInterface(ctx, v.Hosts) }
			if !v.TlsCertificateId.IsNull() && !v.TlsCertificateId.IsUnknown() { sub["tlsCertificateId"] = v.TlsCertificateId.ValueString() }
			if !v.TlsProtocols.IsNull() && !v.TlsProtocols.IsUnknown() { sub["tlsProtocols"] = stringListToInterface(ctx, v.TlsProtocols) }
			if !v.TlsAutogenerateCertificate.IsNull() && !v.TlsAutogenerateCertificate.IsUnknown() { sub["tlsAutogenerateCertificate"] = v.TlsAutogenerateCertificate.ValueBool() }
			items = append(items, sub)
		}
		m["loadbalancerTlsListeners"] = items
	}
	if !plan.LoadbalancerTcpListeners.IsNull() && !plan.LoadbalancerTcpListeners.IsUnknown() {
		var mp map[string]TxnLoadbalancerTcpListenerModel
		plan.LoadbalancerTcpListeners.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.LoadbalancerId.IsNull() && !v.LoadbalancerId.IsUnknown() { sub["loadbalancerId"] = v.LoadbalancerId.ValueString() }
			if !v.Interface.IsNull() && !v.Interface.IsUnknown() { sub["interface"] = v.Interface.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.Ports.IsNull() && !v.Ports.IsUnknown() { sub["ports"] = stringListToInterface(ctx, v.Ports) }
			items = append(items, sub)
		}
		m["loadbalancerTcpListeners"] = items
	}
	if !plan.LoadbalancerUdpListeners.IsNull() && !plan.LoadbalancerUdpListeners.IsUnknown() {
		var mp map[string]TxnLoadbalancerUdpListenerModel
		plan.LoadbalancerUdpListeners.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.LoadbalancerId.IsNull() && !v.LoadbalancerId.IsUnknown() { sub["loadbalancerId"] = v.LoadbalancerId.ValueString() }
			if !v.Interface.IsNull() && !v.Interface.IsUnknown() { sub["interface"] = v.Interface.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.Ports.IsNull() && !v.Ports.IsUnknown() { sub["ports"] = stringListToInterface(ctx, v.Ports) }
			items = append(items, sub)
		}
		m["loadbalancerUdpListeners"] = items
	}
	if !plan.LoadbalancerHttpListenerRules.IsNull() && !plan.LoadbalancerHttpListenerRules.IsUnknown() {
		var mp map[string]TxnLoadbalancerHttpListenerRuleModel
		plan.LoadbalancerHttpListenerRules.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.HttpListenerId.IsNull() && !v.HttpListenerId.IsUnknown() { sub["httpListenerId"] = v.HttpListenerId.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.MatchPath.IsNull() && !v.MatchPath.IsUnknown() { sub["matchPath"] = v.MatchPath.ValueString() }
			if !v.MatchPathMatchType.IsNull() && !v.MatchPathMatchType.IsUnknown() { sub["matchPathMatchType"] = v.MatchPathMatchType.ValueString() }
			if !v.ActionType.IsNull() && !v.ActionType.IsUnknown() { sub["actionType"] = v.ActionType.ValueString() }
			if !v.ActionJson.IsNull() && !v.ActionJson.IsUnknown() { sub["actionJson"] = v.ActionJson.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancerHttpListenerRules"] = items
	}
	if !plan.LoadbalancerHttpsListenerRules.IsNull() && !plan.LoadbalancerHttpsListenerRules.IsUnknown() {
		var mp map[string]TxnLoadbalancerHttpsListenerRuleModel
		plan.LoadbalancerHttpsListenerRules.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.HttpsListenerId.IsNull() && !v.HttpsListenerId.IsUnknown() { sub["httpsListenerId"] = v.HttpsListenerId.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.MatchPath.IsNull() && !v.MatchPath.IsUnknown() { sub["matchPath"] = v.MatchPath.ValueString() }
			if !v.MatchPathMatchType.IsNull() && !v.MatchPathMatchType.IsUnknown() { sub["matchPathMatchType"] = v.MatchPathMatchType.ValueString() }
			if !v.ActionType.IsNull() && !v.ActionType.IsUnknown() { sub["actionType"] = v.ActionType.ValueString() }
			if !v.ActionJson.IsNull() && !v.ActionJson.IsUnknown() { sub["actionJson"] = v.ActionJson.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancerHttpsListenerRules"] = items
	}
	if !plan.LoadbalancerTlsListenerRules.IsNull() && !plan.LoadbalancerTlsListenerRules.IsUnknown() {
		var mp map[string]TxnLoadbalancerTlsListenerRuleModel
		plan.LoadbalancerTlsListenerRules.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.TlsListenerId.IsNull() && !v.TlsListenerId.IsUnknown() { sub["tlsListenerId"] = v.TlsListenerId.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.ActionType.IsNull() && !v.ActionType.IsUnknown() { sub["actionType"] = v.ActionType.ValueString() }
			if !v.ActionJson.IsNull() && !v.ActionJson.IsUnknown() { sub["actionJson"] = v.ActionJson.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancerTlsListenerRules"] = items
	}
	if !plan.LoadbalancerTcpListenerRules.IsNull() && !plan.LoadbalancerTcpListenerRules.IsUnknown() {
		var mp map[string]TxnLoadbalancerTcpListenerRuleModel
		plan.LoadbalancerTcpListenerRules.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.TcpListenerId.IsNull() && !v.TcpListenerId.IsUnknown() { sub["tcpListenerId"] = v.TcpListenerId.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.ActionType.IsNull() && !v.ActionType.IsUnknown() { sub["actionType"] = v.ActionType.ValueString() }
			if !v.ActionJson.IsNull() && !v.ActionJson.IsUnknown() { sub["actionJson"] = v.ActionJson.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancerTcpListenerRules"] = items
	}
	if !plan.LoadbalancerUdpListenerRules.IsNull() && !plan.LoadbalancerUdpListenerRules.IsUnknown() {
		var mp map[string]TxnLoadbalancerUdpListenerRuleModel
		plan.LoadbalancerUdpListenerRules.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.UdpListenerId.IsNull() && !v.UdpListenerId.IsUnknown() { sub["udpListenerId"] = v.UdpListenerId.ValueString() }
			if !v.Order.IsNull() && !v.Order.IsUnknown() { sub["order"] = v.Order.ValueInt64() }
			if !v.ActionJson.IsNull() && !v.ActionJson.IsUnknown() { sub["actionJson"] = v.ActionJson.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancerUdpListenerRules"] = items
	}
	if !plan.KubernetesNodeGroups.IsNull() && !plan.KubernetesNodeGroups.IsUnknown() {
		var mp map[string]TxnKubernetesNodeGroupModel
		plan.KubernetesNodeGroups.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.KubernetesId.IsNull() && !v.KubernetesId.IsUnknown() { sub["kubernetesId"] = v.KubernetesId.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.VmOfferId.IsNull() && !v.VmOfferId.IsUnknown() { sub["vmOfferId"] = v.VmOfferId.ValueString() }
			if !v.VolumeOfferId.IsNull() && !v.VolumeOfferId.IsUnknown() { sub["volumeOfferId"] = v.VolumeOfferId.ValueString() }
			if !v.VolumeSizeGib.IsNull() && !v.VolumeSizeGib.IsUnknown() { sub["volumeSizeGib"] = v.VolumeSizeGib.ValueInt64() }
			if !v.DesiredNodeCount.IsNull() && !v.DesiredNodeCount.IsUnknown() { sub["desiredNodeCount"] = v.DesiredNodeCount.ValueInt64() }
			if !v.VmState.IsNull() && !v.VmState.IsUnknown() { sub["vmState"] = v.VmState.ValueString() }
			if !v.CreatePublicIpv4.IsNull() && !v.CreatePublicIpv4.IsUnknown() { sub["createPublicIpv4"] = v.CreatePublicIpv4.ValueBool() }
			items = append(items, sub)
		}
		m["kubernetesNodeGroups"] = items
	}
	if !plan.KubernetesUserRoles.IsNull() && !plan.KubernetesUserRoles.IsUnknown() {
		var mp map[string]TxnKubernetesUserRoleModel
		plan.KubernetesUserRoles.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.ApiGroups.IsNull() && !v.ApiGroups.IsUnknown() { sub["apiGroups"] = stringListToInterface(ctx, v.ApiGroups) }
			if !v.Resources.IsNull() && !v.Resources.IsUnknown() { sub["resources"] = stringListToInterface(ctx, v.Resources) }
			if !v.Verbs.IsNull() && !v.Verbs.IsUnknown() { sub["verbs"] = stringListToInterface(ctx, v.Verbs) }
			if !v.Namespaces.IsNull() && !v.Namespaces.IsUnknown() { sub["namespaces"] = stringListToInterface(ctx, v.Namespaces) }
			items = append(items, sub)
		}
		m["kubernetesUserRoles"] = items
	}
	if !plan.OpenVpns.IsNull() && !plan.OpenVpns.IsUnknown() {
		var mp map[string]TxnOpenVpnModel
		plan.OpenVpns.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			items = append(items, sub)
		}
		m["openVpns"] = items
	}
	if !plan.PostgresqlParametersSets.IsNull() && !plan.PostgresqlParametersSets.IsUnknown() {
		var mp map[string]TxnPostgresqlParametersSetModel
		plan.PostgresqlParametersSets.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Parameters.IsNull() && !v.Parameters.IsUnknown() { sub["parameters"] = strMapFromTF(ctx, v.Parameters) }
			items = append(items, sub)
		}
		m["postgreSqlParametersSets"] = items
	}
	if !plan.SupportPlans.IsNull() && !plan.SupportPlans.IsUnknown() {
		var mp map[string]TxnSupportPlanModel
		plan.SupportPlans.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			items = append(items, sub)
		}
		m["supportPlans"] = items
	}
	if !plan.SupportTickets.IsNull() && !plan.SupportTickets.IsUnknown() {
		var mp map[string]TxnSupportTicketModel
		plan.SupportTickets.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Kind.IsNull() && !v.Kind.IsUnknown() { sub["kind"] = v.Kind.ValueString() }
			if !v.Severity.IsNull() && !v.Severity.IsUnknown() { sub["severity"] = v.Severity.ValueString() }
			if !v.Status.IsNull() && !v.Status.IsUnknown() { sub["status"] = v.Status.ValueString() }
			items = append(items, sub)
		}
		m["supportTickets"] = items
	}
	if !plan.SupportTicketComments.IsNull() && !plan.SupportTicketComments.IsUnknown() {
		var mp map[string]TxnSupportTicketCommentModel
		plan.SupportTicketComments.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.TicketId.IsNull() && !v.TicketId.IsUnknown() { sub["ticketId"] = v.TicketId.ValueString() }
			if !v.Content.IsNull() && !v.Content.IsUnknown() { sub["content"] = v.Content.ValueString() }
			if !v.AttachmentsIds.IsNull() && !v.AttachmentsIds.IsUnknown() { sub["attachmentsIds"] = stringListToInterface(ctx, v.AttachmentsIds) }
			items = append(items, sub)
		}
		m["supportTicketComments"] = items
	}
	if !plan.GitlabRunners.IsNull() && !plan.GitlabRunners.IsUnknown() {
		var mp map[string]TxnGitlabRunnerModel
		plan.GitlabRunners.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			if !v.VmState.IsNull() && !v.VmState.IsUnknown() { sub["vmState"] = v.VmState.ValueString() }
			if !v.VmOfferId.IsNull() && !v.VmOfferId.IsUnknown() { sub["vmOfferId"] = v.VmOfferId.ValueString() }
			if !v.VolumeOfferId.IsNull() && !v.VolumeOfferId.IsUnknown() { sub["volumeOfferId"] = v.VolumeOfferId.ValueString() }
			if !v.VolumeSizeGib.IsNull() && !v.VolumeSizeGib.IsUnknown() { sub["volumeSizeGib"] = v.VolumeSizeGib.ValueInt64() }
			if !v.Concurrency.IsNull() && !v.Concurrency.IsUnknown() { sub["concurrency"] = v.Concurrency.ValueInt64() }
			if !v.Version.IsNull() && !v.Version.IsUnknown() { sub["version"] = v.Version.ValueString() }
			if !v.DockerOptionsJsonString.IsNull() && !v.DockerOptionsJsonString.IsUnknown() { sub["dockerOptionsJsonString"] = v.DockerOptionsJsonString.ValueString() }
			items = append(items, sub)
		}
		m["gitlabRunners"] = items
	}
	if !plan.OpenVpnUserSettings.IsNull() && !plan.OpenVpnUserSettings.IsUnknown() {
		var mp map[string]TxnOpenVpnUserSettingsModel
		plan.OpenVpnUserSettings.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.AllowedIpV4Cidrs.IsNull() && !v.AllowedIpV4Cidrs.IsUnknown() { sub["allowedIpV4Cidrs"] = stringListToInterface(ctx, v.AllowedIpV4Cidrs) }
			if !v.AllowedIpV6Cidrs.IsNull() && !v.AllowedIpV6Cidrs.IsUnknown() { sub["allowedIpV6Cidrs"] = stringListToInterface(ctx, v.AllowedIpV6Cidrs) }
			if !v.DeniedIpV4Cidrs.IsNull() && !v.DeniedIpV4Cidrs.IsUnknown() { sub["deniedIpV4Cidrs"] = stringListToInterface(ctx, v.DeniedIpV4Cidrs) }
			if !v.DeniedIpV6Cidrs.IsNull() && !v.DeniedIpV6Cidrs.IsUnknown() { sub["deniedIpV6Cidrs"] = stringListToInterface(ctx, v.DeniedIpV6Cidrs) }
			if !v.AllowedDomains.IsNull() && !v.AllowedDomains.IsUnknown() { sub["allowedDomains"] = stringListToInterface(ctx, v.AllowedDomains) }
			if !v.DeniedDomains.IsNull() && !v.DeniedDomains.IsUnknown() { sub["deniedDomains"] = stringListToInterface(ctx, v.DeniedDomains) }
			items = append(items, sub)
		}
		m["openVpnUserSettings"] = items
	}
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		var mp map[string]TxnIamUserModel
		plan.Users.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Email.IsNull() && !v.Email.IsUnknown() { sub["email"] = v.Email.ValueString() }
			if !v.AccessPolicyIds.IsNull() && !v.AccessPolicyIds.IsUnknown() { sub["accessPolicyIds"] = stringListToInterface(ctx, v.AccessPolicyIds) }
			items = append(items, sub)
		}
		m["users"] = items
	}
	if !plan.UserTokens.IsNull() && !plan.UserTokens.IsUnknown() {
		var mp map[string]TxnUserTokenModel
		plan.UserTokens.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.UserId.IsNull() && !v.UserId.IsUnknown() { sub["userId"] = v.UserId.ValueString() }
			if !v.SendToEmail.IsNull() && !v.SendToEmail.IsUnknown() { sub["sendToEmail"] = v.SendToEmail.ValueBool() }
			items = append(items, sub)
		}
		m["userTokens"] = items
	}
	if !plan.FloatingIps.IsNull() && !plan.FloatingIps.IsUnknown() {
		var mp map[string]TxnFloatingIpModel
		plan.FloatingIps.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.HostingProviderId.IsNull() && !v.HostingProviderId.IsUnknown() { sub["hostingProviderId"] = v.HostingProviderId.ValueString() }
			items = append(items, sub)
		}
		m["floatingIps"] = items
	}
	if !plan.Vpcs.IsNull() && !plan.Vpcs.IsUnknown() {
		var mp map[string]TxnVpcModel
		plan.Vpcs.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.HostingProviderId.IsNull() && !v.HostingProviderId.IsUnknown() { sub["hostingProviderId"] = v.HostingProviderId.ValueString() }
			if !v.Ipv4Cidr.IsNull() && !v.Ipv4Cidr.IsUnknown() { sub["ipv4Cidr"] = v.Ipv4Cidr.ValueString() }
			if !v.NatFloatingIpId.IsNull() && !v.NatFloatingIpId.IsUnknown() { sub["natFloatingIpId"] = v.NatFloatingIpId.ValueString() }
			if !v.SecurityGroupIds.IsNull() && !v.SecurityGroupIds.IsUnknown() { sub["securityGroupIds"] = stringListToInterface(ctx, v.SecurityGroupIds) }
			if !v.ExternallyManaged.IsNull() && !v.ExternallyManaged.IsUnknown() { sub["externallyManaged"] = v.ExternallyManaged.ValueBool() }
			items = append(items, sub)
		}
		m["vpcs"] = items
	}
	if !plan.VpcPeeringPeers.IsNull() && !plan.VpcPeeringPeers.IsUnknown() {
		var mp map[string]TxnVpcPeeringPeerModel
		plan.VpcPeeringPeers.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.VpcPeeringId.IsNull() && !v.VpcPeeringId.IsUnknown() { sub["vpcPeeringId"] = v.VpcPeeringId.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			items = append(items, sub)
		}
		m["vpcPeeringPeers"] = items
	}
	if !plan.Loadbalancers.IsNull() && !plan.Loadbalancers.IsUnknown() {
		var mp map[string]TxnLoadbalancerModel
		plan.Loadbalancers.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			items = append(items, sub)
		}
		m["loadbalancers"] = items
	}
	if !plan.Kubernetes.IsNull() && !plan.Kubernetes.IsUnknown() {
		var mp map[string]TxnKubernetesModel
		plan.Kubernetes.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.AssignPublicIpV4.IsNull() && !v.AssignPublicIpV4.IsUnknown() { sub["assignPublicIpV4"] = v.AssignPublicIpV4.ValueBool() }
			if !v.Version.IsNull() && !v.Version.IsUnknown() { sub["version"] = v.Version.ValueString() }
			if !v.ControlPlaneLocations.IsNull() && !v.ControlPlaneLocations.IsUnknown() {
				type cplModel struct{ VpcSubnetId types.String `tfsdk:"vpc_subnet_id"` }
				var locs []cplModel
				v.ControlPlaneLocations.ElementsAs(ctx, &locs, true)
				cplItems := make([]interface{}, 0, len(locs))
				for _, loc := range locs {
					cplItems = append(cplItems, map[string]interface{}{"vpcSubnetId": loc.VpcSubnetId.ValueString()})
				}
				sub["controlPlaneLocations"] = cplItems
			}
			items = append(items, sub)
		}
		m["kuberneteses"] = items
	}
	if !plan.KubernetesUsers.IsNull() && !plan.KubernetesUsers.IsUnknown() {
		var mp map[string]TxnKubernetesUserModel
		plan.KubernetesUsers.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.KubernetesId.IsNull() && !v.KubernetesId.IsUnknown() { sub["kubernetesId"] = v.KubernetesId.ValueString() }
			if !v.RoleIds.IsNull() && !v.RoleIds.IsUnknown() { sub["roleIds"] = stringListToInterface(ctx, v.RoleIds) }
			items = append(items, sub)
		}
		m["kubernetesUsers"] = items
	}
	if !plan.PostgresqlStandalones.IsNull() && !plan.PostgresqlStandalones.IsUnknown() {
		var mp map[string]TxnPostgresqlStandaloneModel
		plan.PostgresqlStandalones.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.Version.IsNull() && !v.Version.IsUnknown() { sub["version"] = v.Version.ValueString() }
			if !v.RootPassword.IsNull() && !v.RootPassword.IsUnknown() { sub["rootPassword"] = v.RootPassword.ValueString() }
			if !v.ParametersSetId.IsNull() && !v.ParametersSetId.IsUnknown() { sub["parametersSetId"] = v.ParametersSetId.ValueString() }
			if !v.BackupRetentionDays.IsNull() && !v.BackupRetentionDays.IsUnknown() { sub["backupRetentionDays"] = v.BackupRetentionDays.ValueInt64() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.VmState.IsNull() && !v.VmState.IsUnknown() { sub["vmState"] = v.VmState.ValueString() }
			if !v.VmOfferId.IsNull() && !v.VmOfferId.IsUnknown() { sub["vmOfferId"] = v.VmOfferId.ValueString() }
			if !v.VolumeOfferId.IsNull() && !v.VolumeOfferId.IsUnknown() { sub["volumeOfferId"] = v.VolumeOfferId.ValueString() }
			if !v.VolumeSizeGib.IsNull() && !v.VolumeSizeGib.IsUnknown() { sub["volumeSizeGib"] = v.VolumeSizeGib.ValueInt64() }
			items = append(items, sub)
		}
		m["postgreSqlStandalones"] = items
	}
	if !plan.OpenVpnUsers.IsNull() && !plan.OpenVpnUsers.IsUnknown() {
		var mp map[string]TxnOpenVpnUserModel
		plan.OpenVpnUsers.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.OpenVpnId.IsNull() && !v.OpenVpnId.IsUnknown() { sub["openVpnId"] = v.OpenVpnId.ValueString() }
			if !v.OpenVpnSettingsIds.IsNull() && !v.OpenVpnSettingsIds.IsUnknown() { sub["openVpnSettingsIds"] = stringListToInterface(ctx, v.OpenVpnSettingsIds) }
			items = append(items, sub)
		}
		m["openVpnUsers"] = items
	}
	if !plan.BillingAccounts.IsNull() && !plan.BillingAccounts.IsUnknown() {
		var mp map[string]TxnBillingAccountModel
		plan.BillingAccounts.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.ResourceName.IsNull() && !v.ResourceName.IsUnknown() { sub["resourceName"] = v.ResourceName.ValueString() }
			items = append(items, sub)
		}
		m["billingAccounts"] = items
	}
	if !plan.Quotas.IsNull() && !plan.Quotas.IsUnknown() {
		var mp map[string]TxnQuotaModel
		plan.Quotas.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Product.IsNull() && !v.Product.IsUnknown() { sub["product"] = v.Product.ValueString() }
			if !v.Resource.IsNull() && !v.Resource.IsUnknown() { sub["resource"] = v.Resource.ValueString() }
			if !v.Parameter.IsNull() && !v.Parameter.IsUnknown() { sub["parameter"] = v.Parameter.ValueString() }
			if !v.Limit.IsNull() && !v.Limit.IsUnknown() { sub["limit"] = v.Limit.ValueInt64() }
			items = append(items, sub)
		}
		m["quotas"] = items
	}
	if !plan.QuotaChangeRequests.IsNull() && !plan.QuotaChangeRequests.IsUnknown() {
		var mp map[string]TxnQuotaChangeRequestModel
		plan.QuotaChangeRequests.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.QuotaId.IsNull() && !v.QuotaId.IsUnknown() { sub["quotaId"] = v.QuotaId.ValueString() }
			if !v.NewQuotaLimit.IsNull() && !v.NewQuotaLimit.IsUnknown() { sub["newQuotaLimit"] = v.NewQuotaLimit.ValueInt64() }
			items = append(items, sub)
		}
		m["quotaChangeRequests"] = items
	}
	if !plan.VictoriaMetrics.IsNull() && !plan.VictoriaMetrics.IsUnknown() {
		var mp map[string]TxnVictoriaMetricsModel
		plan.VictoriaMetrics.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.VpcId.IsNull() && !v.VpcId.IsUnknown() { sub["vpcId"] = v.VpcId.ValueString() }
			if !v.CreatePublicIpv4.IsNull() && !v.CreatePublicIpv4.IsUnknown() { sub["createPublicIpv4"] = v.CreatePublicIpv4.ValueBool() }
			if !v.CreatePublicIpv6.IsNull() && !v.CreatePublicIpv6.IsUnknown() { sub["createPublicIpv6"] = v.CreatePublicIpv6.ValueBool() }
			if !v.DnsRecordName.IsNull() && !v.DnsRecordName.IsUnknown() { sub["dnsRecordName"] = v.DnsRecordName.ValueString() }
			items = append(items, sub)
		}
		m["victoriaMetricss"] = items
	}
	if !plan.Gitlabs.IsNull() && !plan.Gitlabs.IsUnknown() {
		var mp map[string]TxnGitlabModel
		plan.Gitlabs.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.Tier.IsNull() && !v.Tier.IsUnknown() { sub["tier"] = v.Tier.ValueString() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.Version.IsNull() && !v.Version.IsUnknown() { sub["version"] = v.Version.ValueString() }
			if !v.RootPassword.IsNull() && !v.RootPassword.IsUnknown() { sub["rootPassword"] = v.RootPassword.ValueString() }
			if !v.VmState.IsNull() && !v.VmState.IsUnknown() { sub["vmState"] = v.VmState.ValueString() }
			if !v.VmOfferId.IsNull() && !v.VmOfferId.IsUnknown() { sub["vmOfferId"] = v.VmOfferId.ValueString() }
			if !v.VolumeOfferId.IsNull() && !v.VolumeOfferId.IsUnknown() { sub["volumeOfferId"] = v.VolumeOfferId.ValueString() }
			if !v.VolumeSizeGib.IsNull() && !v.VolumeSizeGib.IsUnknown() { sub["volumeSizeGib"] = v.VolumeSizeGib.ValueInt64() }
			if !v.Edition.IsNull() && !v.Edition.IsUnknown() { sub["edition"] = v.Edition.ValueString() }
			if !v.RecordName.IsNull() && !v.RecordName.IsUnknown() { sub["recordName"] = v.RecordName.ValueString() }
			items = append(items, sub)
		}
		m["gitlabs"] = items
	}
	if !plan.SupportTicketCommentAttachments.IsNull() && !plan.SupportTicketCommentAttachments.IsUnknown() {
		var mp map[string]TxnSupportTicketCommentAttachmentModel
		plan.SupportTicketCommentAttachments.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.FileName.IsNull() && !v.FileName.IsUnknown() { sub["fileName"] = v.FileName.ValueString() }
			if !v.FileType.IsNull() && !v.FileType.IsUnknown() { sub["fileType"] = v.FileType.ValueString() }
			if !v.FileContentBase64.IsNull() && !v.FileContentBase64.IsUnknown() { sub["fileContentBase64"] = v.FileContentBase64.ValueString() }
			items = append(items, sub)
		}
		m["supportTicketCommentAttachments"] = items
	}
	if !plan.Images.IsNull() && !plan.Images.IsUnknown() {
		var mp map[string]TxnImageModel
		plan.Images.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.VmId.IsNull() && !v.VmId.IsUnknown() { sub["vmId"] = v.VmId.ValueString() }
			items = append(items, sub)
		}
		m["images"] = items
	}
	if !plan.Vms.IsNull() && !plan.Vms.IsUnknown() {
		var mp map[string]TxnVmModel
		plan.Vms.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			if !v.VmState.IsNull() && !v.VmState.IsUnknown() { sub["vmState"] = v.VmState.ValueString() }
			if !v.VpcSubnetId.IsNull() && !v.VpcSubnetId.IsUnknown() { sub["vpcSubnetId"] = v.VpcSubnetId.ValueString() }
			if !v.FloatingIpId.IsNull() && !v.FloatingIpId.IsUnknown() { sub["floatingIpId"] = v.FloatingIpId.ValueString() }
			if !v.ImageId.IsNull() && !v.ImageId.IsUnknown() { sub["imageId"] = v.ImageId.ValueString() }
			if !v.OfferId.IsNull() && !v.OfferId.IsUnknown() { sub["offerId"] = v.OfferId.ValueString() }
			if !v.ImageBootVolumeDeviceIndex.IsNull() && !v.ImageBootVolumeDeviceIndex.IsUnknown() { sub["imageBootVolumeDeviceIndex"] = v.ImageBootVolumeDeviceIndex.ValueInt64() }
			if !v.SshKeyIds.IsNull() && !v.SshKeyIds.IsUnknown() { sub["sshKeyIds"] = stringListToInterface(ctx, v.SshKeyIds) }
			if !v.ImageScheduleIds.IsNull() && !v.ImageScheduleIds.IsUnknown() { sub["imageScheduleIds"] = stringListToInterface(ctx, v.ImageScheduleIds) }
			if !v.BootstrapCommand.IsNull() && !v.BootstrapCommand.IsUnknown() {
				type bcModel struct {
					Command           types.String `tfsdk:"command"`
					SuccessReturnCode types.Int64  `tfsdk:"success_return_code"`
					TimeoutSeconds    types.Int64  `tfsdk:"timeout_seconds"`
				}
				var cmds []bcModel
				v.BootstrapCommand.ElementsAs(ctx, &cmds, true)
				bcItems := make([]interface{}, 0, len(cmds))
				for _, c := range cmds {
					bc := map[string]interface{}{"command": c.Command.ValueString()}
					if !c.SuccessReturnCode.IsNull() && !c.SuccessReturnCode.IsUnknown() { bc["successReturnCode"] = c.SuccessReturnCode.ValueInt64() }
					if !c.TimeoutSeconds.IsNull() && !c.TimeoutSeconds.IsUnknown() { bc["timeoutSeconds"] = c.TimeoutSeconds.ValueInt64() }
					bcItems = append(bcItems, bc)
				}
				sub["bootstrapCommand"] = bcItems
			}
			items = append(items, sub)
		}
		m["vms"] = items
	}
	if !plan.SecurityGroups.IsNull() && !plan.SecurityGroups.IsUnknown() {
		var mp map[string]TxnSecurityGroupModel
		plan.SecurityGroups.ElementsAs(ctx, &mp, true)
		items := make([]interface{}, 0, len(mp))
		for _, v := range mp {
			sub := subBase(v.ID, v.Name, v.Description, v.FolderID, v.DeleteProtection, v.Labels)
			buildRules := func(src types.List) []interface{} {
				if src.IsNull() || src.IsUnknown() { return nil }
				type ruleModel struct {
					Ports      types.List   `tfsdk:"ports"`
					Ipv4Blocks types.List   `tfsdk:"ipv4_blocks"`
					Ipv6Blocks types.List   `tfsdk:"ipv6_blocks"`
					Action     types.String `tfsdk:"action"`
				}
				var rules []ruleModel
				src.ElementsAs(ctx, &rules, true)
				out := make([]interface{}, 0, len(rules))
				for _, r := range rules {
					rm := map[string]interface{}{}
					if !r.Ports.IsNull() && !r.Ports.IsUnknown() { rm["ports"] = stringListToInterface(ctx, r.Ports) }
					if !r.Ipv4Blocks.IsNull() && !r.Ipv4Blocks.IsUnknown() { rm["ipv4Blocks"] = stringListToInterface(ctx, r.Ipv4Blocks) }
					if !r.Ipv6Blocks.IsNull() && !r.Ipv6Blocks.IsUnknown() { rm["ipv6Blocks"] = stringListToInterface(ctx, r.Ipv6Blocks) }
					if !r.Action.IsNull() && !r.Action.IsUnknown() { rm["action"] = r.Action.ValueString() }
					out = append(out, rm)
				}
				return out
			}
			if rules := buildRules(v.Ingress); rules != nil { sub["ingress"] = rules }
			if rules := buildRules(v.Egress); rules != nil { sub["egress"] = rules }
			items = append(items, sub)
		}
		m["securityGroups"] = items
	}

	return m
}

// assignSubIDsMap pre-generates ULIDs for map entries that have no id yet.
func assignSubIDsMap(m types.Map, attrTypes map[string]attr.Type) types.Map {
	if m.IsNull() || m.IsUnknown() { return m }
	elems := m.Elements()
	if len(elems) == 0 { return m }
	updated := make(map[string]attr.Value, len(elems))
	for key, e := range elems {
		obj, ok := e.(types.Object)
		if !ok { updated[key] = e; continue }
		attrs := obj.Attributes()
		if idVal, ok2 := attrs["id"].(types.String); ok2 && !idVal.IsNull() && !idVal.IsUnknown() && idVal.ValueString() != "" {
			updated[key] = e; continue
		}
		newAttrs := make(map[string]attr.Value, len(attrs))
		for k, v := range attrs { newAttrs[k] = v }
		newAttrs["id"] = types.StringValue(newULID())
		newObj, diags := types.ObjectValue(attrTypes, newAttrs)
		if diags.HasError() { updated[key] = e; continue }
		updated[key] = newObj
	}
	result, _ := types.MapValue(types.ObjectType{AttrTypes: attrTypes}, updated)
	return result
}

func ulidOrNew(id types.String) string {
	if id.IsNull() || id.IsUnknown() || id.ValueString() == "" {
		return newULID()
	}
	return id.ValueString()
}

func strMapFromTF(ctx context.Context, m types.Map) map[string]string {
	var result map[string]string
	m.ElementsAs(ctx, &result, false)
	return result
}

// ---------------------------------------------------------------------------
// State population
// ---------------------------------------------------------------------------

func (r *TransactionResource) populateState(ctx context.Context, data map[string]interface{}, state *TransactionResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.DeleteResourcesOnTransactionDelete = getBool(data, "deleteResourcesOnTransactionDelete")
	state.State = getStringFromInfo(data, "state")

	state.Folders = populateTxnFolders(ctx, data, state.Folders)
	state.SshKeys = populateTxnSshKeys(ctx, data, state.SshKeys)
	state.S3Buckets = populateTxnS3Buckets(ctx, data, state.S3Buckets)
	state.S3UserAccessPolicies = populateTxnS3Policies(ctx, data, state.S3UserAccessPolicies)
	state.S3Users = populateTxnS3Users(ctx, data, state.S3Users)
	state.Volumes = populateTxnVolumes(ctx, data, state.Volumes)
	state.VolumeAttachments = populateTxnVolumeAttachments(ctx, data, state.VolumeAttachments)
	state.AccessPolicies = populateTxnAccessPolicies(ctx, data, state.AccessPolicies)
	state.HostingProviders = populateTxnHostingProviders(ctx, data, state.HostingProviders)
	state.SshPrivateKeys = populateTxnSshPrivateKeys(ctx, data, state.SshPrivateKeys)
	state.Certificates = populateTxnCertificates(ctx, data, state.Certificates)
	state.VpcSubnets = populateTxnVpcSubnets(ctx, data, state.VpcSubnets)
	state.VpcPeerings = populateTxnVpcPeerings(ctx, data, state.VpcPeerings)
	state.VpcPeeringExternalPeers = populateTxnVpcPeeringExternalPeers(ctx, data, state.VpcPeeringExternalPeers)
	state.RouteTables = populateTxnRouteTables(ctx, data, state.RouteTables)
	state.RouteTableRoutes = populateTxnRouteTableRoutes(ctx, data, state.RouteTableRoutes)
	state.RouteTableAttachments = populateTxnRouteTableAttachments(ctx, data, state.RouteTableAttachments)
	state.ImageSchedules = populateTxnImageSchedules(ctx, data, state.ImageSchedules)
	state.LoadbalancerTargetGroups = populateTxnLoadbalancerTargetGroups(ctx, data, state.LoadbalancerTargetGroups)
	state.LoadbalancerTargetGroupStaticTargets = populateTxnLoadbalancerTargetGroupStaticTargets(ctx, data, state.LoadbalancerTargetGroupStaticTargets)
	state.LoadbalancerTargetGroupServiceDiscoveryTargets = populateTxnLoadbalancerTargetGroupServiceDiscoveryTargets(ctx, data, state.LoadbalancerTargetGroupServiceDiscoveryTargets)
	state.LoadbalancerHttpListeners = populateTxnLoadbalancerHttpListeners(ctx, data, state.LoadbalancerHttpListeners)
	state.LoadbalancerHttpsListeners = populateTxnLoadbalancerHttpsListeners(ctx, data, state.LoadbalancerHttpsListeners)
	state.LoadbalancerTlsListeners = populateTxnLoadbalancerTlsListeners(ctx, data, state.LoadbalancerTlsListeners)
	state.LoadbalancerTcpListeners = populateTxnLoadbalancerTcpListeners(ctx, data, state.LoadbalancerTcpListeners)
	state.LoadbalancerUdpListeners = populateTxnLoadbalancerUdpListeners(ctx, data, state.LoadbalancerUdpListeners)
	state.LoadbalancerHttpListenerRules = populateTxnLoadbalancerHttpListenerRules(ctx, data, state.LoadbalancerHttpListenerRules)
	state.LoadbalancerHttpsListenerRules = populateTxnLoadbalancerHttpsListenerRules(ctx, data, state.LoadbalancerHttpsListenerRules)
	state.LoadbalancerTlsListenerRules = populateTxnLoadbalancerTlsListenerRules(ctx, data, state.LoadbalancerTlsListenerRules)
	state.LoadbalancerTcpListenerRules = populateTxnLoadbalancerTcpListenerRules(ctx, data, state.LoadbalancerTcpListenerRules)
	state.LoadbalancerUdpListenerRules = populateTxnLoadbalancerUdpListenerRules(ctx, data, state.LoadbalancerUdpListenerRules)
	state.KubernetesNodeGroups = populateTxnKubernetesNodeGroups(ctx, data, state.KubernetesNodeGroups)
	state.KubernetesUserRoles = populateTxnKubernetesUserRoles(ctx, data, state.KubernetesUserRoles)
	state.OpenVpns = populateTxnOpenVpns(ctx, data, state.OpenVpns)
	state.PostgresqlParametersSets = populateTxnPostgresqlParametersSets(ctx, data, state.PostgresqlParametersSets)
	state.SupportPlans = populateTxnSupportPlans(ctx, data, state.SupportPlans)
	state.SupportTickets = populateTxnSupportTickets(ctx, data, state.SupportTickets)
	state.SupportTicketComments = populateTxnSupportTicketComments(ctx, data, state.SupportTicketComments)
	state.GitlabRunners = populateTxnGitlabRunners(ctx, data, state.GitlabRunners)
	state.OpenVpnUserSettings = populateTxnOpenVpnUserSettings(ctx, data, state.OpenVpnUserSettings)
	state.Users = populateTxnUsers(ctx, data, state.Users)
	state.UserTokens = populateTxnUserTokens(ctx, data, state.UserTokens)
	state.FloatingIps = populateTxnFloatingIps(ctx, data, state.FloatingIps)
	state.Vpcs = populateTxnVpcs(ctx, data, state.Vpcs)
	state.VpcPeeringPeers = populateTxnVpcPeeringPeers(ctx, data, state.VpcPeeringPeers)
	state.Loadbalancers = populateTxnLoadbalancers(ctx, data, state.Loadbalancers)
	state.Kubernetes = populateTxnKubernetes(ctx, data, state.Kubernetes)
	state.KubernetesUsers = populateTxnKubernetesUsers(ctx, data, state.KubernetesUsers)
	state.PostgresqlStandalones = populateTxnPostgresqlStandalones(ctx, data, state.PostgresqlStandalones)
	state.OpenVpnUsers = populateTxnOpenVpnUsers(ctx, data, state.OpenVpnUsers)
	state.BillingAccounts = populateTxnBillingAccounts(ctx, data, state.BillingAccounts)
	state.Quotas = populateTxnQuotas(ctx, data, state.Quotas)
	state.QuotaChangeRequests = populateTxnQuotaChangeRequests(ctx, data, state.QuotaChangeRequests)
	state.VictoriaMetrics = populateTxnVictoriaMetrics(ctx, data, state.VictoriaMetrics)
	state.Gitlabs = populateTxnGitlabs(ctx, data, state.Gitlabs)
	state.SupportTicketCommentAttachments = populateTxnSupportTicketCommentAttachments(ctx, data, state.SupportTicketCommentAttachments)
	state.Images = populateTxnImages(ctx, data, state.Images)
	state.Vms = populateTxnVms(ctx, data, state.Vms)
	state.SecurityGroups = populateTxnSecurityGroups(ctx, data, state.SecurityGroups)
	return nil
}

// findMapKeyByID returns the map key whose object has the given "id" value, or ("", false).
func findMapKeyByID(m types.Map, id string) (string, bool) {
	if m.IsNull() || m.IsUnknown() || id == "" {
		return "", false
	}
	for k, v := range m.Elements() {
		obj, ok := v.(types.Object)
		if !ok { continue }
		if idAttr, ok2 := obj.Attributes()["id"].(types.String); ok2 && idAttr.ValueString() == id {
			return k, true
		}
	}
	return "", false
}

// resolveMapKey returns the stable map key for a sub-resource.
// Priority: (1) existing state key matched by ID — preserves user-chosen key names across updates;
// (2) name from API response — handles new sub-resources never saved to state (e.g. after Ctrl+C);
// (3) raw ID as last resort.
func resolveMapKey(existing types.Map, id string, item map[string]interface{}) string {
	if k, found := findMapKeyByID(existing, id); found {
		return k
	}
	if name, _ := item["name"].(string); name != "" {
		return name
	}
	return id
}

// simpleStateInfo returns a types.Object info block for sub-resources
func simpleStateInfo(data map[string]interface{}) types.Object {
	return simpleStateInfoObj(data)
}

// populateSimpleMap is a generic helper for sub-resources with only simple scalar fields.
// It reduces boilerplate for the many types that share the same populate pattern.
func populateSimpleMap(data map[string]interface{}, apiKey string, existing types.Map, attrTypes map[string]attr.Type, buildObj func(m map[string]interface{}) map[string]attr.Value) types.Map {
	raw, ok := data[apiKey].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: attrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		fm, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := fm["id"].(string)
		attrs := buildObj(fm)
		obj, _ := types.ObjectValue(attrTypes, attrs)
		items[resolveMapKey(existing, id, fm)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: attrTypes}, items)
}

// --- Original 5 populate functions (unchanged) ---

func populateTxnFolders(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["folders"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnFolderAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		fm, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := fm["id"].(string)
		obj, _ := types.ObjectValue(txnFolderAttrTypes, map[string]attr.Value{
			"id": getString(fm, "id"), "name": getString(fm, "name"),
			"description": getString(fm, "description"), "folder_id": getString(fm, "folderId"),
			"delete_protection": getBool(fm, "deleteProtection"), "labels": getStringMap(fm, "labels"),
			"info": simpleStateInfo(fm),
		})
		items[resolveMapKey(existing, id, fm)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnFolderAttrTypes}, items)
}

func populateTxnSshKeys(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["sshKeys"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnSshKeyAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		km, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := km["id"].(string)
		obj, _ := types.ObjectValue(txnSshKeyAttrTypes, map[string]attr.Value{
			"id": getString(km, "id"), "name": getString(km, "name"),
			"description": getString(km, "description"), "folder_id": getString(km, "folderId"),
			"delete_protection": getBool(km, "deleteProtection"), "labels": getStringMap(km, "labels"),
			"public_key": getString(km, "publicKey"), "resource_name": getString(km, "resourceName"),
			"info": simpleStateInfo(km),
		})
		items[resolveMapKey(existing, id, km)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnSshKeyAttrTypes}, items)
}

func populateTxnS3Buckets(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["s3Buckets"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnS3BucketAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		bm, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := bm["id"].(string)
		bucketInfo, _ := types.ObjectValue(map[string]attr.Type{"state": types.StringType, "endpoint_url": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(bm, "state"), "endpoint_url": getStringFromInfo(bm, "endpointUrl")})
		obj, _ := types.ObjectValue(txnS3BucketAttrTypes, map[string]attr.Value{
			"id": getString(bm, "id"), "name": getString(bm, "name"),
			"description": getString(bm, "description"), "folder_id": getString(bm, "folderId"),
			"delete_protection": getBool(bm, "deleteProtection"), "labels": getStringMap(bm, "labels"),
			"tier": getString(bm, "tier"), "region": getString(bm, "region"),
			"is_public": getBool(bm, "isPublic"), "is_versioned": getBool(bm, "isVersioned"),
			"is_lock_enabled": getBool(bm, "isLockEnabled"), "quota_gib": getInt64(bm, "quotaGiB"),
			"object_expiration_days": getInt64(bm, "objectExpirationDays"),
			"compliance_retention_days": getInt64(bm, "complianceRetentionDays"),
			"info": bucketInfo,
		})
		items[resolveMapKey(existing, id, bm)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnS3BucketAttrTypes}, items)
}

func populateTxnS3Policies(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["s3UserAccessPolicies"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnS3PolicyAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		pm, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := pm["id"].(string)
		obj, _ := types.ObjectValue(txnS3PolicyAttrTypes, map[string]attr.Value{
			"id": getString(pm, "id"), "name": getString(pm, "name"),
			"description": getString(pm, "description"), "folder_id": getString(pm, "folderId"),
			"delete_protection": getBool(pm, "deleteProtection"), "labels": getStringMap(pm, "labels"),
			"policy_json": getString(pm, "policyJson"), "info": simpleStateInfo(pm),
		})
		items[resolveMapKey(existing, id, pm)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnS3PolicyAttrTypes}, items)
}

func populateTxnS3Users(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, _ := data["s3Users"].([]interface{})
	if len(raw) == 0 {
		if !existing.IsNull() && !existing.IsUnknown() { return existing }
		return types.MapValueMust(types.ObjectType{AttrTypes: txnS3UserAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		um, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := um["id"].(string)
		policyIds := getStringList(ctx, um, "accessPolicyIds")
		userInfo, _ := types.ObjectValue(map[string]attr.Type{"state": types.StringType, "access_key": types.StringType, "secret_key": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(um, "state"), "access_key": getStringFromInfo(um, "accessKey"), "secret_key": getStringFromInfo(um, "secretKey")})
		obj, _ := types.ObjectValue(txnS3UserAttrTypes, map[string]attr.Value{
			"id": getString(um, "id"), "name": getString(um, "name"),
			"description": getString(um, "description"), "folder_id": getString(um, "folderId"),
			"delete_protection": getBool(um, "deleteProtection"), "labels": getStringMap(um, "labels"),
			"bucket_id": getString(um, "bucketId"), "access_policy_ids": policyIds,
			"info": userInfo,
		})
		items[resolveMapKey(existing, id, um)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnS3UserAttrTypes}, items)
}

// --- New 54 populate functions ---

func populateTxnVolumes(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "volumes", existing, txnVolumeAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"hosting_provider_id": getString(m, "hostingProviderId"), "offer_id": getString(m, "offerId"),
			"size_gib": getInt64(m, "sizeGib"), "os_image_id": getString(m, "osImageId"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnVolumeAttachments(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "volumeAttachments", existing, txnVolumeAttachmentAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"volume_id": getString(m, "volumeId"), "vm_id": getString(m, "vmId"),
			"vm_device_index": getInt64(m, "vmDeviceIndex"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnAccessPolicies(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "accessPolicies", existing, txnAccessPolicyAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"content": getString(m, "content"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnHostingProviders(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["hostingProviders"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnHostingProviderAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		kf := getStringList(ctx, m, "keyFeatures")
		obj, _ := types.ObjectValue(txnHostingProviderAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"country": getString(m, "country"), "country_iso_code": getString(m, "countryIsoCode"),
			"city": getString(m, "city"), "cloud": getString(m, "cloud"),
			"sla": getFloat64(m, "sla"), "data_center_index": getInt64(m, "dataCenterIndex"),
			"key_features": kf, "disabled": getBool(m, "disabled"),
			"info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnHostingProviderAttrTypes}, items)
}

func populateTxnSshPrivateKeys(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "sshPrivateKeys", existing, txnSshPrivateKeyAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"private_key": getString(m, "privateKey"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnCertificates(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "certificates", existing, txnCertificateAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"certificate_pem": getString(m, "certificatePem"), "private_key_pem": getString(m, "privateKeyPem"),
			"resource_name": getString(m, "resourceName"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnVpcSubnets(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "vpcSubnets", existing, txnVpcSubnetAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"vpc_id": getString(m, "vpcId"), "ipv4_cidr": getString(m, "ipv4Cidr"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnVpcPeerings(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "vpcPeerings", existing, txnVpcPeeringAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnVpcPeeringExternalPeers(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["vpcPeeringExternalPeers"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnVpcPeeringExternalPeerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnVpcPeeringExternalPeerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"vpc_peering_id": getString(m, "vpcPeeringId"), "ssh_user": getString(m, "sshUser"),
			"ssh_port": getInt64(m, "sshPort"), "ssh_ip_v4": getString(m, "sshIpV4"),
			"private_ip_v4": getString(m, "privateIpV4"), "ip_v4_cidrs": getStringList(ctx, m, "ipV4Cidrs"),
			"ssh_private_key_id": getString(m, "sshPrivateKeyId"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnVpcPeeringExternalPeerAttrTypes}, items)
}

func populateTxnRouteTables(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "routeTables", existing, txnRouteTableAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnRouteTableRoutes(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "routeTableRoutes", existing, txnRouteTableRouteAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"route_table_id": getString(m, "routeTableId"), "destination_cidr": getString(m, "destinationCidr"),
			"target_ip": getString(m, "targetIp"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnRouteTableAttachments(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "routeTableAttachments", existing, txnRouteTableAttachmentAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"route_table_id": getString(m, "routeTableId"), "vpc_id": getString(m, "vpcId"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnImageSchedules(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "imageSchedules", existing, txnImageScheduleAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"enabled": getBool(m, "enabled"), "schedule_format": getString(m, "scheduleFormat"),
			"schedule": getString(m, "schedule"), "retention_count": getInt64(m, "retentionCount"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerTargetGroups(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerTargetGroups", existing, txnLoadbalancerTargetGroupAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerTargetGroupStaticTargets(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerTargetGroupStaticTargets", existing, txnLoadbalancerTargetGroupStaticTargetAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"target_group_id": getString(m, "targetGroupId"), "ip_or_hostname": getString(m, "ipOrHostname"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerTargetGroupServiceDiscoveryTargets(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancerTargetGroupServiceDiscoveryTargets"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"target_group_id": getString(m, "targetGroupId"), "label_selectors": getStringMap(m, "labelSelectors"),
			"info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes}, items)
}

func populateTxnLoadbalancerHttpListeners(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancerHttpListeners"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerHttpListenerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerHttpListenerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"loadbalancer_id": getString(m, "loadbalancerId"), "interface": getString(m, "interface"),
			"order": getInt64(m, "order"), "ports": getStringList(ctx, m, "ports"),
			"hosts": getStringList(ctx, m, "hosts"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerHttpListenerAttrTypes}, items)
}

func populateTxnLoadbalancerHttpsListeners(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancerHttpsListeners"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerHttpsListenerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerHttpsListenerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"loadbalancer_id": getString(m, "loadbalancerId"), "interface": getString(m, "interface"),
			"order": getInt64(m, "order"), "ports": getStringList(ctx, m, "ports"),
			"hosts": getStringList(ctx, m, "hosts"), "enable_http2_support": getBool(m, "enableHttp2Support"),
			"tls_certificate_id": getString(m, "tlsCertificateId"), "tls_protocols": getStringList(ctx, m, "tlsProtocols"),
			"tls_autogenerate_certificate": getBool(m, "tlsAutogenerateCertificate"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerHttpsListenerAttrTypes}, items)
}

func populateTxnLoadbalancerTlsListeners(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancerTlsListeners"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerTlsListenerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerTlsListenerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"loadbalancer_id": getString(m, "loadbalancerId"), "interface": getString(m, "interface"),
			"order": getInt64(m, "order"), "ports": getStringList(ctx, m, "ports"),
			"hosts": getStringList(ctx, m, "hosts"), "tls_certificate_id": getString(m, "tlsCertificateId"),
			"tls_protocols": getStringList(ctx, m, "tlsProtocols"),
			"tls_autogenerate_certificate": getBool(m, "tlsAutogenerateCertificate"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerTlsListenerAttrTypes}, items)
}

func populateTxnLoadbalancerTcpListeners(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancerTcpListeners"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerTcpListenerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerTcpListenerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"loadbalancer_id": getString(m, "loadbalancerId"), "interface": getString(m, "interface"),
			"order": getInt64(m, "order"), "ports": getStringList(ctx, m, "ports"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerTcpListenerAttrTypes}, items)
}

func populateTxnLoadbalancerUdpListeners(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancerUdpListeners"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerUdpListenerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerUdpListenerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"loadbalancer_id": getString(m, "loadbalancerId"), "interface": getString(m, "interface"),
			"order": getInt64(m, "order"), "ports": getStringList(ctx, m, "ports"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerUdpListenerAttrTypes}, items)
}

func populateTxnLoadbalancerHttpListenerRules(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerHttpListenerRules", existing, txnLoadbalancerHttpListenerRuleAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"http_listener_id": getString(m, "httpListenerId"), "order": getInt64(m, "order"),
			"match_path": getString(m, "matchPath"), "match_path_match_type": getString(m, "matchPathMatchType"),
			"action_type": getString(m, "actionType"), "action_json": getString(m, "actionJson"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerHttpsListenerRules(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerHttpsListenerRules", existing, txnLoadbalancerHttpsListenerRuleAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"https_listener_id": getString(m, "httpsListenerId"), "order": getInt64(m, "order"),
			"match_path": getString(m, "matchPath"), "match_path_match_type": getString(m, "matchPathMatchType"),
			"action_type": getString(m, "actionType"), "action_json": getString(m, "actionJson"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerTlsListenerRules(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerTlsListenerRules", existing, txnLoadbalancerTlsListenerRuleAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tls_listener_id": getString(m, "tlsListenerId"), "order": getInt64(m, "order"),
			"action_type": getString(m, "actionType"), "action_json": getString(m, "actionJson"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerTcpListenerRules(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerTcpListenerRules", existing, txnLoadbalancerTcpListenerRuleAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tcp_listener_id": getString(m, "tcpListenerId"), "order": getInt64(m, "order"),
			"action_type": getString(m, "actionType"), "action_json": getString(m, "actionJson"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnLoadbalancerUdpListenerRules(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "loadbalancerUdpListenerRules", existing, txnLoadbalancerUdpListenerRuleAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"udp_listener_id": getString(m, "udpListenerId"), "order": getInt64(m, "order"),
			"action_json": getString(m, "actionJson"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnKubernetesNodeGroups(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "kubernetesNodeGroups", existing, txnKubernetesNodeGroupAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"kubernetes_id": getString(m, "kubernetesId"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"vm_offer_id": getString(m, "vmOfferId"), "volume_offer_id": getString(m, "volumeOfferId"),
			"volume_size_gib": getInt64(m, "volumeSizeGib"), "desired_node_count": getInt64(m, "desiredNodeCount"),
			"vm_state": getString(m, "vmState"), "create_public_ipv4": getBool(m, "createPublicIpv4"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnKubernetesUserRoles(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["kubernetesUserRoles"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnKubernetesUserRoleAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnKubernetesUserRoleAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"api_groups": getStringList(ctx, m, "apiGroups"), "resources": getStringList(ctx, m, "resources"),
			"verbs": getStringList(ctx, m, "verbs"), "namespaces": getStringList(ctx, m, "namespaces"),
			"info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnKubernetesUserRoleAttrTypes}, items)
}

func populateTxnOpenVpns(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "openVpns", existing, txnOpenVpnAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"floating_ip_id": getString(m, "floatingIpId"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnPostgresqlParametersSets(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["postgreSqlParametersSets"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnPostgresqlParametersSetAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnPostgresqlParametersSetAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"parameters": getStringMap(m, "parameters"), "info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnPostgresqlParametersSetAttrTypes}, items)
}

func populateTxnSupportPlans(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "supportPlans", existing, txnSupportPlanAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "info": simpleStateInfo(m),
		}
	})
}

func populateTxnSupportTickets(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "supportTickets", existing, txnSupportTicketAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"kind": getString(m, "kind"), "severity": getString(m, "severity"), "status": getString(m, "status"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnSupportTicketComments(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["supportTicketComments"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnSupportTicketCommentAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnSupportTicketCommentAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"ticket_id": getString(m, "ticketId"), "content": getString(m, "content"),
			"attachments_ids": getStringList(ctx, m, "attachmentsIds"),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnSupportTicketCommentAttrTypes}, items)
}

func populateTxnGitlabRunners(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	return populateSimpleMap(data, "gitlabRunners", existing, txnGitlabRunnerAttrTypes, func(m map[string]interface{}) map[string]attr.Value {
		return map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"floating_ip_id": getString(m, "floatingIpId"), "vm_state": getString(m, "vmState"),
			"vm_offer_id": getString(m, "vmOfferId"), "volume_offer_id": getString(m, "volumeOfferId"),
			"volume_size_gib": getInt64(m, "volumeSizeGib"), "concurrency": getInt64(m, "concurrency"),
			"version": getString(m, "version"), "docker_options_json_string": getString(m, "dockerOptionsJsonString"),
			"info": simpleStateInfo(m),
		}
	})
}

func populateTxnOpenVpnUserSettings(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["openVpnUserSettings"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnOpenVpnUserSettingsAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnOpenVpnUserSettingsAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"allowed_ip_v4_cidrs": getStringList(ctx, m, "allowedIpV4Cidrs"),
			"allowed_ip_v6_cidrs": getStringList(ctx, m, "allowedIpV6Cidrs"),
			"denied_ip_v4_cidrs":  getStringList(ctx, m, "deniedIpV4Cidrs"),
			"denied_ip_v6_cidrs":  getStringList(ctx, m, "deniedIpV6Cidrs"),
			"allowed_domains":     getStringList(ctx, m, "allowedDomains"),
			"denied_domains":      getStringList(ctx, m, "deniedDomains"),
			"info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnOpenVpnUserSettingsAttrTypes}, items)
}

func populateTxnUsers(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["users"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnIamUserAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnIamUserAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"email": getString(m, "email"), "access_policy_ids": getStringList(ctx, m, "accessPolicyIds"),
			"info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnIamUserAttrTypes}, items)
}

func populateTxnUserTokens(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["userTokens"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnUserTokenAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnUserTokenAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"user_id": getString(m, "userId"), "send_to_email": getBool(m, "sendToEmail"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnUserTokenAttrTypes}, items)
}

func populateTxnFloatingIps(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["floatingIps"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnFloatingIpAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnFloatingIpAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"hosting_provider_id": getString(m, "hostingProviderId"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnFloatingIpAttrTypes}, items)
}

func populateTxnVpcs(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["vpcs"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnVpcAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnVpcAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"hosting_provider_id": getString(m, "hostingProviderId"), "ipv4_cidr": getString(m, "ipv4Cidr"),
			"nat_floating_ip_id": getString(m, "natFloatingIpId"),
			"security_group_ids": getStringList(ctx, m, "securityGroupIds"),
			"externally_managed": getBool(m, "externallyManaged"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnVpcAttrTypes}, items)
}

func populateTxnVpcPeeringPeers(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["vpcPeeringPeers"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnVpcPeeringPeerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnVpcPeeringPeerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"vpc_peering_id": getString(m, "vpcPeeringId"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"floating_ip_id": getString(m, "floatingIpId"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnVpcPeeringPeerAttrTypes}, items)
}

func populateTxnLoadbalancers(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["loadbalancers"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnLoadbalancerAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"floating_ip_id": getString(m, "floatingIpId"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnLoadbalancerAttrTypes}, items)
}

func populateTxnKubernetes(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["kuberneteses"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnKubernetesAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		// Build control_plane_locations list
		var cplVals []attr.Value
		if rawLocs, ok2 := m["controlPlaneLocations"].([]interface{}); ok2 {
			for _, l := range rawLocs {
				lm, ok3 := l.(map[string]interface{})
				if !ok3 { continue }
				locObj, _ := types.ObjectValue(txnControlPlaneLocationAttrTypes, map[string]attr.Value{
					"vpc_subnet_id": getString(lm, "vpcSubnetId"),
				})
				cplVals = append(cplVals, locObj)
			}
		}
		cplList, _ := types.ListValue(types.ObjectType{AttrTypes: txnControlPlaneLocationAttrTypes}, cplVals)
		obj, _ := types.ObjectValue(txnKubernetesAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "assign_public_ip_v4": getBool(m, "assignPublicIpV4"),
			"version": getString(m, "version"), "control_plane_locations": cplList, "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnKubernetesAttrTypes}, items)
}

func populateTxnKubernetesUsers(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["kubernetesUsers"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnKubernetesUserAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnKubernetesUserAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"kubernetes_id": getString(m, "kubernetesId"), "role_ids": getStringList(ctx, m, "roleIds"),
			"info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnKubernetesUserAttrTypes}, items)
}

func populateTxnPostgresqlStandalones(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["postgreSqlStandalones"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnPostgresqlStandaloneAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnPostgresqlStandaloneAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "version": getString(m, "version"), "root_password": getString(m, "rootPassword"),
			"parameters_set_id": getString(m, "parametersSetId"), "backup_retention_days": getInt64(m, "backupRetentionDays"),
			"floating_ip_id": getString(m, "floatingIpId"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"vm_state": getString(m, "vmState"), "vm_offer_id": getString(m, "vmOfferId"),
			"volume_offer_id": getString(m, "volumeOfferId"), "volume_size_gib": getInt64(m, "volumeSizeGib"),
			"info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnPostgresqlStandaloneAttrTypes}, items)
}

func populateTxnOpenVpnUsers(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["openVpnUsers"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnOpenVpnUserAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnOpenVpnUserAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"open_vpn_id": getString(m, "openVpnId"), "open_vpn_settings_ids": getStringList(ctx, m, "openVpnSettingsIds"),
			"info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnOpenVpnUserAttrTypes}, items)
}

func populateTxnBillingAccounts(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["billingAccounts"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnBillingAccountAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnBillingAccountAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"resource_name": getString(m, "resourceName"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnBillingAccountAttrTypes}, items)
}

func populateTxnQuotas(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["quotas"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnQuotaAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnQuotaAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"product": getString(m, "product"), "resource": getString(m, "resource"),
			"parameter": getString(m, "parameter"), "limit": getInt64(m, "limit"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnQuotaAttrTypes}, items)
}

func populateTxnQuotaChangeRequests(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["quotaChangeRequests"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnQuotaChangeRequestAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnQuotaChangeRequestAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"quota_id": getString(m, "quotaId"), "new_quota_limit": getInt64(m, "newQuotaLimit"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnQuotaChangeRequestAttrTypes}, items)
}

func populateTxnVictoriaMetrics(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["victoriaMetricss"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnVictoriaMetricsAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnVictoriaMetricsAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "vpc_id": getString(m, "vpcId"),
			"create_public_ipv4": getBool(m, "createPublicIpv4"), "create_public_ipv6": getBool(m, "createPublicIpv6"),
			"dns_record_name": getString(m, "dnsRecordName"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnVictoriaMetricsAttrTypes}, items)
}

func populateTxnGitlabs(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["gitlabs"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnGitlabAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnGitlabAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"tier": getString(m, "tier"), "floating_ip_id": getString(m, "floatingIpId"),
			"vpc_subnet_id": getString(m, "vpcSubnetId"), "version": getString(m, "version"),
			"root_password": getString(m, "rootPassword"), "vm_state": getString(m, "vmState"),
			"vm_offer_id": getString(m, "vmOfferId"), "volume_offer_id": getString(m, "volumeOfferId"),
			"volume_size_gib": getInt64(m, "volumeSizeGib"), "edition": getString(m, "edition"),
			"record_name": getString(m, "recordName"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnGitlabAttrTypes}, items)
}

func populateTxnSupportTicketCommentAttachments(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["supportTicketCommentAttachments"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnSupportTicketCommentAttachmentAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnSupportTicketCommentAttachmentAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"file_name": getString(m, "fileName"), "file_type": getString(m, "fileType"),
			"file_content_base64": getString(m, "fileContentBase64"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnSupportTicketCommentAttachmentAttrTypes}, items)
}

func populateTxnImages(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["images"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnImageAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		obj, _ := types.ObjectValue(txnImageAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"vm_id": getString(m, "vmId"), "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnImageAttrTypes}, items)
}

func populateTxnVms(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["vms"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnVmAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		// Build bootstrap_command list
		var bcVals []attr.Value
		if rawCmds, ok2 := m["bootstrapCommand"].([]interface{}); ok2 {
			for _, c := range rawCmds {
				cm, ok3 := c.(map[string]interface{})
				if !ok3 { continue }
				bcObj, _ := types.ObjectValue(txnBootstrapCmdAttrTypes, map[string]attr.Value{
					"command":             getString(cm, "command"),
					"success_return_code": getInt64(cm, "successReturnCode"),
					"timeout_seconds":     getInt64(cm, "timeoutSeconds"),
				})
				bcVals = append(bcVals, bcObj)
			}
		}
		bcList, _ := types.ListValue(types.ObjectType{AttrTypes: txnBootstrapCmdAttrTypes}, bcVals)
		obj, _ := types.ObjectValue(txnVmAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"vm_state": getString(m, "vmState"), "vpc_subnet_id": getString(m, "vpcSubnetId"),
			"floating_ip_id": getString(m, "floatingIpId"), "image_id": getString(m, "imageId"),
			"offer_id": getString(m, "offerId"),
			"image_boot_volume_device_index": getInt64(m, "imageBootVolumeDeviceIndex"),
			"ssh_key_ids": getStringList(ctx, m, "sshKeyIds"),
			"image_schedule_ids": getStringList(ctx, m, "imageScheduleIds"),
			"bootstrap_command": bcList, "info": simpleStateInfoObj(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnVmAttrTypes}, items)
}

func populateTxnSecurityGroups(ctx context.Context, data map[string]interface{}, existing types.Map) types.Map {
	raw, ok := data["securityGroups"].([]interface{})
	if !ok || len(raw) == 0 {
		return types.MapValueMust(types.ObjectType{AttrTypes: txnSecurityGroupAttrTypes}, map[string]attr.Value{})
	}
	items := make(map[string]attr.Value, len(raw))
	buildRuleList := func(src []interface{}) types.List {
		var ruleVals []attr.Value
		for _, r := range src {
			rm, ok2 := r.(map[string]interface{})
			if !ok2 { continue }
			ruleObj, _ := types.ObjectValue(txnSgRuleAttrTypes, map[string]attr.Value{
				"ports":       getStringList(ctx, rm, "ports"),
				"ipv4_blocks": getStringList(ctx, rm, "ipv4Blocks"),
				"ipv6_blocks": getStringList(ctx, rm, "ipv6Blocks"),
				"action":      getString(rm, "action"),
			})
			ruleVals = append(ruleVals, ruleObj)
		}
		lst, _ := types.ListValue(types.ObjectType{AttrTypes: txnSgRuleAttrTypes}, ruleVals)
		return lst
	}
	for _, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok { continue }
		id, _ := m["id"].(string)
		var ingressRaw, egressRaw []interface{}
		if ir, ok2 := m["ingress"].([]interface{}); ok2 { ingressRaw = ir }
		if er, ok2 := m["egress"].([]interface{}); ok2 { egressRaw = er }
		obj, _ := types.ObjectValue(txnSecurityGroupAttrTypes, map[string]attr.Value{
			"id": getString(m, "id"), "name": getString(m, "name"),
			"description": getString(m, "description"), "folder_id": getString(m, "folderId"),
			"delete_protection": getBool(m, "deleteProtection"), "labels": getStringMap(m, "labels"),
			"ingress": buildRuleList(ingressRaw), "egress": buildRuleList(egressRaw),
			"info": simpleStateInfo(m),
		})
		items[resolveMapKey(existing, id, m)] = obj
	}
	return types.MapValueMust(types.ObjectType{AttrTypes: txnSecurityGroupAttrTypes}, items)
}

// ---------------------------------------------------------------------------
// CRUD
// ---------------------------------------------------------------------------

func (r *TransactionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TransactionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }

	plan.ID = types.StringValue(newULID())

	// The framework marks Computed+Optional nested attrs as unknown during Create (no prior
	// state), even when the user explicitly set them in config. Read req.Config to retrieve
	// user-provided values for sub-resource IDs (and bucket_id/access_policy_ids for s3_users).
	var cfg TransactionResourceModel
	req.Config.Get(ctx, &cfg) // diagnostics intentionally ignored — we only need known values
	plan.SshKeys = fillUnknownIDsFromConfig(ctx, plan.SshKeys, cfg.SshKeys, txnSshKeyAttrTypes, nil)
	plan.S3Buckets = fillUnknownIDsFromConfig(ctx, plan.S3Buckets, cfg.S3Buckets, txnS3BucketAttrTypes, nil)
	plan.S3UserAccessPolicies = fillUnknownIDsFromConfig(ctx, plan.S3UserAccessPolicies, cfg.S3UserAccessPolicies, txnS3PolicyAttrTypes, nil)
	plan.S3Users = fillUnknownIDsFromConfig(ctx, plan.S3Users, cfg.S3Users, txnS3UserAttrTypes, []string{"bucket_id", "access_policy_ids"})
	plan.Folders = fillUnknownIDsFromConfig(ctx, plan.Folders, cfg.Folders, txnFolderAttrTypes, nil)
	plan.Volumes = fillUnknownIDsFromConfig(ctx, plan.Volumes, cfg.Volumes, txnVolumeAttrTypes, nil)
	plan.VolumeAttachments = fillUnknownIDsFromConfig(ctx, plan.VolumeAttachments, cfg.VolumeAttachments, txnVolumeAttachmentAttrTypes, nil)
	plan.AccessPolicies = fillUnknownIDsFromConfig(ctx, plan.AccessPolicies, cfg.AccessPolicies, txnAccessPolicyAttrTypes, nil)
	plan.HostingProviders = fillUnknownIDsFromConfig(ctx, plan.HostingProviders, cfg.HostingProviders, txnHostingProviderAttrTypes, nil)
	plan.SshPrivateKeys = fillUnknownIDsFromConfig(ctx, plan.SshPrivateKeys, cfg.SshPrivateKeys, txnSshPrivateKeyAttrTypes, nil)
	plan.Certificates = fillUnknownIDsFromConfig(ctx, plan.Certificates, cfg.Certificates, txnCertificateAttrTypes, nil)
	plan.VpcSubnets = fillUnknownIDsFromConfig(ctx, plan.VpcSubnets, cfg.VpcSubnets, txnVpcSubnetAttrTypes, nil)
	plan.VpcPeerings = fillUnknownIDsFromConfig(ctx, plan.VpcPeerings, cfg.VpcPeerings, txnVpcPeeringAttrTypes, nil)
	plan.VpcPeeringExternalPeers = fillUnknownIDsFromConfig(ctx, plan.VpcPeeringExternalPeers, cfg.VpcPeeringExternalPeers, txnVpcPeeringExternalPeerAttrTypes, nil)
	plan.RouteTables = fillUnknownIDsFromConfig(ctx, plan.RouteTables, cfg.RouteTables, txnRouteTableAttrTypes, nil)
	plan.RouteTableRoutes = fillUnknownIDsFromConfig(ctx, plan.RouteTableRoutes, cfg.RouteTableRoutes, txnRouteTableRouteAttrTypes, nil)
	plan.RouteTableAttachments = fillUnknownIDsFromConfig(ctx, plan.RouteTableAttachments, cfg.RouteTableAttachments, txnRouteTableAttachmentAttrTypes, nil)
	plan.ImageSchedules = fillUnknownIDsFromConfig(ctx, plan.ImageSchedules, cfg.ImageSchedules, txnImageScheduleAttrTypes, nil)
	plan.LoadbalancerTargetGroups = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTargetGroups, cfg.LoadbalancerTargetGroups, txnLoadbalancerTargetGroupAttrTypes, nil)
	plan.LoadbalancerTargetGroupStaticTargets = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTargetGroupStaticTargets, cfg.LoadbalancerTargetGroupStaticTargets, txnLoadbalancerTargetGroupStaticTargetAttrTypes, nil)
	plan.LoadbalancerTargetGroupServiceDiscoveryTargets = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTargetGroupServiceDiscoveryTargets, cfg.LoadbalancerTargetGroupServiceDiscoveryTargets, txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes, nil)
	plan.LoadbalancerHttpListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpListeners, cfg.LoadbalancerHttpListeners, txnLoadbalancerHttpListenerAttrTypes, nil)
	plan.LoadbalancerHttpsListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpsListeners, cfg.LoadbalancerHttpsListeners, txnLoadbalancerHttpsListenerAttrTypes, nil)
	plan.LoadbalancerTlsListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTlsListeners, cfg.LoadbalancerTlsListeners, txnLoadbalancerTlsListenerAttrTypes, nil)
	plan.LoadbalancerTcpListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTcpListeners, cfg.LoadbalancerTcpListeners, txnLoadbalancerTcpListenerAttrTypes, nil)
	plan.LoadbalancerUdpListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerUdpListeners, cfg.LoadbalancerUdpListeners, txnLoadbalancerUdpListenerAttrTypes, nil)
	plan.LoadbalancerHttpListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpListenerRules, cfg.LoadbalancerHttpListenerRules, txnLoadbalancerHttpListenerRuleAttrTypes, nil)
	plan.LoadbalancerHttpsListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpsListenerRules, cfg.LoadbalancerHttpsListenerRules, txnLoadbalancerHttpsListenerRuleAttrTypes, nil)
	plan.LoadbalancerTlsListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTlsListenerRules, cfg.LoadbalancerTlsListenerRules, txnLoadbalancerTlsListenerRuleAttrTypes, nil)
	plan.LoadbalancerTcpListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTcpListenerRules, cfg.LoadbalancerTcpListenerRules, txnLoadbalancerTcpListenerRuleAttrTypes, nil)
	plan.LoadbalancerUdpListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerUdpListenerRules, cfg.LoadbalancerUdpListenerRules, txnLoadbalancerUdpListenerRuleAttrTypes, nil)
	plan.KubernetesNodeGroups = fillUnknownIDsFromConfig(ctx, plan.KubernetesNodeGroups, cfg.KubernetesNodeGroups, txnKubernetesNodeGroupAttrTypes, nil)
	plan.KubernetesUserRoles = fillUnknownIDsFromConfig(ctx, plan.KubernetesUserRoles, cfg.KubernetesUserRoles, txnKubernetesUserRoleAttrTypes, nil)
	plan.OpenVpns = fillUnknownIDsFromConfig(ctx, plan.OpenVpns, cfg.OpenVpns, txnOpenVpnAttrTypes, nil)
	plan.PostgresqlParametersSets = fillUnknownIDsFromConfig(ctx, plan.PostgresqlParametersSets, cfg.PostgresqlParametersSets, txnPostgresqlParametersSetAttrTypes, nil)
	plan.SupportPlans = fillUnknownIDsFromConfig(ctx, plan.SupportPlans, cfg.SupportPlans, txnSupportPlanAttrTypes, nil)
	plan.SupportTickets = fillUnknownIDsFromConfig(ctx, plan.SupportTickets, cfg.SupportTickets, txnSupportTicketAttrTypes, nil)
	plan.SupportTicketComments = fillUnknownIDsFromConfig(ctx, plan.SupportTicketComments, cfg.SupportTicketComments, txnSupportTicketCommentAttrTypes, nil)
	plan.GitlabRunners = fillUnknownIDsFromConfig(ctx, plan.GitlabRunners, cfg.GitlabRunners, txnGitlabRunnerAttrTypes, nil)
	plan.OpenVpnUserSettings = fillUnknownIDsFromConfig(ctx, plan.OpenVpnUserSettings, cfg.OpenVpnUserSettings, txnOpenVpnUserSettingsAttrTypes, nil)
	plan.Users = fillUnknownIDsFromConfig(ctx, plan.Users, cfg.Users, txnIamUserAttrTypes, nil)
	plan.UserTokens = fillUnknownIDsFromConfig(ctx, plan.UserTokens, cfg.UserTokens, txnUserTokenAttrTypes, nil)
	plan.FloatingIps = fillUnknownIDsFromConfig(ctx, plan.FloatingIps, cfg.FloatingIps, txnFloatingIpAttrTypes, nil)
	plan.Vpcs = fillUnknownIDsFromConfig(ctx, plan.Vpcs, cfg.Vpcs, txnVpcAttrTypes, nil)
	plan.VpcPeeringPeers = fillUnknownIDsFromConfig(ctx, plan.VpcPeeringPeers, cfg.VpcPeeringPeers, txnVpcPeeringPeerAttrTypes, nil)
	plan.Loadbalancers = fillUnknownIDsFromConfig(ctx, plan.Loadbalancers, cfg.Loadbalancers, txnLoadbalancerAttrTypes, nil)
	plan.Kubernetes = fillUnknownIDsFromConfig(ctx, plan.Kubernetes, cfg.Kubernetes, txnKubernetesAttrTypes, nil)
	plan.KubernetesUsers = fillUnknownIDsFromConfig(ctx, plan.KubernetesUsers, cfg.KubernetesUsers, txnKubernetesUserAttrTypes, nil)
	plan.PostgresqlStandalones = fillUnknownIDsFromConfig(ctx, plan.PostgresqlStandalones, cfg.PostgresqlStandalones, txnPostgresqlStandaloneAttrTypes, nil)
	plan.OpenVpnUsers = fillUnknownIDsFromConfig(ctx, plan.OpenVpnUsers, cfg.OpenVpnUsers, txnOpenVpnUserAttrTypes, nil)
	plan.BillingAccounts = fillUnknownIDsFromConfig(ctx, plan.BillingAccounts, cfg.BillingAccounts, txnBillingAccountAttrTypes, nil)
	plan.Quotas = fillUnknownIDsFromConfig(ctx, plan.Quotas, cfg.Quotas, txnQuotaAttrTypes, nil)
	plan.QuotaChangeRequests = fillUnknownIDsFromConfig(ctx, plan.QuotaChangeRequests, cfg.QuotaChangeRequests, txnQuotaChangeRequestAttrTypes, nil)
	plan.VictoriaMetrics = fillUnknownIDsFromConfig(ctx, plan.VictoriaMetrics, cfg.VictoriaMetrics, txnVictoriaMetricsAttrTypes, nil)
	plan.Gitlabs = fillUnknownIDsFromConfig(ctx, plan.Gitlabs, cfg.Gitlabs, txnGitlabAttrTypes, nil)
	plan.SupportTicketCommentAttachments = fillUnknownIDsFromConfig(ctx, plan.SupportTicketCommentAttachments, cfg.SupportTicketCommentAttachments, txnSupportTicketCommentAttachmentAttrTypes, nil)
	plan.Images = fillUnknownIDsFromConfig(ctx, plan.Images, cfg.Images, txnImageAttrTypes, nil)
	plan.Vms = fillUnknownIDsFromConfig(ctx, plan.Vms, cfg.Vms, txnVmAttrTypes, nil)
	plan.SecurityGroups = fillUnknownIDsFromConfig(ctx, plan.SecurityGroups, cfg.SecurityGroups, txnSecurityGroupAttrTypes, nil)

	// Pre-assign ULIDs for entries whose id was not provided in config so that
	// populateState can find the user-chosen map key after the API round-trip.
	plan.Folders = assignSubIDsMap(plan.Folders, txnFolderAttrTypes)
	plan.SshKeys = assignSubIDsMap(plan.SshKeys, txnSshKeyAttrTypes)
	plan.S3Buckets = assignSubIDsMap(plan.S3Buckets, txnS3BucketAttrTypes)
	plan.S3UserAccessPolicies = assignSubIDsMap(plan.S3UserAccessPolicies, txnS3PolicyAttrTypes)
	plan.S3Users = assignSubIDsMap(plan.S3Users, txnS3UserAttrTypes)
	plan.Volumes = assignSubIDsMap(plan.Volumes, txnVolumeAttrTypes)
	plan.VolumeAttachments = assignSubIDsMap(plan.VolumeAttachments, txnVolumeAttachmentAttrTypes)
	plan.AccessPolicies = assignSubIDsMap(plan.AccessPolicies, txnAccessPolicyAttrTypes)
	plan.HostingProviders = assignSubIDsMap(plan.HostingProviders, txnHostingProviderAttrTypes)
	plan.SshPrivateKeys = assignSubIDsMap(plan.SshPrivateKeys, txnSshPrivateKeyAttrTypes)
	plan.Certificates = assignSubIDsMap(plan.Certificates, txnCertificateAttrTypes)
	plan.VpcSubnets = assignSubIDsMap(plan.VpcSubnets, txnVpcSubnetAttrTypes)
	plan.VpcPeerings = assignSubIDsMap(plan.VpcPeerings, txnVpcPeeringAttrTypes)
	plan.VpcPeeringExternalPeers = assignSubIDsMap(plan.VpcPeeringExternalPeers, txnVpcPeeringExternalPeerAttrTypes)
	plan.RouteTables = assignSubIDsMap(plan.RouteTables, txnRouteTableAttrTypes)
	plan.RouteTableRoutes = assignSubIDsMap(plan.RouteTableRoutes, txnRouteTableRouteAttrTypes)
	plan.RouteTableAttachments = assignSubIDsMap(plan.RouteTableAttachments, txnRouteTableAttachmentAttrTypes)
	plan.ImageSchedules = assignSubIDsMap(plan.ImageSchedules, txnImageScheduleAttrTypes)
	plan.LoadbalancerTargetGroups = assignSubIDsMap(plan.LoadbalancerTargetGroups, txnLoadbalancerTargetGroupAttrTypes)
	plan.LoadbalancerTargetGroupStaticTargets = assignSubIDsMap(plan.LoadbalancerTargetGroupStaticTargets, txnLoadbalancerTargetGroupStaticTargetAttrTypes)
	plan.LoadbalancerTargetGroupServiceDiscoveryTargets = assignSubIDsMap(plan.LoadbalancerTargetGroupServiceDiscoveryTargets, txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes)
	plan.LoadbalancerHttpListeners = assignSubIDsMap(plan.LoadbalancerHttpListeners, txnLoadbalancerHttpListenerAttrTypes)
	plan.LoadbalancerHttpsListeners = assignSubIDsMap(plan.LoadbalancerHttpsListeners, txnLoadbalancerHttpsListenerAttrTypes)
	plan.LoadbalancerTlsListeners = assignSubIDsMap(plan.LoadbalancerTlsListeners, txnLoadbalancerTlsListenerAttrTypes)
	plan.LoadbalancerTcpListeners = assignSubIDsMap(plan.LoadbalancerTcpListeners, txnLoadbalancerTcpListenerAttrTypes)
	plan.LoadbalancerUdpListeners = assignSubIDsMap(plan.LoadbalancerUdpListeners, txnLoadbalancerUdpListenerAttrTypes)
	plan.LoadbalancerHttpListenerRules = assignSubIDsMap(plan.LoadbalancerHttpListenerRules, txnLoadbalancerHttpListenerRuleAttrTypes)
	plan.LoadbalancerHttpsListenerRules = assignSubIDsMap(plan.LoadbalancerHttpsListenerRules, txnLoadbalancerHttpsListenerRuleAttrTypes)
	plan.LoadbalancerTlsListenerRules = assignSubIDsMap(plan.LoadbalancerTlsListenerRules, txnLoadbalancerTlsListenerRuleAttrTypes)
	plan.LoadbalancerTcpListenerRules = assignSubIDsMap(plan.LoadbalancerTcpListenerRules, txnLoadbalancerTcpListenerRuleAttrTypes)
	plan.LoadbalancerUdpListenerRules = assignSubIDsMap(plan.LoadbalancerUdpListenerRules, txnLoadbalancerUdpListenerRuleAttrTypes)
	plan.KubernetesNodeGroups = assignSubIDsMap(plan.KubernetesNodeGroups, txnKubernetesNodeGroupAttrTypes)
	plan.KubernetesUserRoles = assignSubIDsMap(plan.KubernetesUserRoles, txnKubernetesUserRoleAttrTypes)
	plan.OpenVpns = assignSubIDsMap(plan.OpenVpns, txnOpenVpnAttrTypes)
	plan.PostgresqlParametersSets = assignSubIDsMap(plan.PostgresqlParametersSets, txnPostgresqlParametersSetAttrTypes)
	plan.SupportPlans = assignSubIDsMap(plan.SupportPlans, txnSupportPlanAttrTypes)
	plan.SupportTickets = assignSubIDsMap(plan.SupportTickets, txnSupportTicketAttrTypes)
	plan.SupportTicketComments = assignSubIDsMap(plan.SupportTicketComments, txnSupportTicketCommentAttrTypes)
	plan.GitlabRunners = assignSubIDsMap(plan.GitlabRunners, txnGitlabRunnerAttrTypes)
	plan.OpenVpnUserSettings = assignSubIDsMap(plan.OpenVpnUserSettings, txnOpenVpnUserSettingsAttrTypes)
	plan.Users = assignSubIDsMap(plan.Users, txnIamUserAttrTypes)
	plan.UserTokens = assignSubIDsMap(plan.UserTokens, txnUserTokenAttrTypes)
	plan.FloatingIps = assignSubIDsMap(plan.FloatingIps, txnFloatingIpAttrTypes)
	plan.Vpcs = assignSubIDsMap(plan.Vpcs, txnVpcAttrTypes)
	plan.VpcPeeringPeers = assignSubIDsMap(plan.VpcPeeringPeers, txnVpcPeeringPeerAttrTypes)
	plan.Loadbalancers = assignSubIDsMap(plan.Loadbalancers, txnLoadbalancerAttrTypes)
	plan.Kubernetes = assignSubIDsMap(plan.Kubernetes, txnKubernetesAttrTypes)
	plan.KubernetesUsers = assignSubIDsMap(plan.KubernetesUsers, txnKubernetesUserAttrTypes)
	plan.PostgresqlStandalones = assignSubIDsMap(plan.PostgresqlStandalones, txnPostgresqlStandaloneAttrTypes)
	plan.OpenVpnUsers = assignSubIDsMap(plan.OpenVpnUsers, txnOpenVpnUserAttrTypes)
	plan.BillingAccounts = assignSubIDsMap(plan.BillingAccounts, txnBillingAccountAttrTypes)
	plan.Quotas = assignSubIDsMap(plan.Quotas, txnQuotaAttrTypes)
	plan.QuotaChangeRequests = assignSubIDsMap(plan.QuotaChangeRequests, txnQuotaChangeRequestAttrTypes)
	plan.VictoriaMetrics = assignSubIDsMap(plan.VictoriaMetrics, txnVictoriaMetricsAttrTypes)
	plan.Gitlabs = assignSubIDsMap(plan.Gitlabs, txnGitlabAttrTypes)
	plan.SupportTicketCommentAttachments = assignSubIDsMap(plan.SupportTicketCommentAttachments, txnSupportTicketCommentAttachmentAttrTypes)
	plan.Images = assignSubIDsMap(plan.Images, txnImageAttrTypes)
	plan.Vms = assignSubIDsMap(plan.Vms, txnVmAttrTypes)
	plan.SecurityGroups = assignSubIDsMap(plan.SecurityGroups, txnSecurityGroupAttrTypes)

	hasUsers := !plan.S3Users.IsNull() && !plan.S3Users.IsUnknown() && len(plan.S3Users.Elements()) > 0

	// Phase 1: create transaction with folders + buckets + policies (no s3Users).
	// s3Users require bucketId as an API-assigned ULID which isn't known until phase 1 completes.
	//
	// Retry loop: if 422 ResourceIsScheduling is returned, a sub-resource from a previous
	// failed attempt is still being provisioned. Wait for it to stabilize, then retry.
	body := buildTxnBody(ctx, plan, true, true, true, false)
	var modResp *client.ModificationResponse
	var lastPutErr error
	for attempt := 0; attempt < 50; attempt++ {
		var putErr error
		modResp, putErr = r.client.Put(ctx, "/api/v1/transaction", body)
		if putErr == nil { break }
		lastPutErr = putErr
		blockingID, blockingType := parseSchedulingConflict(putErr.Error())
		if blockingID == "" {
			resp.Diagnostics.AddError("Create Error", putErr.Error()); return
		}
		// Wait for the conflicting resource to leave Scheduling/Reconciling state.
		if waitErr := waitForResourceStable(ctx, r.client, blockingType, blockingID); waitErr != nil {
			resp.Diagnostics.AddError("Create Error (scheduling wait)", fmt.Sprintf("waiting for %s/%s: %v", blockingType, blockingID, waitErr)); return
		}
		// The API has a dual-type quirk: a single ULID may exist as both an s3-bucket and
		// an s3-user-access-policy. Delete from ALL resource types so both are cleared.
		// Also delete regardless of state (not just schedulingfailed): a "stable" orphan from
		// a prior failed run still blocks Phase 1 internally in the API scheduler.
		for _, rt := range []string{"s3-bucket", "s3-user-access-policy", "folder", "transaction"} {
			deleteAnyExistingResource(ctx, r.client, rt, blockingID)
		}
		// Delay to let API-internal state propagate before the next PUT.
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Create Error", "context cancelled during retry"); return
		case <-time.After(10 * time.Second):
		}
	}
	if modResp == nil {
		errDetail := "phase 1 failed after retries"
		if lastPutErr != nil {
			errDetail = fmt.Sprintf("phase 1 failed after retries; last error: %s", lastPutErr.Error())
		}
		resp.Diagnostics.AddError("Create Error", errDetail); return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error()); return
	}

	txnID := modResp.ResourceId
	if txnID == "" { txnID = plan.ID.ValueString() }

	if !hasUsers {
		// No users — single phase is sufficient.
		apiData, err := r.client.Get(ctx, "/api/v1/transaction", txnID)
		if err != nil { resp.Diagnostics.AddError("Read After Create Error", err.Error()); return }
		if err := r.populateState(ctx, apiData, &plan); err != nil {
			resp.Diagnostics.AddError("State Population Error", err.Error()); return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...); return
	}

	// Wait for all phase-1 sub-resources (buckets, policies) to leave Scheduling state.
	// The transaction request completes before sub-resources finish provisioning; phase 2
	// PUT will fail with 422 if any sub-resource is still Scheduling.
	if err := waitForSubResourcesStable(ctx, r.client, txnID); err != nil {
		cleanupTransaction(ctx, r.client, txnID)
		resp.Diagnostics.AddError("Phase1 Sub-Resource Wait Error", err.Error()); return
	}

	// Phase 2: read back ULIDs of created buckets/policies and wire them into s3Users.
	apiData, err := r.client.Get(ctx, "/api/v1/transaction", txnID)
	if err != nil { resp.Diagnostics.AddError("Read After Phase1 Error", err.Error()); return }

	// populateState stores API-assigned bucket/policy ULIDs in plan.S3Buckets/S3UserAccessPolicies.
	// For s3Users: phase 1 response has no users yet, so populateTxnS3Users returns plan.S3Users unchanged.
	if err := r.populateState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error (phase1)", err.Error()); return
	}

	// Wire s3_users to use the actual Phase 1 bucket/policy IDs. Users may have unknown/null
	// bucket_id/access_policy_ids when IDs are not pre-specified in config.
	plan.S3Users = autoWireUsers(ctx, plan.S3Users, plan.S3Buckets, plan.S3UserAccessPolicies)

	// Phase 2 PUT: include all sub-resources with wired-up user bucket/policy references.
	body2 := buildTxnBody(ctx, plan, true, true, true, true)
	modResp2, err := r.client.Put(ctx, "/api/v1/transaction", body2)
	if err != nil {
		cleanupTransaction(ctx, r.client, txnID)
		resp.Diagnostics.AddError("Create Phase2 Error", err.Error()); return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp2.RequestId); err != nil {
		cleanupTransaction(ctx, r.client, txnID)
		resp.Diagnostics.AddError("Create Phase2 Poll Error", err.Error()); return
	}

	apiData2, err := r.client.Get(ctx, "/api/v1/transaction", txnID)
	if err != nil { resp.Diagnostics.AddError("Read After Phase2 Error", err.Error()); return }
	if err := r.populateState(ctx, apiData2, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error (phase2)", err.Error()); return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TransactionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state TransactionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }

	apiData, err := r.client.Get(ctx, "/api/v1/transaction", state.ID.ValueString())
	if err != nil { resp.Diagnostics.AddError("Read Error", err.Error()); return }
	if apiData == nil { resp.State.RemoveResource(ctx); return }
	if err := r.populateState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error()); return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *TransactionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TransactionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() { return }
	var state TransactionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }
	plan.ID = state.ID

	var cfgU TransactionResourceModel
	req.Config.Get(ctx, &cfgU)
	plan.SshKeys = fillUnknownIDsFromConfig(ctx, plan.SshKeys, cfgU.SshKeys, txnSshKeyAttrTypes, nil)
	plan.S3Buckets = fillUnknownIDsFromConfig(ctx, plan.S3Buckets, cfgU.S3Buckets, txnS3BucketAttrTypes, nil)
	plan.S3UserAccessPolicies = fillUnknownIDsFromConfig(ctx, plan.S3UserAccessPolicies, cfgU.S3UserAccessPolicies, txnS3PolicyAttrTypes, nil)
	plan.S3Users = fillUnknownIDsFromConfig(ctx, plan.S3Users, cfgU.S3Users, txnS3UserAttrTypes, []string{"bucket_id", "access_policy_ids"})
	plan.Folders = fillUnknownIDsFromConfig(ctx, plan.Folders, cfgU.Folders, txnFolderAttrTypes, nil)
	plan.Volumes = fillUnknownIDsFromConfig(ctx, plan.Volumes, cfgU.Volumes, txnVolumeAttrTypes, nil)
	plan.VolumeAttachments = fillUnknownIDsFromConfig(ctx, plan.VolumeAttachments, cfgU.VolumeAttachments, txnVolumeAttachmentAttrTypes, nil)
	plan.AccessPolicies = fillUnknownIDsFromConfig(ctx, plan.AccessPolicies, cfgU.AccessPolicies, txnAccessPolicyAttrTypes, nil)
	plan.HostingProviders = fillUnknownIDsFromConfig(ctx, plan.HostingProviders, cfgU.HostingProviders, txnHostingProviderAttrTypes, nil)
	plan.SshPrivateKeys = fillUnknownIDsFromConfig(ctx, plan.SshPrivateKeys, cfgU.SshPrivateKeys, txnSshPrivateKeyAttrTypes, nil)
	plan.Certificates = fillUnknownIDsFromConfig(ctx, plan.Certificates, cfgU.Certificates, txnCertificateAttrTypes, nil)
	plan.VpcSubnets = fillUnknownIDsFromConfig(ctx, plan.VpcSubnets, cfgU.VpcSubnets, txnVpcSubnetAttrTypes, nil)
	plan.VpcPeerings = fillUnknownIDsFromConfig(ctx, plan.VpcPeerings, cfgU.VpcPeerings, txnVpcPeeringAttrTypes, nil)
	plan.VpcPeeringExternalPeers = fillUnknownIDsFromConfig(ctx, plan.VpcPeeringExternalPeers, cfgU.VpcPeeringExternalPeers, txnVpcPeeringExternalPeerAttrTypes, nil)
	plan.RouteTables = fillUnknownIDsFromConfig(ctx, plan.RouteTables, cfgU.RouteTables, txnRouteTableAttrTypes, nil)
	plan.RouteTableRoutes = fillUnknownIDsFromConfig(ctx, plan.RouteTableRoutes, cfgU.RouteTableRoutes, txnRouteTableRouteAttrTypes, nil)
	plan.RouteTableAttachments = fillUnknownIDsFromConfig(ctx, plan.RouteTableAttachments, cfgU.RouteTableAttachments, txnRouteTableAttachmentAttrTypes, nil)
	plan.ImageSchedules = fillUnknownIDsFromConfig(ctx, plan.ImageSchedules, cfgU.ImageSchedules, txnImageScheduleAttrTypes, nil)
	plan.LoadbalancerTargetGroups = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTargetGroups, cfgU.LoadbalancerTargetGroups, txnLoadbalancerTargetGroupAttrTypes, nil)
	plan.LoadbalancerTargetGroupStaticTargets = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTargetGroupStaticTargets, cfgU.LoadbalancerTargetGroupStaticTargets, txnLoadbalancerTargetGroupStaticTargetAttrTypes, nil)
	plan.LoadbalancerTargetGroupServiceDiscoveryTargets = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTargetGroupServiceDiscoveryTargets, cfgU.LoadbalancerTargetGroupServiceDiscoveryTargets, txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes, nil)
	plan.LoadbalancerHttpListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpListeners, cfgU.LoadbalancerHttpListeners, txnLoadbalancerHttpListenerAttrTypes, nil)
	plan.LoadbalancerHttpsListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpsListeners, cfgU.LoadbalancerHttpsListeners, txnLoadbalancerHttpsListenerAttrTypes, nil)
	plan.LoadbalancerTlsListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTlsListeners, cfgU.LoadbalancerTlsListeners, txnLoadbalancerTlsListenerAttrTypes, nil)
	plan.LoadbalancerTcpListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTcpListeners, cfgU.LoadbalancerTcpListeners, txnLoadbalancerTcpListenerAttrTypes, nil)
	plan.LoadbalancerUdpListeners = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerUdpListeners, cfgU.LoadbalancerUdpListeners, txnLoadbalancerUdpListenerAttrTypes, nil)
	plan.LoadbalancerHttpListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpListenerRules, cfgU.LoadbalancerHttpListenerRules, txnLoadbalancerHttpListenerRuleAttrTypes, nil)
	plan.LoadbalancerHttpsListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerHttpsListenerRules, cfgU.LoadbalancerHttpsListenerRules, txnLoadbalancerHttpsListenerRuleAttrTypes, nil)
	plan.LoadbalancerTlsListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTlsListenerRules, cfgU.LoadbalancerTlsListenerRules, txnLoadbalancerTlsListenerRuleAttrTypes, nil)
	plan.LoadbalancerTcpListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerTcpListenerRules, cfgU.LoadbalancerTcpListenerRules, txnLoadbalancerTcpListenerRuleAttrTypes, nil)
	plan.LoadbalancerUdpListenerRules = fillUnknownIDsFromConfig(ctx, plan.LoadbalancerUdpListenerRules, cfgU.LoadbalancerUdpListenerRules, txnLoadbalancerUdpListenerRuleAttrTypes, nil)
	plan.KubernetesNodeGroups = fillUnknownIDsFromConfig(ctx, plan.KubernetesNodeGroups, cfgU.KubernetesNodeGroups, txnKubernetesNodeGroupAttrTypes, nil)
	plan.KubernetesUserRoles = fillUnknownIDsFromConfig(ctx, plan.KubernetesUserRoles, cfgU.KubernetesUserRoles, txnKubernetesUserRoleAttrTypes, nil)
	plan.OpenVpns = fillUnknownIDsFromConfig(ctx, plan.OpenVpns, cfgU.OpenVpns, txnOpenVpnAttrTypes, nil)
	plan.PostgresqlParametersSets = fillUnknownIDsFromConfig(ctx, plan.PostgresqlParametersSets, cfgU.PostgresqlParametersSets, txnPostgresqlParametersSetAttrTypes, nil)
	plan.SupportPlans = fillUnknownIDsFromConfig(ctx, plan.SupportPlans, cfgU.SupportPlans, txnSupportPlanAttrTypes, nil)
	plan.SupportTickets = fillUnknownIDsFromConfig(ctx, plan.SupportTickets, cfgU.SupportTickets, txnSupportTicketAttrTypes, nil)
	plan.SupportTicketComments = fillUnknownIDsFromConfig(ctx, plan.SupportTicketComments, cfgU.SupportTicketComments, txnSupportTicketCommentAttrTypes, nil)
	plan.GitlabRunners = fillUnknownIDsFromConfig(ctx, plan.GitlabRunners, cfgU.GitlabRunners, txnGitlabRunnerAttrTypes, nil)
	plan.OpenVpnUserSettings = fillUnknownIDsFromConfig(ctx, plan.OpenVpnUserSettings, cfgU.OpenVpnUserSettings, txnOpenVpnUserSettingsAttrTypes, nil)
	plan.Users = fillUnknownIDsFromConfig(ctx, plan.Users, cfgU.Users, txnIamUserAttrTypes, nil)
	plan.UserTokens = fillUnknownIDsFromConfig(ctx, plan.UserTokens, cfgU.UserTokens, txnUserTokenAttrTypes, nil)
	plan.FloatingIps = fillUnknownIDsFromConfig(ctx, plan.FloatingIps, cfgU.FloatingIps, txnFloatingIpAttrTypes, nil)
	plan.Vpcs = fillUnknownIDsFromConfig(ctx, plan.Vpcs, cfgU.Vpcs, txnVpcAttrTypes, nil)
	plan.VpcPeeringPeers = fillUnknownIDsFromConfig(ctx, plan.VpcPeeringPeers, cfgU.VpcPeeringPeers, txnVpcPeeringPeerAttrTypes, nil)
	plan.Loadbalancers = fillUnknownIDsFromConfig(ctx, plan.Loadbalancers, cfgU.Loadbalancers, txnLoadbalancerAttrTypes, nil)
	plan.Kubernetes = fillUnknownIDsFromConfig(ctx, plan.Kubernetes, cfgU.Kubernetes, txnKubernetesAttrTypes, nil)
	plan.KubernetesUsers = fillUnknownIDsFromConfig(ctx, plan.KubernetesUsers, cfgU.KubernetesUsers, txnKubernetesUserAttrTypes, nil)
	plan.PostgresqlStandalones = fillUnknownIDsFromConfig(ctx, plan.PostgresqlStandalones, cfgU.PostgresqlStandalones, txnPostgresqlStandaloneAttrTypes, nil)
	plan.OpenVpnUsers = fillUnknownIDsFromConfig(ctx, plan.OpenVpnUsers, cfgU.OpenVpnUsers, txnOpenVpnUserAttrTypes, nil)
	plan.BillingAccounts = fillUnknownIDsFromConfig(ctx, plan.BillingAccounts, cfgU.BillingAccounts, txnBillingAccountAttrTypes, nil)
	plan.Quotas = fillUnknownIDsFromConfig(ctx, plan.Quotas, cfgU.Quotas, txnQuotaAttrTypes, nil)
	plan.QuotaChangeRequests = fillUnknownIDsFromConfig(ctx, plan.QuotaChangeRequests, cfgU.QuotaChangeRequests, txnQuotaChangeRequestAttrTypes, nil)
	plan.VictoriaMetrics = fillUnknownIDsFromConfig(ctx, plan.VictoriaMetrics, cfgU.VictoriaMetrics, txnVictoriaMetricsAttrTypes, nil)
	plan.Gitlabs = fillUnknownIDsFromConfig(ctx, plan.Gitlabs, cfgU.Gitlabs, txnGitlabAttrTypes, nil)
	plan.SupportTicketCommentAttachments = fillUnknownIDsFromConfig(ctx, plan.SupportTicketCommentAttachments, cfgU.SupportTicketCommentAttachments, txnSupportTicketCommentAttachmentAttrTypes, nil)
	plan.Images = fillUnknownIDsFromConfig(ctx, plan.Images, cfgU.Images, txnImageAttrTypes, nil)
	plan.Vms = fillUnknownIDsFromConfig(ctx, plan.Vms, cfgU.Vms, txnVmAttrTypes, nil)
	plan.SecurityGroups = fillUnknownIDsFromConfig(ctx, plan.SecurityGroups, cfgU.SecurityGroups, txnSecurityGroupAttrTypes, nil)

	plan.Folders = assignSubIDsMap(plan.Folders, txnFolderAttrTypes)
	plan.SshKeys = assignSubIDsMap(plan.SshKeys, txnSshKeyAttrTypes)
	plan.S3Buckets = assignSubIDsMap(plan.S3Buckets, txnS3BucketAttrTypes)
	plan.S3UserAccessPolicies = assignSubIDsMap(plan.S3UserAccessPolicies, txnS3PolicyAttrTypes)
	plan.S3Users = assignSubIDsMap(plan.S3Users, txnS3UserAttrTypes)
	plan.Volumes = assignSubIDsMap(plan.Volumes, txnVolumeAttrTypes)
	plan.VolumeAttachments = assignSubIDsMap(plan.VolumeAttachments, txnVolumeAttachmentAttrTypes)
	plan.AccessPolicies = assignSubIDsMap(plan.AccessPolicies, txnAccessPolicyAttrTypes)
	plan.HostingProviders = assignSubIDsMap(plan.HostingProviders, txnHostingProviderAttrTypes)
	plan.SshPrivateKeys = assignSubIDsMap(plan.SshPrivateKeys, txnSshPrivateKeyAttrTypes)
	plan.Certificates = assignSubIDsMap(plan.Certificates, txnCertificateAttrTypes)
	plan.VpcSubnets = assignSubIDsMap(plan.VpcSubnets, txnVpcSubnetAttrTypes)
	plan.VpcPeerings = assignSubIDsMap(plan.VpcPeerings, txnVpcPeeringAttrTypes)
	plan.VpcPeeringExternalPeers = assignSubIDsMap(plan.VpcPeeringExternalPeers, txnVpcPeeringExternalPeerAttrTypes)
	plan.RouteTables = assignSubIDsMap(plan.RouteTables, txnRouteTableAttrTypes)
	plan.RouteTableRoutes = assignSubIDsMap(plan.RouteTableRoutes, txnRouteTableRouteAttrTypes)
	plan.RouteTableAttachments = assignSubIDsMap(plan.RouteTableAttachments, txnRouteTableAttachmentAttrTypes)
	plan.ImageSchedules = assignSubIDsMap(plan.ImageSchedules, txnImageScheduleAttrTypes)
	plan.LoadbalancerTargetGroups = assignSubIDsMap(plan.LoadbalancerTargetGroups, txnLoadbalancerTargetGroupAttrTypes)
	plan.LoadbalancerTargetGroupStaticTargets = assignSubIDsMap(plan.LoadbalancerTargetGroupStaticTargets, txnLoadbalancerTargetGroupStaticTargetAttrTypes)
	plan.LoadbalancerTargetGroupServiceDiscoveryTargets = assignSubIDsMap(plan.LoadbalancerTargetGroupServiceDiscoveryTargets, txnLoadbalancerTargetGroupServiceDiscoveryTargetAttrTypes)
	plan.LoadbalancerHttpListeners = assignSubIDsMap(plan.LoadbalancerHttpListeners, txnLoadbalancerHttpListenerAttrTypes)
	plan.LoadbalancerHttpsListeners = assignSubIDsMap(plan.LoadbalancerHttpsListeners, txnLoadbalancerHttpsListenerAttrTypes)
	plan.LoadbalancerTlsListeners = assignSubIDsMap(plan.LoadbalancerTlsListeners, txnLoadbalancerTlsListenerAttrTypes)
	plan.LoadbalancerTcpListeners = assignSubIDsMap(plan.LoadbalancerTcpListeners, txnLoadbalancerTcpListenerAttrTypes)
	plan.LoadbalancerUdpListeners = assignSubIDsMap(plan.LoadbalancerUdpListeners, txnLoadbalancerUdpListenerAttrTypes)
	plan.LoadbalancerHttpListenerRules = assignSubIDsMap(plan.LoadbalancerHttpListenerRules, txnLoadbalancerHttpListenerRuleAttrTypes)
	plan.LoadbalancerHttpsListenerRules = assignSubIDsMap(plan.LoadbalancerHttpsListenerRules, txnLoadbalancerHttpsListenerRuleAttrTypes)
	plan.LoadbalancerTlsListenerRules = assignSubIDsMap(plan.LoadbalancerTlsListenerRules, txnLoadbalancerTlsListenerRuleAttrTypes)
	plan.LoadbalancerTcpListenerRules = assignSubIDsMap(plan.LoadbalancerTcpListenerRules, txnLoadbalancerTcpListenerRuleAttrTypes)
	plan.LoadbalancerUdpListenerRules = assignSubIDsMap(plan.LoadbalancerUdpListenerRules, txnLoadbalancerUdpListenerRuleAttrTypes)
	plan.KubernetesNodeGroups = assignSubIDsMap(plan.KubernetesNodeGroups, txnKubernetesNodeGroupAttrTypes)
	plan.KubernetesUserRoles = assignSubIDsMap(plan.KubernetesUserRoles, txnKubernetesUserRoleAttrTypes)
	plan.OpenVpns = assignSubIDsMap(plan.OpenVpns, txnOpenVpnAttrTypes)
	plan.PostgresqlParametersSets = assignSubIDsMap(plan.PostgresqlParametersSets, txnPostgresqlParametersSetAttrTypes)
	plan.SupportPlans = assignSubIDsMap(plan.SupportPlans, txnSupportPlanAttrTypes)
	plan.SupportTickets = assignSubIDsMap(plan.SupportTickets, txnSupportTicketAttrTypes)
	plan.SupportTicketComments = assignSubIDsMap(plan.SupportTicketComments, txnSupportTicketCommentAttrTypes)
	plan.GitlabRunners = assignSubIDsMap(plan.GitlabRunners, txnGitlabRunnerAttrTypes)
	plan.OpenVpnUserSettings = assignSubIDsMap(plan.OpenVpnUserSettings, txnOpenVpnUserSettingsAttrTypes)
	plan.Users = assignSubIDsMap(plan.Users, txnIamUserAttrTypes)
	plan.UserTokens = assignSubIDsMap(plan.UserTokens, txnUserTokenAttrTypes)
	plan.FloatingIps = assignSubIDsMap(plan.FloatingIps, txnFloatingIpAttrTypes)
	plan.Vpcs = assignSubIDsMap(plan.Vpcs, txnVpcAttrTypes)
	plan.VpcPeeringPeers = assignSubIDsMap(plan.VpcPeeringPeers, txnVpcPeeringPeerAttrTypes)
	plan.Loadbalancers = assignSubIDsMap(plan.Loadbalancers, txnLoadbalancerAttrTypes)
	plan.Kubernetes = assignSubIDsMap(plan.Kubernetes, txnKubernetesAttrTypes)
	plan.KubernetesUsers = assignSubIDsMap(plan.KubernetesUsers, txnKubernetesUserAttrTypes)
	plan.PostgresqlStandalones = assignSubIDsMap(plan.PostgresqlStandalones, txnPostgresqlStandaloneAttrTypes)
	plan.OpenVpnUsers = assignSubIDsMap(plan.OpenVpnUsers, txnOpenVpnUserAttrTypes)
	plan.BillingAccounts = assignSubIDsMap(plan.BillingAccounts, txnBillingAccountAttrTypes)
	plan.Quotas = assignSubIDsMap(plan.Quotas, txnQuotaAttrTypes)
	plan.QuotaChangeRequests = assignSubIDsMap(plan.QuotaChangeRequests, txnQuotaChangeRequestAttrTypes)
	plan.VictoriaMetrics = assignSubIDsMap(plan.VictoriaMetrics, txnVictoriaMetricsAttrTypes)
	plan.Gitlabs = assignSubIDsMap(plan.Gitlabs, txnGitlabAttrTypes)
	plan.SupportTicketCommentAttachments = assignSubIDsMap(plan.SupportTicketCommentAttachments, txnSupportTicketCommentAttachmentAttrTypes)
	plan.Images = assignSubIDsMap(plan.Images, txnImageAttrTypes)
	plan.Vms = assignSubIDsMap(plan.Vms, txnVmAttrTypes)
	plan.SecurityGroups = assignSubIDsMap(plan.SecurityGroups, txnSecurityGroupAttrTypes)

	body := buildTxnBody(ctx, plan, true, true, true, true)
	modResp, err := r.client.Put(ctx, "/api/v1/transaction", body)
	if err != nil { resp.Diagnostics.AddError("Update Error", err.Error()); return }
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error()); return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/transaction", plan.ID.ValueString())
	if err != nil { resp.Diagnostics.AddError("Read After Update Error", err.Error()); return }
	if err := r.populateState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error()); return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TransactionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TransactionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() { return }

	modResp, err := r.client.Delete(ctx, "/api/v1/transaction", state.ID.ValueString())
	if err != nil { resp.Diagnostics.AddError("Delete Error", err.Error()); return }
	if err := r.client.PollUntilDone(ctx, "/api/v1/transaction", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error()); return
	}
}

func (r *TransactionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state TransactionResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/transaction", req.ID)
	if err != nil { resp.Diagnostics.AddError("Import Error", err.Error()); return }
	if apiData == nil { resp.Diagnostics.AddError("Import Error", "resource not found"); return }
	if err := r.populateState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error()); return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// ---------------------------------------------------------------------------
// Helpers for two-phase create
// ---------------------------------------------------------------------------

// parseSchedulingConflict extracts the ULID and resource type from a ResourceIsScheduling or
// ResourceIsReconciling 422 error so the caller can wait and retry.
func parseSchedulingConflict(errMsg string) (ulid, resType string) {
	if !strings.Contains(errMsg, "ResourceIsScheduling") && !strings.Contains(errMsg, "is in state Scheduling") &&
		!strings.Contains(errMsg, "Reconciling state") && !strings.Contains(errMsg, "ResourceIsReconciling") {
		return "", ""
	}
	// Map API type names to path segments
	typeMap := map[string]string{
		"OrganizationS3UserAccessPolicy": "s3-user-access-policy",
		"OrganizationS3Bucket":           "s3-bucket",
		"OrganizationFolder":             "folder",
		"OrganizationTransaction":        "transaction",
	}
	for apiType, path := range typeMap {
		if strings.Contains(errMsg, apiType) { resType = path; break }
	}
	// Extract ULID: text inside parentheses after the type name.
	start := strings.Index(errMsg, " (")
	end := strings.Index(errMsg, ") is in state Scheduling")
	if end < 0 { end = strings.Index(errMsg, ") is in state Reconciling") }
	if start < 0 || end < 0 || end <= start { return "", "" }
	ulid = strings.TrimSpace(errMsg[start+2 : end])
	return ulid, resType
}

// waitForResourceStable polls a specific resource until its info.state leaves "scheduling".
func waitForResourceStable(ctx context.Context, c *client.Client, resType, id string) error {
	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		apiData, err := c.Get(ctx, "/api/v1/"+resType, id)
		if err != nil || apiData == nil { return nil }
		info, _ := apiData["info"].(map[string]interface{})
		state, _ := info["state"].(string)
		s := strings.ToLower(state)
		if s != "scheduling" && s != "reconciling" { return nil }
		select {
		case <-ctx.Done(): return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("timed out waiting for %s/%s to leave scheduling state", resType, id)
}

// deleteAnyExistingResource deletes a resource regardless of its state, silently ignoring
// "not found" responses. Used in the Phase 1 retry loop to clear orphaned blocking resources.
func deleteAnyExistingResource(ctx context.Context, c *client.Client, resType, id string) {
	apiData, err := c.Get(ctx, "/api/v1/"+resType, id)
	if err != nil || apiData == nil { return }
	_ = waitForResourceStable(ctx, c, resType, id)
	for i := 0; i < 10; i++ {
		delResp, delErr := c.Delete(ctx, "/api/v1/"+resType, id)
		if delErr == nil && delResp != nil {
			_ = c.PollUntilDone(ctx, "/api/v1/"+resType, delResp.RequestId)
			return
		}
		select {
		case <-ctx.Done(): return
		case <-time.After(10 * time.Second):
		}
	}
}

// cleanupTransaction explicitly deletes all sub-resources of a transaction before deleting
// the transaction container.
func cleanupTransaction(ctx context.Context, c *client.Client, txnID string) {
	apiData, err := c.Get(ctx, "/api/v1/transaction", txnID)
	if err == nil && apiData != nil {
		for _, key := range []string{"s3Buckets", "s3UserAccessPolicies", "folders"} {
			resType := map[string]string{
				"s3Buckets": "s3-bucket", "s3UserAccessPolicies": "s3-user-access-policy", "folders": "folder",
			}[key]
			items, _ := apiData[key].([]interface{})
			for _, raw := range items {
				m, _ := raw.(map[string]interface{})
				if m == nil { continue }
				id, _ := m["id"].(string)
				if id != "" {
					for _, rt := range []string{resType, "s3-bucket", "s3-user-access-policy"} {
						deleteAnyExistingResource(ctx, c, rt, id)
					}
				}
			}
		}
	}
	if delResp, delErr := c.Delete(ctx, "/api/v1/transaction", txnID); delErr == nil {
		_ = c.PollUntilDone(ctx, "/api/v1/transaction", delResp.RequestId)
	}
}

// deleteIfSchedulingFailed deletes a resource that is in schedulingfailed state.
func deleteIfSchedulingFailed(ctx context.Context, c *client.Client, resType, id string) {
	apiData, err := c.Get(ctx, "/api/v1/"+resType, id)
	if err != nil || apiData == nil { return }
	info, _ := apiData["info"].(map[string]interface{})
	state, _ := info["state"].(string)
	if strings.ToLower(state) != "schedulingfailed" { return }
	for i := 0; i < 10; i++ {
		delResp, delErr := c.Delete(ctx, "/api/v1/"+resType, id)
		if delErr == nil && delResp != nil {
			_ = c.PollUntilDone(ctx, "/api/v1/"+resType, delResp.RequestId)
			return
		}
		select {
		case <-ctx.Done(): return
		case <-time.After(10 * time.Second):
		}
	}
}

// waitForSubResourcesStable polls all bucket and policy sub-resources of the given transaction
// until none of them is in "scheduling" state.
func waitForSubResourcesStable(ctx context.Context, c *client.Client, txnID string) error {
	deadline := time.Now().Add(30 * time.Minute)
	type subRes struct{ resType, id string }
	for time.Now().Before(deadline) {
		apiData, err := c.Get(ctx, "/api/v1/transaction", txnID)
		if err != nil { return err }
		var pending []subRes
		for _, key := range []string{"s3Buckets", "s3UserAccessPolicies", "folders"} {
			resType := map[string]string{
				"s3Buckets": "s3-bucket", "s3UserAccessPolicies": "s3-user-access-policy", "folders": "folder",
			}[key]
			items, _ := apiData[key].([]interface{})
			for _, raw := range items {
				m, _ := raw.(map[string]interface{})
				if m == nil { continue }
				id, _ := m["id"].(string)
				info, _ := m["info"].(map[string]interface{})
				state, _ := info["state"].(string)
				switch strings.ToLower(state) {
				case "scheduling", "reconciling":
					pending = append(pending, subRes{resType, id})
				case "deleted", "schedulingfailed":
					return fmt.Errorf("sub-resource %s/%s entered failed state %q — provisioning failed", resType, id, state)
				}
			}
		}
		if len(pending) == 0 { return nil }
		select {
		case <-ctx.Done(): return ctx.Err()
		case <-time.After(10 * time.Second):
		}
	}
	return fmt.Errorf("timed out waiting for transaction sub-resources to leave scheduling state")
}

// fillUnknownIDsFromConfig copies known values from cfgMap into planMap at matching keys
// for the "id" field and any extraFields.
func fillUnknownIDsFromConfig(_ context.Context, planMap, cfgMap types.Map, attrTypes map[string]attr.Type, extraFields []string) types.Map {
	if cfgMap.IsNull() || cfgMap.IsUnknown() { return planMap }
	if planMap.IsNull() || planMap.IsUnknown() { return planMap }
	planElems := planMap.Elements()
	cfgElems := cfgMap.Elements()
	if len(planElems) == 0 { return planMap }

	updated := make(map[string]attr.Value, len(planElems))
	for key, pe := range planElems {
		planObj, ok := pe.(types.Object)
		if !ok { updated[key] = pe; continue }

		cfgVal, hasCfg := cfgElems[key]
		if !hasCfg { updated[key] = pe; continue }
		cfgObj, ok2 := cfgVal.(types.Object)
		if !ok2 { updated[key] = pe; continue }

		origAttrs := planObj.Attributes()
		newAttrs := make(map[string]attr.Value, len(origAttrs))
		for k, v := range origAttrs { newAttrs[k] = v }

		cfgAttrs := cfgObj.Attributes()
		overrideIfConfigKnown(newAttrs, cfgAttrs, "id")
		for _, f := range extraFields { overrideIfConfigKnown(newAttrs, cfgAttrs, f) }

		newObj, diags := types.ObjectValue(attrTypes, newAttrs)
		if diags.HasError() { updated[key] = pe; continue }
		updated[key] = newObj
	}

	result, _ := types.MapValue(types.ObjectType{AttrTypes: attrTypes}, updated)
	return result
}

// overrideIfConfigKnown copies cfgAttrs[key] into planAttrs[key] whenever the config has a
// known non-null non-empty value.
func overrideIfConfigKnown(planAttrs, cfgAttrs map[string]attr.Value, key string) {
	cfgVal, hasCfg := cfgAttrs[key]
	if !hasCfg { return }
	if _, hasPlan := planAttrs[key]; !hasPlan { return }

	var cfgIsKnown bool
	switch cv := cfgVal.(type) {
	case types.String: cfgIsKnown = !cv.IsNull() && !cv.IsUnknown() && cv.ValueString() != ""
	case types.List:   cfgIsKnown = !cv.IsNull() && !cv.IsUnknown()
	default: return
	}
	if !cfgIsKnown { return }

	// Normalize ULIDs to lowercase.
	switch cv := cfgVal.(type) {
	case types.String:
		planAttrs[key] = types.StringValue(strings.ToLower(cv.ValueString()))
	case types.List:
		elems := cv.Elements()
		lowerElems := make([]attr.Value, len(elems))
		for i, e := range elems {
			if sv, ok := e.(types.String); ok {
				lowerElems[i] = types.StringValue(strings.ToLower(sv.ValueString()))
			} else {
				lowerElems[i] = e
			}
		}
		lowerList, diags := types.ListValue(types.StringType, lowerElems)
		if diags.HasError() { planAttrs[key] = cfgVal; return }
		planAttrs[key] = lowerList
	default:
		planAttrs[key] = cfgVal
	}
}

// autoWireUsers populates s3_users.bucket_id with actual Phase 1 bucket ID for users that
// have unknown/null bucket_id.
func autoWireUsers(_ context.Context, users, buckets, _ types.Map) types.Map {
	if users.IsNull() || users.IsUnknown() || len(users.Elements()) == 0 { return users }

	var firstBucketID types.String
	if !buckets.IsNull() && !buckets.IsUnknown() {
		for _, bVal := range buckets.Elements() {
			if bObj, ok := bVal.(types.Object); ok {
				if idVal, ok2 := bObj.Attributes()["id"].(types.String); ok2 && !idVal.IsNull() && !idVal.IsUnknown() && idVal.ValueString() != "" {
					firstBucketID = idVal
					break
				}
			}
		}
	}

	userElems := users.Elements()
	updated := make(map[string]attr.Value, len(userElems))
	for key, ue := range userElems {
		uObj, ok := ue.(types.Object)
		if !ok { updated[key] = ue; continue }

		attrs := uObj.Attributes()
		newAttrs := make(map[string]attr.Value, len(attrs))
		for k, v := range attrs { newAttrs[k] = v }

		if bv, ok2 := newAttrs["bucket_id"].(types.String); ok2 && (bv.IsNull() || bv.IsUnknown()) {
			if !firstBucketID.IsNull() && !firstBucketID.IsUnknown() {
				newAttrs["bucket_id"] = firstBucketID
			}
		}

		newObj, diags := types.ObjectValue(txnS3UserAttrTypes, newAttrs)
		if diags.HasError() { updated[key] = ue; continue }
		updated[key] = newObj
	}

	result, _ := types.MapValue(types.ObjectType{AttrTypes: txnS3UserAttrTypes}, updated)
	return result
}
