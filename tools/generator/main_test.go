package main

import (
	"strings"
	"testing"
)

// camelToSnake must keep IP-version suffixes as one segment (public_ipv4, not public_ip_v4) so the
// generated Terraform schema keys match the platform-wide convention. Regression guard for the
// ipVersionAcronyms normalization.
func TestCamelToSnakeIpVersion(t *testing.T) {
	cases := map[string]string{
		"publicIpV4":       "public_ipv4",
		"privateIpV6":      "private_ipv6",
		"natPublicIpV4":    "nat_public_ipv4",
		"sshIpV4":          "ssh_ipv4",
		"ipV4Cidrs":        "ipv4_cidrs",
		"allowedIpV4Cidrs": "allowed_ipv4_cidrs",
		"assignPublicIpV4": "assign_public_ipv4",
		"createPublicIpv4": "create_public_ipv4", // already-correct form is left untouched
		"volumeSizeGiB":    "volume_size_gib",    // existing GiB normalization still holds
	}
	for in, want := range cases {
		if got := camelToSnake(in); got != want {
			t.Errorf("camelToSnake(%q) = %q, want %q", in, got, want)
		}
	}
}

// extractFields must unwrap the envelope shape (apiVersion/kind/metadata/spec/status on the
// PUT-body schema) and return the resolved spec schema's own properties, not the envelope keys
// themselves. Regression guard: a prior version nested apiVersion/kind/metadata/spec/status
// inside the resource's own spec, and even fabricated a spec.spec field, for every resource.
func TestExtractFields_UnwrapsEnvelopeSpec(t *testing.T) {
	schemas := map[string]SchemaObject{
		"WidgetResource": {
			Properties: map[string]*SchemaRef{
				"apiVersion": {Type: "string"},
				"kind":       {Type: "string"},
				"metadata":   {Ref: "#/components/schemas/ResourceMetadata"},
				"spec":       {Ref: "#/components/schemas/WidgetSpec"},
				"status":     {Ref: "#/components/schemas/WidgetResourceInfo"},
			},
		},
		"WidgetSpec": {
			Properties: map[string]*SchemaRef{
				"tier":      {Type: "string"},
				"vmOfferId": {Type: "string"},
			},
		},
		"WidgetResourceInfo": {
			Properties: map[string]*SchemaRef{
				"state":      {Type: "string"}, // base info field, filtered by baseInfoFields
				"createTime": {Type: "string"}, // base info field, filtered by baseInfoFields
				"host":       {Type: "string"}, // resource-specific extra
			},
		},
	}

	fields, infoFields := extractFields(schemas["WidgetResource"], schemas)

	names := map[string]bool{}
	for _, f := range fields {
		names[f.TFName] = true
	}
	for _, envelopeKey := range []string{"api_version", "kind", "metadata", "spec", "status"} {
		if names[envelopeKey] {
			t.Errorf("fields must not contain envelope key %q, got %+v", envelopeKey, fields)
		}
	}
	if !names["tier"] || !names["vm_offer_id"] {
		t.Errorf("fields must contain the unwrapped spec's own properties, got %+v", fields)
	}

	if len(infoFields) != 1 || infoFields[0].TFName != "host" {
		t.Errorf("infoFields should contain only the resource-specific status extra %q, got %+v", "host", infoFields)
	}
}

// A resource whose top-level schema has no "spec" key at all (e.g. Folder) has no spec fields of
// its own — it must NOT fall back to treating apiVersion/kind/metadata/status as spec fields.
func TestExtractFields_SpecLessResourceYieldsNoFields(t *testing.T) {
	schemas := map[string]SchemaObject{
		"FolderResource": {
			Properties: map[string]*SchemaRef{
				"apiVersion": {Type: "string"},
				"kind":       {Type: "string"},
				"metadata":   {Ref: "#/components/schemas/ResourceMetadataWithOptionalFolder"},
				"status":     {Ref: "#/components/schemas/ResourceInfo"},
			},
		},
	}

	fields, _ := extractFields(schemas["FolderResource"], schemas)
	if len(fields) != 0 {
		t.Errorf("spec-less resource should yield no fields, got %+v", fields)
	}
}

