resource "kvindo_postgresql_parameters_set" "example" {
  metadata = {
    name = "my-pg-params"
    labels = {
      env = "production"
    }
  }
}
