package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vmSpecSchema returns the spec block's nested attributes for assertions.
func vmSpecSchema(t *testing.T) map[string]schema.Attribute {
	t.Helper()
	r := NewVmResource().(*VmResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("spec is not a SingleNestedAttribute")
	}
	return spec.Attributes
}

func TestVmSchema_TopLevelBlocks(t *testing.T) {
	r := NewVmResource().(*VmResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	for _, attr := range []string{"id", "metadata", "spec", "status"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected top-level attribute %q in vm schema", attr)
		}
	}
	// No flat fields at the root.
	for _, gone := range []string{"name", "vm_state", "offer_id", "info", "resource_name", "security_group_ids"} {
		if _, ok := resp.Schema.Attributes[gone]; ok {
			t.Errorf("attribute %q should not be at the root (belongs in metadata/spec)", gone)
		}
	}
}

func TestVmSpec_HasExpectedFields(t *testing.T) {
	spec := vmSpecSchema(t)
	for _, attr := range []string{
		"vm_state", "vpc_subnet_id", "floating_ip_id", "image_id", "offer_id",
		"ssh_key_ids", "bootstrap_command", "security_group_ids", "os_type",
		"boot_volume_attachment",
	} {
		if _, ok := spec[attr]; !ok {
			t.Errorf("expected spec attribute %q in vm schema", attr)
		}
	}
	if _, ok := spec["resource_name"]; ok {
		t.Error("resource_name was removed and must not appear in spec")
	}
}

func TestVmSpec_BootVolumeAttachmentFields(t *testing.T) {
	spec := vmSpecSchema(t)
	bva, ok := spec["boot_volume_attachment"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("boot_volume_attachment is not a SingleNestedAttribute")
	}
	volumeId, ok := bva.Attributes["volume_id"].(schema.StringAttribute)
	if !ok || !volumeId.Required {
		t.Error("boot_volume_attachment.volume_id should be Required")
	}
	attachmentId, ok := bva.Attributes["attachment_id"].(schema.StringAttribute)
	if !ok || !attachmentId.Computed || attachmentId.Required {
		t.Error("boot_volume_attachment.attachment_id should be Computed-only (internal bookkeeping)")
	}
}

// ---- boot_volume_attachment orchestration (behind-the-scenes volume_attachment) ----

func TestResolvedVmState_DefaultsToRunning(t *testing.T) {
	if got := resolvedVmState(VmSpecModel{VmState: types.StringNull()}); got != "running" {
		t.Errorf("expected default \"running\", got %q", got)
	}
	if got := resolvedVmState(VmSpecModel{VmState: types.StringValue("stopped")}); got != "stopped" {
		t.Errorf("expected \"stopped\", got %q", got)
	}
}

func TestVmCreateRequiresBootVolumeAttachment(t *testing.T) {
	running := VmSpecModel{VmState: types.StringValue("running"), BootVolumeAttachment: types.ObjectNull(bootVolumeAttachmentAttrTypes)}
	if !vmCreateRequiresBootVolumeAttachment(running) {
		t.Error("running VM with no boot_volume_attachment should require it")
	}

	stopped := VmSpecModel{VmState: types.StringValue("stopped"), BootVolumeAttachment: types.ObjectNull(bootVolumeAttachmentAttrTypes)}
	if vmCreateRequiresBootVolumeAttachment(stopped) {
		t.Error("stopped VM should not require boot_volume_attachment")
	}

	obj, _ := types.ObjectValue(bootVolumeAttachmentAttrTypes, map[string]attr.Value{
		"volume_id": types.StringValue("01vol"), "attachment_id": types.StringValue("01att"),
	})
	runningWithAttachment := VmSpecModel{VmState: types.StringValue("running"), BootVolumeAttachment: obj}
	if vmCreateRequiresBootVolumeAttachment(runningWithAttachment) {
		t.Error("running VM with boot_volume_attachment set should not require it again")
	}
}

