data "kvindo_folder" "default" {
  name = "default"
}

data "kvindo_clickhouse" "analytics" {
  name = "analytics-cluster"
}

resource "kvindo_clickhouse_database" "events" {
  metadata = {
    name      = "events"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    clickhouse_id = data.kvindo_clickhouse.analytics.metadata.id
  }
}
