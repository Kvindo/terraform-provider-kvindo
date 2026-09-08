package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Regression test for the postgresql_user/valkey_user password bugs found 2026-09-08, in two
// passes:
//
//  1. "Provider produced inconsistent result after apply" - populate<X>State used to
//     unconditionally overwrite the in-memory model's password with the backend's GET response
//     (which is always empty - password is write-only), including when populating a
//     plan-derived model in Create/Update. The first time anyone configured a real password,
//     Terraform Core's post-apply consistency check compared the now-blank final state against
//     the non-blank planned value and failed, even though the resource was created/updated
//     correctly server-side.
//  2. A permanent phantom diff, found live immediately after shipping the fix for (1): that fix
//     only special-cased Create/Update, so Read() kept unconditionally resetting the real
//     configured password back to empty on every refresh - every SUBSEQUENT `terraform plan`
//     then showed the same "+ password" diff forever, since refreshed state could never match
//     the still-configured value.
//
// Fix: populate<X>State preserves whatever's ALREADY known (non-null, non-unknown) in the model
// being populated, and only takes the GET response when nothing is known yet - applied
// unconditionally in all four call paths (Create, Update, Read, ImportState), not just
// Create/Update. This test exercises populate<X>State directly against all three real shapes:
// a plan carrying a configured password (Create/Update), a fresh/zero-value model with nothing
// known yet (initial Create with no configured password, or ImportState), and a refreshed state
// that already holds a real value from a prior Create/Update (Read - this is the case (2) fix).
func TestPopulatePostgresqlUserState_PreservesKnownPassword(t *testing.T) {
	ctx := context.Background()
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01test", "name": "dev-kvindo-cloud"},
		"spec":     map[string]interface{}{"password": "", "login": true},
	}

	// Create/Update: plan already carries the configured password - must survive.
	plan := PostgresqlUserResourceModel{Spec: PostgresqlUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populatePostgresqlUserState(ctx, apiData, &plan); err != nil {
		t.Fatalf("populatePostgresqlUserState: %v", err)
	}
	if got := plan.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("model with a configured password: got %q, want it preserved (%q)", got, "s3cr3t")
	}

	// Nothing known yet (initial Create with no configured password, or a fresh ImportState):
	// the empty GET response is the only source of truth, so it must be applied, not left
	// permanently unknown/null.
	fresh := PostgresqlUserResourceModel{Spec: PostgresqlUserSpecModel{Password: types.StringUnknown()}}
	if err := populatePostgresqlUserState(ctx, apiData, &fresh); err != nil {
		t.Fatalf("populatePostgresqlUserState: %v", err)
	}
	if got := fresh.Spec.Password; got.ValueString() != "" || got.IsUnknown() {
		t.Errorf("model with no configured password: got %#v, want the GET response's empty value applied (not left unknown)", got)
	}

	// Read: a refreshed state that already holds a real value from a prior Create/Update must
	// KEEP that value - this is bug (2). Regressing this reintroduces the permanent phantom diff.
	state := PostgresqlUserResourceModel{Spec: PostgresqlUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populatePostgresqlUserState(ctx, apiData, &state); err != nil {
		t.Fatalf("populatePostgresqlUserState: %v", err)
	}
	if got := state.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("Read() on state with an already-known password: got %q, want it preserved (%q) - a mismatch here means every subsequent plan shows a phantom diff forever", got, "s3cr3t")
	}
}

func TestPopulateValkeyUserState_PreservesKnownPassword(t *testing.T) {
	ctx := context.Background()
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01test", "name": "dev-valkey-user"},
		"spec":     map[string]interface{}{"password": "", "enabled": true},
	}

	plan := ValkeyUserResourceModel{Spec: ValkeyUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateValkeyUserState(ctx, apiData, &plan); err != nil {
		t.Fatalf("populateValkeyUserState: %v", err)
	}
	if got := plan.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("model with a configured password: got %q, want it preserved (%q)", got, "s3cr3t")
	}

	state := ValkeyUserResourceModel{Spec: ValkeyUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateValkeyUserState(ctx, apiData, &state); err != nil {
		t.Fatalf("populateValkeyUserState: %v", err)
	}
	if got := state.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("Read() on state with an already-known password: got %q, want it preserved (%q) - a mismatch here means every subsequent plan shows a phantom diff forever", got, "s3cr3t")
	}
}
