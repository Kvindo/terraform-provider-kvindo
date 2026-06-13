package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// PostgresqlStandaloneResourceModel describes the resource data model.
type PostgresqlStandaloneResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	FolderID            types.String `tfsdk:"folder_id"`
	DeleteProtection    types.Bool   `tfsdk:"delete_protection"`
	Labels              types.Map    `tfsdk:"labels"`
	Tier                types.String `tfsdk:"tier"`
	Version             types.String `tfsdk:"version"`
	RootPassword        types.String `tfsdk:"root_password"`
	ParametersSetId     types.String `tfsdk:"parameters_set_id"`
	BackupRetentionDays types.Int64  `tfsdk:"backup_retention_days"`
	FloatingIpId        types.String `tfsdk:"floating_ip_id"`
	VpcSubnetId         types.String `tfsdk:"vpc_subnet_id"`
	VmState             types.String `tfsdk:"vm_state"`
	VmOfferId           types.String `tfsdk:"vm_offer_id"`
	VolumeOfferId       types.String `tfsdk:"volume_offer_id"`
	VolumeSizeGib       types.Int64  `tfsdk:"volume_size_gib"`
	Info types.Object `tfsdk:"info"`
}

// PostgresqlStandaloneResource defines the resource implementation.
type PostgresqlStandaloneResource struct {
	client *client.Client
}

func NewPostgresqlStandaloneResource() resource.Resource {
	return &PostgresqlStandaloneResource{}
}

func (r *PostgresqlStandaloneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_standalone"
}

func (r *PostgresqlStandaloneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["version"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["root_password"] = schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
		}
	attrs["parameters_set_id"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["backup_retention_days"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["floating_ip_id"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["vpc_subnet_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vm_state"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["vm_offer_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["volume_offer_id"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["volume_size_gib"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}, "root_user_name": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "port": schema.Int64Attribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *PostgresqlStandaloneResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildPostgresqlStandaloneRequestMap(ctx context.Context, plan PostgresqlStandaloneResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.Tier.IsNull() && !plan.Tier.IsUnknown() {
		m["tier"] = plan.Tier.ValueString()
	}
	if !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		m["version"] = plan.Version.ValueString()
	}
	if !plan.RootPassword.IsNull() && !plan.RootPassword.IsUnknown() {
		m["rootPassword"] = plan.RootPassword.ValueString()
	}
	if !plan.ParametersSetId.IsNull() && !plan.ParametersSetId.IsUnknown() {
		m["parametersSetId"] = plan.ParametersSetId.ValueString()
	}
	if !plan.BackupRetentionDays.IsNull() && !plan.BackupRetentionDays.IsUnknown() {
		m["backupRetentionDays"] = plan.BackupRetentionDays.ValueInt64()
	}
	if !plan.FloatingIpId.IsNull() && !plan.FloatingIpId.IsUnknown() {
		m["floatingIpId"] = plan.FloatingIpId.ValueString()
	}
	if !plan.VpcSubnetId.IsNull() && !plan.VpcSubnetId.IsUnknown() {
		m["vpcSubnetId"] = plan.VpcSubnetId.ValueString()
	}
	if !plan.VmState.IsNull() && !plan.VmState.IsUnknown() {
		m["vmState"] = plan.VmState.ValueString()
	}
	if !plan.VmOfferId.IsNull() && !plan.VmOfferId.IsUnknown() {
		m["vmOfferId"] = plan.VmOfferId.ValueString()
	}
	if !plan.VolumeOfferId.IsNull() && !plan.VolumeOfferId.IsUnknown() {
		m["volumeOfferId"] = plan.VolumeOfferId.ValueString()
	}
	if !plan.VolumeSizeGib.IsNull() && !plan.VolumeSizeGib.IsUnknown() {
		m["volumeSizeGiB"] = plan.VolumeSizeGib.ValueInt64()
	}
	return m
}

func populatePostgresqlStandaloneState(ctx context.Context, data map[string]interface{}, state *PostgresqlStandaloneResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.Tier = getString(data, "tier")
	state.Version = getString(data, "version")
	state.ParametersSetId = getString(data, "parametersSetId")
	state.BackupRetentionDays = getInt64(data, "backupRetentionDays")
	state.FloatingIpId = getString(data, "floatingIpId")
	state.VpcSubnetId = getString(data, "vpcSubnetId")
	state.VmState = getString(data, "vmState")
	state.VmOfferId = getString(data, "vmOfferId")
	state.VolumeOfferId = getString(data, "volumeOfferId")
	state.VolumeSizeGib = getInt64(data, "volumeSizeGiB")
	state.Info, _ = types.ObjectValue(map[string]attr.Type{"state": types.StringType, "root_user_name": types.StringType, "public_ip_v4": types.StringType, "private_ip_v4": types.StringType, "port": types.Int64Type}, map[string]attr.Value{"state": getStringFromInfo(data, "state"), "root_user_name": getStringFromInfo(data, "rootUserName"), "public_ip_v4": getStringFromInfo(data, "publicipv4"), "private_ip_v4": getStringFromInfo(data, "privateipv4"), "port": getInt64FromInfo(data, "port")})
	return nil
}

func (r *PostgresqlStandaloneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PostgresqlStandaloneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildPostgresqlStandaloneRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/postgresql-standalone", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/postgresql-standalone", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/postgresql-standalone", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populatePostgresqlStandaloneState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresqlStandaloneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PostgresqlStandaloneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/postgresql-standalone", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populatePostgresqlStandaloneState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresqlStandaloneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PostgresqlStandaloneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state PostgresqlStandaloneResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildPostgresqlStandaloneRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/postgresql-standalone", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/postgresql-standalone", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/postgresql-standalone", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populatePostgresqlStandaloneState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresqlStandaloneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PostgresqlStandaloneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/postgresql-standalone", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/postgresql-standalone", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *PostgresqlStandaloneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state PostgresqlStandaloneResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/postgresql-standalone", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populatePostgresqlStandaloneState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
