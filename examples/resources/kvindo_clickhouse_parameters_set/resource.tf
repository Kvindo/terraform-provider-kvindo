data "kvindo_folder" "default" {
  name = "default"
}

resource "kvindo_clickhouse_parameters_set" "analytics" {
  metadata = {
    name      = "analytics"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    parameters = {
      "max_server_memory_usage" = "34359738368"
      "max_concurrent_queries"  = "100"
    }
  }
}
