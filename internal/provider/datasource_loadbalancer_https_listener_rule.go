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

// LoadbalancerHttpsListenerRuleDataSourceModel describes the data source data model.
type LoadbalancerHttpsListenerRuleDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	FolderID         types.String `tfsdk:"folder_id"`
	DeleteProtection types.Bool   `tfsdk:"delete_protection"`
	Labels           types.Map    `tfsdk:"labels"`
	HttpsListenerId types.String `tfsdk:"https_listener_id"`
	Order types.Int64 `tfsdk:"order"`
	MatchPath types.String `tfsdk:"match_path"`
	MatchPathMatchType types.String `tfsdk:"match_path_match_type"`
	ActionType types.String `tfsdk:"action_type"`
	ActionJson types.String `tfsdk:"action_json"`
	InfoState types.String `tfsdk:"info_state"`
}

type LoadbalancerHttpsListenerRuleDataSource struct {
	client *client.Client
}

func NewLoadbalancerHttpsListenerRuleDataSource() datasource.DataSource {
	return &LoadbalancerHttpsListenerRuleDataSource{}
}

func (d *LoadbalancerHttpsListenerRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_https_listener_rule"
}

func (d *LoadbalancerHttpsListenerRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := commonDatasourceSchemaAttributes()

	attrs["https_listener_id"] = schema.StringAttribute{Computed: true}
	attrs["order"] = schema.Int64Attribute{Computed: true}
	attrs["match_path"] = schema.StringAttribute{Computed: true}
	attrs["match_path_match_type"] = schema.StringAttribute{Computed: true}
	attrs["action_type"] = schema.StringAttribute{Computed: true}
	attrs["action_json"] = schema.StringAttribute{Computed: true}
	attrs["info_state"] = schema.StringAttribute{Computed: true}

	resp.Schema = schema.Schema{Attributes: attrs}
}

func (d *LoadbalancerHttpsListenerRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LoadbalancerHttpsListenerRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state LoadbalancerHttpsListenerRuleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiData, err := d.client.Get(ctx, "/api/v1/loadbalancer-https-listener-rule", state.ID.ValueString())
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
	state.HttpsListenerId = getString(apiData, "httpsListenerId")
	state.Order = getInt64(apiData, "order")
	state.MatchPath = getString(apiData, "matchPath")
	state.MatchPathMatchType = getString(apiData, "matchPathMatchType")
	state.ActionType = getString(apiData, "actionType")
	state.ActionJson = getString(apiData, "actionJson")
	state.InfoState = getStringFromInfo(apiData, "state")
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