func TestBuildBootVolumeAttachmentPlan(t *testing.T) {
	vmPlan := VmResourceModel{
		Metadata: metadataModel{Name: types.StringValue("web-1"), FolderID: types.StringValue("01folder")},
	}
	att := buildBootVolumeAttachmentPlan(vmPlan, "01vm", "01att", "01vol")

	if att.ID.ValueString() != "01att" {
		t.Errorf("attachment id: got %q", att.ID.ValueString())
	}
	if att.Metadata.Name.ValueString() != "web-1-boot" {
		t.Errorf("attachment name: got %q", att.Metadata.Name.ValueString())
	}
	if att.Metadata.FolderID.ValueString() != "01folder" {
		t.Errorf("attachment folder_id should match the vm's: got %q", att.Metadata.FolderID.ValueString())
	}
	if att.Spec.VmId.ValueString() != "01vm" || att.Spec.VolumeId.ValueString() != "01vol" {
		t.Errorf("attachment vm_id/volume_id: got %q/%q", att.Spec.VmId.ValueString(), att.Spec.VolumeId.ValueString())
	}
	if att.Spec.VmDeviceIndex.ValueInt64() != 0 {
		t.Errorf("boot volume must be device index 0, got %d", att.Spec.VmDeviceIndex.ValueInt64())
	}
}

// bootstrap_command dropped success_return_code/timeout_seconds: neither field was ever enforced
// server-side, and the VM now reports its actual execution result via status.bootstrap_command_*.
func TestVmSpec_BootstrapCommandOnlyHasCommandField(t *testing.T) {
	spec := vmSpecSchema(t)
	bc, ok := spec["bootstrap_command"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("bootstrap_command is not a SingleNestedAttribute")
	}
	if _, ok := bc.Attributes["command"]; !ok {
		t.Error("bootstrap_command.command should be present")
	}
	for _, gone := range []string{"success_return_code", "timeout_seconds"} {
		if _, ok := bc.Attributes[gone]; ok {
			t.Errorf("bootstrap_command.%s was removed and must not appear", gone)
		}
	}
}

