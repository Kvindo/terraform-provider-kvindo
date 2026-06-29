data "kvindo_victoria_metrics" "example" {
  id = "01abc123def456gh789012345"
}

output "info_fqdn" {
  value = data.kvindo_victoria_metrics.example.status.fqdn
}
