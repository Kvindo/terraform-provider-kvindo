package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Regression test for the kvindo_vpc security_group_ids drift bug: TerraformMockTests'
// terraform-plan drift check (apply -> plan -detailed-exitcode) caught spec.security_group_ids
// flipping from [] to "known after apply" on every plan after a clean apply, even with no
// config or backend change. Root cause: resourceAttrDef's list_string/map_string cases in
// tools/generator/main.go never added a UseStateForUnknown plan modifier, unlike every scalar
// (string/bool/int64/float64) case, which all did. Without it, terraform-plugin-framework's
// default behavior is to mark an unconfigured Optional+Computed attribute as unknown on every
// plan, producing phantom drift forever. This asserts every top-level Optional+Computed
// List/Map spec attribute across all resources carries a plan modifier, so a future generator
// change (or a hand-patched resource file) can't reintroduce the same gap.
//
// Deliberately out of scope: the SingleNestedAttribute/ListNestedAttribute containers
// themselves (for "object"/"list_object" spec fields) carry no plan modifier of their own —
// only their leaf children do. terraform-plugin-framework derives a compound object/list-of-
// objects value's unknown-ness from its children, so no container-level modifier is needed
// once every leaf is stable.
func TestResourceSchemas_OptionalComputedListAndMapHaveUseStateForUnknown(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	for _, newResource := range p.Resources(context.Background()) {
		r := newResource()

		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)

		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Errorf("%s: schema build produced diagnostics: %v", metaResp.TypeName, schemaResp.Diagnostics)
			continue
		}

		checkTopLevelAttrsForMissingPlanModifiers(t, metaResp.TypeName, "", schemaResp.Schema.Attributes)
	}
}

func checkTopLevelAttrsForMissingPlanModifiers(t *testing.T, resourceName, path string, attrs map[string]schema.Attribute) {
	t.Helper()
	for name, a := range attrs {
		full := path + name
		switch attr := a.(type) {
		case schema.ListAttribute:
			if attr.Optional && attr.Computed && len(attr.PlanModifiers) == 0 {
				t.Errorf("%s: attribute %q is Optional+Computed ListAttribute with no PlanModifiers - "+
					"will report phantom drift on every plan after apply", resourceName, full)
			}
		case schema.MapAttribute:
			if attr.Optional && attr.Computed && len(attr.PlanModifiers) == 0 {
				t.Errorf("%s: attribute %q is Optional+Computed MapAttribute with no PlanModifiers - "+
					"will report phantom drift on every plan after apply", resourceName, full)
			}
		case schema.SingleNestedAttribute:
			checkTopLevelAttrsForMissingPlanModifiers(t, resourceName, full+".", attr.Attributes)
		case schema.ListNestedAttribute:
			checkTopLevelAttrsForMissingPlanModifiers(t, resourceName, full+"[].", attr.NestedObject.Attributes)
		}
	}
}
