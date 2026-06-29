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

type SshPrivateKeyDataSourceModel struct {
	ID       types.String           `tfsdk:"id"`
	Metadata metadataModel          `tfsdk:"metadata"`
	Spec     SshPrivateKeySpecModel `tfsdk:"spec"`
	Status   types.Object           `tfsdk:"status"`
}

type SshPrivateKeyDataSource struct{ client *client.Client }

func NewSshPrivateKeyDataSource() datasource.DataSource { return &SshPrivateKeyDataSource{} }

func (d *SshPrivateKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_private_key"
}

func (d *SshPrivateKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"private_key": schema.StringAttribute{Computed: true, Sensitive: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(nil),
	}}
}

func (d *SshPrivateKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SshPrivateKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SshPrivateKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := d.client.Get(ctx, "/api/v1/ssh-private-key", state.ID.ValueString())
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
	state.Spec.PrivateKey = getString(spec, "privateKey")
	state.Status = simpleStateInfoObj(apiData)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
