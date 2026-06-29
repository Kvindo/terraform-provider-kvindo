data "kvindo_s3_user_access_policy" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_s3_user_access_policy.example.status.state
}
