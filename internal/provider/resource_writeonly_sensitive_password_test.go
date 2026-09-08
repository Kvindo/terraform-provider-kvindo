package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Regression test for the postgresql_user/valkey_user "Provider produced inconsistent result
// after apply" bug (2026-09-08): the backend never returns password on GET (write-only field),
// but populate<X>State used to unconditionally overwrite the in-memory model's password with
// that empty GET response - including when populating a plan-derived model in Create/Update, so
// the first time anyone actually configured a real password, Terraform Core's post-apply
// consistency check compared the final state's (now-blank) password against the plan's
// (non-blank) configured value and failed, even though the resource was created/updated
// correctly server-side. Fix: populate<X>State takes a preserveSensitive bool - true for
// Create/Update (populating a plan, must keep the configured value), false for Read/ImportState
// (populating a genuine refreshed state, must always reflect the live API - unchanged behavior).
func TestPopulatePostgresqlUserState_PreservesConfiguredPasswordOnCreateUpdate(t *testing.T) {
	ctx := context.Background()
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01test", "name": "dev-kvindo-cloud"},
		"spec":     map[string]interface{}{"password": "", "login": true},
	}

	// Create/Update path: plan already carries the configured password - must survive.
	plan := PostgresqlUserResourceModel{Spec: PostgresqlUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populatePostgresqlUserState(ctx, apiData, &plan, true); err != nil {
		t.Fatalf("populatePostgresqlUserState: %v", err)
	}
	if got := plan.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("preserveSensitive=true with a configured password: got %q, want the configured value preserved (%q)", got, "s3cr3t")
	}

	// Create path, nothing configured (server-generated password): the empty GET response is the
	// only source of truth, so it must still be used - not left permanently unknown/null.
	planNoPassword := PostgresqlUserResourceModel{Spec: PostgresqlUserSpecModel{Password: types.StringUnknown()}}
	if err := populatePostgresqlUserState(ctx, apiData, &planNoPassword, true); err != nil {
		t.Fatalf("populatePostgresqlUserState: %v", err)
	}
	if got := planNoPassword.Spec.Password; got.ValueString() != "" || got.IsUnknown() {
		t.Errorf("preserveSensitive=true with no configured password: got %#v, want the GET response's empty value applied (not left unknown)", got)
	}

	// Read/ImportState path (preserveSensitive=false): must always reflect the live API,
	// regardless of whatever the prior state happened to hold - unchanged, established behavior.
	state := PostgresqlUserResourceModel{Spec: PostgresqlUserSpecModel{Password: types.StringValue("stale-value")}}
	if err := populatePostgresqlUserState(ctx, apiData, &state, false); err != nil {
		t.Fatalf("populatePostgresqlUserState: %v", err)
	}
	if got := state.Spec.Password.ValueString(); got != "" {
		t.Errorf("preserveSensitive=false: got %q, want the live API's empty value to win over stale state (%q)", got, "")
	}
}

func TestPopulateValkeyUserState_PreservesConfiguredPasswordOnCreateUpdate(t *testing.T) {
	ctx := context.Background()
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01test", "name": "dev-valkey-user"},
		"spec":     map[string]interface{}{"password": "", "enabled": true},
	}

	plan := ValkeyUserResourceModel{Spec: ValkeyUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateValkeyUserState(ctx, apiData, &plan, true); err != nil {
		t.Fatalf("populateValkeyUserState: %v", err)
	}
	if got := plan.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("preserveSensitive=true with a configured password: got %q, want the configured value preserved (%q)", got, "s3cr3t")
	}

	state := ValkeyUserResourceModel{Spec: ValkeyUserSpecModel{Password: types.StringValue("stale-value")}}
	if err := populateValkeyUserState(ctx, apiData, &state, false); err != nil {
		t.Fatalf("populateValkeyUserState: %v", err)
	}
	if got := state.Spec.Password.ValueString(); got != "" {
		t.Errorf("preserveSensitive=false: got %q, want the live API's empty value to win over stale state (%q)", got, "")
	}
}
