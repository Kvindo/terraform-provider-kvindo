package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// why: the TF attribute is `status`, but the API property is TicketStatus — the wire key must be
// "ticketStatus" or the API silently ignores it (the swagger the generator consumes predates the
// C# rename, so a naive regen emits "status"). See the support_ticket override in tools/generator.
func TestBuildSupportTicketRequestMap_TicketStatusWireKey(t *testing.T) {
	plan := SupportTicketResourceModel{
		ID: types.StringValue("01abc"),
		Metadata: metadataModel{
			Name:     types.StringValue("test-ticket"),
			FolderID: types.StringNull(),
		},
		Spec: SupportTicketSpecModel{
			Kind:     types.StringValue("technical"),
			Severity: types.StringValue("high"),
			Status:   types.StringValue("opened"),
		},
	}

	m := buildSupportTicketRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if spec["ticketStatus"] != "opened" {
		t.Errorf("expected spec.ticketStatus='opened', got %v", spec["ticketStatus"])
	}
	if _, ok := spec["status"]; ok {
		t.Error("spec must not contain the key 'status' — the API only reads 'ticketStatus'")
	}
	if spec["kind"] != "technical" || spec["severity"] != "high" {
		t.Errorf("kind/severity keys regressed: %v", spec)
	}
}

func TestPopulateSupportTicketState_ReadsTicketStatus(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "test-ticket"},
		"spec": map[string]interface{}{
			"kind":         "technical",
			"severity":     "high",
			"ticketStatus": "opened",
		},
	}
	var state SupportTicketResourceModel
	if err := populateSupportTicketState(context.Background(), data, &state); err != nil {
		t.Fatalf("populate failed: %v", err)
	}
	if state.Spec.Status.ValueString() != "opened" {
		t.Errorf("expected Spec.Status='opened' from ticketStatus, got %v", state.Spec.Status)
	}
}

// Regression test for the generator's envelope-filter bug: extractFields unwraps the "spec"
// $ref into specProps, then filtered those unwrapped fields through envelopeFieldNames — which
// contains "kind" (the OUTER apiVersion/kind/metadata/spec/status envelope key). SupportTicketSpec
// is the only spec in the whole API with its own "kind" property, so spec.kind was silently
// dropped from resource_support_ticket.go on every full regen and had to be hand-re-added, with
// no error or warning to signal it. Fixed by applying envelopeFieldNames only when the spec was
// NOT unwrapped (the spec-less fallback, e.g. Folder, where specProps really are envelope keys).
//
// The sibling tests above guard the MODEL (they reference Spec.Kind, so a dropped field breaks
// compilation); this one guards the SCHEMA attribute the provider actually exposes to users,
// which is the thing the generator emits and the thing a practitioner writes in HCL.
func TestSupportTicketSchema_ExposesSpecKind(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	for _, newResource := range p.Resources(context.Background()) {
		r := newResource()

		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)
		if metaResp.TypeName != "kvindo_support_ticket" {
			continue
		}

		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Fatalf("schema build produced diagnostics: %v", schemaResp.Diagnostics)
		}

		specAttr, ok := schemaResp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
		if !ok {
			t.Fatal("kvindo_support_ticket: expected a SingleNestedAttribute 'spec'")
		}
		if _, ok := specAttr.Attributes["kind"]; !ok {
			t.Error("kvindo_support_ticket: spec.kind is missing from the schema — the generator's " +
				"envelopeFieldNames filter has regressed and is dropping it again")
		}
		return
	}
	t.Fatal("kvindo_support_ticket resource not found in the provider's resource list")
}
