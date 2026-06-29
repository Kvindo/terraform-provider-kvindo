data "kvindo_loadbalancer_tcp_listener" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_loadbalancer_tcp_listener.example.status.state
}
