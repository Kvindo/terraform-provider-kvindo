package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// commonDatasourceSchemaAttributes returns the common datasource schema attributes.
func commonDatasourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			Required: true,
		},
		"name": dschema.StringAttribute{
			Computed: true,
		},
		"description": dschema.StringAttribute{
			Computed: true,
		},
		"folder_id": dschema.StringAttribute{
			Computed: true,
		},
		"delete_protection": dschema.BoolAttribute{
			Computed: true,
		},
		"labels": dschema.MapAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
	}
}
