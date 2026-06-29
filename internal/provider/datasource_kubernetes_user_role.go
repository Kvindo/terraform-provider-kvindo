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

type KubernetesUserRoleDataSourceModel struct {
	ID       types.String                `tfsdk:"id"`
	Metadata metadataModel               `tfsdk:"metadata"`
	Spec     KubernetesUserRoleSpecModel `tfsdk:"spec"`
	Status   types.Object                `tfsdk:"status"`
}

type KubernetesUserRoleDataSource struct{ client *client.Client }

func NewKubernetesUserRoleDataSource() datasource.DataSource { return &KubernetesUserRoleDataSource{} }

func (d *KubernetesUserRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_user_role"
}

func (d *KubernetesUserRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"api_groups": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"namespaces": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"resources":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"verbs":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	spec := getSpec(apiData)
	state.Spec.ApiGroups = getStringList(ctx, spec, "apiGroups")
	state.Spec.Namespaces = getStringList(ctx, spec, "namespaces")
	state.Spec.Resources = getStringList(ctx, spec, "resources")
	state.Spec.Verbs = getStringList(ctx, spec, "verbs")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
