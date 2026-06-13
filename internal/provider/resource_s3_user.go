package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf

// S3UserResourceModel describes the resource data model.
type S3UserResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	BucketId         types.String `tfsdk:"bucket_id"`
	AccessPolicyIds  types.List   `tfsdk:"access_policy_ids"`
	Info types.Object `tfsdk:"info"`
}

// S3UserResource defines the resource implementation.
type S3UserResource struct {
	client *client.Client
}

func NewS3UserResource() resource.Resource {
	return &S3UserResource{}
}

func (r *S3UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_user"
}

func (r *S3UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["bucket_id"] = schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["access_policy_ids"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "access_key": schema.StringAttribute{Computed: true}, "secret_key": schema.StringAttribute{Computed: true, Sensitive: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *S3UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildS3UserRequestMap(ctx context.Context, plan S3UserResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.BucketId.IsNull() && !plan.BucketId.IsUnknown() {
		m["bucketId"] = plan.BucketId.ValueString()
	}
	if !plan.AccessPolicyIds.IsNull() && !plan.AccessPolicyIds.IsUnknown() {
		m["accessPolicyIds"] = stringListToInterface(ctx, plan.AccessPolicyIds)
	}
	return m
}

func populateS3UserState(ctx context.Context, data map[string]interface{}, state *S3UserResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.BucketId = getString(data, "bucketId")
	state.AccessPolicyIds = getStringList(ctx, data, "accessPolicyIds")
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "access_key": types.StringType, "secret_key": types.StringType}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "access_key": getStringFromInfo(data, "accessKey"), "secret_key": getStringFromInfo(data, "secretKey")})
	return nil
}

func (r *S3UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan S3UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildS3UserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/s3-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/s3-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/s3-user", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateS3UserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *S3UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state S3UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/s3-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateS3UserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *S3UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan S3UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state S3UserResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildS3UserRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/s3-user", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/s3-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/s3-user", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateS3UserState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *S3UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state S3UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/s3-user", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/s3-user", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *S3UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state S3UserResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/s3-user", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateS3UserState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
