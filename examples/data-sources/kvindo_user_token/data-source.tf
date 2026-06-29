data "kvindo_user_token" "example" {
  id = "01abc123def456gh789012345"
}

output "info_token" {
  value     = data.kvindo_user_token.example.status.token
  sensitive = true
}
