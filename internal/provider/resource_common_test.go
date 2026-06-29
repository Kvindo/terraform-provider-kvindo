package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---- buildUserInfoObj ----

func TestBuildUserInfoObj_Nil(t *testing.T) {
	obj := buildUserInfoObj(nil)
	if !obj.IsNull() {
		t.Error("expected null object for nil input")
	}
}

func TestBuildUserInfoObj_Populated(t *testing.T) {
	obj := buildUserInfoObj(map[string]interface{}{"id": "u1", "name": "Alice"})
	if obj.IsNull() {
		t.Fatal("expected non-null object")
	}
	attrs := obj.Attributes()
	if v, ok := attrs["id"].(types.String); !ok || v.ValueString() != "u1" {
		t.Errorf("expected id='u1', got %v", attrs["id"])
	}
	if v, ok := attrs["name"].(types.String); !ok || v.ValueString() != "Alice" {
		t.Errorf("expected name='Alice', got %v", attrs["name"])
	}
}

func TestBuildUserInfoObj_MissingFields(t *testing.T) {
	obj := buildUserInfoObj(map[string]interface{}{})
	if obj.IsNull() {
		t.Fatal("expected non-null object with empty map")
	}
	attrs := obj.Attributes()
	if v, ok := attrs["id"].(types.String); !ok || !v.IsNull() {
		t.Errorf("expected id=null for missing key, got %v", attrs["id"])
	}
}

// ---- buildPricingObj ----

func TestBuildPricingObj_Nil(t *testing.T) {
	obj := buildPricingObj(nil)
	if !obj.IsNull() {
		t.Error("expected null object for nil input")
	}
}

func TestBuildPricingObj_Populated(t *testing.T) {
	obj := buildPricingObj(map[string]interface{}{
		"month": float64(29.99),
		"day":   float64(1.0),
		"hour":  float64(0.042),
	})
	if obj.IsNull() {
		t.Fatal("expected non-null object")
	}
	attrs := obj.Attributes()
	if v, ok := attrs["month"].(types.Float64); !ok || v.ValueFloat64() != 29.99 {
		t.Errorf("expected month=29.99, got %v", attrs["month"])
	}
	if v, ok := attrs["day"].(types.Float64); !ok || v.ValueFloat64() != 1.0 {
		t.Errorf("expected day=1.0, got %v", attrs["day"])
	}
	if v, ok := attrs["hour"].(types.Float64); !ok || v.ValueFloat64() != 0.042 {
		t.Errorf("expected hour=0.042, got %v", attrs["hour"])
	}
}

// ---- buildLcrInfoObj ----

func TestBuildLcrInfoObj_Nil(t *testing.T) {
	obj := buildLcrInfoObj(nil)
	if !obj.IsNull() {
		t.Error("expected null object for nil input")
	}
}

func TestBuildLcrInfoObj_Populated(t *testing.T) {
	obj := buildLcrInfoObj(map[string]interface{}{
		"state":         "reconcilling",
		"createTime":    "2026-06-27T10:00:00Z",
		"errorMessage":  "some error",
		"createdByUser": map[string]interface{}{"id": "u2", "name": "Bob"},
	})
	if obj.IsNull() {
		t.Fatal("expected non-null object")
	}
	attrs := obj.Attributes()
	if v, ok := attrs["state"].(types.String); !ok || v.ValueString() != "reconcilling" {
		t.Errorf("expected state='reconcilling', got %v", attrs["state"])
	}
	if v, ok := attrs["create_time"].(types.String); !ok || v.ValueString() != "2026-06-27T10:00:00Z" {
		t.Errorf("expected create_time='2026-06-27T10:00:00Z', got %v", attrs["create_time"])
	}
	if v, ok := attrs["error_message"].(types.String); !ok || v.ValueString() != "some error" {
		t.Errorf("expected error_message='some error', got %v", attrs["error_message"])
	}
	// nested created_by_user
	cbuObj, ok := attrs["created_by_user"].(types.Object)
	if !ok || cbuObj.IsNull() {
		t.Fatal("expected non-null created_by_user in lcr")
	}
	if v, ok2 := cbuObj.Attributes()["name"].(types.String); !ok2 || v.ValueString() != "Bob" {
		t.Errorf("expected lcr.created_by_user.name='Bob', got %v", cbuObj.Attributes()["name"])
	}
}

func TestBuildLcrInfoObj_NilNestedUser(t *testing.T) {
	// createdByUser absent → nested object should be null, not panic
	obj := buildLcrInfoObj(map[string]interface{}{
		"state": "reconcilled",
	})
	if obj.IsNull() {
		t.Fatal("expected non-null lcr object")
	}
	cbuObj, ok := obj.Attributes()["created_by_user"].(types.Object)
	if !ok || !cbuObj.IsNull() {
		t.Errorf("expected null created_by_user when key absent, got %v", obj.Attributes()["created_by_user"])
	}
}

// ---- simpleStateInfoObj / buildInfoObj ----

