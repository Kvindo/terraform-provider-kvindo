data "kvindo_postgresql_database" "example" {
  id = "01abc123def456gh789012345"
}

output "postgresql_database_state" {
  value = data.kvindo_postgresql_database.example.status.state
}