// The clobber guard exists because a regen used to overwrite generated files unconditionally,
// silently deleting hand-written code they had accumulated. It compares two axes, and this test
// pins why BOTH are needed.
//
// The declaration axis alone was shipped first and proved insufficient within the day:
// datasource_vm.go would have lost the bootstrap_command status attributes (duration_ms, output,
// return_code) while losing zero top-level declarations, because they live inside a function the
// regen keeps. Only the key axis catches that.
func TestLostContent_BothAxesAreNecessary(t *testing.T) {
	// A helper deleted wholesale — the declaration axis catches it, the key axis does not.
	oldDecl := "package provider\n\nfunc buildBootVolumeAttachmentPlan() {}\n"
	newDecl := "package provider\n"
	if got := missingFrom(topLevelDeclNames(oldDecl), topLevelDeclNames(newDecl)); len(got) != 1 || got[0] != "buildBootVolumeAttachmentPlan" {
		t.Errorf("declaration axis: want [buildBootVolumeAttachmentPlan], got %v", got)
	}

	// Attributes deleted from INSIDE a surviving function — the real datasource_vm.go shape.
	// The declaration axis sees nothing; the key axis must catch all three.
	oldKeys := `package provider
func (d *VmDataSource) Schema() {
	m := map[string]schema.Attribute{
		"bootstrap_command": schema.SingleNestedAttribute{Attributes: map[string]schema.Attribute{
			"return_code": schema.Int64Attribute{Computed: true},
			"output":      schema.StringAttribute{Computed: true},
			"duration_ms": schema.Int64Attribute{Computed: true},
		}},
	}
}
`
	newKeys := `package provider
func (d *VmDataSource) Schema() {
	m := map[string]schema.Attribute{
		"bootstrap_command": schema.SingleNestedAttribute{},
	}
}
`
	if got := missingFrom(topLevelDeclNames(oldKeys), topLevelDeclNames(newKeys)); len(got) != 0 {
		t.Errorf("declaration axis should see nothing here, got %v", got)
	}
	gotKeys := missingFrom(quotedKeyNames(oldKeys), quotedKeyNames(newKeys))
	want := map[string]bool{"duration_ms": true, "output": true, "return_code": true}
	if len(gotKeys) != len(want) {
		t.Fatalf("key axis: want %d lost keys, got %v", len(want), gotKeys)
	}
	for _, k := range gotKeys {
		if !want[k] {
			t.Errorf("key axis: unexpected lost key %q", k)
		}
	}
}

// Reformatting must not trip the guard. These files pack long single-line attribute maps, so a
// line-level diff reports a merely-rewrapped key as both removed and added; set comparison is the
// only reading that survives it. A false positive here would train people to reach for
// --allow-clobber reflexively, which defeats the whole guard.
func TestLostContent_ReformattingIsNotALoss(t *testing.T) {
	oneLine := `package provider
var x = map[string]A{"alpha": A{}, "beta": B{}, "gamma": C{}}
`
	wrapped := `package provider
var x = map[string]A{
	"gamma": C{},
	"alpha": A{},
	"beta":  B{},
}
`
	if got := missingFrom(quotedKeyNames(oneLine), quotedKeyNames(wrapped)); len(got) != 0 {
		t.Errorf("reordering/rewrapping must not report losses, got %v", got)
	}
	// Adding a key is likewise not a loss.
	added := wrapped + `var y = map[string]A{"delta": D{}}` + "\n"
	if got := missingFrom(quotedKeyNames(oneLine), quotedKeyNames(added)); len(got) != 0 {
		t.Errorf("pure additions must not report losses, got %v", got)
	}
	// camelCase API field names are tracked too, not just snake_case schema keys.
	if !quotedKeyNames(`x := map[string]any{"returnCode": 1}`)["returnCode"] {
		t.Error("camelCase API field names must be tracked")
	}
}