func TestSimpleStateInfoObj_OnlyState(t *testing.T) {
	data := map[string]interface{}{
		"status": map[string]interface{}{"state": "stable"},
	}
	obj := simpleStateInfoObj(data)
	if obj.IsNull() {
		t.Fatal("expected non-null info object")
	}
	if v, ok := obj.Attributes()["state"].(types.String); !ok || v.ValueString() != "stable" {
		t.Errorf("expected state='stable', got %v", obj.Attributes()["state"])
	}
}

func TestSimpleStateInfoObj_NilStatus(t *testing.T) {
	// Missing status key → all fields null, object still valid
	obj := simpleStateInfoObj(map[string]interface{}{})
	if obj.IsNull() {
		t.Fatal("expected non-null info object even without status key")
	}
	if v, ok := obj.Attributes()["state"].(types.String); !ok || !v.IsNull() {
		t.Errorf("expected null state when status absent, got %v", obj.Attributes()["state"])
	}
	// pricing should be null
	if p, ok := obj.Attributes()["pricing"].(types.Object); !ok || !p.IsNull() {
		t.Errorf("expected null pricing when status absent, got %v", obj.Attributes()["pricing"])
	}
}

func TestSimpleStateInfoObj_FullStatus(t *testing.T) {
	data := map[string]interface{}{
		"status": map[string]interface{}{
			"state":      "stable",
			"createTime": "2026-06-27T12:00:00Z",
			"createdByUser": map[string]interface{}{
				"id": "usr1", "name": "Eve",
			},
			"lastChangeRequest": map[string]interface{}{
				"state":      "reconcilled",
				"createTime": "2026-06-27T11:00:00Z",
			},
			"pricing": map[string]interface{}{
				"month": float64(100.0),
				"day":   float64(3.33),
				"hour":  float64(0.139),
			},
		},
	}
	obj := simpleStateInfoObj(data)
	attrs := obj.Attributes()

	// state
	if v, ok := attrs["state"].(types.String); !ok || v.ValueString() != "stable" {
		t.Errorf("state: got %v", attrs["state"])
	}
	// create_time
	if v, ok := attrs["create_time"].(types.String); !ok || v.ValueString() != "2026-06-27T12:00:00Z" {
		t.Errorf("create_time: got %v", attrs["create_time"])
	}
	// created_by_user
	userObj, ok := attrs["created_by_user"].(types.Object)
	if !ok || userObj.IsNull() {
		t.Fatal("created_by_user is null or wrong type")
	}
	if v, ok2 := userObj.Attributes()["name"].(types.String); !ok2 || v.ValueString() != "Eve" {
		t.Errorf("created_by_user.name: got %v", userObj.Attributes()["name"])
	}
	// last_change_request
	lcrObj, ok := attrs["last_change_request"].(types.Object)
	if !ok || lcrObj.IsNull() {
		t.Fatal("last_change_request is null or wrong type")
	}
	if v, ok2 := lcrObj.Attributes()["state"].(types.String); !ok2 || v.ValueString() != "reconcilled" {
		t.Errorf("last_change_request.state: got %v", lcrObj.Attributes()["state"])
	}
	// pricing
	pricingObj, ok := attrs["pricing"].(types.Object)
	if !ok || pricingObj.IsNull() {
		t.Fatal("pricing is null or wrong type")
	}
	if v, ok2 := pricingObj.Attributes()["month"].(types.Float64); !ok2 || v.ValueFloat64() != 100.0 {
		t.Errorf("pricing.month: got %v", pricingObj.Attributes()["month"])
	}
}

func TestBuildInfoObj_ExtraAttrs(t *testing.T) {
	data := map[string]interface{}{
		"status": map[string]interface{}{"state": "stable"},
	}
	obj := buildInfoObj(data,
		map[string]attr.Type{"public_ip": types.StringType},
		map[string]attr.Value{"public_ip": types.StringValue("1.2.3.4")},
	)
	if obj.IsNull() {
		t.Fatal("expected non-null object")
	}
	if v, ok := obj.Attributes()["public_ip"].(types.String); !ok || v.ValueString() != "1.2.3.4" {
		t.Errorf("extra attr public_ip: got %v", obj.Attributes()["public_ip"])
	}
	// base state still present
	if v, ok := obj.Attributes()["state"].(types.String); !ok || v.ValueString() != "stable" {
		t.Errorf("base state attr: got %v", obj.Attributes()["state"])
	}
}

// ---- commonInfoSchema ----

func TestCommonInfoSchema_HasBaseAttrs(t *testing.T) {
	s := commonInfoSchema(nil)
	nested, ok := s.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("commonInfoSchema did not return a SingleNestedAttribute")
	}
	for _, want := range []string{"state", "create_time", "created_by_user", "last_change_request", "pricing"} {
		if _, ok := nested.Attributes[want]; !ok {
			t.Errorf("expected base attr %q in commonInfoSchema", want)
		}
	}
}

