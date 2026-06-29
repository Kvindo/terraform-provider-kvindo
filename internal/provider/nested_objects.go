package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// objField describes one attribute of a nested object (or element of a list of objects) for
// converting between the API's JSON (camelCase keys) and Terraform values (snake_case keys).
// The generator emits a []objField descriptor per nested spec field; all conversion and schema
// generation is driven by these descriptors at runtime, so the generated code stays compact and
// the (recursive) logic lives in one place.
type objField struct {
	TF        string // terraform attribute name (snake_case)
	API       string // API JSON key (camelCase)
	Kind      string // string|bool|int64|float64|list_string|map_string|object|list_object
	Sensitive bool
	Obj       []objField // sub-fields for Kind == object / list_object
}

// ---- attr.Type ----

func attrTypeOf(kind string, obj []objField) attr.Type {
	switch kind {
	case "bool":
		return types.BoolType
	case "int64":
		return types.Int64Type
	case "float64":
		return types.Float64Type
	case "list_string":
		return types.ListType{ElemType: types.StringType}
	case "map_string":
		return types.MapType{ElemType: types.StringType}
	case "object":
		return types.ObjectType{AttrTypes: objAttrTypes(obj)}
	case "list_object":
		return types.ListType{ElemType: types.ObjectType{AttrTypes: objAttrTypes(obj)}}
	default:
		return types.StringType
	}
}

func objAttrTypes(fields []objField) map[string]attr.Type {
	m := make(map[string]attr.Type, len(fields))
	for _, f := range fields {
		m[f.TF] = attrTypeOf(f.Kind, f.Obj)
	}
	return m
}

// ---- TF value -> API JSON ----

// objToAPI converts a TF object value into a JSON map keyed by API names. Null/unknown
// attributes are omitted so the API applies its own defaults.
func objToAPI(o types.Object, fields []objField) map[string]interface{} {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	attrs := o.Attributes()
	out := make(map[string]interface{})
	for _, f := range fields {
		v, ok := attrs[f.TF]
		if !ok || v.IsNull() || v.IsUnknown() {
			continue
		}
		out[f.API] = tfAttrToAPI(v, f)
	}
	return out
}

// listObjToAPI converts a TF list-of-objects into a JSON array.
func listObjToAPI(l types.List, fields []objField) []interface{} {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]interface{}, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		if o, ok := e.(types.Object); ok {
			out = append(out, objToAPI(o, fields))
		}
	}
	return out
}

func tfAttrToAPI(v attr.Value, f objField) interface{} {
	switch f.Kind {
	case "bool":
		if b, ok := v.(types.Bool); ok {
			return b.ValueBool()
		}
	case "int64":
		if i, ok := v.(types.Int64); ok {
			return i.ValueInt64()
		}
	case "float64":
		if n, ok := v.(types.Float64); ok {
			return n.ValueFloat64()
		}
	case "list_string":
		if l, ok := v.(types.List); ok {
			out := make([]interface{}, 0, len(l.Elements()))
			for _, e := range l.Elements() {
				if s, ok := e.(types.String); ok && !s.IsNull() {
					out = append(out, s.ValueString())
				}
			}
			return out
		}
	case "map_string":
		if m, ok := v.(types.Map); ok {
			out := make(map[string]interface{})
			for k, e := range m.Elements() {
				if s, ok := e.(types.String); ok && !s.IsNull() {
					out[k] = s.ValueString()
				}
			}
			return out
		}
	case "object":
		if o, ok := v.(types.Object); ok {
			return objToAPI(o, f.Obj)
		}
	case "list_object":
		if l, ok := v.(types.List); ok {
			return listObjToAPI(l, f.Obj)
		}
	default:
		if s, ok := v.(types.String); ok {
			return s.ValueString()
		}
	}
	return nil
}

// ---- API JSON -> TF value ----

// objFromAPI converts a JSON map into a TF object value of the given fields. A nil map yields a
// null object so optional nested blocks round-trip cleanly.
func objFromAPI(raw map[string]interface{}, fields []objField) types.Object {
	at := objAttrTypes(fields)
	if raw == nil {
		return types.ObjectNull(at)
	}
	vals := make(map[string]attr.Value, len(fields))
	for _, f := range fields {
		vals[f.TF] = apiToTfAttr(raw[f.API], f)
	}
	obj, diags := types.ObjectValue(at, vals)
	if diags.HasError() {
		return types.ObjectNull(at)
	}
	return obj
}

// listObjFromAPI converts a JSON array into a TF list-of-objects (empty list when absent).
func listObjFromAPI(raw []interface{}, fields []objField) types.List {
	elemType := types.ObjectType{AttrTypes: objAttrTypes(fields)}
	vals := make([]attr.Value, 0, len(raw))
	for _, it := range raw {
		m, _ := it.(map[string]interface{})
		vals = append(vals, objFromAPI(m, fields))
	}
	l, diags := types.ListValue(elemType, vals)
	if diags.HasError() {
		return types.ListValueMust(elemType, []attr.Value{})
	}
	return l
}

