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

// KubernetesUserDataSourceModel describes the data source data model.
type KubernetesUserDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	KubernetesId types.String `tfsdk:"kubernetes_id"`
	RoleIds types.List `tfsdk:"role_ids"`
	InfoState types.String `tfsdk:"info_state"`
	InfoKubeconfig types.String `tfsdk:"info_kubeconfig"`
}

type KubernetesUserDataSource struct {
	client *client.Client
}

func NewKubernetesUserDataSource() datasource.DataSource {
	return &KubernetesUserDataSource{}
}

func (d *KubernetesUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_user"
}

func (d *KubernetesUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["kubernetes_id"] = schema.StringAttribute{Computed: true}
	attrs["role_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_kubeconfig"] = schema.StringAttribute{Computed: true, Sensitive: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *KubernetesUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KubernetesUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KubernetesUserDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/kubernetes-user", state.ID.ValueString())
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
	state.KubernetesId = getString(apiData, "kubernetesId")
	state.RoleIds = getStringList(ctx, apiData, "roleIds")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoKubeconfig = getStringFromInfo(apiData, "kubeconfig")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
