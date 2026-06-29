package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// txnObjType derives the nested object type of a transaction sub-resource map element from its
// resource schema attributes (the same VmResourceSchemaAttrs() etc. the standalone resource uses).
func txnObjType(attrs map[string]schema.Attribute) types.ObjectType {
	m := make(map[string]attr.Type, len(attrs))
	for k, a := range attrs {
		m[k] = a.GetType()
	}
	return types.ObjectType{AttrTypes: m}
}

// txnBuild adapts a standalone build<Sn>RequestMap into a transaction sub-item builder: decode the
// map element object into the resource model and run the resource's own request builder, which
// already returns the {metadata, spec} wire shape a transaction sub-item needs.
func txnBuild[T any](buildFn func(context.Context, T) map[string]interface{}) func(context.Context, types.Object) map[string]interface{} {
	return func(ctx context.Context, o types.Object) map[string]interface{} {
		var e T
		o.As(ctx, &e, basetypes.ObjectAsOptions{})
		return buildFn(ctx, e)
	}
}

// txnPop adapts a standalone populate<Sn>State into a transaction sub-item populator: run the
// resource's own populate, then materialize the resulting model back into a types.Object (using
// the schema-derived element type) and return it together with the resource id (for map keying).
func txnPop[T any](popFn func(context.Context, map[string]interface{}, *T) error, attrs func() map[string]schema.Attribute) func(context.Context, map[string]interface{}) (types.Object, string) {
	return func(ctx context.Context, item map[string]interface{}) (types.Object, string) {
		var e T
		_ = popFn(ctx, item, &e)
		ot := txnObjType(attrs())
		obj, diags := types.ObjectValueFrom(ctx, ot.AttrTypes, e)
		if diags.HasError() {
			return types.ObjectNull(ot.AttrTypes), ""
		}
		id := ""
		if v, ok := obj.Attributes()["id"].(types.String); ok {
			id = v.ValueString()
		}
		return obj, id
	}
}

// txnSchemaAttrs builds the full transaction schema. Like every other resource it exposes exactly
// three blocks: metadata (identity), spec (the delete-on-destroy flag plus one map-of-objects per
// transactable sub-resource, reusing each resource's schema), and status. The root id is kept for
// import and mirrored at metadata.id.
func txnSchemaAttrs() map[string]schema.Attribute {
	specAttrs := map[string]schema.Attribute{
		"delete_resources_on_transaction_delete": schema.BoolAttribute{Optional: true, Computed: true},
	}
	for _, s := range txnSubs {
		elemAttrs := txnRelaxAttrs(s.attrs())
		// The entry's root id is a computed mirror; users set the id (for cross-referencing) on
		// metadata.id, which stays settable. Forcing root id Computed-only keeps "id" out of the
		// configurable surface so the only place to set it is metadata.id.
		elemAttrs["id"] = schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
		specAttrs[s.tfKey] = schema.MapNestedAttribute{
			Optional:     true,
			NestedObject: schema.NestedAttributeObject{Attributes: elemAttrs},
		}
	}
	return map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"metadata": metadataResourceSchema(),
		"spec":     schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: specAttrs},
		"status":   commonInfoSchema(nil),
	}
}

// txnRelaxAttrs makes every attribute of a reused resource schema Optional+Computed (recursively).
// In a transaction, sub-resource ids and cross-reference fields must be user-settable — a producer
// pre-assigns its id (id = local.x) and a consumer references it (bucket_id = local.x) — which the
// standalone schema forbids (id is Computed-only, cross-ref fields are Required). Relaxing does not
// change attribute TYPES, so the schema-derived element type (txnObjType) and build/populate are
// unaffected; only what the user may set changes.
func txnRelaxAttrs(attrs map[string]schema.Attribute) map[string]schema.Attribute {
	out := make(map[string]schema.Attribute, len(attrs))
	for k, a := range attrs {
		out[k] = txnRelaxAttr(a)
	}
	return out
}

func txnRelaxAttr(a schema.Attribute) schema.Attribute {
	switch v := a.(type) {
	case schema.StringAttribute:
		return schema.StringAttribute{Optional: true, Computed: true, Sensitive: v.Sensitive}
	case schema.BoolAttribute:
		return schema.BoolAttribute{Optional: true, Computed: true, Sensitive: v.Sensitive}
	case schema.Int64Attribute:
		return schema.Int64Attribute{Optional: true, Computed: true, Sensitive: v.Sensitive}
	case schema.Float64Attribute:
		return schema.Float64Attribute{Optional: true, Computed: true, Sensitive: v.Sensitive}
	case schema.ListAttribute:
		return schema.ListAttribute{Optional: true, Computed: true, ElementType: v.ElementType, Sensitive: v.Sensitive}
	case schema.MapAttribute:
		return schema.MapAttribute{Optional: true, Computed: true, ElementType: v.ElementType, Sensitive: v.Sensitive}
	case schema.SingleNestedAttribute:
		return schema.SingleNestedAttribute{Optional: true, Computed: true, Sensitive: v.Sensitive, Attributes: txnRelaxAttrs(v.Attributes)}
	case schema.ListNestedAttribute:
		return schema.ListNestedAttribute{Optional: true, Computed: true, Sensitive: v.Sensitive, NestedObject: schema.NestedAttributeObject{Attributes: txnRelaxAttrs(v.NestedObject.Attributes)}}
	default:
		return a
	}
}

