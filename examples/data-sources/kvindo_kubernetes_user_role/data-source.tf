data "kvindo_kubernetes_user_role" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_kubernetes_user_role.example.status.state
}
