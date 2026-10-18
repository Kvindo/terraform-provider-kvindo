package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Two gaps that shipped silently and are invisible to every other test in this package:
//
//  1. KafkaUserSpec gained `clusterOperations` (and KafkaTopicGrant gained `extraOperations`) in the
//     API on 2026-10-18, so a tenant could run a Schema Registry or a Kafka UI without the Admin role
//     that v1 deliberately does not offer. The provider's schema is GENERATED, so nothing failed when
//     the API grew the field — the provider simply could not express it, and the /sdk page documented
//     a capability Terraform users did not have.
//
//  2. The transaction registry (transaction_registry.go) is HAND-WRITTEN while TransactionSpec.cs is
//     generated from the API models. When the OpenSearch family shipped, its three array properties
//     (OpenSearches/OpenSearchParametersSets/OpenSearchUsers) landed in the API and in
//     kvindo_opensearch* but not in txnSubs, so `kvindo_transaction` could not carry an OpenSearch
//     child at all — while it could carry the Kafka and ClickHouse equivalents.
//
// Both are asserted through the built schema rather than the source, so a regeneration that drops the
// field (or a registry edit that never reaches txnSchemaAttrs) fails here.
func TestKafkaUserSchema_CarriesTheClusterScopedAclFields(t *testing.T) {
	attrs := KafkaUserResourceSchemaAttrs()

	spec, ok := attrs["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("kvindo_kafka_user.spec is not a SingleNestedAttribute")
	}
	clusterOps, ok := spec.Attributes["cluster_operations"]
	if !ok {
		t.Fatal("kvindo_kafka_user.spec.cluster_operations is missing: the API's KafkaUserSpec.ClusterOperations " +
			"cannot be expressed, so a tenant-run Schema Registry / Kafka UI is impossible through Terraform")
	}
	if _, ok := clusterOps.(schema.ListAttribute); !ok {
		t.Errorf("cluster_operations should be a ListAttribute (the API takes KafkaClusterOperation[]), got %T", clusterOps)
	}

	grants, ok := spec.Attributes["topic_grants"].(schema.ListNestedAttribute)
	if !ok {
		t.Fatal("kvindo_kafka_user.spec.topic_grants is not a ListNestedAttribute")
	}
	if _, ok := grants.NestedObject.Attributes["extra_operations"]; !ok {
		t.Error("topic_grants[].extra_operations is missing: KafkaTopicGrant.ExtraOperations " +
			"(DescribeConfigs / AlterConfigs) cannot be expressed through Terraform")
	}

	// The same two fields must be readable back, or a plan can never converge. The data source is
	// generated separately AND against a different schema package (datasource/schema, whose
	// SingleNestedAttribute is a distinct type from resource/schema's), so it is asserted separately.
	var dsSpec dsschema.SingleNestedAttribute
	var found bool
	p := &KvindoProvider{version: "test"}
	for _, newDs := range p.DataSources(context.Background()) {
		d := newDs()
		var metaResp datasource.MetadataResponse
		d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)
		if metaResp.TypeName != "kvindo_kafka_user" {
			continue
		}
		var schemaResp datasource.SchemaResponse
		d.Schema(context.Background(), datasource.SchemaRequest{}, &schemaResp)
		dsSpec, found = schemaResp.Schema.Attributes["spec"].(dsschema.SingleNestedAttribute)
		break
	}
	if !found {
		t.Fatal("the kvindo_kafka_user data source has no spec block")
	}
	if _, ok := dsSpec.Attributes["cluster_operations"]; !ok {
		t.Error("the kvindo_kafka_user DATA SOURCE is missing cluster_operations, so the value can be " +
			"written but never read back")
	}
}

func TestTransactionSchema_CarriesTheOpenSearchFamily(t *testing.T) {
	p := &KvindoProvider{version: "test"}

	var txnAttrs map[string]schema.Attribute
	for _, newResource := range p.Resources(context.Background()) {
		r := newResource()
		var metaResp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "kvindo"}, &metaResp)
		if metaResp.TypeName != "kvindo_transaction" {
			continue
		}
		var schemaResp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		txnAttrs = schemaResp.Schema.Attributes
		break
	}
	if txnAttrs == nil {
		t.Fatal("kvindo_transaction not found in the provider's resource list")
	}

	spec, ok := txnAttrs["spec"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("kvindo_transaction.spec is not a SingleNestedAttribute")
	}

	// Named keys, not a count: the count assertion in transaction_test.go would also pass if these
	// three were added while three unrelated entries were dropped.
	for _, want := range []string{"opensearch_parameters_sets", "opensearches", "opensearch_users"} {
		attr, ok := spec.Attributes[want]
		if !ok {
			t.Errorf("kvindo_transaction.spec.%s is missing: TransactionSpec.OpenSearches/"+
				"OpenSearchParametersSets/OpenSearchUsers are supported by the API "+
				"(TransactionResourceController creates and deletes all three) but the hand-written "+
				"txnSubs registry never gained them", want)
			continue
		}
		if _, ok := attr.(schema.MapNestedAttribute); !ok {
			t.Errorf("kvindo_transaction.spec.%s should be a MapNestedAttribute, got %T", want, attr)
		}
	}

	// Dependency order: the cluster references a parameters set, so the set must be listed first —
	// txnSubs is walked in order when a transaction is submitted.
	var order []string
	for _, s := range txnSubs {
		switch s.tfKey {
		case "opensearch_parameters_sets", "opensearches", "opensearch_users":
			order = append(order, s.tfKey)
		}
	}
	want := []string{"opensearch_parameters_sets", "opensearches", "opensearch_users"}
	if len(order) != len(want) {
		t.Fatalf("expected %d OpenSearch txnSubs entries, got %d (%v)", len(want), len(order), order)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("OpenSearch txnSubs order is %v, want %v (the set must precede the cluster, "+
				"and the user must follow it — it references open_search_id)", order, want)
			break
		}
	}
}
