resource "kvindo_valkey_parameters_set" "example" {
  metadata = {
    name = "my-valkey-params"
    labels = {
      env = "production"
    }
  }
  spec = {
    parameters = {
      "maxmemory-policy" = "allkeys-lru"
    }
  }
}
