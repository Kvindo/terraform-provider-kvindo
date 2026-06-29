data "kvindo_kubernetes" "example" {
  id = "01abc123def456gh789012345"
}

output "info_api_server_url" {
  value = data.kvindo_kubernetes.example.status.api_server_url
}
