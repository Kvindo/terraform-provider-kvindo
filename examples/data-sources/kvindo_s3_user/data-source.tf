data "kvindo_s3_user" "example" {
  id = "01abc123def456gh789012345"
}

output "info_access_key" {
  value = data.kvindo_s3_user.example.status.access_key
}
