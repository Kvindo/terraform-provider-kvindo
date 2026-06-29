data "kvindo_vm" "example" {
  id = "01abc123def456gh789012345"
}

output "info_private_ipv4" {
  value = data.kvindo_vm.example.status.private_ipv4
}
