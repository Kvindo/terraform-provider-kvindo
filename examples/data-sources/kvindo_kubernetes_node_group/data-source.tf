data "kvindo_kubernetes_node_group" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_kubernetes_node_group.example.status.state
}
