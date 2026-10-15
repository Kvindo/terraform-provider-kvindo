package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// postgresqlSpecAttribute pulls kvindo_postgresql's "spec" nested object out of its schema.
func postgresqlSpecAttribute(t *testing.T) schema.SingleNestedAttribute {
	t.Helper()

	p := &KvindoProvider{version: "test"}
	for _, newResource := range p.Resources(context.Background()) {
		r := newResource()

		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)
		if metaResp.TypeName != "kvindo_postgresql" {
			continue
		}

		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		if schemaResp.Diagnostics.HasError() {
			t.Fatalf("kvindo_postgresql: schema build produced diagnostics: %v", schemaResp.Diagnostics)
		}

		spec, ok := schemaResp.Schema.Attributes["spec"].(schema.SingleNestedAttribute)
		if !ok {
			t.Fatalf("kvindo_postgresql: spec is %T, not a SingleNestedAttribute",
				schemaResp.Schema.Attributes["spec"])
		}
		return spec
	}

	t.Fatal("kvindo_postgresql was not registered in the provider")
	return schema.SingleNestedAttribute{}
}

// kvindo_postgresql.spec.etcd_instances is the caller-supplied DCS (etcd) placement. Its Optional
// shape is the load-bearing part: omitting it is the SUPPORTED request that asks Kvindo Cloud to
// self-provision the DCS, and it is what every cluster created before the field existed does. A
// Required field would break all of them; the UseStateForUnknown modifier is what stops an unset
// field from being reported "known after apply" on every plan (see
// resource_schema_plan_modifiers_test.go for the drift bug that modifier exists to prevent).
func TestPostgresqlEtcdInstances_IsOptionalWithAStableUnsetValue(t *testing.T) {
	spec := postgresqlSpecAttribute(t)
	attr, ok := spec.Attributes["etcd_instances"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("spec.etcd_instances is %T, not a ListNestedAttribute", spec.Attributes["etcd_instances"])
	}

	if !attr.Optional {
		t.Error("spec.etcd_instances must be Optional - omitting it is the supported " +
			"self-provisioning request, not an error")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Error("spec.etcd_instances has no PlanModifiers - an unset value would be reported as " +
			"'known after apply' on every plan, i.e. phantom drift forever")
	}

	// The two halves of a member, and nothing else: the folder the members land in is resolved
	// server-side from the organization, so exposing it here would let a config claim a value the
	// backend ignores.
	for _, want := range []string{"id", "vpc_subnet_id"} {
		if _, ok := attr.NestedObject.Attributes[want]; !ok {
			t.Errorf("spec.etcd_instances is missing the %q attribute", want)
		}
	}
	if len(attr.NestedObject.Attributes) != 2 {
		t.Errorf("spec.etcd_instances has %d attributes, expected exactly id and vpc_subnet_id",
			len(attr.NestedObject.Attributes))
	}
}

// Moving a live DCS member to another subnet is not something the backend will do - it would
// silently re-place the member holding Patroni's leader state, so the API rejects it with a 400.
// Terraform must never even reach that request: the attribute has to force replacement
// (destroy + create), exactly like the shard-group subnet it sits beside.
//
// This pins the "immutability comes free from the generator's immutableNestedFields table" claim.
// If postgresql is dropped from that table, or the nested field is renamed, the provider would
// silently start planning an in-place apply that always fails at the API.
func TestPostgresqlEtcdInstancesVpcSubnetId_ForcesReplacement(t *testing.T) {
	spec := postgresqlSpecAttribute(t)
	instances, ok := spec.Attributes["etcd_instances"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatalf("spec.etcd_instances is %T, not a ListNestedAttribute", spec.Attributes["etcd_instances"])
	}

	subnet, ok := instances.NestedObject.Attributes["vpc_subnet_id"].(schema.StringAttribute)
	if !ok {
		t.Fatalf("spec.etcd_instances[].vpc_subnet_id is %T, not a StringAttribute",
			instances.NestedObject.Attributes["vpc_subnet_id"])
	}

	// Detected by the modifier's own Description, not by its concrete type: the type
	// (stringplanmodifier.requiresReplaceModifier) is unexported and its name is the framework's
	// business, whereas this sentence is the stable, user-visible contract we are asserting.
	for _, mod := range subnet.PlanModifiers {
		if strings.Contains(mod.Description(context.Background()), "destroy and recreate") {
			return
		}
	}

	t.Errorf("spec.etcd_instances[].vpc_subnet_id has no RequiresReplace plan modifier "+
		"(modifiers: %v) - changing it would be planned as an in-place update, which the Kvindo "+
		"Cloud API rejects rather than silently re-placing a live DCS member",
		subnet.PlanModifiers)
}
