data "kvindo_loadbalancer_target_group_static_target" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_loadbalancer_target_group_static_target.example.status.state
}
