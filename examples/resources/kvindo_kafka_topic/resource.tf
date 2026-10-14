data "kvindo_kafka" "main" {
  name = "main"
}

resource "kvindo_kafka_topic" "orders" {
  metadata = {
    name      = "orders"
    folder_id = data.kvindo_kafka.main.metadata.folder_id
  }
  spec = {
    kafka_id   = data.kvindo_kafka.main.metadata.id
    partitions = 6

    # Omit to keep the topic's existing factor, or to take the cluster default on create.
    # Set it only to pin an explicit value - sending 1 is a real request for a single replica,
    # not a way to say "unset".
    # replication_factor = 3

    config = {
      "retention.ms" = "604800000"
    }
  }
}
