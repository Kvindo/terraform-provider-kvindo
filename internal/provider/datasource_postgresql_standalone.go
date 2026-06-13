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

// PostgresqlStandaloneDataSourceModel describes the data source data model.
type PostgresqlStandaloneDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	Version types.String `tfsdk:"version"`
	RootPassword types.String `tfsdk:"root_password"`
	ParametersSetId types.String `tfsdk:"parameters_set_id"`
	BackupRetentionDays types.Int64 `tfsdk:"backup_retention_days"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	VmState types.String `tfsdk:"vm_state"`
	VmOfferId types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib types.Int64 `tfsdk:"volume_size_gib"`
	InfoState types.String `tfsdk:"info_state"`
	InfoRootUserName types.String `tfsdk:"info_root_user_name"`
	InfoPublicIpV4 types.String `tfsdk:"info_public_ip_v4"`
	InfoPrivateIpV4 types.String `tfsdk:"info_private_ip_v4"`
	InfoPort types.Int64 `tfsdk:"info_port"`
}

type PostgresqlStandaloneDataSource struct {
	client *client.Client
}

func NewPostgresqlStandaloneDataSource() datasource.DataSource {
	return &PostgresqlStandaloneDataSource{}
}

func (d *PostgresqlStandaloneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_standalone"
}

func (d *PostgresqlStandaloneDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{Computed: true}
	attrs["version"] = schema.StringAttribute{Computed: true}
	attrs["root_password"] = schema.StringAttribute{Computed: true, Sensitive: true}
	attrs["parameters_set_id"] = schema.StringAttribute{Computed: true}
	attrs["backup_retention_days"] = schema.Int64Attribute{Computed: true}
	attrs["floating_ip_id"] = schema.StringAttribute{Computed: true}
	attrs["vpc_subnet_id"] = schema.StringAttribute{Computed: true}
	attrs["vm_state"] = schema.StringAttribute{Computed: true}
	attrs["vm_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_size_gib"] = schema.Int64Attribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_root_user_name"] = schema.StringAttribute{Computed: true}
	attrs["info_public_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["info_private_ip_v4"] = schema.StringAttribute{Computed: true}
	attrs["info_port"] = schema.Int64Attribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *PostgresqlStandaloneDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PostgresqlStandaloneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state PostgresqlStandaloneDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/postgresql-standalone", state.ID.ValueString())
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
	state.Version = getString(apiData, "version")
	state.RootPassword = getString(apiData, "rootPassword")
	state.ParametersSetId = getString(apiData, "parametersSetId")
	state.BackupRetentionDays = getInt64(apiData, "backupRetentionDays")
	state.FloatingIpId = getString(apiData, "floatingIpId")
	state.VpcSubnetId = getString(apiData, "vpcSubnetId")
	state.VmState = getString(apiData, "vmState")
	state.VmOfferId = getString(apiData, "vmOfferId")
	state.VolumeOfferId = getString(apiData, "volumeOfferId")
	state.VolumeSizeGib = getInt64(apiData, "volumeSizeGib")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoRootUserName = getStringFromInfo(apiData, "rootusername")
	state.InfoPublicIpV4 = getStringFromInfo(apiData, "publicipv4")
	state.InfoPrivateIpV4 = getStringFromInfo(apiData, "privateipv4")
	state.InfoPort = getInt64FromInfo(apiData, "port")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