// vmLikeResourceDef is a minimal stand-in for the real "vm" ResourceDef swagger extraction would
// produce: one ordinary swagger-driven field, matching the shape emitVmCreateDelete is written
// against. boot_volume_attachment itself is deliberately absent from Fields — it has no swagger
// counterpart at all (see vmBootVolumeAttachmentSchemaAttr's doc comment) and must appear purely
// via the r.Name == "vm" special case in generateResourceFile/emitSpecModel/emitResourceImports.
func vmLikeResourceDef() ResourceDef {
	return ResourceDef{
		Name:    "vm",
		APIPath: "/api/v1/vm",
		Fields: []FieldDef{
			{TFName: "os_type", APIName: "osType", FieldType: "string"},
		},
	}
}

// Regression guard for the resource_vm.go regeneration hazard the clobber guard exists to catch:
// before this fix, regenerating resource_vm.go silently dropped the whole hand-written
// boot_volume_attachment feature (schema field, its 3 helper functions, and most of Create/
// Delete's orchestration logic) — no error, no warning. Verified against live dev swagger via a
// full generator run (see this session's manual verification) to reproduce the committed file
// exactly, modulo cosmetic field ordering; this test pins the specific markers that regen must
// keep reproducing, independent of any live swagger fetch.
func TestGenerateResourceFile_VmSpecialCase_ReproducesBootVolumeAttachment(t *testing.T) {
	got := generateResourceFile(vmLikeResourceDef())

	mustContain := []string{
		// boot_volume_attachment schema attribute, with its own (not just its parent's) stability
		// plan modifier.
		`"boot_volume_attachment": schema.SingleNestedAttribute{`,
		`PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()}`,
		`"attachment_id": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}`,
		`"volume_id":     schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}}`,
		// the model field and the 3 hand-written helpers/vars.
		`BootVolumeAttachment types.Object`,
		"var bootVolumeAttachmentAttrTypes",
		"func resolvedVmState(",
		"func vmCreateRequiresBootVolumeAttachment(",
		"func buildBootVolumeAttachmentPlan(",
		// Create's extra validation + behind-the-scenes attachment creation.
		"Missing boot_volume_attachment",
		"Boot Volume Attachment Create Error",
		// Delete's extra cleanup.
		"Boot Volume Attachment Delete Error",
		"Boot Volume Attachment Delete Poll Error",
	}
	for _, marker := range mustContain {
		if !strings.Contains(got, marker) {
			t.Errorf("vm-generated resource file missing expected content: %s", marker)
		}
	}

	// The ordinary swagger-driven field must still go through the normal generic path untouched.
	if !strings.Contains(got, `"os_type": schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}`) {
		t.Error("vm's ordinary swagger-driven field (os_type) was not generated via the normal path")
	}
}

// The vm special case must not leak into any other resource's generated file.
func TestGenerateResourceFile_NonVmResource_HasNoBootVolumeAttachment(t *testing.T) {
	other := vmLikeResourceDef()
	other.Name = "volume"
	other.APIPath = "/api/v1/volume"

	got := generateResourceFile(other)
	leaked := []string{
		"boot_volume_attachment",
		"bootVolumeAttachmentAttrTypes",
		"resolvedVmState",
		"vmCreateRequiresBootVolumeAttachment",
		"buildBootVolumeAttachmentPlan",
		"objectplanmodifier",
	}
	for _, marker := range leaked {
		if strings.Contains(got, marker) {
			t.Errorf("non-vm resource file unexpectedly contains vm-specific content: %s", marker)
		}
	}
	// And it must still get a normal generated Delete (vm is the only resource that skips it).
	if !strings.Contains(got, "func (r *VolumeResource) Delete(") {
		t.Error("non-vm resource is missing its normal generated Delete method")
	}
}
