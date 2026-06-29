data "kvindo_open_vpn" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_open_vpn.example.status.state
}
