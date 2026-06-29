data "kvindo_kubernetes_user" "example" {
  id = "01abc123def456gh789012345"
}

output "info_kubeconfig" {
  value     = data.kvindo_kubernetes_user.example.status.kubeconfig
  sensitive = true
}
