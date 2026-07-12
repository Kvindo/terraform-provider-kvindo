data "kvindo_vm_command_schedule" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_vm_command_schedule.example.status.state
}
