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
// attr package used for list/object types

// SecurityGroupDataSourceModel describes the data source data model.
type SecurityGroupDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Ingress types.List `tfsdk:"ingress"`
	Egress types.List `tfsdk:"egress"`
	InfoState types.String `tfsdk:"info_state"`
}

type SecurityGroupDataSource struct {
	client *client.Client
}

func NewSecurityGroupDataSource() datasource.DataSource {
	return &SecurityGroupDataSource{}
}

func (d *SecurityGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (d *SecurityGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["ingress"] = schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"ports": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"ipv4_blocks": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"ipv6_blocks": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"action": schema.StringAttribute{Computed: true},
				},
			},
		}
	attrs["egress"] = schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"ports": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"ipv4_blocks": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"ipv6_blocks": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"action": schema.StringAttribute{Computed: true},
				},
			},
		}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *SecurityGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecurityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SecurityGroupDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/security-group", state.ID.ValueString())
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
	{
		rawIngress, _ := apiData["ingress"].([]interface{})
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
		rawEgress, _ := apiData["egress"].([]interface{})
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
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
