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

type ValkeyDataSourceModel struct {
	ID       types.String     `tfsdk:"id"`
	Name     types.String     `tfsdk:"name"`
	Metadata *metadataModel   `tfsdk:"metadata"`
	Spec     *ValkeySpecModel `tfsdk:"spec"`
	Status   types.Object     `tfsdk:"status"`
}

type ValkeyDataSource struct{ client *client.Client }

func NewValkeyDataSource() datasource.DataSource { return &ValkeyDataSource{} }

func (d *ValkeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_valkey"
}

func (d *ValkeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"create_public_ipv4": schema.BoolAttribute{Computed: true},
		"parameters_set_id":  schema.StringAttribute{Computed: true},
		"replicas_per_shard": schema.Int64Attribute{Computed: true},
		"root_password":      schema.StringAttribute{Computed: true, Sensitive: true},
		"shards":             listObjDatasourceSchema(valkeyShardsObjFields),
		"use_fqdn":           schema.BoolAttribute{Computed: true},
		"valkey_version":     schema.StringAttribute{Computed: true},
		"vm_offer_id":        schema.StringAttribute{Computed: true},
		"volume_offer_id":    schema.StringAttribute{Computed: true},
		"volume_size_gib":    schema.Int64Attribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"anti_affinity_message": schema.StringAttribute{Computed: true}, "anti_affinity_ok": schema.BoolAttribute{Computed: true}, "cluster_endpoints": schema.StringAttribute{Computed: true}, "cluster_state": schema.StringAttribute{Computed: true}, "connection_uri": schema.StringAttribute{Computed: true}, "dns_seed_fqdn": schema.StringAttribute{Computed: true}, "nodes": listObjDatasourceSchema(valkeyStatusNodesObjFields), "password": schema.StringAttribute{Computed: true}, "port": schema.Int64Attribute{Computed: true}, "primary_endpoints": schema.StringAttribute{Computed: true}, "shards": listObjDatasourceSchema(valkeyStatusShardsObjFields)}),
	}}
}

func (d *ValkeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ValkeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ValkeyDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/valkey", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/valkey", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	state.Metadata = &metadataModel{}
	if err := setCommonFieldsNested(ctx, apiData, state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	state.Spec = &ValkeySpecModel{}
	spec := getSpec(apiData)
	state.Spec.CreatePublicIpv4 = getBool(spec, "createPublicIpv4")
	state.Spec.ParametersSetId = getString(spec, "parametersSetId")
	state.Spec.ReplicasPerShard = getInt64(spec, "replicasPerShard")
	state.Spec.RootPassword = getString(spec, "rootPassword")
	state.Spec.Shards = listObjFromAPI(objList(spec, "shards"), valkeyShardsObjFields)
	state.Spec.UseFqdn = getBool(spec, "useFqdn")
	state.Spec.ValkeyVersion = getString(spec, "valkeyVersion")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"anti_affinity_message": types.StringType,
			"anti_affinity_ok":      types.BoolType,
			"cluster_endpoints":     types.StringType,
			"cluster_state":         types.StringType,
			"connection_uri":        types.StringType,
			"dns_seed_fqdn":         types.StringType,
			"nodes":                 attrTypeOf("list_object", valkeyStatusNodesObjFields),
			"password":              types.StringType,
			"port":                  types.Int64Type,
			"primary_endpoints":     types.StringType,
			"shards":                attrTypeOf("list_object", valkeyStatusShardsObjFields),
		},
		map[string]attr.Value{
			"anti_affinity_message": getStringFromInfo(apiData, "antiAffinityMessage"),
			"anti_affinity_ok":      getBoolFromInfo(apiData, "antiAffinityOk"),
			"cluster_endpoints":     getStringFromInfo(apiData, "clusterEndpoints"),
			"cluster_state":         getStringFromInfo(apiData, "clusterState"),
			"connection_uri":        getStringFromInfo(apiData, "connectionUri"),
			"dns_seed_fqdn":         getStringFromInfo(apiData, "dnsSeedFqdn"),
			"nodes":                 getListObjFromInfo(apiData, "nodes", valkeyStatusNodesObjFields),
			"password":              getStringFromInfo(apiData, "password"),
			"port":                  getInt64FromInfo(apiData, "port"),
			"primary_endpoints":     getStringFromInfo(apiData, "primaryEndpoints"),
			"shards":                getListObjFromInfo(apiData, "shards", valkeyStatusShardsObjFields),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
