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
