data "kvindo_valkey_parameters_set" "example" {
  id = "01abc123def456gh789012345"
}

output "valkey_parameters_set_state" {
  value = data.kvindo_valkey_parameters_set.example.status.state
}
