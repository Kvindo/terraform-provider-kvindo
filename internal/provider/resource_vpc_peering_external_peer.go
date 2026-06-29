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

type VpcPeeringExternalPeerSpecModel struct {
	IpV4Cidrs       types.List   `tfsdk:"ip_v4_cidrs"`
	PrivateIpV4     types.String `tfsdk:"private_ip_v4"`
	SshIpV4         types.String `tfsdk:"ssh_ip_v4"`
	SshPort         types.Int64  `tfsdk:"ssh_port"`
	SshPrivateKeyId types.String `tfsdk:"ssh_private_key_id"`
	SshUser         types.String `tfsdk:"ssh_user"`
	VpcPeeringId    types.String `tfsdk:"vpc_peering_id"`
}

type VpcPeeringExternalPeerResourceModel struct {
	ID       types.String                    `tfsdk:"id"`
	Metadata metadataModel                   `tfsdk:"metadata"`
	Spec     VpcPeeringExternalPeerSpecModel `tfsdk:"spec"`
	Status   types.Object                    `tfsdk:"status"`
}

type VpcPeeringExternalPeerResource struct{ client *client.Client }

func NewVpcPeeringExternalPeerResource() resource.Resource { return &VpcPeeringExternalPeerResource{} }

func (r *VpcPeeringExternalPeerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_peering_external_peer"
}

func VpcPeeringExternalPeerResourceSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"ip_v4_cidrs":        schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType},
		"private_ip_v4":      schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"ssh_ip_v4":          schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"ssh_port":           schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
		"ssh_private_key_id": schema.StringAttribute{Optional: true},
		"ssh_user":           schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"vpc_peering_id":     schema.StringAttribute{Required: true},
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Required: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

func (r *VpcPeeringExternalPeerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: VpcPeeringExternalPeerResourceSchemaAttrs()}
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
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Metadata.Name.ValueString(), plan.Metadata.Description, plan.Metadata.FolderID, plan.Metadata.DeleteProtection, plan.Metadata.Labels, ctx)
	spec := m["spec"].(map[string]interface{})
	if !plan.Spec.IpV4Cidrs.IsNull() && !plan.Spec.IpV4Cidrs.IsUnknown() {
		spec["ipV4Cidrs"] = stringListToInterface(ctx, plan.Spec.IpV4Cidrs)
	}
	if !plan.Spec.PrivateIpV4.IsNull() && !plan.Spec.PrivateIpV4.IsUnknown() {
		spec["privateIpV4"] = plan.Spec.PrivateIpV4.ValueString()
	}
	if !plan.Spec.SshIpV4.IsNull() && !plan.Spec.SshIpV4.IsUnknown() {
		spec["sshIpV4"] = plan.Spec.SshIpV4.ValueString()
	}
	if !plan.Spec.SshPort.IsNull() && !plan.Spec.SshPort.IsUnknown() {
		spec["sshPort"] = plan.Spec.SshPort.ValueInt64()
	}
	if !plan.Spec.SshPrivateKeyId.IsNull() && !plan.Spec.SshPrivateKeyId.IsUnknown() {
		spec["sshPrivateKeyId"] = plan.Spec.SshPrivateKeyId.ValueString()
	}
	if !plan.Spec.SshUser.IsNull() && !plan.Spec.SshUser.IsUnknown() {
		spec["sshUser"] = plan.Spec.SshUser.ValueString()
	}
	if !plan.Spec.VpcPeeringId.IsNull() && !plan.Spec.VpcPeeringId.IsUnknown() {
		spec["vpcPeeringId"] = plan.Spec.VpcPeeringId.ValueString()
	}
	return m
}

func populateVpcPeeringExternalPeerState(ctx context.Context, data map[string]interface{}, state *VpcPeeringExternalPeerResourceModel) error {
	if err := setCommonFieldsNested(ctx, data, &state.Metadata); err != nil {
		return err
	}
	state.ID = state.Metadata.ID
	spec := getSpec(data)
	state.Spec.IpV4Cidrs = getStringList(ctx, spec, "ipV4Cidrs")
	state.Spec.PrivateIpV4 = getString(spec, "privateIpV4")
	state.Spec.SshIpV4 = getString(spec, "sshIpV4")
	state.Spec.SshPort = getInt64(spec, "sshPort")
	state.Spec.SshPrivateKeyId = getString(spec, "sshPrivateKeyId")
	state.Spec.SshUser = getString(spec, "sshUser")
	state.Spec.VpcPeeringId = getString(spec, "vpcPeeringId")
	state.Status = simpleStateInfoObj(data)
	return nil
}

func (r *VpcPeeringExternalPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VpcPeeringExternalPeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *VpcPeeringExternalPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VpcPeeringExternalPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *VpcPeeringExternalPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VpcPeeringExternalPeerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
		resp.Diagnostics.AddError("Read After Update Error", "not found")
		return
	}
	if err := populateVpcPeeringExternalPeerState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *VpcPeeringExternalPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VpcPeeringExternalPeerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
	var state VpcPeeringExternalPeerResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/vpc-peering-external-peer", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "not found")
		return
	}
	if err := populateVpcPeeringExternalPeerState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
