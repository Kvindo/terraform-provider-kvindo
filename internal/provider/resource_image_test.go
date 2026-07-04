package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func imageSpecSchema(t *testing.T) map[string]schema.Attribute {
	t.Helper()
	r := NewImageResource().(*ImageResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("spec is not a SingleNestedAttribute")
	}
	return spec.Attributes
}

func TestImageSpec_VolumeIdAttributeName(t *testing.T) {
	spec := imageSpecSchema(t)
	if _, ok := spec["volume_id"]; !ok {
		t.Error("expected spec attribute 'volume_id'")
	}
}

func TestBuildImageRequestMap_VolumeId(t *testing.T) {
	plan := ImageResourceModel{
		ID:       types.StringValue("01abc"),
		Metadata: metadataModel{Name: types.StringValue("test-img"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: ImageSpecModel{
			VolumeId: types.StringValue("01vol123"),
		},
	}

	m := buildImageRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if spec["volumeId"] != "01vol123" {
		t.Errorf("expected volumeId=01vol123, got %v", spec["volumeId"])
	}
	if _, ok := spec["vmId"]; ok {
		t.Error("expected no vmId key when VmId is unset")
	}
}

func TestPopulateImageState_VolumeIdAndIsVmImage(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "test-img"},
		"spec":     map[string]interface{}{"volumeId": "01vol123"},
		"status":   map[string]interface{}{"state": "stable", "isVmImage": false, "sizeBytes": float64(1024), "volumes": "[]"},
	}

	var state ImageResourceModel
	if err := populateImageState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateImageState returned error: %v", err)
	}
	if state.Spec.VolumeId.ValueString() != "01vol123" {
		t.Errorf("expected spec.volume_id=01vol123, got %q", state.Spec.VolumeId.ValueString())
	}
	if v, ok := state.Status.Attributes()["is_vm_image"].(types.Bool); !ok || v.ValueBool() != false {
		t.Errorf("expected status.is_vm_image=false, got %v", state.Status.Attributes()["is_vm_image"])
	}
}
