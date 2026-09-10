data "kvindo_valkey" "example" {
  id = "01abc123def456gh789012345"
}

output "valkey_state" {
  value = data.kvindo_valkey.example.status.state
}
