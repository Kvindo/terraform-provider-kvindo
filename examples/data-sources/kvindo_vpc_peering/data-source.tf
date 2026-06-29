data "kvindo_vpc_peering" "example" {
  id = "01abc123def456gh789012345"
}

output "info_state" {
  value = data.kvindo_vpc_peering.example.status.state
}
