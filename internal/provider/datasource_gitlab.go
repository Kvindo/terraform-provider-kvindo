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

// GitlabDataSourceModel describes the data source data model.
type GitlabDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	Version types.String `tfsdk:"version"`
	RootPassword types.String `tfsdk:"root_password"`
	VmState types.String `tfsdk:"vm_state"`
	VmOfferId types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib types.Int64 `tfsdk:"volume_size_gib"`
	Edition types.String `tfsdk:"edition"`
	RecordName types.String `tfsdk:"record_name"`
	InfoState types.String `tfsdk:"info_state"`
	InfoPublicIpV4 types.String `tfsdk:"info_public_ip_v4"`
	InfoPublicIpV6 types.String `tfsdk:"info_public_ip_v6"`
	InfoPrivateIpV4 types.String `tfsdk:"info_private_ip_v4"`
	InfoPrivateIpV6 types.String `tfsdk:"info_private_ip_v6"`
	InfoFqdn types.String `tfsdk:"info_fqdn"`
}

type GitlabDataSource struct {
	client *client.Client
}

func NewGitlabDataSource() datasource.DataSource {
	return &GitlabDataSource{}
}

func (d *GitlabDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab"
}

func (d *GitlabDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{Computed: true}
	attrs["floating_ip_id"] = schema.StringAttribute{Computed: true}
	attrs["vpc_subnet_id"] = schema.StringAttribute{Computed: true}
	attrs["version"] = schema.StringAttribute{Computed: true}
	attrs["root_password"] = schema.StringAttribute{Computed: true, Sensitive: true}
	attrs["vm_state"] = schema.StringAttribute{Computed: true}
	attrs["vm_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_size_gib"] = schema.Int64Attribute{Computed: true}
	attrs["edition"] = schema.StringAttribute{Computed: true}
	attrs["record_name"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ip_v6"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ip_v6"] = schema.StringAttribute{Computed: true}
	attrs["info_fqdn"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *GitlabDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GitlabDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GitlabDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/gitlab", state.ID.ValueString())
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
	state.FloatingIpId = getString(apiData, "floatingIpId")
	state.VpcSubnetId = getString(apiData, "vpcSubnetId")
	state.Version = getString(apiData, "version")
	state.RootPassword = getString(apiData, "rootPassword")
	state.VmState = getString(apiData, "vmState")
	state.VmOfferId = getString(apiData, "vmOfferId")
	state.VolumeOfferId = getString(apiData, "volumeOfferId")
	state.VolumeSizeGib = getInt64(apiData, "volumeSizeGib")
	state.Edition = getString(apiData, "edition")
	state.RecordName = getString(apiData, "recordName")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoPublicIpV4 = getStringFromInfo(apiData, "publicipv4")
	state.InfoPublicIpV6 = getStringFromInfo(apiData, "publicipv6")
	state.InfoPrivateIpV4 = getStringFromInfo(apiData, "privateipv4")
	state.InfoPrivateIpV6 = getStringFromInfo(apiData, "privateipv6")
	state.InfoFqdn = getStringFromInfo(apiData, "fqdn")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
