data "kvindo_loadbalancer_https_listener" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_loadbalancer_https_listener.example.status.state
}
