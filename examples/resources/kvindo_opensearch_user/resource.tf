data "kvindo_opensearch" "search" {
  name = "search-cluster"
}

resource "kvindo_opensearch_user" "dashboard_reader" {
  metadata = {
    name      = "dashboard-reader"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    # Immutable - a security-plugin user lives inside exactly one cluster for its whole lifetime.
    open_search_id = data.kvindo_opensearch.search.metadata.id

    # The backend never returns this on read. A user change re-applies the WHOLE cluster, so the
    # value is re-sent on every dispatch and is never treated as "lock the credential" by omission.
    password = "correct-horse-battery-staple"
    enabled  = true

    # Read-only gets `read`; read-write adds read/write/create_index/delete/manage. A user with no
    # grants can authenticate and do nothing. Cluster administration is never granted.
    index_grants = [
      { index_pattern = "logs-2026", pattern_type = "Prefixed", role = "ReadOnly" },
      { index_pattern = "app-metrics", pattern_type = "Literal", role = "ReadWrite" },
    ]
  }
}
