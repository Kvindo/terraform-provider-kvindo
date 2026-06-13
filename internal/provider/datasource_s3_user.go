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

// S3UserDataSourceModel describes the data source data model.
type S3UserDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	BucketId types.String `tfsdk:"bucket_id"`
	AccessPolicyIds types.List `tfsdk:"access_policy_ids"`
	InfoState types.String `tfsdk:"info_state"`
	InfoAccessKey types.String `tfsdk:"info_access_key"`
	InfoSecretKey types.String `tfsdk:"info_secret_key"`
}

type S3UserDataSource struct {
	client *client.Client
}

func NewS3UserDataSource() datasource.DataSource {
	return &S3UserDataSource{}
}

func (d *S3UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_user"
}

func (d *S3UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["bucket_id"] = schema.StringAttribute{Computed: true}
	attrs["access_policy_ids"] = schema.ListAttribute{Computed: true, ElementType: types.StringType}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_access_key"] = schema.StringAttribute{Computed: true, Sensitive: true}
	attrs["info_secret_key"] = schema.StringAttribute{Computed: true, Sensitive: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *S3UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *S3UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state S3UserDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/s3-user", state.ID.ValueString())
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
	state.BucketId = getString(apiData, "bucketId")
	state.AccessPolicyIds = getStringList(ctx, apiData, "accessPolicyIds")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoAccessKey = getStringFromInfo(apiData, "accesskey")
	state.InfoSecretKey = getStringFromInfo(apiData, "secretkey")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
