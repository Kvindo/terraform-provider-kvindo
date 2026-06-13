package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types

// VpcPeeringExternalPeerResourceModel describes the resource data model.
type VpcPeeringExternalPeerResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	VpcPeeringId types.String `tfsdk:"vpc_peering_id"`
	SshUser types.String `tfsdk:"ssh_user"`
	SshPort types.Int64 `tfsdk:"ssh_port"`
	SshIpV4 types.String `tfsdk:"ssh_ip_v4"`
	PrivateIpV4 types.String `tfsdk:"private_ip_v4"`
	IpV4Cidrs types.List `tfsdk:"ip_v4_cidrs"`
	SshPrivateKeyId types.String `tfsdk:"ssh_private_key_id"`
	Info types.Object `tfsdk:"info"`
}

// VpcPeeringExternalPeerResource defines the resource implementation.
type VpcPeeringExternalPeerResource struct {
	client *client.Client
}

func NewVpcPeeringExternalPeerResource() resource.Resource {
	return &VpcPeeringExternalPeerResource{}
}

func (r *VpcPeeringExternalPeerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peering_external_peer"
}

func (r *VpcPeeringExternalPeerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["vpc_peering_id"] = schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["ssh_user"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["ssh_port"] = schema.Int64Attribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		}
	attrs["ssh_ip_v4"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["private_ip_v4"] = schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	attrs["ip_v4_cidrs"] = schema.ListAttribute{
			Optional: true,
				Computed: true,
				ElementType: types.StringType,
		}
	attrs["ssh_private_key_id"] = schema.StringAttribute{
			Optional: true,
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *VpcPeeringExternalPeerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildVpcPeeringExternalPeerRequestMap(ctx context.Context, plan VpcPeeringExternalPeerResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.VpcPeeringId.IsNull() && !plan.VpcPeeringId.IsUnknown() {
		m["vpcPeeringId"] = plan.VpcPeeringId.ValueString()
	}
	if !plan.SshUser.IsNull() && !plan.SshUser.IsUnknown() {
		m["sshUser"] = plan.SshUser.ValueString()
	}
	if !plan.SshPort.IsNull() && !plan.SshPort.IsUnknown() {
		m["sshPort"] = plan.SshPort.ValueInt64()
	}
	if !plan.SshIpV4.IsNull() && !plan.SshIpV4.IsUnknown() {
		m["sshIpV4"] = plan.SshIpV4.ValueString()
	}
	if !plan.PrivateIpV4.IsNull() && !plan.PrivateIpV4.IsUnknown() {
		m["privateIpV4"] = plan.PrivateIpV4.ValueString()
	}
	if !plan.IpV4Cidrs.IsNull() && !plan.IpV4Cidrs.IsUnknown() {
		m["ipV4Cidrs"] = stringListToInterface(ctx, plan.IpV4Cidrs)
	}
	if !plan.SshPrivateKeyId.IsNull() && !plan.SshPrivateKeyId.IsUnknown() {
		m["sshPrivateKeyId"] = plan.SshPrivateKeyId.ValueString()
	}
	return m
}

func populateVpcPeeringExternalPeerState(ctx context.Context, data map[string]interface{}, state *VpcPeeringExternalPeerResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	state.VpcPeeringId = getString(data, "vpcPeeringId")
	state.SshUser = getString(data, "sshUser")
	state.SshPort = getInt64(data, "sshPort")
	state.SshIpV4 = getString(data, "sshIpV4")
	state.PrivateIpV4 = getString(data, "privateIpV4")
	state.IpV4Cidrs = getStringList(ctx, data, "ipV4Cidrs")
	state.SshPrivateKeyId = getString(data, "sshPrivateKeyId")
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *VpcPeeringExternalPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VpcPeeringExternalPeerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildVpcPeeringExternalPeerRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/vpc-peering-external-peer", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/vpc-peering-external-peer", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/vpc-peering-external-peer", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateVpcPeeringExternalPeerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VpcPeeringExternalPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VpcPeeringExternalPeerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/vpc-peering-external-peer", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateVpcPeeringExternalPeerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VpcPeeringExternalPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VpcPeeringExternalPeerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state VpcPeeringExternalPeerResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildVpcPeeringExternalPeerRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/vpc-peering-external-peer", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/vpc-peering-external-peer", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/vpc-peering-external-peer", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateVpcPeeringExternalPeerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VpcPeeringExternalPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VpcPeeringExternalPeerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/vpc-peering-external-peer", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/vpc-peering-external-peer", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *VpcPeeringExternalPeerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state VpcPeeringExternalPeerResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/vpc-peering-external-peer", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateVpcPeeringExternalPeerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
