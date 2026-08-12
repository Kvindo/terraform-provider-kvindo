package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// A descriptor exercising every kind, including a nested object and a list of objects.
var testObjFields = []objField{
	{TF: "name", API: "name", Kind: "string"},
	{TF: "enabled", API: "enabled", Kind: "bool"},
	{TF: "weight", API: "weight", Kind: "int64"},
	{TF: "ratio", API: "ratio", Kind: "float64"},
	{TF: "hosts", API: "hosts", Kind: "list_string"},
	{TF: "headers", API: "headers", Kind: "map_string"},
	{TF: "tls", API: "tls", Kind: "object", Obj: []objField{
		{TF: "min_version", API: "minVersion", Kind: "string"},
	}},
	{TF: "targets", API: "targets", Kind: "list_object", Obj: []objField{
		{TF: "ip", API: "ip", Kind: "string"},
		{TF: "port", API: "port", Kind: "int64"},
	}},
}

func TestObjRoundTrip(t *testing.T) {
	raw := map[string]interface{}{
		"name":    "rule-1",
		"enabled": true,
		"weight":  float64(7),
		"ratio":   float64(0.5),
		"hosts":   []interface{}{"a.com", "b.com"},
		"headers": map[string]interface{}{"X-Env": "prod"},
		"tls":     map[string]interface{}{"minVersion": "1.2"},
		"targets": []interface{}{
			map[string]interface{}{"ip": "10.0.0.1", "port": float64(80)},
			map[string]interface{}{"ip": "10.0.0.2", "port": float64(443)},
		},
	}

	obj := objFromAPI(raw, testObjFields)
	if obj.IsNull() {
		t.Fatal("objFromAPI returned null for populated map")
	}
	// nested object
	tls, _ := obj.Attributes()["tls"].(types.Object)
	if v, ok := tls.Attributes()["min_version"].(types.String); !ok || v.ValueString() != "1.2" {
		t.Errorf("tls.min_version: got %v", tls.Attributes()["min_version"])
	}
	// list of objects
	targets, _ := obj.Attributes()["targets"].(types.List)
	if len(targets.Elements()) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets.Elements()))
	}

	// Convert back to API and check keys are camelCase and values round-trip.
	back := objToAPI(obj, testObjFields)
	if back["name"] != "rule-1" || back["enabled"] != true {
		t.Errorf("scalar round-trip failed: %v", back)
	}
	if back["weight"] != int64(7) {
		t.Errorf("int64 round-trip: got %v (%T)", back["weight"], back["weight"])
	}
	tlsBack, ok := back["tls"].(map[string]interface{})
	if !ok || tlsBack["minVersion"] != "1.2" {
		t.Errorf("nested object camelCase key round-trip failed: %v", back["tls"])
	}
	targetsBack, ok := back["targets"].([]interface{})
	if !ok || len(targetsBack) != 2 {
		t.Fatalf("list_object round-trip failed: %v", back["targets"])
	}
	first := targetsBack[0].(map[string]interface{})
	if first["ip"] != "10.0.0.1" || first["port"] != int64(80) {
		t.Errorf("list_object element round-trip: got %v", first)
	}
}

func TestObjFromAPI_NilIsNull(t *testing.T) {
	obj := objFromAPI(nil, testObjFields)
	if !obj.IsNull() {
		t.Error("expected null object for nil input")
	}
}

