package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// The transaction schema must expose each sub-resource as a map of nested metadata/spec/status
// objects (reusing the standalone resource schema), not the old flat shape.
func TestTransactionSchema_SubResourcesAreNested(t *testing.T) {
	r := NewTransactionResource().(*TransactionResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	// The transaction itself exposes the three blocks; sub-resources live under spec.
	for _, want := range []string{"id", "metadata", "spec", "status"} {
		if _, ok := resp.Schema.Attributes[want]; !ok {
			t.Errorf("transaction schema missing top-level %q", want)
		}
	}
	txnSpec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("transaction spec must be a SingleNestedAttribute")
	}
	if _, ok := txnSpec.Attributes["delete_resources_on_transaction_delete"]; !ok {
		t.Error("transaction.spec missing delete_resources_on_transaction_delete")
	}

	// One entry per registry type under spec, each a map of nested metadata/spec/status objects.
	for _, s := range txnSubs {
		attr, ok := txnSpec.Attributes[s.tfKey]
		if !ok {
			t.Errorf("transaction.spec missing sub-resource %q", s.tfKey)
			continue
		}
		mapAttr, ok := attr.(schema.MapNestedAttribute)
		if !ok {
			t.Errorf("spec.%q should be a MapNestedAttribute, got %T", s.tfKey, attr)
			continue
		}
		nested := mapAttr.NestedObject.Attributes
		for _, want := range []string{"id", "metadata", "status"} {
			if _, ok := nested[want]; !ok {
				t.Errorf("transaction.spec.%s element missing %q block", s.tfKey, want)
			}
		}
		if _, ok := nested["metadata"].(schema.SingleNestedAttribute); !ok {
			t.Errorf("transaction.spec.%s.metadata should be a SingleNestedAttribute", s.tfKey)
		}
	}

	// vms must carry the spec block with the nested fields (proves reuse of VmResourceSchemaAttrs).
	vms := txnSpec.Attributes["vms"].(schema.MapNestedAttribute)
	spec, ok := vms.NestedObject.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("transaction.spec.vms element missing spec block")
	}
	if _, ok := spec.Attributes["security_group_ids"]; !ok {
		t.Error("transaction.spec.vms.spec should include security_group_ids (reused from VmResourceSchemaAttrs)")
	}
}

func TestTransactionRegistry_CoversAllMaps(t *testing.T) {
	// Every registry tfKey must be a real schema attribute. There are 67 transactable sub-types
	// (the transaction's own "labels" map is not a sub-resource) - bumped from 61 when
	// on_off_schedules/vm_command_schedules/postgresqls/postgresql_users/postgresql_databases/
	// valkey_users were added (review finding #13: these 6 backend-supported TransactionSpec.cs
	// array properties had no txnSubs entry at all). If TransactionSpec.cs (avant.cloud) gains
	// another array property, update this count AND add its txnSubs entry together - nothing
	// automated catches one without the other.
	if len(txnSubs) != 67 {
		t.Errorf("expected 67 transactable sub-types, got %d", len(txnSubs))
	}
	seen := map[string]bool{}
	for _, s := range txnSubs {
		if seen[s.tfKey] {
			t.Errorf("duplicate registry tfKey %q", s.tfKey)
		}
		seen[s.tfKey] = true
		if s.apiKey == "" || s.build == nil || s.populate == nil || s.attrs == nil || s.field == nil {
			t.Errorf("registry entry %q has nil fields", s.tfKey)
		}
	}
	// Named-key check, not just the count above: a count-only assertion could stay green even if
	// one of these 6 were mis-added while an unrelated pre-existing entry was accidentally dropped.
	for _, want := range []string{
		"on_off_schedules", "vm_command_schedules", "postgresqls",
		"postgresql_users", "postgresql_databases", "valkey_users",
	} {
		if !seen[want] {
			t.Errorf("expected txnSubs to contain tfKey %q", want)
		}
	}
}

// The generic cross-ref recovery lowercases ULID id fields (so plan==state) but must NOT mangle
// other spec fields (region, names, ...). overrideIfConfigKnown drives that.
func TestOverrideIfConfigKnown_LowercasesOnlyIds(t *testing.T) {
	if !isIdKey("id") || !isIdKey("vpc_id") || !isIdKey("access_policy_ids") {
		t.Error("isIdKey should match id / *_id / *_ids")
	}
	if isIdKey("region") || isIdKey("name") || isIdKey("policy_json") {
		t.Error("isIdKey should not match non-id fields")
	}

	plan := map[string]attr.Value{
		"bucket_id": types.StringValue("placeholder"),
		"region":    types.StringValue("placeholder"),
	}
	cfg := map[string]attr.Value{
		"bucket_id": types.StringValue("01ABCDEF"), // uppercase ULID-ish
		"region":    types.StringValue("RU-MSK-1"), // must be preserved verbatim
	}
	overrideIfConfigKnown(plan, cfg, "bucket_id")
	overrideIfConfigKnown(plan, cfg, "region")

	if v := plan["bucket_id"].(types.String).ValueString(); v != "01abcdef" {
		t.Errorf("id field should be lowercased, got %q", v)
	}
	if v := plan["region"].(types.String).ValueString(); v != "RU-MSK-1" {
		t.Errorf("non-id field must be preserved verbatim, got %q", v)
	}
}

