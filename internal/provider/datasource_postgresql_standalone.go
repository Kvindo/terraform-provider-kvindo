package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type PostgresqlStandaloneDataSourceModel struct {
	ID       types.String                  `tfsdk:"id"`
	Metadata metadataModel                 `tfsdk:"metadata"`
	Spec     PostgresqlStandaloneSpecModel `tfsdk:"spec"`
	Status   types.Object                  `tfsdk:"status"`
}

type PostgresqlStandaloneDataSource struct{ client *client.Client }

func NewPostgresqlStandaloneDataSource() datasource.DataSource {
	return &PostgresqlStandaloneDataSource{}
}

func (d *PostgresqlStandaloneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_standalone"
}

func (d *PostgresqlStandaloneDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"backup_retention_days": schema.Int64Attribute{Computed: true},
		"floating_ip_id":        schema.StringAttribute{Computed: true},
		"parameters_set_id":     schema.StringAttribute{Computed: true},
		"root_password":         schema.StringAttribute{Computed: true, Sensitive: true},
		"tier":                  schema.StringAttribute{Computed: true},
		"version":               schema.StringAttribute{Computed: true},
		"vm_offer_id":           schema.StringAttribute{Computed: true},
		"vm_state":              schema.StringAttribute{Computed: true},
		"volume_offer_id":       schema.StringAttribute{Computed: true},
		"volume_size_gib":       schema.Int64Attribute{Computed: true},
		"vpc_subnet_id":         schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"port": schema.Int64Attribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "root_user_name": schema.StringAttribute{Computed: true}}),
	}}
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	spec := getSpec(apiData)
	state.Spec.BackupRetentionDays = getInt64(spec, "backupRetentionDays")
	state.Spec.FloatingIpId = getString(spec, "floatingIpId")
	state.Spec.ParametersSetId = getString(spec, "parametersSetId")
	state.Spec.RootPassword = getString(spec, "rootPassword")
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.Version = getString(spec, "version")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"port":           types.Int64Type,
			"private_ip_v4":  types.StringType,
			"public_ip_v4":   types.StringType,
			"root_user_name": types.StringType,
		},
		map[string]attr.Value{
			"port":           getInt64FromInfo(apiData, "port"),
			"private_ip_v4":  getStringFromInfo(apiData, "privateIpV4"),
			"public_ip_v4":   getStringFromInfo(apiData, "publicIpV4"),
			"root_user_name": getStringFromInfo(apiData, "rootUserName"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