func apiToTfAttr(raw interface{}, f objField) attr.Value {
	switch f.Kind {
	case "bool":
		if b, ok := raw.(bool); ok {
			return types.BoolValue(b)
		}
		return types.BoolNull()
	case "int64":
		switch n := raw.(type) {
		case float64:
			return types.Int64Value(int64(n))
		case int64:
			return types.Int64Value(n)
		case int:
			return types.Int64Value(int64(n))
		}
		return types.Int64Null()
	case "float64":
		switch n := raw.(type) {
		case float64:
			return types.Float64Value(n)
		case int64:
			return types.Float64Value(float64(n))
		case int:
			return types.Float64Value(float64(n))
		}
		return types.Float64Null()
	case "list_string":
		arr, _ := raw.([]interface{})
		vals := make([]attr.Value, 0, len(arr))
		for _, e := range arr {
			if s, ok := e.(string); ok {
				vals = append(vals, types.StringValue(s))
			}
		}
		return types.ListValueMust(types.StringType, vals)
	case "map_string":
		m, _ := raw.(map[string]interface{})
		vals := make(map[string]attr.Value, len(m))
		for k, e := range m {
			if s, ok := e.(string); ok {
				vals[k] = types.StringValue(s)
			}
		}
		return types.MapValueMust(types.StringType, vals)
	case "object":
		m, _ := raw.(map[string]interface{})
		return objFromAPI(m, f.Obj)
	case "list_object":
		arr, _ := raw.([]interface{})
		return listObjFromAPI(arr, f.Obj)
	default:
		if s, ok := raw.(string); ok {
			return types.StringValue(s)
		}
		return types.StringNull()
	}
}

// ---- schema builders ----

func objLeafResourceSchema(f objField) rschema.Attribute {
	switch f.Kind {
	case "bool":
		return rschema.BoolAttribute{Optional: true, Computed: true}
	case "int64":
		return rschema.Int64Attribute{Optional: true, Computed: true}
	case "float64":
		return rschema.Float64Attribute{Optional: true, Computed: true}
	case "list_string":
		return rschema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType}
	case "map_string":
		return rschema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType}
	case "object":
		return objResourceSchema(f.Obj)
	case "list_object":
		return listObjResourceSchema(f.Obj)
	default:
		return rschema.StringAttribute{Optional: true, Computed: true, Sensitive: f.Sensitive}
	}
}

// objResourceSchema builds an Optional+Computed SingleNestedAttribute for a nested object.
func objResourceSchema(fields []objField) rschema.Attribute {
	attrs := make(map[string]rschema.Attribute, len(fields))
	for _, f := range fields {
		attrs[f.TF] = objLeafResourceSchema(f)
	}
	return rschema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: attrs}
}

// listObjResourceSchema builds an Optional+Computed ListNestedAttribute for a list of objects.
func listObjResourceSchema(fields []objField) rschema.Attribute {
	attrs := make(map[string]rschema.Attribute, len(fields))
	for _, f := range fields {
		attrs[f.TF] = objLeafResourceSchema(f)
	}
	return rschema.ListNestedAttribute{
		Optional:     true,
		Computed:     true,
		NestedObject: rschema.NestedAttributeObject{Attributes: attrs},
	}
}

func objLeafDatasourceSchema(f objField) dschema.Attribute {
	switch f.Kind {
	case "bool":
		return dschema.BoolAttribute{Computed: true}
	case "int64":
		return dschema.Int64Attribute{Computed: true}
	case "float64":
		return dschema.Float64Attribute{Computed: true}
	case "list_string":
		return dschema.ListAttribute{Computed: true, ElementType: types.StringType}
	case "map_string":
		return dschema.MapAttribute{Computed: true, ElementType: types.StringType}
	case "object":
		return objDatasourceSchema(f.Obj)
	case "list_object":
		return listObjDatasourceSchema(f.Obj)
	default:
		return dschema.StringAttribute{Computed: true, Sensitive: f.Sensitive}
	}
}

func objDatasourceSchema(fields []objField) dschema.Attribute {
	attrs := make(map[string]dschema.Attribute, len(fields))
	for _, f := range fields {
		attrs[f.TF] = objLeafDatasourceSchema(f)
	}
	return dschema.SingleNestedAttribute{Computed: true, Attributes: attrs}
}

func listObjDatasourceSchema(fields []objField) dschema.Attribute {
	attrs := make(map[string]dschema.Attribute, len(fields))
	for _, f := range fields {
		attrs[f.TF] = objLeafDatasourceSchema(f)
	}
	return dschema.ListNestedAttribute{
		Computed:     true,
		NestedObject: dschema.NestedAttributeObject{Attributes: attrs},
	}
}

// objMap / objList extract a raw nested value from a spec/parent map for objFromAPI/listObjFromAPI.
func objMap(parent map[string]interface{}, key string) map[string]interface{} {
	m, _ := parent[key].(map[string]interface{})
	return m
}

func objList(parent map[string]interface{}, key string) []interface{} {
	a, _ := parent[key].([]interface{})
	return a
}