func TestCommonInfoSchema_MergesExtraAttrs(t *testing.T) {
	s := commonInfoSchema(map[string]schema.Attribute{
		"custom_field": schema.StringAttribute{Computed: true},
	})
	nested, ok := s.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("expected SingleNestedAttribute")
	}
	if _, ok := nested.Attributes["custom_field"]; !ok {
		t.Error("extra attr 'custom_field' not merged into commonInfoSchema")
	}
	// base attrs still present
	if _, ok := nested.Attributes["state"]; !ok {
		t.Error("base attr 'state' missing after merging extras")
	}
}

func TestCommonInfoSchema_NestedObjectsHaveExpectedShape(t *testing.T) {
	s := commonInfoSchema(nil)
	nested := s.(schema.SingleNestedAttribute)

	// created_by_user must be SingleNestedAttribute with id + name
	cbu, ok := nested.Attributes["created_by_user"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("created_by_user should be SingleNestedAttribute")
	}
	for _, k := range []string{"id", "name"} {
		if _, ok := cbu.Attributes[k]; !ok {
			t.Errorf("created_by_user missing attr %q", k)
		}
	}

	// pricing must be SingleNestedAttribute with month/day/hour
	pricing, ok := nested.Attributes["pricing"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("pricing should be SingleNestedAttribute")
	}
	for _, k := range []string{"month", "day", "hour"} {
		if _, ok := pricing.Attributes[k]; !ok {
			t.Errorf("pricing missing attr %q", k)
		}
	}

	// last_change_request has nested created_by_user
	lcr, ok := nested.Attributes["last_change_request"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("last_change_request should be SingleNestedAttribute")
	}
	if _, ok := lcr.Attributes["created_by_user"]; !ok {
		t.Error("last_change_request missing nested created_by_user")
	}
	if _, ok := lcr.Attributes["error_message"]; !ok {
		t.Error("last_change_request missing error_message")
	}
}

// ---- populate propagates full info block (integration check via folder) ----

func TestPopulateFolderState_InfoBlockHasAllBaseFields(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "my-folder"},
		"status": map[string]interface{}{
			"state":      "stable",
			"createTime": "2026-06-28T09:00:00Z",
			"createdByUser": map[string]interface{}{
				"id": "user-1", "name": "Admin",
			},
			"lastChangeRequest": map[string]interface{}{
				"state":      "reconcilled",
				"createTime": "2026-06-28T08:00:00Z",
				"createdByUser": map[string]interface{}{
					"id": "user-1", "name": "Admin",
				},
			},
			"pricing": map[string]interface{}{
				"month": float64(5.0),
				"day":   float64(0.17),
				"hour":  float64(0.007),
			},
		},
	}

	var state FolderResourceModel
	if err := populateFolderState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateFolderState error: %v", err)
	}
	if state.Status.IsNull() {
		t.Fatal("info should not be null")
	}

	attrs := state.Status.Attributes()

	if v, ok := attrs["create_time"].(types.String); !ok || v.ValueString() != "2026-06-28T09:00:00Z" {
		t.Errorf("info.create_time: got %v", attrs["create_time"])
	}

	userObj, _ := attrs["created_by_user"].(types.Object)
	if userObj.IsNull() {
		t.Fatal("info.created_by_user is null")
	}
	if v, ok := userObj.Attributes()["name"].(types.String); !ok || v.ValueString() != "Admin" {
		t.Errorf("info.created_by_user.name: got %v", userObj.Attributes()["name"])
	}

	pObj, _ := attrs["pricing"].(types.Object)
	if pObj.IsNull() {
		t.Fatal("info.pricing is null")
	}
	if v, ok := pObj.Attributes()["month"].(types.Float64); !ok || v.ValueFloat64() != 5.0 {
		t.Errorf("info.pricing.month: got %v", pObj.Attributes()["month"])
	}

	lcrObj, _ := attrs["last_change_request"].(types.Object)
	if lcrObj.IsNull() {
		t.Fatal("info.last_change_request is null")
	}
	if v, ok := lcrObj.Attributes()["state"].(types.String); !ok || v.ValueString() != "reconcilled" {
		t.Errorf("info.last_change_request.state: got %v", lcrObj.Attributes()["state"])
	}
}

func TestPopulateFolderState_InfoNullFieldsWhenStatusEmpty(t *testing.T) {
	// Status present but no optional fields → nested objects should be null
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "f"},
		"status":   map[string]interface{}{"state": "stable"},
	}

	var state FolderResourceModel
	if err := populateFolderState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateFolderState error: %v", err)
	}

	attrs := state.Status.Attributes()
	if p, ok := attrs["pricing"].(types.Object); ok && !p.IsNull() {
		t.Error("pricing should be null when not in status")
	}
	if u, ok := attrs["created_by_user"].(types.Object); ok && !u.IsNull() {
		t.Error("created_by_user should be null when not in status")
	}
	if l, ok := attrs["last_change_request"].(types.Object); ok && !l.IsNull() {
		t.Error("last_change_request should be null when not in status")
	}
}
