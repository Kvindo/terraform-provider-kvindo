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

// GitlabRunnerDataSourceModel describes the data source data model.
type GitlabRunnerDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	VpcSubnetId types.String `tfsdk:"vpc_subnet_id"`
	FloatingIpId types.String `tfsdk:"floating_ip_id"`
	VmState types.String `tfsdk:"vm_state"`
	VmOfferId types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib types.Int64 `tfsdk:"volume_size_gib"`
	Concurrency types.Int64 `tfsdk:"concurrency"`
	Version types.String `tfsdk:"version"`
	DockerOptionsJsonString types.String `tfsdk:"docker_options_json_string"`
	InfoState types.String `tfsdk:"info_state"`
}

type GitlabRunnerDataSource struct {
	client *client.Client
}

func NewGitlabRunnerDataSource() datasource.DataSource {
	return &GitlabRunnerDataSource{}
}

func (d *GitlabRunnerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab_runner"
}

func (d *GitlabRunnerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{Computed: true}
	attrs["vpc_subnet_id"] = schema.StringAttribute{Computed: true}
	attrs["floating_ip_id"] = schema.StringAttribute{Computed: true}
	attrs["vm_state"] = schema.StringAttribute{Computed: true}
	attrs["vm_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_offer_id"] = schema.StringAttribute{Computed: true}
	attrs["volume_size_gib"] = schema.Int64Attribute{Computed: true}
	attrs["concurrency"] = schema.Int64Attribute{Computed: true}
	attrs["version"] = schema.StringAttribute{Computed: true}
	attrs["docker_options_json_string"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *GitlabRunnerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GitlabRunnerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GitlabRunnerDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/gitlab-runner", state.ID.ValueString())
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
	state.VpcSubnetId = getString(apiData, "vpcSubnetId")
	state.FloatingIpId = getString(apiData, "floatingIpId")
	state.VmState = getString(apiData, "vmState")
	state.VmOfferId = getString(apiData, "vmOfferId")
	state.VolumeOfferId = getString(apiData, "volumeOfferId")
	state.VolumeSizeGib = getInt64(apiData, "volumeSizeGib")
	state.Concurrency = getInt64(apiData, "concurrency")
	state.Version = getString(apiData, "version")
	state.DockerOptionsJsonString = getString(apiData, "dockerOptionsJsonString")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
