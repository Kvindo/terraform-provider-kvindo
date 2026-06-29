package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

type S3BucketSpecModel struct {
	ComplianceRetentionDays types.Int64  `tfsdk:"compliance_retention_days"`
	IsLockEnabled           types.Bool   `tfsdk:"is_lock_enabled"`
	IsPublic                types.Bool   `tfsdk:"is_public"`
	IsVersioned             types.Bool   `tfsdk:"is_versioned"`
	ObjectExpirationDays    types.Int64  `tfsdk:"object_expiration_days"`
	QuotaGib                types.Int64  `tfsdk:"quota_gib"`
	Region                  types.String `tfsdk:"region"`
	Tier                    types.String `tfsdk:"tier"`
}

type S3BucketResourceModel struct {
	ID       types.String      `tfsdk:"id"`
	Metadata metadataModel     `tfsdk:"metadata"`
	Spec     S3BucketSpecModel `tfsdk:"spec"`
	Status   types.Object      `tfsdk:"status"`
}

type S3BucketResource struct{ client *client.Client }

func NewS3BucketResource() resource.Resource { return &S3BucketResource{} }

func (r *S3BucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_bucket"
}

func S3BucketResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"compliance_retention_days": schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"is_lock_enabled":           schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"is_public":                 schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"is_versioned":              schema.BoolAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
		"object_expiration_days":    schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"quota_gib":                 schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"region":                    schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"tier":                      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(map[string]schema.Attribute{"endpoint_url": schema.StringAttribute{Computed: true}}),
	}
}

func (r *S3BucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: S3BucketResourceSchemaAttrs()}
}

func (r *S3BucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*KvindoProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *KvindoProviderData, got %T", req.ProviderData))
		return
	}
	r.client = pd.Client
}

func buildS3BucketRequestMap(ctx context.Context, plan S3BucketResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.ComplianceRetentionDays.IsNull() && !plan.Spec.ComplianceRetentionDays.IsUnknown() {
		spec["complianceRetentionDays"] = plan.Spec.ComplianceRetentionDays.ValueInt64()
	}
	if !plan.Spec.IsLockEnabled.IsNull() && !plan.Spec.IsLockEnabled.IsUnknown() {
		spec["isLockEnabled"] = plan.Spec.IsLockEnabled.ValueBool()
	}
	if !plan.Spec.IsPublic.IsNull() && !plan.Spec.IsPublic.IsUnknown() {
		spec["isPublic"] = plan.Spec.IsPublic.ValueBool()
	}
	if !plan.Spec.IsVersioned.IsNull() && !plan.Spec.IsVersioned.IsUnknown() {
		spec["isVersioned"] = plan.Spec.IsVersioned.ValueBool()
	}
	if !plan.Spec.ObjectExpirationDays.IsNull() && !plan.Spec.ObjectExpirationDays.IsUnknown() {
		spec["objectExpirationDays"] = plan.Spec.ObjectExpirationDays.ValueInt64()
	}
	if !plan.Spec.QuotaGib.IsNull() && !plan.Spec.QuotaGib.IsUnknown() {
		spec["quotaGiB"] = plan.Spec.QuotaGib.ValueInt64()
	}
	if !plan.Spec.Region.IsNull() && !plan.Spec.Region.IsUnknown() {
		spec["region"] = plan.Spec.Region.ValueString()
	}
	if !plan.Spec.Tier.IsNull() && !plan.Spec.Tier.IsUnknown() {
		spec["tier"] = plan.Spec.Tier.ValueString()
	}
	return m
}

func populateS3BucketState(ctx context.Context, data map[string]interface{}, state *S3BucketResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.ComplianceRetentionDays = getInt64(spec, "complianceRetentionDays")
	state.Spec.IsLockEnabled = getBool(spec, "isLockEnabled")
	state.Spec.IsPublic = getBool(spec, "isPublic")
	state.Spec.IsVersioned = getBool(spec, "isVersioned")
	state.Spec.ObjectExpirationDays = getInt64(spec, "objectExpirationDays")
	state.Spec.QuotaGib = getInt64(spec, "quotaGiB")
	state.Spec.Region = getString(spec, "region")
	state.Spec.Tier = getString(spec, "tier")
	state.Status = buildInfoObj(data,
		map[string]attr.Type{
			"endpoint_url": types.StringType,
		},
		map[string]attr.Value{
			"endpoint_url": getStringFromInfo(data, "endpointUrl"),
		})
	return nil
}

func (r *S3BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan S3BucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(newULID())
	body := buildS3BucketRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/s3-bucket", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/s3-bucket", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}
	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/s3-bucket", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateS3BucketState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *S3BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state S3BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/s3-bucket", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateS3BucketState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *S3BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state S3BucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildS3BucketRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/s3-bucket", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/s3-bucket", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}
	apiData, err := r.client.Get(ctx, "/api/v1/s3-bucket", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateS3BucketState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *S3BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state S3BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	modResp, err := r.client.Delete(ctx, "/api/v1/s3-bucket", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/s3-bucket", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *S3BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state S3BucketResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/s3-bucket", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateS3BucketState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