// TestGetListObjFromInfo_ReadsStatusArray reproduces the real bug this covers: a status field
// that's an array of objects (e.g. etcd's status.instances) used to always read back as an
// empty string via getStringFromInfo's `v.(string)` assertion silently failing. This proves the
// list_object path through infoFieldRaw -> getListObjFromInfo -> listObjFromAPI works end to end.
func TestGetListObjFromInfo_ReadsStatusArray(t *testing.T) {
	instanceFields := []objField{
		{TF: "id", API: "id", Kind: "string"},
		{TF: "private_ipv4", API: "privateIpV4", Kind: "string"},
		{TF: "public_ipv4", API: "publicIpV4", Kind: "string"},
	}
	data := map[string]interface{}{
		"status": map[string]interface{}{
			"instances": []interface{}{
				map[string]interface{}{"id": "i1", "privateIpV4": "10.0.0.5"},
				map[string]interface{}{"id": "i2", "privateIpV4": "10.0.0.9"},
			},
		},
	}
	l := getListObjFromInfo(data, "instances", instanceFields)
	if l.IsNull() || l.IsUnknown() {
		t.Fatal("expected a non-null, known list")
	}
	if len(l.Elements()) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(l.Elements()))
	}
	first, ok := l.Elements()[0].(types.Object)
	if !ok {
		t.Fatalf("element 0 is not an Object: %T", l.Elements()[0])
	}
	if v, ok := first.Attributes()["id"].(types.String); !ok || v.ValueString() != "i1" {
		t.Errorf("element 0 id: got %v", first.Attributes()["id"])
	}
	if v, ok := first.Attributes()["private_ipv4"].(types.String); !ok || v.ValueString() != "10.0.0.5" {
		t.Errorf("element 0 private_ipv4: got %v", first.Attributes()["private_ipv4"])
	}
	// public_ipv4 was absent on the wire (private-only cluster) - must come back null, not an
	// empty string, mirroring the real API's "field omitted when disabled" contract.
	if v, ok := first.Attributes()["public_ipv4"].(types.String); !ok || !v.IsNull() {
		t.Errorf("element 0 public_ipv4: expected null, got %v", first.Attributes()["public_ipv4"])
	}
}

// TestGetListObjFromInfo_AbsentFieldIsEmptyNotPanic covers the field-missing case (e.g. a
// resource whose status predates this field, or a wire hiccup) - must degrade to an empty list,
// never panic on the type assertion.
func TestGetListObjFromInfo_AbsentFieldIsEmptyNotPanic(t *testing.T) {
	l := getListObjFromInfo(map[string]interface{}{"status": map[string]interface{}{}}, "instances", testObjFields)
	if l.IsNull() {
		t.Error("expected an empty (non-null) list when the field is absent, matching listObjFromAPI's convention")
	}
	if len(l.Elements()) != 0 {
		t.Errorf("expected 0 elements, got %d", len(l.Elements()))
	}
}

// TestGetObjFromInfo_ReadsStatusObject covers the singular nested-object status case
// (object, not list_object) through the same infoFieldRaw path.
func TestGetObjFromInfo_ReadsStatusObject(t *testing.T) {
	fields := []objField{{TF: "min_version", API: "minVersion", Kind: "string"}}
	data := map[string]interface{}{
		"status": map[string]interface{}{
			"tls": map[string]interface{}{"minVersion": "1.3"},
		},
	}
	o := getObjFromInfo(data, "tls", fields)
	if o.IsNull() {
		t.Fatal("expected a non-null object")
	}
	if v, ok := o.Attributes()["min_version"].(types.String); !ok || v.ValueString() != "1.3" {
		t.Errorf("min_version: got %v", o.Attributes()["min_version"])
	}
}

// TestListObjStatusSchema_IsComputedOnly pins the semantic point of adding a separate
// status-side schema builder instead of reusing listObjResourceSchema (spec-field, Optional+
// Computed): a status field is always server-computed, so it and every leaf under it must be
// Computed only - Optional would let terraform-plugin-framework accept user config for a
// read-only sub-tree at schema-build time, only to break at plan time.
func TestListObjStatusSchema_IsComputedOnly(t *testing.T) {
	attr := listObjStatusSchema([]objField{
		{TF: "id", API: "id", Kind: "string"},
	})
	ln, ok := attr.(interface {
		IsOptional() bool
		IsComputed() bool
	})
	if !ok {
		t.Fatalf("unexpected attribute type: %T", attr)
	}
	if ln.IsOptional() {
		t.Error("status list attribute must not be Optional")
	}
	if !ln.IsComputed() {
		t.Error("status list attribute must be Computed")
	}
}

func TestObjToAPI_OmitsNullAttrs(t *testing.T) {
	// Build an object with only one attribute set; the rest null.
	at := objAttrTypes(testObjFields)
	vals := map[string]interface{}{"name": "only-name"}
	obj := objFromAPI(vals, testObjFields)
	_ = at
	out := objToAPI(obj, testObjFields)
	if out["name"] != "only-name" {
		t.Errorf("name should survive: %v", out)
	}
	if _, ok := out["enabled"]; ok {
		t.Error("null bool attr should be omitted from API map")
	}
}
