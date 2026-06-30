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

type S3BucketDataSourceModel struct {
	ID       types.String      `tfsdk:"id"`
	Name     types.String      `tfsdk:"name"`
	Metadata metadataModel     `tfsdk:"metadata"`
	Spec     S3BucketSpecModel `tfsdk:"spec"`
	Status   types.Object      `tfsdk:"status"`
}

type S3BucketDataSource struct{ client *client.Client }

func NewS3BucketDataSource() datasource.DataSource { return &S3BucketDataSource{} }

func (d *S3BucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_bucket"
}

func (d *S3BucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"compliance_retention_days": schema.Int64Attribute{Computed: true},
		"is_lock_enabled":           schema.BoolAttribute{Computed: true},
		"is_public":                 schema.BoolAttribute{Computed: true},
		"is_versioned":              schema.BoolAttribute{Computed: true},
		"object_expiration_days":    schema.Int64Attribute{Computed: true},
		"quota_gib":                 schema.Int64Attribute{Computed: true},
		"region":                    schema.StringAttribute{Computed: true},
		"tier":                      schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"endpoint_url": schema.StringAttribute{Computed: true}}),
	}}
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
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var apiData map[string]interface{}
	var err error
	idSet := !state.ID.IsNull() && state.ID.ValueString() != ""
	nameSet := !state.Name.IsNull() && state.Name.ValueString() != ""
	if idSet == nameSet {
		resp.Diagnostics.AddError("Invalid lookup", "exactly one of \"id\" or \"name\" must be set")
		return
	}
	if idSet {
		apiData, err = d.client.Get(ctx, "/api/v1/s3-bucket", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/s3-bucket", state.Name.ValueString())
	}
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
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	spec := getSpec(apiData)
	state.Spec.ComplianceRetentionDays = getInt64(spec, "complianceRetentionDays")
	state.Spec.IsLockEnabled = getBool(spec, "isLockEnabled")
	state.Spec.IsPublic = getBool(spec, "isPublic")
	state.Spec.IsVersioned = getBool(spec, "isVersioned")
	state.Spec.ObjectExpirationDays = getInt64(spec, "objectExpirationDays")
	state.Spec.QuotaGib = getInt64(spec, "quotaGiB")
	state.Spec.Region = getString(spec, "region")
	state.Spec.Tier = getString(spec, "tier")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"endpoint_url": types.StringType,
		},
		map[string]attr.Value{
			"endpoint_url": getStringFromInfo(apiData, "endpointUrl"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
