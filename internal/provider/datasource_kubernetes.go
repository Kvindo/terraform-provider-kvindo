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

// KubernetesDataSourceModel describes the data source data model.
type KubernetesDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	Tier types.String `tfsdk:"tier"`
	AssignPublicIpV4 types.Bool `tfsdk:"assign_public_ip_v4"`
	Version types.String `tfsdk:"version"`
	ControlPlaneLocations types.List `tfsdk:"control_plane_locations"`
	InfoState types.String `tfsdk:"info_state"`
	InfoApiServerUrl types.String `tfsdk:"info_api_server_url"`
}

type KubernetesDataSource struct {
	client *client.Client
}

func NewKubernetesDataSource() datasource.DataSource {
	return &KubernetesDataSource{}
}

func (d *KubernetesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes"
}

func (d *KubernetesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["tier"] = schema.StringAttribute{Computed: true}
	attrs["assign_public_ip_v4"] = schema.BoolAttribute{Computed: true}
	attrs["version"] = schema.StringAttribute{Computed: true}
	attrs["control_plane_locations"] = schema.ListNestedAttribute{
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"vpc_subnet_id": schema.StringAttribute{Computed: true},
				},
			},
		}
	attrs["info_state"] = schema.StringAttribute{Computed: true}
	attrs["info_api_server_url"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *KubernetesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KubernetesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state KubernetesDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/kubernetes", state.ID.ValueString())
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
	state.AssignPublicIpV4 = getBool(apiData, "assignPublicIpV4")
	state.Version = getString(apiData, "version")
	{
		rawControlPlaneLocations, _ := apiData["controlPlaneLocations"].([]interface{})
		attrTypes := map[string]attr.Type{
			"vpc_subnet_id": types.StringType,
		}
		objs := make([]attr.Value, 0, len(rawControlPlaneLocations))
		for _, item := range rawControlPlaneLocations {
			if m, ok := item.(map[string]interface{}); ok {
				attrs := map[string]attr.Value{
					"vpc_subnet_id": getString(m, "vpcSubnetId"),
				}
				obj, _ := types.ObjectValue(attrTypes, attrs)
				objs = append(objs, obj)
			}
		}
		state.ControlPlaneLocations, _ = types.ListValue(types.ObjectType{AttrTypes: attrTypes}, objs)
	}
	state.InfoState = getStringFromInfo(apiData, "state")
	state.InfoApiServerUrl = getStringFromInfo(apiData, "apiserverurl")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
