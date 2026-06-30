package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProviderSchema_HasTokenAndEndpoint(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	if _, ok := resp.Schema.Attributes["token"]; !ok {
		t.Error("expected attribute 'token' in provider schema")
	}
	if _, ok := resp.Schema.Attributes["endpoint"]; !ok {
		t.Error("expected attribute 'endpoint' in provider schema")
	}
}

// Guards the Registry/tooling description: tfplugindocs and `terraform providers schema`
// surface this text, so an empty provider description regresses discoverability (the
// Registry landing page once shipped blank because this was missing).
func TestProviderSchema_HasDescription(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	if resp.Schema.MarkdownDescription == "" {
		t.Error("expected provider schema to have a non-empty MarkdownDescription")
	}
}

func TestProviderSchema_TokenIsSensitive(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	tokenAttr, ok := resp.Schema.Attributes["token"]
	if !ok {
		t.Fatal("token attribute not found")
	}
	// Sensitive() is only on StringAttribute — assert via type assertion.
	type sensitiver interface{ IsSensitive() bool }
	if s, ok := tokenAttr.(sensitiver); !ok || !s.IsSensitive() {
		t.Error("expected token attribute to be sensitive")
	}
}
