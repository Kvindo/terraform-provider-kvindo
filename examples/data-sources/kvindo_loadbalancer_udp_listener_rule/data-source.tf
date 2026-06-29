data "kvindo_loadbalancer_udp_listener_rule" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_loadbalancer_udp_listener_rule.example.status.state
}
