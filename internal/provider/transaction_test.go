package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	// Every registry tfKey must be a real schema attribute. There are 59 transactable sub-types
	// (the transaction's own "labels" map is not a sub-resource).
	if len(txnSubs) != 59 {
		t.Errorf("expected 59 transactable sub-types, got %d", len(txnSubs))
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
