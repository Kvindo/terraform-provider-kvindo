package main

import "testing"

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
