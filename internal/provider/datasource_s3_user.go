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

type S3UserDataSourceModel struct {
	ID       types.String    `tfsdk:"id"`
	Metadata metadataModel   `tfsdk:"metadata"`
	Spec     S3UserSpecModel `tfsdk:"spec"`
	Status   types.Object    `tfsdk:"status"`
}

type S3UserDataSource struct{ client *client.Client }

func NewS3UserDataSource() datasource.DataSource { return &S3UserDataSource{} }

func (d *S3UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_user"
}

func (d *S3UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"access_policy_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		"bucket_id":         schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Required: true},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"access_key": schema.StringAttribute{Computed: true}, "secret_key": schema.StringAttribute{Computed: true, Sensitive: true}}),
	}}
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	spec := getSpec(apiData)
	state.Spec.AccessPolicyIds = getStringList(ctx, spec, "accessPolicyIds")
	state.Spec.BucketId = getString(spec, "bucketId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"access_key": types.StringType,
			"secret_key": types.StringType,
		},
		map[string]attr.Value{
			"access_key": getStringFromInfo(apiData, "accessKey"),
			"secret_key": getStringFromInfo(apiData, "secretKey"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
