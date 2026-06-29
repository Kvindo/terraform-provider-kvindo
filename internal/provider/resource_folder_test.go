package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFolderSchema_HasTopLevelBlocks(t *testing.T) {
	r := NewFolderResource().(*FolderResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	// Folder has no spec fields, so it exposes id + metadata + status (no spec block).
	for _, attr := range []string{"id", "metadata", "status"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q in folder schema", attr)
		}
	}
	if _, ok := resp.Schema.Attributes["spec"]; ok {
		t.Error("folder has no spec fields; spec block should be absent")
	}
	// Old flat fields must be gone.
	for _, gone := range []string{"name", "description", "folder_id", "labels", "info", "resource_name"} {
		if _, ok := resp.Schema.Attributes[gone]; ok {
			t.Errorf("flat attribute %q should not be at the root anymore", gone)
		}
	}
}

func TestBuildFolderRequestMap_Metadata(t *testing.T) {
	plan := FolderResourceModel{
		ID: types.StringValue("01abc"),
		Metadata: metadataModel{
			Name:        types.StringValue("test-folder"),
			Description: types.StringValue("desc"),
			FolderID:    types.StringNull(),
		},
	}

	m := buildFolderRequestMap(context.Background(), plan)
	metadata, ok := m["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'metadata' key with map value in request")
	}
	if metadata["name"] != "test-folder" {
		t.Errorf("expected metadata.name='test-folder', got %v", metadata["name"])
	}
	if metadata["id"] != "01abc" {
		t.Errorf("expected metadata.id='01abc', got %v", metadata["id"])
	}
	if _, ok := m["spec"].(map[string]interface{}); !ok {
		t.Error("request should still carry an (empty) spec map")
	}
}

func TestPopulateFolderState_NestedMetadataAndStatus(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "test-folder"},
		"status":   map[string]interface{}{"state": "stable"},
	}

	var state FolderResourceModel
	if err := populateFolderState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateFolderState returned error: %v", err)
	}
	// Root id mirrors metadata.id.
	if state.ID.ValueString() != "01abc" {
		t.Errorf("expected root ID='01abc', got %q", state.ID.ValueString())
	}
	if state.Metadata.ID.ValueString() != "01abc" {
		t.Errorf("expected metadata.id='01abc', got %q", state.Metadata.ID.ValueString())
	}
	if state.Metadata.Name.ValueString() != "test-folder" {
		t.Errorf("expected metadata.name='test-folder', got %q", state.Metadata.Name.ValueString())
	}
	if state.Status.IsNull() {
		t.Fatal("status should not be null")
	}
	if v, ok := state.Status.Attributes()["state"].(types.String); !ok || v.ValueString() != "stable" {
		t.Errorf("expected status.state='stable', got %v", state.Status.Attributes()["state"])
	}
}

func TestPopulateFolderState_RootFolderHasNullFolderID(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01root", "name": "root"},
	}

	var state FolderResourceModel
	if err := populateFolderState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateFolderState returned error: %v", err)
	}
	if !state.Metadata.FolderID.IsNull() {
		t.Errorf("expected metadata.folder_id null for root folder, got %q", state.Metadata.FolderID.ValueString())
	}
}
