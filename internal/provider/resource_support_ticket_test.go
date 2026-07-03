package provider

import (
	"context"
	"testing"

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