// txnBuildSubResources adds each present sub-resource map to the request spec, honoring the
// two-phase S3 gates. It is called from buildTxnBody after the transaction metadata is assembled.
func txnBuildSubResources(ctx context.Context, plan *TransactionResourceModel, spec map[string]interface{}, includeFolders, includeBuckets, includePolicies, includeUsers bool) {
	gates := map[string]bool{"folder": includeFolders, "bucket": includeBuckets, "policy": includePolicies, "user": includeUsers}
	for _, s := range txnSubs {
		if s.gate != "" && !gates[s.gate] {
			continue
		}
		mp := *s.field(plan)
		if mp.IsNull() || mp.IsUnknown() {
			continue
		}
		elems := mp.Elements()
		items := make([]interface{}, 0, len(elems))
		for _, e := range elems {
			if o, ok := e.(types.Object); ok {
				items = append(items, s.build(ctx, o))
			}
		}
		spec[s.apiKey] = items
	}
}

// txnPopulateSubResources fills each sub-resource map on state from the API response, preserving
// the user's chosen map keys (matched by id via resolveMapKey).
func txnPopulateSubResources(ctx context.Context, data map[string]interface{}, state *TransactionResourceModel) {
	respSpec := getSpec(data)
	for _, s := range txnSubs {
		fp := s.field(state)
		ot := txnObjType(s.attrs())
		// Only overwrite when the response actually returns elements for this type. The transaction
		// GET includes empty arrays for unused types, and the two-phase S3 create returns s3_users
		// empty in phase 1. Overwriting with an empty map in those cases would (a) turn a null into
		// an empty map ("inconsistent result after apply") and (b) wipe the planned s3_users between
		// phase 1 and phase 2 so phase 2 sends no user. Leaving the current value preserves null for
		// unset maps, an explicit empty map, and the planned entries across phases.
		arr, _ := respSpec[s.apiKey].([]interface{})
		if len(arr) == 0 {
			continue
		}
		items := make(map[string]attr.Value, len(arr))
		for _, v := range arr {
			m, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			obj, id := s.populate(ctx, m)
			meta, _ := m["metadata"].(map[string]interface{})
			items[resolveMapKey(*fp, id, meta)] = obj
		}
		*fp = types.MapValueMust(ot, items)
	}
}

// txnAssignAndRecoverIDs runs the create-time id machinery over every sub-resource map: recover
// user-provided ids / cross-reference fields the framework marked unknown, then assign ULIDs to
// any element still missing an id so populateState can match map keys after the round-trip.
func txnAssignAndRecoverIDs(ctx context.Context, plan, cfg *TransactionResourceModel) {
	for _, s := range txnSubs {
		fp := s.field(plan)
		cfp := s.field(cfg)
		at := txnObjType(s.attrs()).AttrTypes
		*fp = fillUnknownIDsFromConfig(ctx, *fp, *cfp, at)
	}
	for _, s := range txnSubs {
		fp := s.field(plan)
		at := txnObjType(s.attrs()).AttrTypes
		*fp = assignSubIDsMap(*fp, at)
	}
}

// setSpecField sets key on the element's nested "spec" object (rebuilding it), used by the s3
// auto-wiring. Returns the updated element object.
func setSpecField(obj types.Object, elemAttrTypes map[string]attr.Type, key string, val attr.Value) types.Object {
	attrs := obj.Attributes()
	specVal, ok := attrs["spec"].(types.Object)
	if !ok || specVal.IsNull() {
		return obj
	}
	specType, ok := elemAttrTypes["spec"].(types.ObjectType)
	if !ok {
		return obj
	}
	specAttrs := specVal.Attributes()
	newSpec := make(map[string]attr.Value, len(specAttrs))
	for k, v := range specAttrs {
		newSpec[k] = v
	}
	newSpec[key] = val
	newSpecObj, diags := types.ObjectValue(specType.AttrTypes, newSpec)
	if diags.HasError() {
		return obj
	}
	newAttrs := make(map[string]attr.Value, len(attrs))
	for k, v := range attrs {
		newAttrs[k] = v
	}
	newAttrs["spec"] = newSpecObj
	out, diags := types.ObjectValue(elemAttrTypes, newAttrs)
	if diags.HasError() {
		return obj
	}
	return out
}

// getSpecString reads a string field from the element's nested "spec" object.
func getSpecString(obj types.Object, key string) (types.String, bool) {
	specVal, ok := obj.Attributes()["spec"].(types.Object)
	if !ok || specVal.IsNull() {
		return types.StringNull(), false
	}
	v, ok := specVal.Attributes()[key].(types.String)
	return v, ok
}

var _ = strings.ToLower
