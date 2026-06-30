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

type GitlabDataSourceModel struct {
	ID       types.String    `tfsdk:"id"`
	Name     types.String    `tfsdk:"name"`
	Metadata metadataModel   `tfsdk:"metadata"`
	Spec     GitlabSpecModel `tfsdk:"spec"`
	Status   types.Object    `tfsdk:"status"`
}

type GitlabDataSource struct{ client *client.Client }

func NewGitlabDataSource() datasource.DataSource { return &GitlabDataSource{} }

func (d *GitlabDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab"
}

func (d *GitlabDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	specAttrs := map[string]schema.Attribute{
		"custom_ingress_configuration": objDatasourceSchema(gitlabCustomIngressConfigurationObjFields),
		"edition":                      schema.StringAttribute{Computed: true},
		"floating_ip_id":               schema.StringAttribute{Computed: true},
		"record_name":                  schema.StringAttribute{Computed: true},
		"root_password":                schema.StringAttribute{Computed: true, Sensitive: true},
		"secret_push_protection":       objDatasourceSchema(gitlabSecretPushProtectionObjFields),
		"tier":                         schema.StringAttribute{Computed: true},
		"version":                      schema.StringAttribute{Computed: true},
		"vm_offer_id":                  schema.StringAttribute{Computed: true},
		"vm_state":                     schema.StringAttribute{Computed: true},
		"volume_offer_id":              schema.StringAttribute{Computed: true},
		"volume_size_gib":              schema.Int64Attribute{Computed: true},
		"vpc_subnet_id":                schema.StringAttribute{Computed: true},
	}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Optional: true, Computed: true, Description: "ID of the resource to look up. Set exactly one of `id` or `name`."},
		"name":     schema.StringAttribute{Optional: true, Computed: true, Description: "Name of the resource to look up. Set exactly one of `id` or `name`."},
		"metadata": metadataDatasourceSchema(),
		"spec":     schema.SingleNestedAttribute{Computed: true, Attributes: specAttrs},
		"status":   commonInfoDatasourceSchema(map[string]schema.Attribute{"fqdn": schema.StringAttribute{Computed: true}, "private_ip_v4": schema.StringAttribute{Computed: true}, "private_ip_v6": schema.StringAttribute{Computed: true}, "public_ip_v4": schema.StringAttribute{Computed: true}, "public_ip_v6": schema.StringAttribute{Computed: true}}),
	}}
}

func (d *GitlabDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GitlabDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GitlabDataSourceModel
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
		apiData, err = d.client.Get(ctx, "/api/v1/gitlab", state.ID.ValueString())
	} else {
		apiData, err = d.client.GetByName(ctx, "/api/v1/gitlab", state.Name.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if apiData == nil {
		resp.Diagnostics.AddError("Not Found", "resource not found")
		return
	}
	if err := setCommonFieldsNested(ctx, apiData, &state.Metadata); err != nil {
		resp.Diagnostics.AddError("State Error", err.Error())
		return
	}
	state.ID = state.Metadata.ID
	state.Name = state.Metadata.Name
	spec := getSpec(apiData)
	state.Spec.CustomIngressConfiguration = objFromAPI(objMap(spec, "customIngressConfiguration"), gitlabCustomIngressConfigurationObjFields)
	state.Spec.Edition = getString(spec, "edition")
	state.Spec.FloatingIpId = getString(spec, "floatingIpId")
	state.Spec.RecordName = getString(spec, "recordName")
	state.Spec.RootPassword = getString(spec, "rootPassword")
	state.Spec.SecretPushProtection = objFromAPI(objMap(spec, "secretPushProtection"), gitlabSecretPushProtectionObjFields)
	state.Spec.Tier = getString(spec, "tier")
	state.Spec.Version = getString(spec, "version")
	state.Spec.VmOfferId = getString(spec, "vmOfferId")
	state.Spec.VmState = getString(spec, "vmState")
	state.Spec.VolumeOfferId = getString(spec, "volumeOfferId")
	state.Spec.VolumeSizeGib = getInt64(spec, "volumeSizeGiB")
	state.Spec.VpcSubnetId = getString(spec, "vpcSubnetId")
	state.Status = buildInfoObj(apiData,
		map[string]attr.Type{
			"fqdn":          types.StringType,
			"private_ip_v4": types.StringType,
			"private_ip_v6": types.StringType,
			"public_ip_v4":  types.StringType,
			"public_ip_v6":  types.StringType,
		},
		map[string]attr.Value{
			"fqdn":          getStringFromInfo(apiData, "fqdn"),
			"private_ip_v4": getStringFromInfo(apiData, "privateIpV4"),
			"private_ip_v6": getStringFromInfo(apiData, "privateIpV6"),
			"public_ip_v4":  getStringFromInfo(apiData, "publicIpV4"),
			"public_ip_v6":  getStringFromInfo(apiData, "publicIpV6"),
		})
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
