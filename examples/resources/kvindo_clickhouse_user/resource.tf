data "kvindo_folder" "default" {
  name = "default"
}

data "kvindo_clickhouse" "analytics" {
  name = "analytics-cluster"
}

resource "kvindo_clickhouse_user" "analyst" {
  metadata = {
    name      = "analyst"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    clickhouse_id = data.kvindo_clickhouse.analytics.metadata.id
    password      = "change-me-please"
    enabled       = true
    database_grants = [
      { database_name = "events", role = "ReadWrite" }
    ]
    settings = {
      "max_memory_usage" = "10000000000"
      "readonly"         = "0"
    }
  }
}
