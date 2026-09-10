data "kvindo_postgresql" "example" {
  id = "01abc123def456gh789012345"
}

output "postgresql_state" {
  value = data.kvindo_postgresql.example.status.state
}
