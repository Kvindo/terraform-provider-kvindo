package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

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
