data "kvindo_folder" "default" {
  name = "default"
}

# Keys are raw Kafka broker property names; values are always strings.
resource "kvindo_kafka_parameters_set" "durable" {
  metadata = {
    name      = "durable"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    parameters = {
      "log.retention.ms"    = "86400000"
      "compression.type"    = "zstd"
      "min.insync.replicas" = "2"
    }
  }
}
