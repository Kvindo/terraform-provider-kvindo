data "kvindo_valkey_user" "example" {
  id = "01abc123def456gh789012345"
}

output "user_state" {
  value = data.kvindo_valkey_user.example.status.state
}
