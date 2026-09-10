data "kvindo_postgresql_user" "example" {
  id = "01abc123def456gh789012345"
}

output "postgresql_user_state" {
  value = data.kvindo_postgresql_user.example.status.state
}
