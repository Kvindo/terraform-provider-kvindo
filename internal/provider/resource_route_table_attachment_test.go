package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vpc_id and vpc_subnet_id are mutually exclusive (server-enforced, no client-side validator -
// same convention as kvindo_volume's os_image_id/image_id). These mirror that resource's test
// shape: each field's build must omit the other's key, and the new field must round-trip through
// state.

func TestBuildRouteTableAttachmentRequestMap_VpcId(t *testing.T) {
	plan := RouteTableAttachmentResourceModel{
		ID:       types.StringValue("01abc"),
		Metadata: metadataModel{Name: types.StringValue("test-rta"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: RouteTableAttachmentSpecModel{
			RouteTableId: types.StringValue("rt-1"),
			VpcId:        types.StringValue("vpc-1"),
		},
	}

	m := buildRouteTableAttachmentRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if spec["vpcId"] != "vpc-1" {
		t.Errorf("expected vpcId=vpc-1, got %v", spec["vpcId"])
	}
	if _, ok := spec["vpcSubnetId"]; ok {
		t.Error("expected no vpcSubnetId key when VpcSubnetId is unset")
	}
}

func TestBuildRouteTableAttachmentRequestMap_VpcSubnetId(t *testing.T) {
	plan := RouteTableAttachmentResourceModel{
		ID:       types.StringValue("01abc"),
		Metadata: metadataModel{Name: types.StringValue("test-rta"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: RouteTableAttachmentSpecModel{
			RouteTableId: types.StringValue("rt-1"),
			VpcSubnetId:  types.StringValue("subnet-1"),
		},
	}

	m := buildRouteTableAttachmentRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if spec["vpcSubnetId"] != "subnet-1" {
		t.Errorf("expected vpcSubnetId=subnet-1, got %v", spec["vpcSubnetId"])
	}
	if _, ok := spec["vpcId"]; ok {
		t.Error("expected no vpcId key when VpcId is unset")
	}
}

func TestPopulateRouteTableAttachmentState_VpcSubnetId(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "test-rta"},
		"spec":     map[string]interface{}{"routeTableId": "rt-1", "vpcSubnetId": "subnet-1"},
		"status":   map[string]interface{}{"state": "stable"},
	}

	var state RouteTableAttachmentResourceModel
	if err := populateRouteTableAttachmentState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateRouteTableAttachmentState returned error: %v", err)
	}
	if state.Spec.VpcSubnetId.ValueString() != "subnet-1" {
		t.Errorf("expected spec.vpc_subnet_id=subnet-1, got %q", state.Spec.VpcSubnetId.ValueString())
	}
	if !state.Spec.VpcId.IsNull() {
		t.Errorf("expected spec.vpc_id to be null, got %q", state.Spec.VpcId.ValueString())
	}
}