func TestVmStatus_HasBootstrapCommandResultObject(t *testing.T) {
	r := NewVmResource().(*VmResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	status, ok := resp.Schema.Attributes["status"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("status is not a SingleNestedAttribute")
	}
	bc, ok := status.Attributes["bootstrap_command"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("status.bootstrap_command is not a SingleNestedAttribute")
	}
	if _, ok := bc.Attributes["return_code"].(schema.Int64Attribute); !ok {
		t.Error("status.bootstrap_command.return_code not found or wrong type")
	}
	if _, ok := bc.Attributes["output"].(schema.StringAttribute); !ok {
		t.Error("status.bootstrap_command.output not found or wrong type")
	}
	if _, ok := bc.Attributes["duration_ms"].(schema.Int64Attribute); !ok {
		t.Error("status.bootstrap_command.duration_ms not found or wrong type")
	}
}

func TestVmStatus_WindowsPasswordSensitive(t *testing.T) {
	r := NewVmResource().(*VmResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	status, ok := resp.Schema.Attributes["status"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("status is not a SingleNestedAttribute")
	}
	wap, ok := status.Attributes["windows_administrator_password"].(schema.StringAttribute)
	if !ok {
		t.Fatal("status.windows_administrator_password not found or wrong type")
	}
	if !wap.Sensitive {
		t.Error("status.windows_administrator_password should be Sensitive")
	}
	// base status fields present
	for _, want := range []string{"state", "create_time", "created_by_user", "last_change_request", "pricing"} {
		if _, ok := status.Attributes[want]; !ok {
			t.Errorf("status missing base attr %q", want)
		}
	}
}

func TestVmMetadata_IsRequiredBlock(t *testing.T) {
	r := NewVmResource().(*VmResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	md, ok := resp.Schema.Attributes["metadata"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("metadata is not a SingleNestedAttribute")
	}
	if !md.Required {
		t.Error("metadata block should be Required")
	}
	if name, ok := md.Attributes["name"].(schema.StringAttribute); !ok || !name.Required {
		t.Error("metadata.name should be Required")
	}
}

// ---- buildVmRequestMap (metadata + spec nesting) ----

func baseVmPlan() VmResourceModel {
	return VmResourceModel{
		ID:       types.StringValue("01vm"),
		Metadata: metadataModel{Name: types.StringValue("my-vm"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec: VmSpecModel{
			SecurityGroupIds:   types.ListNull(types.StringType),
			OsType:             types.StringNull(),
			SshKeyIds:          types.ListNull(types.StringType),
			ImageScheduleIds:   types.ListNull(types.StringType),
			CommandScheduleIds: types.ListNull(types.StringType),
			OnOffScheduleIds:   types.ListNull(types.StringType),
			BootstrapCommand:   types.ObjectNull(objAttrTypes(vmBootstrapCommandObjFields)),
		},
	}
}

func TestBuildVmRequestMap_MetadataAndSpec(t *testing.T) {
	plan := baseVmPlan()
	plan.Spec.OfferId = types.StringValue("offer-1")
	m := buildVmRequestMap(context.Background(), plan)

	md, ok := m["metadata"].(map[string]interface{})
	if !ok || md["name"] != "my-vm" {
		t.Fatalf("metadata.name not set: %v", m["metadata"])
	}
	spec, ok := m["spec"].(map[string]interface{})
	if !ok || spec["offerId"] != "offer-1" {
		t.Fatalf("spec.offerId not set: %v", m["spec"])
	}
}

func TestBuildVmRequestMap_SecurityGroupIds(t *testing.T) {
	plan := baseVmPlan()
	plan.Spec.SecurityGroupIds, _ = types.ListValue(types.StringType, []attr.Value{
		types.StringValue("sg-aaa"), types.StringValue("sg-bbb"),
	})
	m := buildVmRequestMap(context.Background(), plan)
	spec := m["spec"].(map[string]interface{})
	sgRaw, ok := spec["securityGroupIds"].([]interface{})
	if !ok || len(sgRaw) != 2 {
		t.Fatalf("expected 2 securityGroupIds in spec, got %v", spec["securityGroupIds"])
	}
}

func TestBuildVmRequestMap_SecurityGroupIdsNull(t *testing.T) {
	m := buildVmRequestMap(context.Background(), baseVmPlan())
	spec := m["spec"].(map[string]interface{})
	if _, ok := spec["securityGroupIds"]; ok {
		t.Error("securityGroupIds should be omitted when null")
	}
}

func TestBuildVmRequestMap_OsType(t *testing.T) {
	plan := baseVmPlan()
	plan.Spec.OsType = types.StringValue("windows")
	m := buildVmRequestMap(context.Background(), plan)
	spec := m["spec"].(map[string]interface{})
	if spec["osType"] != "windows" {
		t.Errorf("expected osType='windows', got %v", spec["osType"])
	}
}

// ---- populateVmState (nested metadata/spec/status) ----

func makeVmApiData(spec, status map[string]interface{}) map[string]interface{} {
	d := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01vm", "name": "my-vm"},
		"spec":     spec,
	}
	if status != nil {
		d["status"] = status
	}
	return d
}

func TestPopulateVmState_SpecAndMetadata(t *testing.T) {
	data := makeVmApiData(
		map[string]interface{}{"osType": "windows", "securityGroupIds": []interface{}{"sg-1", "sg-2", "sg-3"}},
		map[string]interface{}{"state": "stable"},
	)
	var state VmResourceModel
	if err := populateVmState(context.Background(), data, &state); err != nil {
		t.Fatalf("populateVmState error: %v", err)
	}
	if state.ID.ValueString() != "01vm" || state.Metadata.Name.ValueString() != "my-vm" {
		t.Errorf("metadata not populated: id=%q name=%q", state.ID.ValueString(), state.Metadata.Name.ValueString())
	}
	if state.Spec.OsType.ValueString() != "windows" {
		t.Errorf("spec.os_type: got %q", state.Spec.OsType.ValueString())
	}
	if len(state.Spec.SecurityGroupIds.Elements()) != 3 {
		t.Errorf("expected 3 security_group_ids, got %d", len(state.Spec.SecurityGroupIds.Elements()))
	}
}

func TestPopulateVmState_SecurityGroupIdsEmpty(t *testing.T) {
	data := makeVmApiData(map[string]interface{}{}, map[string]interface{}{"state": "stable"})
	var state VmResourceModel
	if err := populateVmState(context.Background(), data, &state); err != nil {
		t.Fatalf("populateVmState error: %v", err)
	}
	if state.Spec.SecurityGroupIds.IsNull() || len(state.Spec.SecurityGroupIds.Elements()) != 0 {
		t.Errorf("security_group_ids should be empty list, got %v", state.Spec.SecurityGroupIds)
	}
}

// status uses case-insensitive lookup: wire returns lowercase keys, generator emits camelCase.
func TestPopulateVmState_WindowsPassword_CaseInsensitive(t *testing.T) {
	data := makeVmApiData(
		map[string]interface{}{},
		map[string]interface{}{
			"state":                        "stable",
			"windowsAdministratorPassword": "S3cr3t!",
			"publicipv4":                   "1.2.3.4", // wire lowercase, schema key is public_ipv4
		},
	)
	var state VmResourceModel
	if err := populateVmState(context.Background(), data, &state); err != nil {
		t.Fatalf("populateVmState error: %v", err)
	}
	attrs := state.Status.Attributes()
	if v, ok := attrs["windows_administrator_password"].(types.String); !ok || v.ValueString() != "S3cr3t!" {
		t.Errorf("windows_administrator_password: got %v", attrs["windows_administrator_password"])
	}
	if v, ok := attrs["public_ipv4"].(types.String); !ok || v.ValueString() != "1.2.3.4" {
		t.Errorf("public_ipv4 (case-insensitive lookup of 'publicipv4'): got %v", attrs["public_ipv4"])
	}
}

func TestPopulateVmState_StatusFullBaseFields(t *testing.T) {
	data := makeVmApiData(
		map[string]interface{}{},
		map[string]interface{}{
			"state":      "stable",
			"createTime": "2026-06-28T00:00:00Z",
			"pricing":    map[string]interface{}{"month": float64(50.0), "day": float64(1.67), "hour": float64(0.07)},
		},
	)
	var state VmResourceModel
	if err := populateVmState(context.Background(), data, &state); err != nil {
		t.Fatalf("populateVmState error: %v", err)
	}
	attrs := state.Status.Attributes()
	if v, ok := attrs["create_time"].(types.String); !ok || v.ValueString() != "2026-06-28T00:00:00Z" {
		t.Errorf("status.create_time: got %v", attrs["create_time"])
	}
	pObj, _ := attrs["pricing"].(types.Object)
	if pObj.IsNull() {
		t.Fatal("status.pricing should not be null")
	}
	if v, ok := pObj.Attributes()["month"].(types.Float64); !ok || v.ValueFloat64() != 50.0 {
		t.Errorf("status.pricing.month: got %v", pObj.Attributes()["month"])
	}
}

// ---- resource_name removal (ssh_key / certificate / billing_account) ----

func TestSshKeySpec_NoResourceName(t *testing.T) {
	r := NewSshKeyResource().(*SshKeyResource)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	spec := resp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
	if _, ok := spec.Attributes["resource_name"]; ok {
		t.Error("resource_name should not appear in ssh_key spec")
	}
	if _, ok := spec.Attributes["public_key"]; !ok {
		t.Error("ssh_key spec should have public_key")
	}
}

func TestBuildSshKeyRequestMap_NoResourceName(t *testing.T) {
	plan := SshKeyResourceModel{
		ID:       types.StringValue("01key"),
		Metadata: metadataModel{Name: types.StringValue("my-key"), Description: types.StringNull(), FolderID: types.StringNull(), Labels: types.MapNull(types.StringType)},
		Spec:     SshKeySpecModel{PublicKey: types.StringValue("ssh-ed25519 AAAA...")},
	}
	m := buildSshKeyRequestMap(context.Background(), plan)
	spec := m["spec"].(map[string]interface{})
	if _, ok := spec["resourceName"]; ok {
		t.Error("resourceName should not be in spec")
	}
	if spec["publicKey"] != "ssh-ed25519 AAAA..." {
		t.Errorf("publicKey: got %v", spec["publicKey"])
	}
}

func TestPopulateSshKeyState_Nested(t *testing.T) {
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01key", "name": "my-key"},
		"spec":     map[string]interface{}{"publicKey": "ssh-ed25519 AAAA..."},
		"status":   map[string]interface{}{"state": "stable"},
	}
	var state SshKeyResourceModel
	if err := populateSshKeyState(context.Background(), apiData, &state); err != nil {
		t.Fatalf("populateSshKeyState error: %v", err)
	}
	if state.Spec.PublicKey.ValueString() != "ssh-ed25519 AAAA..." {
		t.Errorf("spec.public_key: got %q", state.Spec.PublicKey.ValueString())
	}
	if v, ok := state.Status.Attributes()["state"].(types.String); !ok || v.ValueString() != "stable" {
		t.Errorf("status.state: got %v", state.Status.Attributes()["state"])
	}
}
