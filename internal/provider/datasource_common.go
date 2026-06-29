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

// metadataDatasourceSchema returns the Computed "metadata" block for datasources.
// The root-level id is the lookup key (Required); metadata mirrors it and the rest read-only.
func metadataDatasourceSchema() dschema.Attribute {
	return dschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dschema.Attribute{
			"id":                dschema.StringAttribute{Computed: true},
			"name":              dschema.StringAttribute{Computed: true},
			"description":       dschema.StringAttribute{Computed: true},
			"folder_id":         dschema.StringAttribute{Computed: true},
			"delete_protection": dschema.BoolAttribute{Computed: true},
			"labels":            dschema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

// commonInfoDatasourceSchema returns the Computed "status" block for datasources. It mirrors
// the resource-side commonInfoSchema (base ResourceInfo fields) but in the datasource schema
// package and without plan modifiers (datasources are always recomputed on read). Pass extra
// resource-specific status attrs to merge in.
func commonInfoDatasourceSchema(extraAttrs map[string]dschema.Attribute) dschema.Attribute {
	userAttrs := map[string]dschema.Attribute{
		"id":   dschema.StringAttribute{Computed: true},
		"name": dschema.StringAttribute{Computed: true},
	}
	attrs := map[string]dschema.Attribute{
		"state":           dschema.StringAttribute{Computed: true},
		"create_time":     dschema.StringAttribute{Computed: true},
		"created_by_user": dschema.SingleNestedAttribute{Computed: true, Attributes: userAttrs},
		"last_change_request": dschema.SingleNestedAttribute{Computed: true, Attributes: map[string]dschema.Attribute{
			"state":           dschema.StringAttribute{Computed: true},
			"create_time":     dschema.StringAttribute{Computed: true},
			"error_message":   dschema.StringAttribute{Computed: true},
			"created_by_user": dschema.SingleNestedAttribute{Computed: true, Attributes: userAttrs},
		}},
		"pricing": dschema.SingleNestedAttribute{Computed: true, Attributes: map[string]dschema.Attribute{
			"month": dschema.Float64Attribute{Computed: true},
			"day":   dschema.Float64Attribute{Computed: true},
			"hour":  dschema.Float64Attribute{Computed: true},
		}},
	}
	for k, v := range extraAttrs {
		attrs[k] = v
	}
	return dschema.SingleNestedAttribute{Computed: true, Attributes: attrs}
}
