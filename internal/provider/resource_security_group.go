package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kvindo/terraform-provider-kvindo/internal/client"
)

var _ = fmt.Sprintf
// attr package used for list/object types
var _ = listplanmodifier.UseStateForUnknown

// SecurityGroupResourceModel describes the resource data model.
type SecurityGroupResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Ingress types.List `tfsdk:"ingress"`
	Egress types.List `tfsdk:"egress"`
	Info types.Object `tfsdk:"info"`
}

// SecurityGroupIngressModel is the nested object model for ingress.
type SecurityGroupIngressModel struct {
	Ports types.List `tfsdk:"ports"`
	Ipv4Blocks types.List `tfsdk:"ipv4_blocks"`
	Ipv6Blocks types.List `tfsdk:"ipv6_blocks"`
	Action types.String `tfsdk:"action"`
}

// SecurityGroupEgressModel is the nested object model for egress.
type SecurityGroupEgressModel struct {
	Ports types.List `tfsdk:"ports"`
	Ipv4Blocks types.List `tfsdk:"ipv4_blocks"`
	Ipv6Blocks types.List `tfsdk:"ipv6_blocks"`
	Action types.String `tfsdk:"action"`
}

// SecurityGroupResource defines the resource implementation.
type SecurityGroupResource struct {
	client *client.Client
}

func NewSecurityGroupResource() resource.Resource {
	return &SecurityGroupResource{}
}

func (r *SecurityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *SecurityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := commonSchemaAttributes()

	attrs["ingress"] = schema.ListNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"ports": schema.ListAttribute{
			Computed: true,
				ElementType: types.StringType,
		},
					"ipv4_blocks": schema.ListAttribute{
			Computed: true,
				ElementType: types.StringType,
		},
					"ipv6_blocks": schema.ListAttribute{
			Computed: true,
				ElementType: types.StringType,
		},
					"action": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
				},
			},
		}
	attrs["egress"] = schema.ListNestedAttribute{
			Optional: true,
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"ports": schema.ListAttribute{
			Computed: true,
				ElementType: types.StringType,
		},
					"ipv4_blocks": schema.ListAttribute{
			Computed: true,
				ElementType: types.StringType,
		},
					"ipv6_blocks": schema.ListAttribute{
			Computed: true,
				ElementType: types.StringType,
		},
					"action": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
				},
			},
		}
	attrs["info"] = commonInfoSchema(map[string]schema.Attribute{"state": schema.StringAttribute{Computed: true}})

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (r *SecurityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildSecurityGroupRequestMap(ctx context.Context, plan SecurityGroupResourceModel) map[string]interface{} {
	m := buildCommonRequestMap(plan.ID.ValueString(), plan.Name.ValueString(), plan.Description, plan.FolderID, plan.DeleteProtection, plan.Labels, ctx)
	if !plan.Ingress.IsNull() && !plan.Ingress.IsUnknown() {
		var items []map[string]interface{}
		for _, elem := range plan.Ingress.Elements() {
			if ov, ok := elem.(types.Object); ok {
				item := map[string]interface{}{}
				if v, ok := ov.Attributes()["ports"]; ok {
					if lv, ok := v.(types.List); ok && !lv.IsNull() {
						item["ports"] = stringListToInterface(ctx, lv)
					}
				}
				if v, ok := ov.Attributes()["ipv4_blocks"]; ok {
					if lv, ok := v.(types.List); ok && !lv.IsNull() {
						item["ipv4Blocks"] = stringListToInterface(ctx, lv)
					}
				}
				if v, ok := ov.Attributes()["ipv6_blocks"]; ok {
					if lv, ok := v.(types.List); ok && !lv.IsNull() {
						item["ipv6Blocks"] = stringListToInterface(ctx, lv)
					}
				}
				if v, ok := ov.Attributes()["action"]; ok {
					if sv, ok := v.(types.String); ok && !sv.IsNull() {
						item["action"] = sv.ValueString()
					}
				}
				items = append(items, item)
			}
		}
		m["ingress"] = items
	}
	if !plan.Egress.IsNull() && !plan.Egress.IsUnknown() {
		var items []map[string]interface{}
		for _, elem := range plan.Egress.Elements() {
			if ov, ok := elem.(types.Object); ok {
				item := map[string]interface{}{}
				if v, ok := ov.Attributes()["ports"]; ok {
					if lv, ok := v.(types.List); ok && !lv.IsNull() {
						item["ports"] = stringListToInterface(ctx, lv)
					}
				}
				if v, ok := ov.Attributes()["ipv4_blocks"]; ok {
					if lv, ok := v.(types.List); ok && !lv.IsNull() {
						item["ipv4Blocks"] = stringListToInterface(ctx, lv)
					}
				}
				if v, ok := ov.Attributes()["ipv6_blocks"]; ok {
					if lv, ok := v.(types.List); ok && !lv.IsNull() {
						item["ipv6Blocks"] = stringListToInterface(ctx, lv)
					}
				}
				if v, ok := ov.Attributes()["action"]; ok {
					if sv, ok := v.(types.String); ok && !sv.IsNull() {
						item["action"] = sv.ValueString()
					}
				}
				items = append(items, item)
			}
		}
		m["egress"] = items
	}
	return m
}

func populateSecurityGroupState(ctx context.Context, data map[string]interface{}, state *SecurityGroupResourceModel) error {
	if err := setCommonFields(ctx, data, &state.ID, &state.Name, &state.Description, &state.FolderID, &state.DeleteProtection, &state.Labels); err != nil {
		return err
	}
	{
		rawIngress, _ := data["ingress"].([]interface{})
		attrTypes := map[string]attr.Type{
			"ports": types.ListType{ElemType: types.StringType},
			"ipv4_blocks": types.ListType{ElemType: types.StringType},
			"ipv6_blocks": types.ListType{ElemType: types.StringType},
			"action": types.StringType,
		}
		objs := make([]attr.Value, 0, len(rawIngress))
		for _, item := range rawIngress {
			if m, ok := item.(map[string]interface{}); ok {
				attrs := map[string]attr.Value{
					"ports": getStringList(ctx, m, "ports"),
					"ipv4_blocks": getStringList(ctx, m, "ipv4Blocks"),
					"ipv6_blocks": getStringList(ctx, m, "ipv6Blocks"),
					"action": getString(m, "action"),
				}
				obj, _ := types.ObjectValue(attrTypes, attrs)
				objs = append(objs, obj)
			}
		}
		state.Ingress, _ = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objs)
	}
	{
		rawEgress, _ := data["egress"].([]interface{})
		attrTypes := map[string]attr.Type{
			"ports": types.ListType{ElemType: types.StringType},
			"ipv4_blocks": types.ListType{ElemType: types.StringType},
			"ipv6_blocks": types.ListType{ElemType: types.StringType},
			"action": types.StringType,
		}
		objs := make([]attr.Value, 0, len(rawEgress))
		for _, item := range rawEgress {
			if m, ok := item.(map[string]interface{}); ok {
				attrs := map[string]attr.Value{
					"ports": getStringList(ctx, m, "ports"),
					"ipv4_blocks": getStringList(ctx, m, "ipv4Blocks"),
					"ipv6_blocks": getStringList(ctx, m, "ipv6Blocks"),
					"action": getString(m, "action"),
				}
				obj, _ := types.ObjectValue(attrTypes, attrs)
				objs = append(objs, obj)
			}
		}
		state.Egress, _ = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objs)
	}
	state.Info = simpleStateInfoObj(data)
	return nil
}

