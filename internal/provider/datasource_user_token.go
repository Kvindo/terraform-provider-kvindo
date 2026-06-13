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

// UserTokenDataSourceModel describes the data source data model.
type UserTokenDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	UserId types.String `tfsdk:"user_id"`
	SendToEmail types.Bool `tfsdk:"send_to_email"`
	InfoState types.String `tfsdk:"info_state"`
	InfoToken types.String `tfsdk:"info_token"`
}

type UserTokenDataSource struct {
	client *client.Client
}

func NewUserTokenDataSource() datasource.DataSource {
	return &UserTokenDataSource{}
}

func (d *UserTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_token"
}

func (d *UserTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["user_id"] = schema.StringAttribute{Computed: true}
	attrs["send_to_email"] = schema.BoolAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_token"] = schema.StringAttribute{Computed: true, Sensitive: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *UserTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UserTokenDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/user-token", state.ID.ValueString())
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
	state.UserId = getString(apiData, "userId")
	state.SendToEmail = getBool(apiData, "sendToEmail")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoToken = getStringFromInfo(apiData, "token")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
