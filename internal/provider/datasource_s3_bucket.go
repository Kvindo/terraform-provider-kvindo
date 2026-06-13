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

// S3BucketDataSourceModel describes the data source data model.
type S3BucketDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	Region types.String `tfsdk:"region"`
	IsPublic types.Bool `tfsdk:"is_public"`
	IsLockEnabled types.Bool `tfsdk:"is_lock_enabled"`
	IsVersioned types.Bool `tfsdk:"is_versioned"`
	ObjectExpirationDays types.Int64 `tfsdk:"object_expiration_days"`
	ComplianceRetentionDays types.Int64 `tfsdk:"compliance_retention_days"`
	QuotaGib types.Int64 `tfsdk:"quota_gib"`
	InfoState types.String `tfsdk:"info_state"`
	InfoEndpointUrl types.String `tfsdk:"info_endpoint_url"`
}

type S3BucketDataSource struct {
	client *client.Client
}

func NewS3BucketDataSource() datasource.DataSource {
	return &S3BucketDataSource{}
}

func (d *S3BucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_bucket"
}

func (d *S3BucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{Computed: true}
	attrs["region"] = schema.StringAttribute{Computed: true}
	attrs["is_public"] = schema.BoolAttribute{Computed: true}
	attrs["is_lock_enabled"] = schema.BoolAttribute{Computed: true}
	attrs["is_versioned"] = schema.BoolAttribute{Computed: true}
	attrs["object_expiration_days"] = schema.Int64Attribute{Computed: true}
	attrs["compliance_retention_days"] = schema.Int64Attribute{Computed: true}
	attrs["quota_gib"] = schema.Int64Attribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_endpoint_url"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *S3BucketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *S3BucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state S3BucketDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/s3-bucket", state.ID.ValueString())
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
	state.Region = getString(apiData, "region")
	state.IsPublic = getBool(apiData, "isPublic")
	state.IsLockEnabled = getBool(apiData, "isLockEnabled")
	state.IsVersioned = getBool(apiData, "isVersioned")
	state.ObjectExpirationDays = getInt64(apiData, "objectExpirationDays")
	state.ComplianceRetentionDays = getInt64(apiData, "complianceRetentionDays")
	state.QuotaGib = getInt64(apiData, "quotaGib")
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoEndpointUrl = getStringFromInfo(apiData, "endpointurl")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
