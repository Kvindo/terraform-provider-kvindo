data "kvindo_quota" "example" {
  id = "01abc123def456gh789012345"
}

output "info_current_value" {
  value = data.kvindo_quota.example.status.current_value
}
