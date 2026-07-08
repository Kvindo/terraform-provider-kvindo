package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func vmOnOffMaintenanceActionSpecSchema(t *testing.T) map[string]schema.Attribute {
	t.Helper()
	r := NewVmOnOffMaintenanceActionResource().(*VmOnOffMaintenanceActionResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("spec is not a SingleNestedAttribute")
	}
	return spec.Attributes
}

func TestVmOnOffMaintenanceActionSpec_AttributeNames(t *testing.T) {
	spec := vmOnOffMaintenanceActionSpecSchema(t)
	for _, name := range []string{"enabled", "schedule", "schedule_format", "target_state"} {
		if _, ok := spec[name]; !ok {
			t.Errorf("expected spec attribute %q", name)
		}
	}
}

func TestBuildVmOnOffMaintenanceActionRequestMap(t *testing.T) {
	plan := VmOnOffMaintenanceActionResourceModel{
		ID:       types.StringValue("01abc"),
		Metadata: metadataModel{Name: types.StringValue("evening-stop"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: VmOnOffMaintenanceActionSpecModel{
			Enabled:        types.BoolValue(true),
			Schedule:       types.StringValue("0 20 * * *"),
			ScheduleFormat: types.StringValue("cron"),
			TargetState:    types.StringValue("stopped"),
		},
	}

	m := buildVmOnOffMaintenanceActionRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if spec["targetState"] != "stopped" {
		t.Errorf("expected targetState=stopped, got %v", spec["targetState"])
	}
	if spec["scheduleFormat"] != "cron" {
		t.Errorf("expected scheduleFormat=cron, got %v", spec["scheduleFormat"])
	}
	if spec["schedule"] != "0 20 * * *" {
		t.Errorf("expected schedule='0 20 * * *', got %v", spec["schedule"])
	}
	if spec["enabled"] != true {
		t.Errorf("expected enabled=true, got %v", spec["enabled"])
	}
}

func TestPopulateVmOnOffMaintenanceActionState(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "evening-stop"},
		"spec":     map[string]interface{}{"enabled": true, "schedule": "0 20 * * *", "scheduleFormat": "cron", "targetState": "stopped"},
		"status":   map[string]interface{}{"state": "stable"},
	}

	var state VmOnOffMaintenanceActionResourceModel
	if err := populateVmOnOffMaintenanceActionState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateVmOnOffMaintenanceActionState returned error: %v", err)
	}
	if state.Spec.TargetState.ValueString() != "stopped" {
		t.Errorf("expected spec.target_state=stopped, got %q", state.Spec.TargetState.ValueString())
	}
	if state.Metadata.Name.ValueString() != "evening-stop" {
		t.Errorf("metadata.name: got %q", state.Metadata.Name.ValueString())
	}
}