func (r *SecurityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(newULID())
	body := buildSecurityGroupRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/security-group", body)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/security-group", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Create Poll Error", err.Error())
		return
	}

	resourceId := modResp.ResourceId
	if resourceId == "" {
		resourceId = plan.ID.ValueString()
	}
	apiData, err := r.client.Get(ctx, "/api/v1/security-group", resourceId)
	if err != nil {
		resp.Diagnostics.AddError("Read After Create Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Create Error", "resource not found after creation")
		return
	}
	if err := populateSecurityGroupState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/security-group", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if err := populateSecurityGroupState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state SecurityGroupResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	body := buildSecurityGroupRequestMap(ctx, plan)
	modResp, err := r.client.Put(ctx, "/api/v1/security-group", body)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/security-group", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Update Poll Error", err.Error())
		return
	}

	apiData, err := r.client.Get(ctx, "/api/v1/security-group", plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read After Update Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Read After Update Error", "resource not found after update")
		return
	}
	if err := populateSecurityGroupState(ctx, apiData, &plan); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	modResp, err := r.client.Delete(ctx, "/api/v1/security-group", state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}
	if err := r.client.PollUntilDone(ctx, "/api/v1/security-group", modResp.RequestId); err != nil {
		resp.Diagnostics.AddError("Delete Poll Error", err.Error())
		return
	}
}

func (r *SecurityGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	var state SecurityGroupResourceModel
	state.ID = types.StringValue(req.ID)
	apiData, err := r.client.Get(ctx, "/api/v1/security-group", req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Import Error", "resource not found")
		return
	}
	if err := populateSecurityGroupState(ctx, apiData, &state); err != nil {
		resp.Diagnostics.AddError("State Population Error", err.Error())
		return
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
