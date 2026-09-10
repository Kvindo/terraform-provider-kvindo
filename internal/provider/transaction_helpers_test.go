package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakePopModel/fakePopAttrs back TestTxnPop below - a minimal stand-in for a real
// populate<Sn>State/<Sn>ResourceSchemaAttrs pair, just enough to exercise txnPop's own error
// threading in isolation.
type fakePopModel struct {
	ID types.String `tfsdk:"id"`
}

func fakePopAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{"id": schema.StringAttribute{Computed: true}}
}

// Regression coverage for review finding #14: txnPop used to discard popFn's error entirely
// (`_ = popFn(ctx, item, &e)`), so a genuinely failed per-item populate silently produced a
// half-filled object that txnPopulateSubResources then wrote into state as if it had succeeded.
func TestTxnPop_PropagatesPopFnError(t *testing.T) {
	wantErr := errors.New("boom")
	popFn := func(_ context.Context, _ map[string]interface{}, _ *fakePopModel) error {
		return wantErr
	}
	pop := txnPop[fakePopModel](popFn, fakePopAttrs)

	obj, id, err := pop(context.Background(), map[string]interface{}{"id": "abc"})
	if err == nil {
		t.Fatal("expected the popFn error to propagate, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("expected the exact popFn error to propagate, got %v", err)
	}
	if id != "" {
		t.Errorf("expected empty id on error, got %q", id)
	}
	if !obj.IsNull() {
		t.Errorf("expected a null object on error, not a half-filled one, got %v", obj)
	}
}

func TestTxnPop_SuccessPath(t *testing.T) {
	popFn := func(_ context.Context, item map[string]interface{}, e *fakePopModel) error {
		id, _ := item["id"].(string)
		e.ID = types.StringValue(id)
		return nil
	}
	pop := txnPop[fakePopModel](popFn, fakePopAttrs)

	obj, id, err := pop(context.Background(), map[string]interface{}{"id": "abc123"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if id != "abc123" {
		t.Errorf("expected id abc123, got %q", id)
	}
	if obj.IsNull() {
		t.Error("expected a non-null object on success")
	}
}

// TestTxnRelaxAttr_PreservesPlanModifiers pins the bug fixed 2026-07-20: txnRelaxAttr used to
// rebuild every attribute from scratch as Optional+Computed with no PlanModifiers at all, which
// meant a transaction sub-resource's Computed fields (metadata.id/delete_protection/labels, and
// any nested status block via volatileInfoModifier) could never freeze on an idle re-plan -
// `plan -detailed-exitcode` could never report zero drift for ANY transaction sub-resource.
// Covers every schema.Attribute case txnRelaxAttr switches on.
func TestTxnRelaxAttr_PreservesPlanModifiers(t *testing.T) {
	stringMods := []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
	boolMods := []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}

	cases := []struct {
		name string
		in   schema.Attribute
	}{
		{
			name: "StringAttribute",
			in:   schema.StringAttribute{Computed: true, PlanModifiers: stringMods},
		},
		{
			name: "BoolAttribute",
			in:   schema.BoolAttribute{Computed: true, PlanModifiers: boolMods},
		},
		{
			name: "SingleNestedAttribute",
			in: schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{Computed: true, PlanModifiers: stringMods},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := txnRelaxAttr(c.in)

			switch v := out.(type) {
			case schema.StringAttribute:
				if len(v.PlanModifiers) == 0 {
					t.Error("expected PlanModifiers to be preserved on relaxed StringAttribute, got none")
				}
				if !v.Optional || !v.Computed {
					t.Error("expected relaxed StringAttribute to be Optional+Computed")
				}
			case schema.BoolAttribute:
				if len(v.PlanModifiers) == 0 {
					t.Error("expected PlanModifiers to be preserved on relaxed BoolAttribute, got none")
				}
			case schema.SingleNestedAttribute:
				idAttr, ok := v.Attributes["id"].(schema.StringAttribute)
				if !ok {
					t.Fatalf("expected nested 'id' attribute to remain a StringAttribute, got %T", v.Attributes["id"])
				}
				if len(idAttr.PlanModifiers) == 0 {
					t.Error("expected PlanModifiers to be preserved on a nested leaf inside a relaxed SingleNestedAttribute, got none")
				}
			default:
				t.Fatalf("unexpected relaxed attribute type %T for case %q", out, c.name)
			}
		})
	}
}
