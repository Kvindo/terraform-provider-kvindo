package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// KafkaUser is the third write-only-password resource, and the one whose backend behaviour differs:
// unlike valkey_user/postgresql_user the API REJECTS a create with enabled=true and an empty
// password, so there is no server-generated value to recover - but the read-back is still always
// empty (confirmed live 2026-10-11: three GETs of a live user all returned "password":null), which
// is what makes the preserve-if-known rule just as necessary here. Without it every refresh resets
// the configured password to null and every later plan shows a phantom "+ password" diff forever.
func TestPopulateKafkaUserState_PreservesKnownPassword(t *testing.T) {
	ctx := context.Background()
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01test", "name": "dev-kafka-user"},
		"spec":     map[string]interface{}{"password": "", "enabled": true},
	}

	plan := KafkaUserResourceModel{Spec: KafkaUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateKafkaUserState(ctx, apiData, &plan); err != nil {
		t.Fatalf("populateKafkaUserState: %v", err)
	}
	if got := plan.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("model with a configured password: got %q, want it preserved (%q)", got, "s3cr3t")
	}

	state := KafkaUserResourceModel{Spec: KafkaUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateKafkaUserState(ctx, apiData, &state); err != nil {
		t.Fatalf("populateKafkaUserState: %v", err)
	}
	if got := state.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("Read() on state with an already-known password: got %q, want it preserved (%q) - a mismatch here means every subsequent plan shows a phantom diff forever", got, "s3cr3t")
	}
}

// The schema-level half of the same guarantee: a credential must never be echoed into plan output.
// This is what a missing sensitiveSpecFields entry costs, and it fails at compile time for nobody.
func TestKafkaUserSchema_PasswordIsSensitive(t *testing.T) {
	attrs := KafkaUserResourceSchemaAttrs()
	spec, ok := attrs["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("spec is not a SingleNestedAttribute: %T", attrs["spec"])
	}
	attr, ok := spec.Attributes["password"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("spec.password is not a StringAttribute: %T", spec.Attributes["password"])
	}
	if !attr.Sensitive {
		t.Error("kafka_user spec.password must be Sensitive - valkey_user and postgresql_user both are; " +
			"without it a real SCRAM credential is printed in terraform plan output and stored in plain " +
			"text in the plan file")
	}
}

// ClickHouseUser is the fourth write-only-password resource, and the one that shipped without the
// Sensitive flag at all: unlike valkey/postgresql/kafka, neither its resource nor its datasource
// marked the field, so terraform plan printed a live database credential. Its backend behaves like
// Kafka's (no server-side generation - ClickHouseUserResourceController rejects a create with
// enabled=true and an empty password), and its read-back is always empty, so the preserve-if-known
// rule is just as necessary here.
func TestPopulateClickhouseUserState_PreservesKnownPassword(t *testing.T) {
	ctx := context.Background()
	apiData := map[string]interface{}{
		"metadata": map[string]interface{}{"id": "01test", "name": "dev-clickhouse-user"},
		"spec":     map[string]interface{}{"password": "", "enabled": true},
	}

	plan := ClickhouseUserResourceModel{Spec: ClickhouseUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateClickhouseUserState(ctx, apiData, &plan); err != nil {
		t.Fatalf("populateClickhouseUserState: %v", err)
	}
	if got := plan.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("model with a configured password: got %q, want it preserved (%q)", got, "s3cr3t")
	}

	state := ClickhouseUserResourceModel{Spec: ClickhouseUserSpecModel{Password: types.StringValue("s3cr3t")}}
	if err := populateClickhouseUserState(ctx, apiData, &state); err != nil {
		t.Fatalf("populateClickhouseUserState: %v", err)
	}
	if got := state.Spec.Password.ValueString(); got != "s3cr3t" {
		t.Errorf("Read() on state with an already-known password: got %q, want it preserved (%q) - a mismatch here means every subsequent plan shows a phantom diff forever", got, "s3cr3t")
	}
}

// The schema-level half: a credential must never be echoed into plan output. This is what a missing
// sensitiveSpecFields entry costs, and nothing fails at compile time when it is missing.
func TestClickhouseUserPasswordIsSensitive(t *testing.T) {
	attrs := ClickhouseUserResourceSchemaAttrs()
	spec, ok := attrs["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("spec is not a SingleNestedAttribute: %T", attrs["spec"])
	}
	attr, ok := spec.Attributes["password"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("spec.password is not a StringAttribute: %T", spec.Attributes["password"])
	}
	if !attr.Sensitive {
		t.Error("clickhouse_user spec.password must be Sensitive - valkey_user, postgresql_user and " +
			"kafka_user all are; without it a real database credential is printed in terraform plan " +
			"output and stored in plain text in the plan file")
	}
}

// And the datasource, which has its own schema: it must hide the field too. (A datasource is
// read-only, so there is no preserve-known path - the GET never returns a password, and Sensitive is
// what keeps that empty value out of plan/apply output rather than displaying it.)
func TestClickhouseUserDatasourcePasswordIsSensitive(t *testing.T) {
	d := &ClickhouseUserDataSource{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	spec, ok := resp.Schema.Attributes["spec"].(dschema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("datasource spec is not a SingleNestedAttribute: %T", resp.Schema.Attributes["spec"])
	}
	attr, ok := spec.Attributes["password"].(dschema.StringAttribute)
	if !ok {
		t.Fatalf("datasource spec.password is not a StringAttribute: %T", spec.Attributes["password"])
	}
	if !attr.Sensitive {
		t.Error("clickhouse_user data source spec.password must be Sensitive (datasource_kafka_user.go " +
			"and datasource_valkey_user.go both are) - otherwise the value it returns is displayed")
	}
}
