package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Every entry in the generator's specFieldDescriptions table (tools/generator/main.go) must
// actually reach the built schema, because that table is the ONLY thing keeping these
// descriptions alive: the resource_*.go files are generated, so a Description hand-written into
// one is silently wiped by the next full regen - the exact failure that motivated the table's
// existence in the first place (2026-08-11).
//
// The reverse direction is what this test really guards: dropping or renaming a table entry
// breaks nothing at compile time and produces no diagnostic. It just quietly ships a public
// Registry page with a bare `connection_limit (Number)` and no explanation, and nobody notices
// until a user asks why their apply was rejected.
//
// The table itself lives in package main of the generator and cannot be imported here, so the
// pairs are restated. When you ADD a specFieldDescriptions entry, add it here too.
func TestResourceSchemas_SpecFieldDescriptionsArePresent(t *testing.T) {
	want := map[string][]string{
		"kvindo_route_table_attachment": {"vpc_id", "vpc_subnet_id"},
		"kvindo_valkey_user":            {"key_patterns", "categories", "channels", "password"},
		"kvindo_postgresql_user":        {"connection_limit", "password"},
	}

	seen := map[string]bool{}
	p := &KvindoProvider{version: "test"}
	for _, newResource := range p.Resources(context.Background()) {
		r := newResource()

		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)

		fields, ok := want[metaResp.TypeName]
		if !ok {
			continue
		}
		seen[metaResp.TypeName] = true

		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Fatalf("%s: schema build produced diagnostics: %v", metaResp.TypeName, schemaResp.Diagnostics)
		}

		spec, ok := schemaResp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("%s: expected a spec SingleNestedAttribute", metaResp.TypeName)
		}

		for _, f := range fields {
			attr, ok := spec.Attributes[f]
			if !ok {
				t.Errorf("%s: spec has no attribute %q - was it renamed without updating specFieldDescriptions?", metaResp.TypeName, f)
				continue
			}
			if strings.TrimSpace(attr.GetDescription()) == "" {
				t.Errorf("%s: spec.%s has an empty Description - the specFieldDescriptions entry was lost "+
					"(a regen will not restore it; fix tools/generator/main.go)", metaResp.TypeName, f)
			}
		}
	}

	for name := range want {
		if !seen[name] {
			t.Errorf("resource %q was not registered by the provider at all - this test's table is stale", name)
		}
	}
}

// The rule the connection_limit description states is enforced server-side, so the text is the
// only place a practitioner can learn it before an apply is rejected. Pin the substance, not the
// wording, so a copy-edit is free but silently gutting it is not.
func TestPostgresqlUserConnectionLimitDescription_ExplainsTheRule(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	for _, newResource := range p.Resources(context.Background()) {
		r := newResource()

		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)
		if metaResp.TypeName != "kvindo_postgresql_user" {
			continue
		}

		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		spec := schemaResp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
		got := strings.ToLower(spec.Attributes["connection_limit"].GetDescription())

		// Both halves matter: that leaving it unset is legitimate (it is Optional and the
		// platform assigns a share), and that a value set too low is rejected.
		for _, need := range []string{"unset", "50", "maintenance database"} {
			if !strings.Contains(got, need) {
				t.Errorf("connection_limit description no longer mentions %q: %q", need, got)
			}
		}
		return
	}
	t.Fatal("kvindo_postgresql_user resource not registered")
}
