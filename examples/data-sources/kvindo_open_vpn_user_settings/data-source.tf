data "kvindo_open_vpn_user_settings" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_open_vpn_user_settings.example.status.state
}
