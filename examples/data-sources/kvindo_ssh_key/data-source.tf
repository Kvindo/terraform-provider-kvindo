data "kvindo_ssh_key" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_ssh_key.example.status.state
}
