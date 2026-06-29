package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func volumeSpecSchema(t *testing.T) map[string]schema.Attribute {
	t.Helper()
	r := NewVolumeResource().(*VolumeResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("spec is not a SingleNestedAttribute")
	}
	return spec.Attributes
}

// camelToSnake must keep the GiB acronym intact (size_gib, not size_gi_b).
func TestVolumeSpec_SizeGibAttributeName(t *testing.T) {
	spec := volumeSpecSchema(t)
	if _, ok := spec["size_gib"]; !ok {
		t.Error("expected spec attribute 'size_gib'")
	}
	if _, ok := spec["size_gi_b"]; ok {
		t.Error("found incorrect attribute 'size_gi_b'; should be 'size_gib'")
	}
}

func TestBuildVolumeRequestMap_SizeGib(t *testing.T) {
	plan := VolumeResourceModel{
		ID:       types.StringValue("01abc"),
		Metadata: metadataModel{Name: types.StringValue("test-vol"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: VolumeSpecModel{
			HostingProviderId: types.StringValue("provider-1"),
			OfferId:           types.StringValue("gp3-750"),
			SizeGib:           types.Int64Value(40),
			OsImageId:         types.StringValue("200"),
		},
	}

	m := buildVolumeRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if _, ok := spec["sizeGiB"]; !ok {
		t.Error("expected key 'sizeGiB' in spec map")
	}
	if _, ok := spec["sizeGib"]; ok {
		t.Error("found wrong key 'sizeGib' in spec map; should be 'sizeGiB'")
	}
	if spec["sizeGiB"] != int64(40) {
		t.Errorf("expected sizeGiB=40, got %v", spec["sizeGiB"])
	}
}

func TestPopulateVolumeState_Nested(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "test-vol"},
		"spec":     map[string]interface{}{"hostingProviderId": "provider-1", "offerId": "gp3-750", "sizeGiB": float64(40), "osImageId": "200"},
		"status":   map[string]interface{}{"state": "stable"},
	}

	var state VolumeResourceModel
	if err := populateVolumeState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateVolumeState returned error: %v", err)
	}
	if state.Spec.SizeGib.ValueInt64() != 40 {
		t.Errorf("expected spec.size_gib=40, got %d", state.Spec.SizeGib.ValueInt64())
	}
	if state.Metadata.Name.ValueString() != "test-vol" {
		t.Errorf("metadata.name: got %q", state.Metadata.Name.ValueString())
	}
}