// nullAttrValue recursively builds a well-typed, all-fields-empty attr.Value for t - a concrete
// object/list/map with null-or-empty leaves rather than a bare ObjectNull/ListNull/MapNull, so it
// satisfies any downstream ObjectValue/ObjectValueMust type check regardless of that field's own
// nullability. Test-only helper for building minimal-but-real S3Bucket/S3User-shaped elements
// without hand-writing every one of their many fields.
func nullAttrValue(t attr.Type) attr.Value {
	switch tt := t.(type) {
	case types.ObjectType:
		vals := make(map[string]attr.Value, len(tt.AttrTypes))
		for k, at := range tt.AttrTypes {
			vals[k] = nullAttrValue(at)
		}
		return types.ObjectValueMust(tt.AttrTypes, vals)
	case types.ListType:
		return types.ListValueMust(tt.ElemType, []attr.Value{})
	case types.MapType:
		return types.MapValueMust(tt.ElemType, map[string]attr.Value{})
	case basetypes.BoolType:
		return types.BoolNull()
	case basetypes.Int64Type:
		return types.Int64Null()
	case basetypes.Float64Type:
		return types.Float64Null()
	default:
		// basetypes.StringType and anything else not explicitly handled above.
		return types.StringNull()
	}
}

// buildTxnElement builds a real-shaped transaction sub-resource map element (matching
// txnObjType(schemaAttrs()).AttrTypes exactly, as autoWireUsers/setSpecField require) with every
// field null/empty except "id" (top-level) and, if bucketID is non-empty, spec.bucket_id.
func buildTxnElement(t *testing.T, schemaAttrs map[string]schema.Attribute, id, bucketID string) types.Object {
	t.Helper()
	elemTypes := txnObjType(schemaAttrs).AttrTypes
	vals := make(map[string]attr.Value, len(elemTypes))
	for k, at := range elemTypes {
		vals[k] = nullAttrValue(at)
	}
	vals["id"] = types.StringValue(id)
	if bucketID != "" {
		specType := elemTypes["spec"].(types.ObjectType)
		specVals := vals["spec"].(types.Object).Attributes()
		newSpecVals := make(map[string]attr.Value, len(specVals))
		for k, v := range specVals {
			newSpecVals[k] = v
		}
		newSpecVals["bucket_id"] = types.StringValue(bucketID)
		vals["spec"] = types.ObjectValueMust(specType.AttrTypes, newSpecVals)
	}
	return types.ObjectValueMust(elemTypes, vals)
}

// Regression coverage for review finding #19: autoWireUsers picked "the first bucket in Go map
// iteration order" when an s3_user had no explicit bucket_id - Go map iteration is randomized per
// process, so a transaction with more than one bucket wired a different (arbitrary) bucket on every
// run. Runs the real function many times to make a non-deterministic pre-fix result actually show
// up as flaky (a single run has a real chance of "getting lucky" against the old bug).
func TestAutoWireUsers_PicksDeterministicBucket(t *testing.T) {
	bucketAttrs := S3BucketResourceSchemaAttrs()
	userAttrs := S3UserResourceSchemaAttrs()
	bucketElemType := txnObjType(bucketAttrs)

	buckets, diags := types.MapValue(bucketElemType, map[string]attr.Value{
		"bucket_z": buildTxnElement(t, bucketAttrs, "id-z", ""),
		"bucket_a": buildTxnElement(t, bucketAttrs, "id-a", ""),
		"bucket_m": buildTxnElement(t, bucketAttrs, "id-m", ""),
	})
	if diags.HasError() {
		t.Fatalf("failed to build buckets map: %v", diags)
	}

	var firstResult string
	for i := 0; i < 20; i++ {
		userElemType := txnObjType(userAttrs)
		users, diags := types.MapValue(userElemType, map[string]attr.Value{
			"user1": buildTxnElement(t, userAttrs, "user-id-1", ""), // no bucket_id set - needs auto-wiring
		})
		if diags.HasError() {
			t.Fatalf("failed to build users map: %v", diags)
		}

		result := autoWireUsers(context.Background(), users, buckets, types.MapNull(types.ObjectType{}))
		userObj := result.Elements()["user1"].(types.Object)
		bucketID, _ := getSpecString(userObj, "bucket_id")
		got := bucketID.ValueString()

		if i == 0 {
			firstResult = got
			continue
		}
		if got != firstResult {
			t.Fatalf("autoWireUsers picked a different bucket across runs (non-deterministic): run 0 got %q, run %d got %q", firstResult, i, got)
		}
	}
}
