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

type KubernetesUserDataSourceModel struct {
	ID       types.String            `tfsdk:"id"`
	Metadata metadataModel           `tfsdk:"metadata"`
	Spec     KubernetesUserSpecModel `tfsdk:"spec"`
	Status   types.Object            `tfsdk:"status"`
}

type KubernetesUserDataSource struct{ client *client.Client }

func NewKubernetesUserDataSource() datasource.DataSource { return &KubernetesUserDataSource{} }

func (d *KubernetesUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_user"
}

func (d *KubernetesUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"kubernetes_id": schema.StringAttribute{Computed: true},
		"role_ids":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"kubeconfig": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}}
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	spec := getSpec(apiData)
	state.Spec.KubernetesId = getString(spec, "kubernetesId")
	state.Spec.RoleIds = getStringList(ctx, spec, "roleIds")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"kubeconfig": types.StringType,
		},
		map[string]attr.Value{
			"kubeconfig": getStringFromInfo(apiData, "kubeconfig"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
