package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func vmCommandScheduleSpecSchema(t *testing.T) map[string]schema.Attribute {
	t.Helper()
	r := NewVmCommandScheduleResource().(*VmCommandScheduleResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("spec is not a SingleNestedAttribute")
	}
	return spec.Attributes
}

func TestVmCommandScheduleSpec_AttributeNames(t *testing.T) {
	spec := vmCommandScheduleSpecSchema(t)
	for _, name := range []string{"enabled", "schedule", "schedule_format", "command", "command_timeout_seconds"} {
		if _, ok := spec[name]; !ok {
			t.Errorf("expected spec attribute %q", name)
		}
	}
}

func TestBuildVmCommandScheduleRequestMap(t *testing.T) {
	plan := VmCommandScheduleResourceModel{
		ID:       types.StringValue("01abc"),
		Metadata: metadataModel{Name: types.StringValue("nightly-cleanup"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: VmCommandScheduleSpecModel{
			Enabled:               types.BoolValue(true),
			Schedule:              types.StringValue("0 4 * * *"),
			ScheduleFormat:        types.StringValue("cron"),
			Command:               types.StringValue("journalctl --vacuum-time=7d && apt-get clean"),
			CommandTimeoutSeconds: types.Int64Value(300),
		},
	}

	m := buildVmCommandScheduleRequestMap(context.Background(), plan)
	spec, ok := m["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'spec' key with map value in request")
	}
	if spec["command"] != "journalctl --vacuum-time=7d && apt-get clean" {
		t.Errorf("expected command to match, got %v", spec["command"])
	}
	if spec["commandTimeoutSeconds"] != int64(300) {
		t.Errorf("expected commandTimeoutSeconds=300, got %v", spec["commandTimeoutSeconds"])
	}
	if spec["scheduleFormat"] != "cron" {
		t.Errorf("expected scheduleFormat=cron, got %v", spec["scheduleFormat"])
	}
	if spec["schedule"] != "0 4 * * *" {
		t.Errorf("expected schedule='0 4 * * *', got %v", spec["schedule"])
	}
	if spec["enabled"] != true {
		t.Errorf("expected enabled=true, got %v", spec["enabled"])
	}
}

func TestPopulateVmCommandScheduleState(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01abc", "name": "nightly-cleanup"},
		"spec": map[string]interface{}{
			"enabled": true, "schedule": "0 4 * * *", "scheduleFormat": "cron",
			"command": "journalctl --vacuum-time=7d && apt-get clean", "commandTimeoutSeconds": float64(300),
		},
		"status": map[string]interface{}{"state": "stable"},
	}

	var state VmCommandScheduleResourceModel
	if err := populateVmCommandScheduleState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateVmCommandScheduleState returned error: %v", err)
	}
	if state.Spec.Command.ValueString() != "journalctl --vacuum-time=7d && apt-get clean" {
		t.Errorf("expected spec.command to match, got %q", state.Spec.Command.ValueString())
	}
	if state.Spec.CommandTimeoutSeconds.ValueInt64() != 300 {
		t.Errorf("expected spec.command_timeout_seconds=300, got %d", state.Spec.CommandTimeoutSeconds.ValueInt64())
	}
	if state.Metadata.Name.ValueString() != "nightly-cleanup" {
		t.Errorf("metadata.name: got %q", state.Metadata.Name.ValueString())
	}
}
