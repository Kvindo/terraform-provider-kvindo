data "kvindo_vm_on_off_maintenance_action" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_vm_on_off_maintenance_action.example.status.state
}
