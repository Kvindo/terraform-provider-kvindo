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

// KubernetesUserRoleDataSourceModel describes the data source data model.
type KubernetesUserRoleDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	ApiGroups types.List `tfsdk:"api_groups"`
	Resources types.List `tfsdk:"resources"`
	Verbs types.List `tfsdk:"verbs"`
	Namespaces types.List `tfsdk:"namespaces"`
	InfoState types.String `tfsdk:"info_state"`
}

type KubernetesUserRoleDataSource struct {
	client *client.Client
}

func NewKubernetesUserRoleDataSource() datasource.DataSource {
	return &KubernetesUserRoleDataSource{}
}

func (d *KubernetesUserRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_user_role"
}

func (d *KubernetesUserRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["api_groups"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["resources"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["verbs"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["namespaces"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *KubernetesUserRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KubernetesUserRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KubernetesUserRoleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/kubernetes-user-role", state.ID.ValueString())
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
	state.ApiGroups = getStringList(ctx, apiData, "apiGroups")
	state.Resources = getStringList(ctx, apiData, "resources")
	state.Verbs = getStringList(ctx, apiData, "verbs")
	state.Namespaces = getStringList(ctx, apiData, "namespaces")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
