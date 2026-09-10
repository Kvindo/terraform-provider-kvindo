package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func writeKcProfile(t *testing.T, home, name, yaml string) {
	t.Helper()
	dir := filepath.Join(home, ".kc", "config")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadKcProfile_ReadsServerAndToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	home, _ := os.UserHomeDir()
	writeKcProfile(t, home, "prod", "server: https://cloud-api.kvindo.com\ntoken: from-profile-token\n")

	p, found := loadKcProfile("prod")
	if !found {
		t.Fatal("expected found=true for an existing, readable profile")
	}
	if p.Server != "https://cloud-api.kvindo.com" {
		t.Errorf("Server = %q", p.Server)
	}
	if p.Token != "from-profile-token" {
		t.Errorf("Token = %q", p.Token)
	}
}

func TestLoadKcProfile_MissingFileReturnsZeroValue(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	p, found := loadKcProfile("does-not-exist")
	if found {
		t.Error("expected found=false for a missing profile")
	}
	if p.Server != "" || p.Token != "" {
		t.Errorf("expected zero-value profile, got %+v", p)
	}
}

func TestLoadKcProfile_SanitizesPathTraversal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// A file sitting outside ~/.kc/config, at the exact spot "../secret.yaml" would
	// resolve to without sanitization.
	if err := os.WriteFile(filepath.Join(home, "..", "secret.yaml"), []byte("server: https://escaped.example.com\ntoken: escaped-token\n"), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join(home, "..", "secret.yaml"))

	p, found := loadKcProfile("../secret")
	if found {
		t.Error("expected found=false: '../secret' must not escape ~/.kc/config via filepath.Base sanitization")
	}
	if p.Server != "" || p.Token != "" {
		t.Errorf("expected zero-value profile, got %+v", p)
	}
}

func TestProviderSchema_HasCliProfile(t *testing.T) {
	p := &KvindoProvider{version: "test"}
	var resp provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["cli_profile"]
	if !ok {
		t.Fatal("expected attribute 'cli_profile' in provider schema")
	}
	if attr.IsRequired() {
		t.Error("expected cli_profile to be optional, not required")
	}
}

// configureWith builds a provider.ConfigureRequest from the given top-level string attribute
// values (nil means null/unset) and runs Configure, returning the response.
func configureWith(t *testing.T, endpoint, token, cliProfile *string) provider.ConfigureResponse {
	t.Helper()
	ctx := context.Background()
	p := &KvindoProvider{version: "test"}
	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)

	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("provider schema type is not an object")
	}
	strVal := func(v *string) tftypes.Value {
		if v == nil {
			return tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(tftypes.String, *v)
	}
	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"endpoint":    strVal(endpoint),
		"token":       strVal(token),
		"cli_profile": strVal(cliProfile),
	})

	var resp provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: tfsdk.Config{Raw: raw, Schema: schemaResp.Schema}}, &resp)
	return resp
}

func strp(s string) *string { return &s }

func TestConfigure_CliProfileSuppliesEndpointAndToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KVINDO_ENDPOINT", "")
	t.Setenv("KVINDO_TOKEN", "")
	t.Setenv("KVINDO_CLI_PROFILE", "")
	home, _ := os.UserHomeDir()
	writeKcProfile(t, home, "staging", "server: https://staging-api.example.com\ntoken: staging-token\n")

	resp := configureWith(t, nil, nil, strp("staging"))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	pd, ok := resp.ResourceData.(*KvindoProviderData)
	if !ok {
		t.Fatalf("expected *KvindoProviderData, got %T", resp.ResourceData)
	}
	if pd.Client.BaseURL != "https://staging-api.example.com" {
		t.Errorf("BaseURL = %q", pd.Client.BaseURL)
	}
	if pd.Client.Token != "staging-token" {
		t.Errorf("Token = %q", pd.Client.Token)
	}
}

func TestConfigure_ExplicitAttributesOverrideCliProfile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KVINDO_ENDPOINT", "")
	t.Setenv("KVINDO_TOKEN", "")
	t.Setenv("KVINDO_CLI_PROFILE", "")
	home, _ := os.UserHomeDir()
	writeKcProfile(t, home, "staging", "server: https://staging-api.example.com\ntoken: staging-token\n")

	resp := configureWith(t, strp("https://explicit.example.com"), strp("explicit-token"), strp("staging"))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	pd := resp.ResourceData.(*KvindoProviderData)
	if pd.Client.BaseURL != "https://explicit.example.com" {
		t.Errorf("BaseURL = %q, want explicit endpoint attribute to win", pd.Client.BaseURL)
	}
	if pd.Client.Token != "explicit-token" {
		t.Errorf("Token = %q, want explicit token attribute to win", pd.Client.Token)
	}
}

func TestConfigure_EnvVarCliProfileFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("KVINDO_ENDPOINT", "")
	t.Setenv("KVINDO_TOKEN", "")
	t.Setenv("KVINDO_CLI_PROFILE", "from-env")
	writeKcProfile(t, home, "from-env", "server: https://env-profile.example.com\ntoken: env-profile-token\n")

	resp := configureWith(t, nil, nil, nil)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	pd := resp.ResourceData.(*KvindoProviderData)
	if pd.Client.BaseURL != "https://env-profile.example.com" {
		t.Errorf("BaseURL = %q", pd.Client.BaseURL)
	}
	if pd.Client.Token != "env-profile-token" {
		t.Errorf("Token = %q", pd.Client.Token)
	}
}

func TestConfigure_MissingTokenErrorsEvenWithCliProfileUnset(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KVINDO_ENDPOINT", "")
	t.Setenv("KVINDO_TOKEN", "")
	t.Setenv("KVINDO_CLI_PROFILE", "")

	resp := configureWith(t, nil, nil, nil)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when no token, env var, or cli_profile is available")
	}
}

func TestConfigure_MissingTokenErrorNamesUnreadableCliProfile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KVINDO_ENDPOINT", "")
	t.Setenv("KVINDO_TOKEN", "")
	t.Setenv("KVINDO_CLI_PROFILE", "")

	resp := configureWith(t, nil, nil, strp("nonexistent"))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when the named cli_profile can't be read and no other token source exists")
	}
	found := false
	for _, d := range resp.Diagnostics {
		if strings.Contains(d.Detail(), "nonexistent") && strings.Contains(d.Detail(), "could not be read") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a diagnostic naming the unreadable profile, got: %v", resp.Diagnostics)
	}
}

func TestConfigure_EndpointTrailingSlashTrimmed(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("KVINDO_ENDPOINT", "")
	t.Setenv("KVINDO_TOKEN", "")
	t.Setenv("KVINDO_CLI_PROFILE", "")

	resp := configureWith(t, strp("https://explicit.example.com/"), strp("explicit-token"), nil)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	pd := resp.ResourceData.(*KvindoProviderData)
	if pd.Client.BaseURL != "https://explicit.example.com" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", pd.Client.BaseURL)
	}
}
