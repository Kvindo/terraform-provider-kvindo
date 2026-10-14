data "kvindo_kafka" "main" {
  name = "main"
}

resource "kvindo_kafka_user" "app" {
  metadata = {
    name      = "app-user"
    folder_id = data.kvindo_kafka.main.metadata.folder_id
  }
  spec = {
    kafka_id = data.kvindo_kafka.main.metadata.id
    password = "'change-me'"
    enabled  = true

    # Role-based, not raw ACL tuples: the platform expands "produce to orders" into the
    # operations Kafka actually needs.
    topic_grants = [
      { topic_pattern = "orders", pattern_type = "Literal", role = "Producer" },
      { topic_pattern = "orders-", pattern_type = "Prefixed", role = "Consumer" },
    ]

    # REQUIRED as soon as any grant is a Consumer: a consumer without group access cannot
    # commit offsets, so the request is rejected at submit time.
    consumer_group_prefixes = ["app-"]
  }
}
