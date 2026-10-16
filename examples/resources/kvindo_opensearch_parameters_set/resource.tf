resource "kvindo_opensearch_parameters_set" "search_tuning" {
  metadata = {
    name      = "search-tuning"
    folder_id = data.kvindo_folder.default.metadata.id
  }
  spec = {
    # Only allow-listed OpenSearch settings are accepted; an unknown key or an out-of-range value
    # is rejected at submit rather than failing a provisioning job minutes later.
    parameters = {
      "indices.queries.cache.size"       = "10%"
      "thread_pool.search.queue_size"    = "2000"
      "indices.fielddata.cache.size"     = "20%"
      "indices.breaker.fielddata.limit"  = "40%"
    }
  }
}
